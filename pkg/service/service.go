package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"

	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"

	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/pkg/errors"

	"github.com/devfile/library/pkg/devfile/parser"
	servicebinding "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"

	"github.com/ghodss/yaml"
)

// LinkLabel is the name of the name of the link in the devfile
const LinkLabel = "app.kubernetes.io/link-name"

// ServiceLabel is the name of the service in the service binding object
const ServiceLabel = "app.kubernetes.io/service-name"

// ServiceKind is the kind of the service in the service binding object
const ServiceKind = "app.kubernetes.io/service-kind"

const UriFolder = "kubernetes"

const filePrefix = "odo-service-"

// DeleteOperatorService deletes an Operator backed service
// TODO: make it unlink the service from component as a part of
// https://github.com/openshift/odo/issues/3563
func DeleteOperatorService(client kclient.ClientInterface, serviceName string) error {
	kind, name, err := SplitServiceKindName(serviceName)
	if err != nil {
		return errors.Wrapf(err, "Refer %q to see list of running services", serviceName)
	}

	csv, err := client.GetCSVWithCR(kind)
	if err != nil {
		return err
	}

	if csv == nil {
		return fmt.Errorf("unable to find any Operator providing the service %q", kind)
	}

	crs := client.GetCustomResourcesFromCSV(csv)
	var cr *olm.CRDDescription

	for _, c := range *crs {
		customResource := c
		if customResource.Kind == kind {
			cr = &customResource
			break
		}
	}

	group, version, resource := kclient.GetGVRFromCR(cr)

	return client.DeleteDynamicResource(name, group, version, resource)
}

// ListOperatorServices lists all operator backed services.
// It returns list of services, slice of services that it failed (if any) to list and error (if any)
func ListOperatorServices(client kclient.ClientInterface) ([]unstructured.Unstructured, []string, error) {
	klog.V(4).Info("Getting list of services")

	// First let's get the list of all the operators in the namespace
	csvs, err := client.ListClusterServiceVersions()
	if err != nil {
		return nil, nil, err
	}

	if err != nil {
		return nil, nil, errors.Wrap(err, "Unable to list operator backed services")
	}

	var allCRInstances []unstructured.Unstructured
	var failedListingCR []string

	// let's get the Services a.k.a Custom Resources (CR) defined by each operator, one by one
	for _, csv := range csvs.Items {
		clusterServiceVersion := csv
		klog.V(4).Infof("Getting services started from operator: %s", clusterServiceVersion.Name)
		customResources := client.GetCustomResourcesFromCSV(&clusterServiceVersion)

		// list and write active instances of each service/CR
		var instances []unstructured.Unstructured
		for _, cr := range *customResources {
			customResource := cr

			list, err := GetCRInstances(client, &customResource)
			if err != nil {
				crName := strings.Join([]string{csv.Name, cr.Kind}, "/")
				klog.V(4).Infof("Failed to list instances of %q with error: %s", crName, err.Error())
				failedListingCR = append(failedListingCR, crName)
				break
			}

			if len(list.Items) > 0 {
				instances = append(instances, list.Items...)
			}
		}

		// assuming there are more than one instances of a CR
		allCRInstances = append(allCRInstances, instances...)
	}

	return allCRInstances, failedListingCR, nil
}

func GetGVRFromOperator(csv olm.ClusterServiceVersion, cr string) (string, string, string, error) {
	var group, version, resource string

	for _, customresource := range csv.Spec.CustomResourceDefinitions.Owned {
		custRes := customresource
		if custRes.Kind == cr {
			group, version, resource = kclient.GetGVRFromCR(&custRes)
			return group, version, resource, nil
		}
	}
	return "", "", "", fmt.Errorf("couldn't parse group, version, resource from Operator %q", csv.Name)
}

func GetGVKFromCR(cr *olm.CRDDescription) (group, version, kind string, err error) {
	return getGVKFromCR(cr)
}

