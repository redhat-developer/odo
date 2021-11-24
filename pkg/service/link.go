package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/generator"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
	"github.com/ghodss/yaml"
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/log"
	servicebinding "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	"github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/context"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

// PushLinks updates Link(s) from Kubernetes Inlined component in a devfile by creating new ones or removing old ones
// returns true if the component needs to be restarted (when a link has been created or deleted)
// if service binding operator is not present, it will call pushLinksWithoutOperator to create the links without it.
func PushLinks(client kclient.ClientInterface, k8sComponents []devfile.Component, labels map[string]string, deployment *v1.Deployment, context string) (bool, error) {
	serviceBindingSupport, err := client.IsServiceBindingSupported()
	if err != nil {
		return false, err
	}

	if !serviceBindingSupport {
		return pushLinksWithoutOperator(client, k8sComponents, labels, deployment, context)
	}

	return pushLinksWithOperator(client, k8sComponents, labels, deployment, context)
}

// pushLinksWithOperator creates links or deletes links (if service binding operator is installed) between components and services
// returns true if the component needs to be restarted (a secret was generated and added to the deployment)
func pushLinksWithOperator(client kclient.ClientInterface, k8sComponents []devfile.Component, labels map[string]string, deployment *v1.Deployment, context string) (bool, error) {

	ownerReference := generator.GetOwnerReference(deployment)
	deployed, err := ListDeployedServices(client, labels)
	if err != nil {
		return false, err
	}

	for key, deployedResource := range deployed {
		if !deployedResource.isLinkResource {
			delete(deployed, key)
		}
	}

	restartNeeded := false

	// create an object on the kubernetes cluster for all the Kubernetes Inlined components
	for _, c := range k8sComponents {
		// get the string representation of the YAML definition of a CRD
		strCRD := c.Kubernetes.Inlined
		if c.Kubernetes.Uri != "" {
			strCRD, err = getDataFromURI(c.Kubernetes.Uri, context, devfilefs.DefaultFs{})
			if err != nil {
				return false, err
			}
		}

		// convert the YAML definition into map[string]interface{} since it's needed to create dynamic resource
		u := unstructured.Unstructured{}
		if e := yaml.Unmarshal([]byte(strCRD), &u.Object); e != nil {
			return false, e
		}

		if !isLinkResource(u.GetKind()) {
			// operator hub is not installed on the cluster
			// or it's a service binding related resource
			continue
		}

		crdName := u.GetName()
		u.SetOwnerReferences([]metav1.OwnerReference{ownerReference})
		u.SetLabels(labels)

		err = createOperatorService(client, u)
		delete(deployed, u.GetKind()+"/"+crdName)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				// this could be the case when "odo push" was executed after making change to code but there was no change to the service itself
				// TODO: better way to handle this might be introduced by https://github.com/openshift/odo/issues/4553
				continue // this ensures that services slice is not updated
			} else {
				return false, err
			}
		}

		name := u.GetName()
		log.Successf("Created link %q using Service Binding Operator on the cluster; component will be restarted", name)
		restartNeeded = true
	}

	for key, val := range deployed {
		if !isLinkResource(val.Kind) {
			continue
		}
		err = DeleteOperatorService(client, key)
		if err != nil {
			return false, err

		}

		log.Successf("Deleted link %q using Service Binding Operator on the cluster; component will be restarted", key)
		restartNeeded = true
	}

	if !restartNeeded {
		log.Success("Links are in sync with the cluster, no changes are required")
	}
	return restartNeeded, nil
}

