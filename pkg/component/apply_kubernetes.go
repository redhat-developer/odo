package component

import (
	"fmt"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/service"
)

// ApplyKubernetes contains the logic to create the k8s resources defined by the `apply` command
// mode(Dev, Deploy): the mode in which the resources are deployed
// appName: application name
// devfile: the devfile object
// kubernetes: the kubernetes devfile component to be deployed
// kubeClient: Kubernetes client to be used to deploy the resource
// path: path to the context directory
func ApplyKubernetes(
	mode string,
	appName string,
	componentName string,
	devfile parser.DevfileObj,
	kubernetes devfilev1.Component,
	kubeClient kclient.ClientInterface,
	path string,
) error {
	// TODO: Use GetK8sComponentAsUnstructured here and pass it to ValidateResourcesExistInK8sComponent
	// Validate if the GVRs represented by Kubernetes inlined components are supported by the underlying cluster
	kind, err := ValidateResourcesExistInK8sComponent(kubeClient, devfile, kubernetes, path)
	if err != nil {
		return fmt.Errorf("%s: %w", kind, err)
	}

	// Get the most common labels that's applicable to all resources being deployed.
	// Set the mode. Regardless of what Kubernetes resource we are deploying.
	runtime := GetComponentRuntimeFromDevfileMetadata(devfile.Data.GetMetadata())
	labels := odolabels.GetLabels(componentName, appName, runtime, mode, false)

	klog.V(4).Infof("Injecting labels: %+v into k8s artifact", labels)

	// Create the annotations
	// Retrieve the component type from the devfile and also inject it into the list of annotations
	annotations := make(map[string]string)
	odolabels.SetProjectType(annotations, GetComponentTypeFromDevfileMetadata(devfile.Data.GetMetadata()))

	// Get the Kubernetes component
	uList, err := libdevfile.GetK8sComponentAsUnstructuredList(devfile, kubernetes.Name, path, devfilefs.DefaultFs{})
	if err != nil {
		return err
	}
	for _, u := range uList {
		// Deploy the actual Kubernetes component and error out if there's an issue.
		log.Sectionf("Deploying Kubernetes Component: %s", u.GetName())
		err = service.PushKubernetesResource(kubeClient, u, labels, annotations, mode)
		if err != nil {
			return fmt.Errorf("failed to create service(s) associated with the component: %w", err)
		}
	}
	return nil
}