// getGVKFromCR parses and returns the values for group, version and resource
// for a given Custom Resource (CR).
func getGVKFromCR(cr *olm.CRDDescription) (group, version, kind string, err error) {
	kind = cr.Kind
	version = cr.Version

	gr := strings.SplitN(cr.Name, ".", 2)
	if len(gr) != 2 {
		err = fmt.Errorf("couldn't split Custom Resource's name into two: %s", cr.Name)
		return
	}
	group = gr[1]

	return
}

// GetAlmExample fetches the ALM example from an Operator's definition. This
// example contains the example yaml to be used to spin up a service for a
// given CR in an Operator
func GetAlmExample(csv olm.ClusterServiceVersion, cr, serviceType string) (almExample map[string]interface{}, err error) {
	var almExamples []map[string]interface{}

	val, ok := csv.Annotations["alm-examples"]
	if ok {
		err = json.Unmarshal([]byte(val), &almExamples)
		if err != nil {
			return nil, errors.Wrap(err, "unable to unmarshal alm-examples")
		}
	} else {
		// There's no alm examples in the CSV's definition
		return nil,
			fmt.Errorf("could not find alm-examples in %q Operator's definition", cr)
	}

	almExample, err = getAlmExample(almExamples, cr, serviceType)
	if err != nil {
		return nil, err
	}

	return almExample, nil
}

// getAlmExample returns the alm-example for exact service of an Operator
func getAlmExample(almExamples []map[string]interface{}, crd, operator string) (map[string]interface{}, error) {
	for _, example := range almExamples {
		if example["kind"].(string) == crd {
			// Remove metadata.namespace from example
			if metadata, ok := example["metadata"].(map[string]interface{}); ok {
				delete(metadata, "namespace")
			}
			return example, nil
		}
	}
	return nil, errors.Errorf("could not find example yaml definition for %q service in %q Operator's definition.", crd, operator)
}

// GetCRInstances fetches and returns instances of the CR provided in the
// "customResource" field. It also returns error (if any)
func GetCRInstances(client kclient.ClientInterface, customResource *olm.CRDDescription) (*unstructured.UnstructuredList, error) {
	klog.V(4).Infof("Getting instances of: %s\n", customResource.Name)

	group, version, resource := kclient.GetGVRFromCR(customResource)

	instances, err := client.ListDynamicResource(group, version, resource)
	if err != nil {
		return nil, err
	}

	return instances, nil
}

// IsOperatorServiceNameValid checks if the provided name follows
// <service-type>/<service-name> format. For example: "EtcdCluster/example" is
// a valid service name but "EtcdCluster/", "EtcdCluster", "example" aren't.
func IsOperatorServiceNameValid(name string) (string, string, error) {
	checkName := strings.SplitN(name, "/", 2)

	if len(checkName) != 2 || checkName[0] == "" || checkName[1] == "" {
		return "", "", fmt.Errorf("invalid service name. Must adhere to <service-type>/<service-name> formatting. For example: %q. Execute %q for list of services", "EtcdCluster/example", "odo service list")
	}
	return checkName[0], checkName[1], nil
}

// OperatorSvcExists checks whether an Operator backed service with given name
// exists or not. It takes 'serviceName' of the format
// '<service-kind>/<service-name>'. For example: EtcdCluster/example.
// It doesn't bother about application since
// https://github.com/openshift/odo/issues/2801 is blocked
func OperatorSvcExists(client kclient.ClientInterface, serviceName string) (bool, error) {
	kind, name, err := SplitServiceKindName(serviceName)
	if err != nil {
		return false, errors.Wrapf(err, "Refer %q to see list of running services", serviceName)
	}

	// Get the CSV (Operator) that provides the CR
	csv, err := client.GetCSVWithCR(kind)
	if err != nil {
		return false, err
	}

	// Get the specific CR that matches "kind"
	crs := client.GetCustomResourcesFromCSV(csv)

	var cr *olm.CRDDescription
	for _, custRes := range *crs {
		c := custRes
		if c.Kind == kind {
			cr = &c
			break
		}
	}

	// Get instances of the specific CR
	crInstances, err := GetCRInstances(client, cr)
	if err != nil {
		return false, err
	}

	for _, s := range crInstances.Items {
		if s.GetKind() == kind && s.GetName() == name {
			return true, nil
		}
	}

	return false, nil
}

