package occlient

import (
	"fmt"
	"sync"
	"time"

	"github.com/openshift/odo/pkg/log/fidget"
	"github.com/pkg/errors"

	appsv1 "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Error channels for both the pod and deploymentconfig
var deploymentConfigErrorChannel chan error
var podErrorChannel chan error

// Main wait group for both go routines.
var wg sync.WaitGroup

// waitForDD is a go routine to wait for the DeploymentConfig to successfully complete / come up
func (c *Client) waitForDC(name string, deploymentConfigSpinner *fidget.Spinny) {

	// Create the watchers
	deploymentConfigWatcher, err := c.appsClient.DeploymentConfigs(c.Namespace).Watch(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	})
	if err != nil {
		deploymentConfigErrorChannel <- errors.Wrapf(err, "unable to watch deploymentconfig")
	}

	// This will only happen if the wrong "name" was passed in, or you'll get an invalid memory address / nil pointer reference
	if deploymentConfigWatcher == nil {
		deploymentConfigErrorChannel <- errors.Errorf("unable to retrieve any deploymentconfig by the name %s", name)
	}
	defer deploymentConfigWatcher.Stop()

	for {

		// Wait for some information
		dcVal, ok := <-deploymentConfigWatcher.ResultChan()
		if !ok {
			deploymentConfigErrorChannel <- errors.New("watch channel was closed")
			break
		}

		if e, ok := dcVal.Object.(*appsv1.DeploymentConfig); ok {

			// Output some statuses / current information
			deploymentConfigSpinner.Status(fmt.Sprintf("DeploymentConfig: Requested Replicas (%d) | Available Replicas (%d)", e.Status.Replicas, e.Status.AvailableReplicas))

			// If we're successful
			if e.Status.Replicas > 0 && (e.Status.Replicas == e.Status.AvailableReplicas) {
				deploymentConfigSpinner.Success()
				wg.Done()
				break
			}
		} else {
			deploymentConfigErrorChannel <- errors.New("unable to convert event object to DeploymentConfig")
			break
		}
	}

}

// waitForPod waits for the pod to successfully complete / come up
func (c *Client) waitForPod(podSelector string, podSpinner *fidget.Spinny) {

	// Create the watcher
	podWatcher, err := c.kubeClient.CoreV1().Pods(c.Namespace).Watch(metav1.ListOptions{
		LabelSelector: podSelector,
	})
	if err != nil {
		podErrorChannel <- errors.Wrapf(err, "unable to watch pod")
	}

	// This will only happen if the wrong "name" was passed in, or you'll get an invalid memory address / nil pointer reference
	if podWatcher == nil {
		podErrorChannel <- errors.Errorf("unable to retrieve any pod by the selector %s", podSelector)
	}
	defer podWatcher.Stop()

	for {

		// Wait for something to happen
		val, ok := <-podWatcher.ResultChan()
		if !ok {
			podErrorChannel <- errors.New("watch channel was closed")
			break
		}

		// Evaluate the channel / object you retrieved
		if e, ok := val.Object.(*corev1.Pod); ok {

			// Output some status information
			podSpinner.Status(fmt.Sprintf("Pod: %s", e.Status.Phase))

			switch e.Status.Phase {

			// If we're successful
			case corev1.PodRunning:
				podSpinner.Success()
				wg.Done()
				break

			// If we have failed, we've got to output some more verbose output..
			case corev1.PodFailed, corev1.PodUnknown:
				podSpinner.Fail()
				podErrorChannel <- errors.Errorf("pod %s status %s", e.Name, e.Status.Phase)
				break
			}

		} else {
			podErrorChannel <- errors.New("unable to convert event object to Pod")
			break
		}
	}
	podErrorChannel <- errors.New("unable to convert event object to Pod")
}

// WaitForEverything will wait for EVERYTHING to be RUNNING! (replacement for "WaitAndGetPod" since this far more verbose
// and elaborate..
// TODO:
// Use `kubectl get events -w` for LAST occuring event? Also, we shouldn't really be displaying this if we're redeploying? Else new
// PVC's need to be created, etc.
// For --git build should we output the same stuff? Got to output builconfig too?
// Secrets
// PVC?
// buildconfig?
func (c *Client) WaitForEverything(podSelector string, name string, timeout time.Duration) error {

	// Wait group
	done := make(chan struct{})
	wg.Add(2)

	// Init the channels
	podErrorChannel = make(chan error)
	deploymentConfigErrorChannel = make(chan error)

	// Create the spinners and the initial message information
	podSpinner := fidget.NewSpinny("Pod: Pushing to cluster")
	deploymentConfigSpinner := fidget.NewSpinny("DeploymentConfig: Pushing to cluster")
	spinnerSet := fidget.NewSpinnerSet([]*fidget.Spinny{podSpinner, deploymentConfigSpinner})
	spinnerSet.Title = "Deploying..."
	spinnerSet.Start()
	defer spinnerSet.End(false)

	// Run the checks within go routines
	go c.waitForDC(name, deploymentConfigSpinner)
	go c.waitForPod(podSelector, podSpinner)

	// We will wait until both go routines are done
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for either something to error / fail OR for all goroutines to finish (meaning everything passed..)
	select {
	case <-done:
		spinnerSet.End(true)
		return nil
	case err := <-podErrorChannel:
		return err
	case err := <-deploymentConfigErrorChannel:
		return err
	case <-time.After(timeout):
		return errors.Errorf("timed out, waited %s", timeout)
	}

}
