package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/libdevfile"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/generator"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/ghodss/yaml"
	v1 "k8s.io/api/apps/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	authv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	odolabels "github.com/redhat-developer/odo/pkg/labels"

	"github.com/redhat-developer/odo/pkg/kclient"

	sboApi "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	sboKubernetes "github.com/redhat-developer/service-binding-operator/pkg/client/kubernetes"
	sboPipeline "github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	sboContext "github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/context"
)

// PushLinks updates Link(s) from Kubernetes Inlined component in a devfile by creating new ones or removing old ones
// if service binding operator is not present, it will call pushLinksWithoutOperator to create the links without it.
func PushLinks(client kclient.ClientInterface, devfileObj parser.DevfileObj, k8sComponents []devfile.Component, labels map[string]string, deployment *v1.Deployment, context string) error {
	serviceBindingSupport, err := client.IsServiceBindingSupported()
	if err != nil {
		return err
	}

	if !serviceBindingSupport {
		klog.V(4).Info("Service Binding Operator is not installed on cluster. Service Binding will be created by odo using SB library.")
		return pushLinksWithoutOperator(client, devfileObj, k8sComponents, labels, deployment, context)
	}

	return pushLinksWithOperator(client, devfileObj, k8sComponents, labels, deployment, context)
}

// pushLinksWithOperator creates links or deletes links (if service binding operator is installed) between components and services
func pushLinksWithOperator(client kclient.ClientInterface, devfileObj parser.DevfileObj, k8sComponents []devfile.Component, labels map[string]string, deployment *v1.Deployment, context string) error {

	ownerReference := generator.GetOwnerReference(deployment)
	deployed, err := ListDeployedServices(client, labels)
	if err != nil {
		return err
	}

	for key, deployedResource := range deployed {
		if !deployedResource.isLinkResource {
			delete(deployed, key)
		}
	}

	// create an object on the kubernetes cluster for all the Kubernetes Inlined components
	var strCRD string
	for _, c := range k8sComponents {
		// get the string representation of the YAML definition of a CRD
		strCRD, err = libdevfile.GetK8sManifestWithVariablesSubstituted(devfileObj, c.Name, context, devfilefs.DefaultFs{})
		if err != nil {
			return err
		}

		// convert the YAML definition into map[string]interface{} since it's needed to create dynamic resource
		u := unstructured.Unstructured{}
		if e := yaml.Unmarshal([]byte(strCRD), &u.Object); e != nil {
			return e
		}

		if !isLinkResource(u.GetKind()) {
			// operator hub is not installed on the cluster
			// or it's a service binding related resource
			continue
		}

		crdName := u.GetName()
		u.SetOwnerReferences([]metav1.OwnerReference{ownerReference})
		u.SetLabels(labels)

		_, err = updateOperatorService(client, u)
		delete(deployed, u.GetKind()+"/"+crdName)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				// TODO: better way to handle this might be introduced by https://github.com/redhat-developer/odo/issues/4553
				continue // this ensures that services slice is not updated
			} else {
				return err
			}
		}
	}

	for key, val := range deployed {
		if !isLinkResource(val.Kind) {
			continue
		}
		err = DeleteOperatorService(client, key)
		if err != nil {
			return err

		}
	}

	return nil
}