// SplitServiceKindName splits the service name provided for deletion by the
// user. It has to be of the format <service-kind>/<service-name>. Example: EtcdCluster/myetcd
func SplitServiceKindName(serviceName string) (string, string, error) {
	sn := strings.SplitN(serviceName, "/", 2)
	if len(sn) != 2 || sn[0] == "" || sn[1] == "" {
		return "", "", fmt.Errorf("couldn't split %q into exactly two", serviceName)
	}

	kind := sn[0]
	name := sn[1]

	return kind, name, nil
}

// IsDefined checks if a service with the given name is defined in a DevFile
func IsDefined(name string, devfileObj parser.DevfileObj) (bool, error) {
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return false, err
	}
	for _, c := range components {
		if c.Name == name {
			return true, nil
		}
	}
	return false, nil
}

// ListDevfileLinks returns the names of the links defined in a Devfile
func ListDevfileLinks(devfileObj parser.DevfileObj, context string) ([]string, error) {
	return listDevfileLinks(devfileObj, context, devfilefs.DefaultFs{})
}

func listDevfileLinks(devfileObj parser.DevfileObj, context string, fs devfilefs.Filesystem) ([]string, error) {
	if devfileObj.Data == nil {
		return nil, nil
	}
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: parsercommon.ComponentOptions{ComponentType: devfile.KubernetesComponentType},
	})
	if err != nil {
		return nil, err
	}
	var services []string
	for _, c := range components {
		u, err := GetK8sComponentAsUnstructured(c.Kubernetes, context, fs)
		if err != nil {
			return nil, err
		}
		if !isLinkResource(u.GetKind()) {
			continue
		}
		var sbr servicebinding.ServiceBinding
		js, err := u.MarshalJSON()
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(js, &sbr)
		if err != nil {
			return nil, err
		}
		sbrServices := sbr.Spec.Services
		if len(sbrServices) != 1 {
			return nil, errors.New("ServiceBinding should have only one service")
		}
		service := sbrServices[0]
		if service.Kind == "Service" {
			services = append(services, service.Name)
		} else {
			services = append(services, service.Kind+"/"+service.Name)
		}
	}
	return services, nil
}

// ListDevfileServices returns the names of the services defined in a Devfile
func ListDevfileServices(client kclient.ClientInterface, devfileObj parser.DevfileObj, componentContext string) (map[string]unstructured.Unstructured, error) {
	return listDevfileServices(client, devfileObj, componentContext, devfilefs.DefaultFs{})
}

func listDevfileServices(client kclient.ClientInterface, devfileObj parser.DevfileObj, componentContext string, fs devfilefs.Filesystem) (map[string]unstructured.Unstructured, error) {
	if devfileObj.Data == nil {
		return nil, nil
	}
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: parsercommon.ComponentOptions{ComponentType: devfile.KubernetesComponentType},
	})
	if err != nil {
		return nil, err
	}

	csvSupported, err := client.IsCSVSupported()
	if err != nil {
		return nil, err
	}
	var operatorGVRList []meta.RESTMapping
	if csvSupported {
		operatorGVRList, err = client.GetOperatorGVRList()
		if err != nil {
			return nil, err
		}
	}

	services := map[string]unstructured.Unstructured{}
	for _, c := range components {
		u, err := GetK8sComponentAsUnstructured(c.Kubernetes, componentContext, fs)
		if err != nil {
			return nil, err
		}
		restMapping, err := client.GetRestMappingFromUnstructured(u)
		if err != nil {
			// getting a RestMapping would fail if there are no matches for the Kind field on the cluster
			// this could be a case when an Operator backed service was added to devfile while working on a cluster
			// that had the Operator installed but "odo service list" is run when that Operator is either no longer
			// available or on a different cluster
			services[strings.Join([]string{u.GetKind(), c.Name}, "/")] = u
			continue
		}
		var match bool
		for _, i := range operatorGVRList {
			if i.Resource == restMapping.Resource {
				// if it's an Operator backed service, it will match; if it's Pod, Deployment, etc. it won't
				match = true
				break
			}
		}
		if match {
			services[strings.Join([]string{u.GetKind(), c.Name}, "/")] = u
		}
	}
	// final list of services includes Operator backed services both supported and unsupported by the underlying k8s cluster
	// but it doesn't include things like Pod, Deployment, etc.
	return services, nil
}

