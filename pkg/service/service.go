package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"strings"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/kclient"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog"

	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	servicebinding "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
)

// LinkLabel is the name of the name of the link in the devfile
const LinkLabel = "app.kubernetes.io/link-name"

// ServiceLabel is the name of the service in the service binding object
const ServiceLabel = "app.kubernetes.io/service-name"

// ServiceKind is the kind of the service in the service binding object
const ServiceKind = "app.kubernetes.io/service-kind"

// DeleteOperatorService deletes an Operator backed service
// TODO: make it unlink the service from component as a part of
// https://github.com/redhat-developer/odo/issues/3563
func DeleteOperatorService(client kclient.ClientInterface, serviceName string) error {
	kind, name, err := SplitServiceKindName(serviceName)
	if err != nil {
		return fmt.Errorf("Refer %q to see list of running services: %w", serviceName, err)
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
		return nil, nil, fmt.Errorf("Unable to list operator backed services: %w", err)
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
				continue
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
		u, err := libdevfile.GetK8sComponentAsUnstructured(c.Kubernetes, context, fs)
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

// PushKubernetesResources updates service(s) from Kubernetes Inlined component in a devfile by creating new ones or removing old ones
func PushKubernetesResources(client kclient.ClientInterface, k8sComponents []devfile.Component, labels map[string]string, annotations map[string]string, context string) error {
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

	// create an object on the kubernetes cluster for all the Kubernetes Inlined components
	for _, c := range k8sComponents {
		u, er := libdevfile.GetK8sComponentAsUnstructured(c.Kubernetes, context, devfilefs.DefaultFs{})
		if er != nil {
			return er
		}
		_, er = PushKubernetesResource(client, u, labels, annotations)
		if er != nil {
			return er
		}
		if csvSupported {
			delete(deployed, u.GetKind()+"/"+u.GetName())
		}
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
		}
	}

	return nil
}

// PushKubernetesResource pushes a Kubernetes resource (u) to the cluster using client
// adding labels to the resource
func PushKubernetesResource(client kclient.ClientInterface, u unstructured.Unstructured, labels map[string]string, annotations map[string]string) (bool, error) {
	if isLinkResource(u.GetKind()) {
		// it's a service binding related resource
		return false, nil
	}

	isOp, err := isOperatorBackedService(client, u)
	if err != nil {
		return false, err
	}

	// Add all passed in labels to the k8s resource regardless if it's an operator or not
	u.SetLabels(mergeMaps(u.GetLabels(), labels))

	// Pass in all annotations to the k8s resource
	u.SetAnnotations(mergeMaps(u.GetAnnotations(), annotations))

	err = createOperatorService(client, u)
	return isOp, err
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

func mergeMaps(maps ...map[string]string) map[string]string {
	mergedMaps := map[string]string{}

	for _, l := range maps {
		for k, v := range l {
			mergedMaps[k] = v
		}
	}

	return mergedMaps
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
		if deployedLabels[applabels.ManagedBy] == "odo" && deployedLabels[componentlabels.KubernetesInstanceLabel] == labels[componentlabels.KubernetesInstanceLabel] {
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
		u, err := libdevfile.GetK8sComponentAsUnstructured(c.Kubernetes, context, devfilefs.DefaultFs{})
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

// ValidateResourcesExist validates if the Kubernetes inlined components are installed on the cluster
func ValidateResourcesExist(client kclient.ClientInterface, k8sComponents []devfile.Component, context string) error {
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

func ValidateResourceExist(client kclient.ClientInterface, k8sComponent devfile.Component, context string) (kindErr string, err error) {
	// get the string representation of the YAML definition of a CRD
	u, err := libdevfile.GetK8sComponentAsUnstructured(k8sComponent.Kubernetes, context, devfilefs.DefaultFs{})
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
