package component

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/kclient"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
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