// FindDevfileServiceBinding returns the name of the ServiceBinding defined in a Devfile matching kind and name
func FindDevfileServiceBinding(devfileObj parser.DevfileObj, kind string, name, context string) (string, bool, error) {
	return findDevfileServiceBinding(devfileObj, kind, name, context, devfilefs.DefaultFs{})
}

func findDevfileServiceBinding(devfileObj parser.DevfileObj, kind string, name, context string, fs devfilefs.Filesystem) (string, bool, error) {
	if devfileObj.Data == nil {
		return "", false, nil
	}
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: parsercommon.ComponentOptions{ComponentType: devfile.KubernetesComponentType},
	})
	if err != nil {
		return "", false, err
	}

	for _, c := range components {
		u, err := GetK8sComponentAsUnstructured(c.Kubernetes, context, fs)
		if err != nil {
			return "", false, err
		}
		if isLinkResource(u.GetKind()) {
			var sbr servicebinding.ServiceBinding
			js, err := u.MarshalJSON()
			if err != nil {
				return "", false, err
			}
			err = json.Unmarshal(js, &sbr)
			if err != nil {
				return "", false, err
			}
			services := sbr.Spec.Services
			if len(services) != 1 {
				continue
			}
			service := services[0]
			if service.Kind == kind && service.Name == name {
				return u.GetName(), true, nil
			}
		}
	}
	return "", false, nil
}

// AddKubernetesComponentToDevfile adds service definition to devfile as an inlined Kubernetes component
func AddKubernetesComponentToDevfile(crd, name string, devfileObj parser.DevfileObj) error {
	err := devfileObj.Data.AddComponents([]devfile.Component{{
		Name: name,
		ComponentUnion: devfile.ComponentUnion{
			Kubernetes: &devfile.KubernetesComponent{
				K8sLikeComponent: devfile.K8sLikeComponent{
					BaseComponent: devfile.BaseComponent{},
					K8sLikeComponentLocation: devfile.K8sLikeComponentLocation{
						Inlined: crd,
					},
				},
			},
		},
	}})
	if err != nil {
		return err
	}

	return devfileObj.WriteYamlDevfile()
}

// AddKubernetesComponent adds the crd information to a separate file and adds the uri information to a devfile component
func AddKubernetesComponent(crd, name, componentContext string, devfile parser.DevfileObj) error {
	return addKubernetesComponent(crd, name, componentContext, devfile, devfilefs.DefaultFs{})
}

// AddKubernetesComponent adds the crd information to a separate file and adds the uri information to a devfile component
func addKubernetesComponent(crd, name, componentContext string, devfileObj parser.DevfileObj, fs devfilefs.Filesystem) error {
	filePath := filepath.Join(componentContext, UriFolder, filePrefix+name+".yaml")
	if _, err := fs.Stat(filepath.Join(componentContext, UriFolder)); os.IsNotExist(err) {
		err = fs.MkdirAll(filepath.Join(componentContext, UriFolder), os.ModePerm)
		if err != nil {
			return err
		}
	}

	if _, err := fs.Stat(filePath); !os.IsNotExist(err) {
		return fmt.Errorf("the file %q already exists", filePath)
	}

	err := fs.WriteFile(filePath, []byte(crd), 0755)
	if err != nil {
		return err
	}

	err = devfileObj.Data.AddComponents([]devfile.Component{{
		Name: name,
		ComponentUnion: devfile.ComponentUnion{
			Kubernetes: &devfile.KubernetesComponent{
				K8sLikeComponent: devfile.K8sLikeComponent{
					BaseComponent: devfile.BaseComponent{},
					K8sLikeComponentLocation: devfile.K8sLikeComponentLocation{
						Uri: filepath.Join(UriFolder, filePrefix+name+".yaml"),
					},
				},
			},
		},
	}})
	if err != nil {
		return err
	}

	return devfileObj.WriteYamlDevfile()
}

