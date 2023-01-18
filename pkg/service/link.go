package service

import (
	"fmt"

	"github.com/devfile/library/v2/pkg/devfile/generator"
	appsv1 "k8s.io/api/apps/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	authv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"

	sboApi "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
	sboKubernetes "github.com/redhat-developer/service-binding-operator/pkg/client/kubernetes"
	sboPipeline "github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline"
	sboContext "github.com/redhat-developer/service-binding-operator/pkg/reconcile/pipeline/context"
)

// pushLinksWithoutOperator creates links (if service binding operator is not installed) between components and services
func pushLinksWithoutOperator(client kclient.ClientInterface, u unstructured.Unstructured, labels map[string]string) error {

	// check csv support before proceeding
	csvSupport, err := client.IsCSVSupported()
	if err != nil {
		return err
	}

	var serviceCompMap map[string]string

	// create the links
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

	// get the string representation of the YAML definition of a CRD
	var serviceBinding sboApi.ServiceBinding
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(u.UnstructuredContent(), &serviceBinding)
	if err != nil {
		return err
	}

	if len(serviceBinding.Spec.Services) != 1 {
		klog.V(4).Infof("cannot create the link; serviceBinding.Spec.Services != 1; ServiceBinding: %v", serviceBinding)
		return nil
	}

	if !csvSupport && !isLinkResource(serviceBinding.Spec.Services[0].Kind) {
		// ignore service binding objects linked to services if csv support is not present on the cluster
		klog.V(4).Infof("cannot create the link; CSV is not supported")
		return nil
	}

	// set the labels and namespace
	serviceBinding.SetLabels(labels)
	serviceBinding.Namespace = client.GetCurrentNamespace()
	ns := client.GetCurrentNamespace()
	serviceBinding.Spec.Services[0].Namespace = &ns

	var processingPipeline sboPipeline.Pipeline
	processingPipeline, err = getPipeline(client)
	if err != nil {
		return err
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
	sbSecret, err := client.GetSecret(serviceBinding.Status.Secret, client.GetCurrentNamespace())
	if err != nil {
		return err
	}
	sbSecret.Labels = labels
	sbSecret.Labels[LinkLabel] = serviceBinding.Name
	if _, ok := serviceCompMap[serviceBinding.Spec.Services[0].Name]; ok {
		sbSecret.Labels[ServiceLabel] = serviceCompMap[serviceBinding.Spec.Services[0].Name]
	} else {
		sbSecret.Labels[ServiceLabel] = serviceBinding.Spec.Services[0].Name
	}
	sbSecret.Labels[ServiceKind] = serviceBinding.Spec.Services[0].Kind

	if serviceBinding.Spec.Services[0].Kind != "Service" {
		// the service name is stored as kind-name as `/` is not a valid char for labels of kubernetes secrets
		sbSecret.Labels[ServiceLabel] = fmt.Sprintf("%v-%v", serviceBinding.Spec.Services[0].Kind, serviceBinding.Spec.Services[0].Name)
	}

	// Obtain the component deployment to set as owner reference; this ensures the secret gets deleted when deployment does
	deployment, err := client.GetOneDeploymentFromSelector(odolabels.GetSelector(odolabels.GetComponentName(labels), odolabels.GetAppName(labels), odolabels.ComponentAnyMode, true))
	if err != nil {
		return err
	}
	ownerReferences := generator.GetOwnerReference(deployment)
	sbSecret.SetOwnerReferences([]metav1.OwnerReference{ownerReferences})

	_, err = client.UpdateSecret(sbSecret, client.GetCurrentNamespace())
	if err != nil {
		return err
	}

	return nil
}

// UnbindWithLibrary unbinds the component and service using the ServiceBinding library; it does not delete the secret
func UnbindWithLibrary(kubeClient kclient.ClientInterface, secretToUnbind unstructured.Unstructured, deployment *appsv1.Deployment) error {
	var processingPipeline sboPipeline.Pipeline
	deploymentGVK, err := kubeClient.GetDeploymentAPIVersion()
	if err != nil {
		return fmt.Errorf("failed to get deployment GVK: %w", err)
	}
	// build the ServiceBinding object to for unbinding
	var newServiceBinding sboApi.ServiceBinding
	newServiceBinding.Name = secretToUnbind.GetLabels()[LinkLabel]
	newServiceBinding.Namespace = kubeClient.GetCurrentNamespace()
	newServiceBinding.Spec.Application = sboApi.Application{
		Ref: sboApi.Ref{
			Name:    deployment.Name,
			Group:   deploymentGVK.Group,
			Version: deploymentGVK.Version,
			Kind:    deploymentGVK.Kind,
		},
	}
	newServiceBinding.Status.Secret = secretToUnbind.GetName()
	// set the deletion time stamp to trigger deletion
	timeNow := metav1.Now()
	newServiceBinding.DeletionTimestamp = &timeNow

	if processingPipeline == nil {
		processingPipeline, err = getPipeline(kubeClient)
		if err != nil {
			return err
		}
	}
	// use the library to perform unbinding;
	// this will remove all the envvars, volume/secret mounts done on the deployment to bind it to the service
	_, err = processingPipeline.Process(&newServiceBinding)
	return err
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
