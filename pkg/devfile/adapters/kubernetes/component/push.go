package component

import (
	"fmt"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/service"
	appsv1 "k8s.io/api/apps/v1"
)

// getComponentDeployment returns the deployment associated with the component, if deployed
// and indicate if the deplyment has been found
func (a *Adapter) getComponentDeployment() (*appsv1.Deployment, bool, error) {
	// Get the Dev deployment:
	// Since `odo deploy` can theoretically deploy a deployment as well with the same instance name
	// we make sure that we are retrieving the deployment with the Dev mode, NOT Deploy.
	selectorLabels := odolabels.GetSelector(a.ComponentName, a.AppName, odolabels.ComponentDevMode)
	deployment, err := a.kubeClient.GetOneDeploymentFromSelector(selectorLabels)

	if err != nil {
		if _, ok := err.(*kclient.DeploymentNotFoundError); !ok {
			return nil, false, fmt.Errorf("unable to determine if component %s exists: %w", a.ComponentName, err)
		}
	}
	componentExists := deployment != nil
	return deployment, componentExists, nil
}

// getPodName returns the name of the pod associated with the component, if any
// An empty name is returned id no pod exists
func (a *Adapter) getPodName() (string, error) {
	var podName string
	// First see if the component does have a pod. it could have been scaled down to zero
	_, err := a.kubeClient.GetOnePodFromSelector(fmt.Sprintf("component=%s", a.ComponentName))
	// If an error occurs, we don't call a.getPod (a blocking function that waits till it finds a pod in "Running" state.)
	// We would rely on a call to a.createOrUpdateComponent to reset the pod count for the component to one.
	if err == nil {
		pod, podErr := a.getPod(nil, true)
		if podErr != nil {
			return "", fmt.Errorf("unable to get pod for component %s: %w", a.ComponentName, podErr)
		}
		podName = pod.GetName()
	}
	return podName, nil
}

// pushKubernetesComponents gets the Kubernetes components from the Devfile and push them to the cluster
// adding the specified labels to them
func (a *Adapter) pushKubernetesComponents(
	labels map[string]string,
) ([]v1alpha2.Component, error) {
	// fetch the "kubernetes inlined components" to create them on cluster
	// from odo standpoint, these components contain yaml manifest of an odo service or an odo link
	k8sComponents, err := devfile.GetKubernetesComponentsToPush(a.Devfile)
	if err != nil {
		return nil, fmt.Errorf("error while trying to fetch service(s) from devfile: %w", err)
	}

	// validate if the GVRs represented by Kubernetes inlined components are supported by the underlying cluster
	err = service.ValidateResourcesExist(a.kubeClient, a.Devfile, k8sComponents, a.Context)
	if err != nil {
		return nil, err
	}

	// Set the annotations for the component type
	annotations := make(map[string]string)
	odolabels.SetProjectType(annotations, component.GetComponentTypeFromDevfileMetadata(a.AdapterContext.Devfile.Data.GetMetadata()))

	// create the Kubernetes objects from the manifest and delete the ones not in the devfile
	err = service.PushKubernetesResources(a.kubeClient, a.Devfile, k8sComponents, labels, annotations, a.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to create service(s) associated with the component: %w", err)
	}
	return k8sComponents, nil
}

func (a *Adapter) getPushDevfileCommands(parameters adapters.PushParameters) (map[devfilev1.CommandGroupKind]devfilev1.Command, error) {
	pushDevfileCommands, err := libdevfile.ValidateAndGetPushCommands(a.Devfile, parameters.DevfileBuildCmd, parameters.DevfileRunCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to validate devfile build and run commands: %w", err)
	}

	if parameters.Debug {
		pushDevfileDebugCommands, e := libdevfile.ValidateAndGetCommand(a.Devfile, parameters.DevfileDebugCmd, devfilev1.DebugCommandGroupKind)
		if e != nil {
			return nil, fmt.Errorf("debug command is not valid: %w", e)
		}
		pushDevfileCommands[devfilev1.DebugCommandGroupKind] = pushDevfileDebugCommands
	}

	return pushDevfileCommands, nil
}