// DeleteKubernetesComponentFromDevfile deletes an inlined Kubernetes component from devfile, if one exists
func DeleteKubernetesComponentFromDevfile(name string, devfileObj parser.DevfileObj, componentContext string) error {
	return deleteKubernetesComponentFromDevfile(name, devfileObj, componentContext, devfilefs.DefaultFs{})
}

// deleteKubernetesComponentFromDevfile deletes an inlined Kubernetes component from devfile, if one exists
func deleteKubernetesComponentFromDevfile(name string, devfileObj parser.DevfileObj, componentContext string, fs devfilefs.Filesystem) error {
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}

	found := false
	for _, c := range components {
		if c.Name == name {
			err = devfileObj.Data.DeleteComponent(c.Name)
			if err != nil {
				return err
			}

			if c.Kubernetes.Uri != "" {
				parsedURL, err := url.Parse(c.Kubernetes.Uri)
				if err != nil {
					return err
				}
				if len(parsedURL.Host) == 0 || len(parsedURL.Scheme) == 0 {
					err := fs.Remove(filepath.Join(componentContext, c.Kubernetes.Uri))
					if err != nil {
						return err
					}
				}
			}
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("could not find the service %q in devfile", name)
	}

	return devfileObj.WriteYamlDevfile()
}

// PushKubernetesResources updates service(s) from Kubernetes Inlined component in a devfile by creating new ones or removing old ones
func PushKubernetesResources(client kclient.ClientInterface, k8sComponents []devfile.Component, labels map[string]string, context string) error {
	// check csv support before proceeding
	csvSupported, err := client.IsCSVSupported()
	if err != nil {
		return err
	}

	var deployed map[string]DeployedInfo

	if csvSupported {
		deployed, err = ListDeployedServices(client, labels)
		if err != nil {
			return err
		}

		for key, deployedResource := range deployed {
			if deployedResource.isLinkResource {
				delete(deployed, key)
			}
		}
	}

	madeChange := false

	// create an object on the kubernetes cluster for all the Kubernetes Inlined components
	for _, c := range k8sComponents {
		u, er := GetK8sComponentAsUnstructured(c.Kubernetes, context, devfilefs.DefaultFs{})
		if er != nil {
			return er
		}

		isOperatorBackedService, er := PushKubernetesResource(client, u, labels)
		if er != nil {
			return er
		}
		if csvSupported {
			delete(deployed, u.GetKind()+"/"+u.GetName())
		}
		if isOperatorBackedService {
			log.Successf("Created service %q on the cluster; refer %q to know how to link it to the component", strings.Join([]string{u.GetKind(), u.GetName()}, "/"), "odo link -h")
		}
		madeChange = true
	}

	if csvSupported {
		for key, val := range deployed {
			if isLinkResource(val.Kind) {
				continue
			}
			err = DeleteOperatorService(client, key)
			if err != nil {
				return err

			}

			log.Successf("Deleted service %q from the cluster", key)
			madeChange = true
		}
	}

	if !madeChange {
		log.Success("Services are in sync with the cluster, no changes are required")
	}

	return nil
}