// pushLinksWithoutOperator creates links or deletes links (if service binding operator is not installed) between components and services
// returns true if the component needs to be restarted (a secret was generated and added to the deployment)
func pushLinksWithoutOperator(client kclient.ClientInterface, k8sComponents []devfile.Component, labels map[string]string, deployment *v1.Deployment, context string) (bool, error) {

	// check csv support before proceeding
	csvSupport, err := client.IsCSVSupported()
	if err != nil {
		return false, err
	}

	secrets, err := client.ListSecrets(componentlabels.GetSelector(labels[componentlabels.ComponentLabel], labels[applabels.ApplicationLabel]))
	if err != nil {
		return false, err
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
	for _, c := range k8sComponents {
		// get the string representation of the YAML definition of a CRD
		strCRD := c.Kubernetes.Inlined
		if c.Kubernetes.Uri != "" {
			strCRD, err = getDataFromURI(c.Kubernetes.Uri, context, devfilefs.DefaultFs{})
			if err != nil {
				return false, err
			}
		}

		// convert the YAML definition into map[string]interface{} since it's needed to create dynamic resource
		u := unstructured.Unstructured{}
		if e := yaml.Unmarshal([]byte(strCRD), &u.Object); e != nil {
			return false, e
		}

		if !isLinkResource(u.GetKind()) {
			// not a service binding object, thus continue
			continue
		}
		localLinksMap[c.Name] = strCRD
	}

	var processingPipeline pipeline.Pipeline

	deploymentGVR, err := client.GetDeploymentAPIVersion()
	if err != nil {
		return false, err
	}

	var restartRequired bool

	// delete the links not present on the devfile
	for linkName, secretName := range clusterLinksMap {
		if _, ok := localLinksMap[linkName]; !ok {

			// recreate parts of the service binding request for deletion
			var newServiceBinding servicebinding.ServiceBinding
			newServiceBinding.Name = linkName
			newServiceBinding.Namespace = client.GetCurrentNamespace()
			newServiceBinding.Spec.Application = servicebinding.Application{
				Ref: servicebinding.Ref{
					Name:     deployment.Name,
					Group:    deploymentGVR.Group,
					Version:  deploymentGVR.Version,
					Resource: deploymentGVR.Resource,
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
					return false, err
				}
			}
			_, err = processingPipeline.Process(&newServiceBinding)
			if err != nil {
				return false, err
			}

			// since the library currently doesn't delete the secret after unbinding
			// delete the secret manually
			err = client.DeleteSecret(secretName, client.GetCurrentNamespace())
			if err != nil {
				return false, err
			}
			restartRequired = true
			log.Successf("Deleted link %q on the cluster; component will be restarted", linkName)
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
					return false, e
				}

				// get the services and get match them against the component
				serviceCompMap = make(map[string]string)
				for _, service := range services {
					serviceCompMap[service.Name] = service.Labels[componentlabels.ComponentLabel]
				}
			}

			// get the string representation of the YAML definition of a CRD
			var serviceBinding servicebinding.ServiceBinding
			err = yaml.Unmarshal([]byte(strCRD), &serviceBinding)
			if err != nil {
				return false, err
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
				return false, err
			}

			if processingPipeline == nil {
				processingPipeline, err = getPipeline(client)
				if err != nil {
					return false, err
				}
			}

			_, err = processingPipeline.Process(&serviceBinding)
			if err != nil {
				if kerrors.IsForbidden(err) {
					// due to https://github.com/redhat-developer/service-binding-operator/issues/1003
					return false, fmt.Errorf("please install the service binding operator")
				}
				return false, err
			}

			if len(serviceBinding.Status.Secret) == 0 {
				return false, fmt.Errorf("no secret was provided by service binding's pipleine")
			}

			// get the generated secret and update it with the labels and owner reference
			secret, err := client.GetSecret(serviceBinding.Status.Secret, client.GetCurrentNamespace())
			if err != nil {
				return false, err
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
				return false, err
			}
			restartRequired = true
			log.Successf("Created link %q on the cluster; component will be restarted", linkName)
		}
	}

	if restartRequired {
		return true, nil
	} else {
		log.Success("Links are in sync with the cluster, no changes are required")
	}

	return false, nil
}

// getPipeline gets the pipeline to process service binding requests
func getPipeline(client kclient.ClientInterface) (pipeline.Pipeline, error) {
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
	return OdoDefaultBuilder.WithContextProvider(context.Provider(client.GetDynamicClient(), context.ResourceLookup(mgr.GetRESTMapper()))).Build(), nil
}