// pushLinksWithoutOperator creates links or deletes links (if service binding operator is not installed) between components and services
// returns true if the component needs to be restarted (a secret was generated and added to the deployment)
func pushLinksWithoutOperator(client kclient.ClientInterface, devfileObj parser.DevfileObj, k8sComponents []devfile.Component, labels map[string]string, deployment *v1.Deployment, context string) error {

	// check csv support before proceeding
	csvSupport, err := client.IsCSVSupported()
	if err != nil {
		return err
	}

	secrets, err := client.ListSecrets(odolabels.GetSelector(odolabels.GetComponentName(labels), odolabels.GetAppName(labels), odolabels.ComponentAnyMode, false))
	if err != nil {
		return err
	}

	ownerReferences := generator.GetOwnerReference(deployment)

	clusterLinksMap := make(map[string]string)
	for _, secret := range secrets {
		if value, ok := secret.GetLabels()[LinkLabel]; ok {
			clusterLinksMap[value] = secret.Name
		}
	}

	localLinksMap := make(map[string]string)
	// create an object on the kubernetes cluster for all the Kubernetes Inlined components
	var strCRD string
	for _, c := range k8sComponents {
		// get the string representation of the YAML definition of a CRD
		strCRD, err = libdevfile.GetK8sManifestWithVariablesSubstituted(devfileObj, c.Name, context, devfilefs.DefaultFs{})
		if err != nil {
			return err
		}

		// convert the YAML definition into map[string]interface{} since it's needed to create dynamic resource
		u := unstructured.Unstructured{}
		if e := yaml.Unmarshal([]byte(strCRD), &u.Object); e != nil {
			return e
		}

		if !isLinkResource(u.GetKind()) {
			// not a service binding object, thus continue
			continue
		}
		localLinksMap[c.Name] = strCRD
	}

	var processingPipeline sboPipeline.Pipeline

	deploymentGVK, err := client.GetDeploymentAPIVersion()
	if err != nil {
		return err
	}

	// delete the links not present on the devfile
	for linkName, secretName := range clusterLinksMap {
		if _, ok := localLinksMap[linkName]; !ok {

			// recreate parts of the service binding request for deletion
			var newServiceBinding sboApi.ServiceBinding
			newServiceBinding.Name = linkName
			newServiceBinding.Namespace = client.GetCurrentNamespace()
			newServiceBinding.Spec.Application = sboApi.Application{
				Ref: sboApi.Ref{
					Name:    deployment.Name,
					Group:   deploymentGVK.Group,
					Version: deploymentGVK.Version,
					Kind:    deploymentGVK.Kind,
				},
			}
			newServiceBinding.Status.Secret = secretName

			// set the deletion time stamp to trigger deletion
			timeNow := metav1.Now()
			newServiceBinding.DeletionTimestamp = &timeNow

			// if the pipeline was created before
			// skip deletion
			if processingPipeline == nil {
				processingPipeline, err = getPipeline(client)
				if err != nil {
					return err
				}
			}
			_, err = processingPipeline.Process(&newServiceBinding)
			if err != nil {
				return err
			}

			// since the library currently doesn't delete the secret after unbinding
			// delete the secret manually
			err = client.DeleteSecret(secretName, client.GetCurrentNamespace())
			if err != nil {
				return err
			}
		}
	}

	var serviceCompMap map[string]string

	// create the links
	for linkName, strCRD := range localLinksMap {
		if _, ok := clusterLinksMap[linkName]; !ok {
			if serviceCompMap == nil {
				// prevent listing of services unless required
				services, e := client.ListServices("")
				if e != nil {
					return e
				}

				// get the services and get match them against the component
				serviceCompMap = make(map[string]string)
				for _, service := range services {
					serviceCompMap[service.Name] = odolabels.GetComponentName(service.Labels)
				}
			}

			// get the string representation of the YAML definition of a CRD
			var serviceBinding sboApi.ServiceBinding
			err = yaml.Unmarshal([]byte(strCRD), &serviceBinding)
			if err != nil {
				return err
			}

			if len(serviceBinding.Spec.Services) != 1 {
				continue
			}

			if !csvSupport && !isLinkResource(serviceBinding.Spec.Services[0].Kind) {
				// ignore service binding objects linked to services if csv support is not present on the cluster
				continue
			}

			// set the labels and namespace
			serviceBinding.SetLabels(labels)
			serviceBinding.Namespace = client.GetCurrentNamespace()
			ns := client.GetCurrentNamespace()
			serviceBinding.Spec.Services[0].Namespace = &ns

			_, err = json.MarshalIndent(serviceBinding, " ", " ")
			if err != nil {
				return err
			}

			if processingPipeline == nil {
				processingPipeline, err = getPipeline(client)
				if err != nil {
					return err
				}
			}

			_, err = processingPipeline.Process(&serviceBinding)
			if err != nil {
				if kerrors.IsForbidden(err) {
					// due to https://github.com/redhat-developer/service-binding-operator/issues/1003
					return fmt.Errorf("please install the service binding operator")
				}
				return err
			}

			if len(serviceBinding.Status.Secret) == 0 {
				return fmt.Errorf("no secret was provided by service binding's pipleine")
			}

			// get the generated secret and update it with the labels and owner reference
			secret, err := client.GetSecret(serviceBinding.Status.Secret, client.GetCurrentNamespace())
			if err != nil {
				return err
			}
			secret.Labels = labels
			secret.Labels[LinkLabel] = linkName
			if _, ok := serviceCompMap[serviceBinding.Spec.Services[0].Name]; ok {
				secret.Labels[ServiceLabel] = serviceCompMap[serviceBinding.Spec.Services[0].Name]
			} else {
				secret.Labels[ServiceLabel] = serviceBinding.Spec.Services[0].Name
			}
			secret.Labels[ServiceKind] = serviceBinding.Spec.Services[0].Kind
			if serviceBinding.Spec.Services[0].Kind != "Service" {
				// the service name is stored as kind-name as `/` is not a valid char for labels of kubernetes secrets
				secret.Labels[ServiceLabel] = fmt.Sprintf("%v-%v", serviceBinding.Spec.Services[0].Kind, serviceBinding.Spec.Services[0].Name)
			}
			secret.SetOwnerReferences([]metav1.OwnerReference{ownerReferences})
			_, err = client.UpdateSecret(secret, client.GetCurrentNamespace())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// getPipeline gets the pipeline to process service binding requests
func getPipeline(client kclient.ClientInterface) (sboPipeline.Pipeline, error) {
	mgr, err := ctrl.NewManager(client.GetClientConfig(), ctrl.Options{
		Scheme: runtime.NewScheme(),
		// disable the health probes to prevent binding to them
		HealthProbeBindAddress: "0",
		// disable the prometheus metrics
		MetricsBindAddress: "0",
	})
	if err != nil {
		return nil, err
	}

	authClient, err := authv1.NewForConfig(client.GetClientConfig())
	if err != nil {
		return nil, err
	}

	return OdoDefaultBuilder.WithContextProvider(
		sboContext.Provider(
			client.GetDynamicClient(),
			authClient.SubjectAccessReviews(),
			sboKubernetes.ResourceLookup(mgr.GetRESTMapper()),
		),
	).Build(), nil
}