// PushKubernetesResource pushes a Kubernetes resource (u) to the cluster using client
// adding labels to the resource
func PushKubernetesResource(client kclient.ClientInterface, u unstructured.Unstructured, labels map[string]string) (bool, error) {
	if isLinkResource(u.GetKind()) {
		// it's a service binding related resource
		return false, nil
	}

	isOp, err := isOperatorBackedService(client, u)
	if err != nil {
		return false, err
	}

	// add labels to the CRD before creation
	existingLabels := u.GetLabels()
	if isOp {
		u.SetLabels(mergeLabels(existingLabels, labels))
	} else {
		// Kubernetes built-in resource; only set managed-by label to it
		u.SetLabels(mergeLabels(existingLabels, map[string]string{"app.kubernetes.io/managed-by": "odo"}))
	}

	e := createOperatorService(client, u)
	if e != nil {
		if strings.Contains(e.Error(), "already exists") {
			// this could be the case when "odo push" was executed after making change to code but there was no change to the service itself
			// TODO: better way to handle this might be introduced by https://github.com/openshift/odo/issues/4553
			return isOp, nil // this ensures that services slice is not updated
		} else {
			return isOp, e
		}
	}
	return isOp, nil
}

func isOperatorBackedService(client kclient.ClientInterface, u unstructured.Unstructured) (bool, error) {
	restMapping, err := client.GetRestMappingFromUnstructured(u)
	if err != nil {
		return false, err
	}
	// check if the GVR of the CRD belongs to any of the CRs provided by any of the Operators
	// if yes, it is an Operator backed service.
	// if no, it is likely a Kubernetes built-in resource.
	operatorGVRList, err := client.GetOperatorGVRList()
	if err != nil {
		return false, err
	}

	for _, i := range operatorGVRList {
		if i.Resource == restMapping.Resource {
			return true, nil
		}
	}
	return false, nil
}

func GetK8sComponentAsUnstructured(component *devfile.KubernetesComponent, context string, fs devfilefs.Filesystem) (unstructured.Unstructured, error) {
	strCRD := component.Inlined
	var err error
	if component.Uri != "" {
		strCRD, err = getDataFromURI(component.Uri, context, fs)
		if err != nil {
			return unstructured.Unstructured{}, err
		}
	}

	// convert the YAML definition into map[string]interface{} since it's needed to create dynamic resource
	u := unstructured.Unstructured{}
	if err = yaml.Unmarshal([]byte(strCRD), &u.Object); err != nil {
		return unstructured.Unstructured{}, err
	}
	return u, nil
}

func mergeLabels(labels ...map[string]string) map[string]string {
	mergedLabels := map[string]string{}

	for _, l := range labels {
		for k, v := range l {
			mergedLabels[k] = v
		}
	}

	return mergedLabels
}

// DeployedInfo holds information about the services present on the cluster
type DeployedInfo struct {
	Kind           string
	Name           string
	isLinkResource bool
}

func ListDeployedServices(client kclient.ClientInterface, labels map[string]string) (map[string]DeployedInfo, error) {
	deployed := map[string]DeployedInfo{}

	deployedServices, _, err := ListOperatorServices(client)
	if err != nil {
		// We ignore ErrNoSuchOperator error as we can deduce Operator Services are not installed
		return nil, err
	}
	for _, svc := range deployedServices {
		name := svc.GetName()
		kind := svc.GetKind()
		deployedLabels := svc.GetLabels()
		if deployedLabels[applabels.ManagedBy] == "odo" && deployedLabels[componentlabels.ComponentLabel] == labels[componentlabels.ComponentLabel] {
			deployed[kind+"/"+name] = DeployedInfo{
				Kind:           kind,
				Name:           name,
				isLinkResource: isLinkResource(kind),
			}
		}
	}

	return deployed, nil
}

// UpdateServicesWithOwnerReferences adds an owner reference to an inlined Kubernetes resource (except service binding objects)
// if not already present in the list of owner references
func UpdateServicesWithOwnerReferences(client kclient.ClientInterface, k8sComponents []devfile.Component, ownerReference metav1.OwnerReference, context string) error {
	for _, c := range k8sComponents {
		// get the string representation of the YAML definition of a CRD
		u, err := GetK8sComponentAsUnstructured(c.Kubernetes, context, devfilefs.DefaultFs{})
		if err != nil {
			return err
		}

		if isLinkResource(u.GetKind()) {
			// ignore service binding resources
			continue
		}

		restMapping, err := client.GetRestMappingFromUnstructured(u)
		if err != nil {
			return err
		}

		d, err := client.GetDynamicResource(restMapping.Resource.Group, restMapping.Resource.Version, restMapping.Resource.Resource, u.GetName())
		if err != nil {
			return err
		}

		found := false
		for _, ownerRef := range d.GetOwnerReferences() {
			if ownerRef.UID == ownerReference.UID {
				found = true
				break
			}
		}
		if found {
			continue
		}
		d.SetOwnerReferences(append(d.GetOwnerReferences(), ownerReference))

		err = client.UpdateDynamicResource(restMapping.Resource.Group, restMapping.Resource.Version, restMapping.Resource.Resource, u.GetName(), d)
		if err != nil {
			return err
		}
	}
	return nil
}

func isLinkResource(kind string) bool {
	return kind == "ServiceBinding"
}

// createOperatorService creates the given operator on the cluster
// it returns the CR,Kind and errors
func createOperatorService(client kclient.ClientInterface, u unstructured.Unstructured) error {
	gvr, err := client.GetRestMappingFromUnstructured(u)
	if err != nil {
		return err
	}

	// create the service on cluster
	err = client.CreateDynamicResource(u, gvr)
	if err != nil {
		return err
	}
	return err
}

// getDataFromURI gets the data from the given URI
// if the uri is a local path, we use the componentContext to complete the local path
func getDataFromURI(uri, componentContext string, fs devfilefs.Filesystem) (string, error) {

	parsedURL, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	if len(parsedURL.Host) != 0 && len(parsedURL.Scheme) != 0 {
		params := util.HTTPRequestParams{
			URL: uri,
		}
		dataBytes, err := util.DownloadFileInMemoryWithCache(params, 1)
		if err != nil {
			return "", err
		}
		return string(dataBytes), nil
	} else {
		dataBytes, err := fs.ReadFile(filepath.Join(componentContext, uri))
		if err != nil {
			return "", err
		}
		return string(dataBytes), nil
	}
}

// ValidateResourcesExist validates if the Kubernetes inlined components are installed on the cluster
func ValidateResourcesExist(client *kclient.Client, k8sComponents []devfile.Component, context string) error {
	if len(k8sComponents) == 0 {
		return nil
	}

	var unsupportedResources []string
	for _, c := range k8sComponents {
		kindErr, err := ValidateResourceExist(client, c, context)
		if err != nil {
			if kindErr != "" {
				unsupportedResources = append(unsupportedResources, kindErr)
			} else {
				return err
			}
		}
	}

	if len(unsupportedResources) > 0 {
		// tell the user about all the unsupported resources in one message
		return fmt.Errorf("following resource(s) in the devfile are not supported by your cluster; please install corresponding Operator(s) before doing \"odo push\": %s", strings.Join(unsupportedResources, ", "))
	}
	return nil
}

func ValidateResourceExist(client *kclient.Client, k8sComponent devfile.Component, context string) (kindErr string, err error) {
	// get the string representation of the YAML definition of a CRD
	u, err := GetK8sComponentAsUnstructured(k8sComponent.Kubernetes, context, devfilefs.DefaultFs{})
	if err != nil {
		return "", err
	}

	_, err = client.GetRestMappingFromUnstructured(u)
	if err != nil && u.GetKind() != "ServiceBinding" {
		// getting a RestMapping would fail if there are no matches for the Kind field on the cluster;
		// but if it's a "ServiceBinding" resource, we don't add it to unsupported list because odo can create links
		// without having SBO installed
		return u.GetKind(), errors.New("resource not supported")
	}
	return "", nil
}
