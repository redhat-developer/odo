package occlient

import (
	"encoding/json"
	"fmt"
	appsv1 "github.com/openshift/api/apps/v1"
	appsschema "github.com/openshift/client-go/apps/clientset/versioned/scheme"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/klog"
	"sort"
	"time"
)

// IsDeploymentConfigSupported checks if DeploymentConfig type is present on the cluster
func (c *Client) IsDeploymentConfigSupported() (bool, error) {
	const Group = "apps.openshift.io"
	const Version = "v1"

	return c.isResourceSupported(Group, Version, "deploymentconfigs")
}

// GetDeploymentConfigFromName returns the Deployment Config resource given
// the Deployment Config name
func (c *Client) GetDeploymentConfigFromName(name string) (*appsv1.DeploymentConfig, error) {
	klog.V(3).Infof("Getting DeploymentConfig: %s", name)
	deploymentConfig, err := c.appsClient.DeploymentConfigs(c.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return deploymentConfig, nil
}

// GetDeploymentConfigFromSelector returns the Deployment Config object associated
// with the given selector.
// An error is thrown when exactly one Deployment Config is not found for the
// selector.
func (c *Client) GetDeploymentConfigFromSelector(selector string) (*appsv1.DeploymentConfig, error) {
	deploymentConfigs, err := c.ListDeploymentConfigs(selector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get DeploymentConfigs for the selector: %v", selector)
	}

	numDC := len(deploymentConfigs)
	if numDC == 0 {
		return nil, fmt.Errorf("no Deployment Config was found for the selector: %v", selector)
	} else if numDC > 1 {
		return nil, fmt.Errorf("multiple Deployment Configs exist for the selector: %v. Only one must be present", selector)
	}

	return &deploymentConfigs[0], nil
}

// ListDeploymentConfigs returns an array of Deployment Config
// resources which match the given selector
func (c *Client) ListDeploymentConfigs(selector string) ([]appsv1.DeploymentConfig, error) {
	var dcList *appsv1.DeploymentConfigList
	var err error

	if selector != "" {
		dcList, err = c.appsClient.DeploymentConfigs(c.Namespace).List(metav1.ListOptions{
			LabelSelector: selector,
		})
	} else {
		dcList, err = c.appsClient.DeploymentConfigs(c.Namespace).List(metav1.ListOptions{
			FieldSelector: fields.Set{"metadata.namespace": c.Namespace}.AsSelector().String(),
		})
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to list DeploymentConfigs")
	}
	return dcList.Items, nil
}

// WaitAndGetDC block and waits until the DeploymentConfig has updated it's annotation
// Parameters:
//	name: Name of DC
//	timeout: Interval of time.Duration to wait for before timing out waiting for its rollout
//	waitCond: Function indicating when to consider dc rolled out
// Returns:
//	Updated DC and errors if any
func (c *Client) WaitAndGetDC(name string, desiredRevision int64, timeout time.Duration, waitCond func(*appsv1.DeploymentConfig, int64) bool) (*appsv1.DeploymentConfig, error) {

	w, err := c.appsClient.DeploymentConfigs(c.Namespace).Watch(metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	})
	defer w.Stop()

	if err != nil {
		return nil, errors.Wrapf(err, "unable to watch dc")
	}

	timeoutChannel := time.After(timeout)
	// Keep trying until we're timed out or got a result or got an error
	for {
		select {

		// Timout after X amount of seconds
		case <-timeoutChannel:
			return nil, errors.New("Timed out waiting for annotation to update")

		// Each loop we check the result
		case val, ok := <-w.ResultChan():

			if !ok {
				break
			}
			if e, ok := val.Object.(*appsv1.DeploymentConfig); ok {
				for _, cond := range e.Status.Conditions {
					// using this just for debugging message, so ignoring error on purpose
					jsonCond, _ := json.Marshal(cond)
					klog.V(3).Infof("DeploymentConfig Condition: %s", string(jsonCond))
				}
				// If the annotation has been updated, let's exit
				if waitCond(e, desiredRevision) {
					return e, nil
				}

			}
		}
	}
}

// GetDeploymentConfigLabelValues get label values of given label from objects in project that are matching selector
// returns slice of unique label values
func (c *Client) GetDeploymentConfigLabelValues(label string, selector string) ([]string, error) {

	// List DeploymentConfig according to selectors
	dcList, err := c.appsClient.DeploymentConfigs(c.Namespace).List(metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list DeploymentConfigs")
	}

	// Grab all the matched strings
	var values []string
	for _, elem := range dcList.Items {
		for key, val := range elem.Labels {
			if key == label {
				values = append(values, val)
			}
		}
	}

	// Sort alphabetically
	sort.Strings(values)

	return values, nil
}

// DisplayDeploymentConfigLog logs the deployment config to stdout
func (c *Client) DisplayDeploymentConfigLog(deploymentConfigName string, followLog bool) error {

	// Set standard log options
	deploymentLogOptions := appsv1.DeploymentLogOptions{Follow: false, NoWait: true}

	// If the log is being followed, set it to follow / don't wait
	if followLog {
		// TODO: https://github.com/kubernetes/kubernetes/pull/60696
		// Unable to set to 0, until openshift/client-go updates their Kubernetes vendoring to 1.11.0
		// Set to 1 for now.
		tailLines := int64(1)
		deploymentLogOptions = appsv1.DeploymentLogOptions{Follow: true, NoWait: false, Previous: false, TailLines: &tailLines}
	}

	// RESTClient call to OpenShift
	rd, err := c.appsClient.RESTClient().Get().
		Namespace(c.Namespace).
		Name(deploymentConfigName).
		Resource("deploymentconfigs").
		SubResource("log").
		VersionedParams(&deploymentLogOptions, appsschema.ParameterCodec).
		Stream()
	if err != nil {
		return errors.Wrapf(err, "unable get deploymentconfigs log %s", deploymentConfigName)
	}
	if rd == nil {
		return errors.New("unable to retrieve DeploymentConfig from OpenShift, does your component exist?")
	}

	return util.DisplayLog(followLog, rd, deploymentConfigName)
}

// StartDeployment instantiates a given deployment
// deploymentName is the name of the deployment to instantiate
func (c *Client) StartDeployment(deploymentName string) (string, error) {
	if deploymentName == "" {
		return "", errors.Errorf("deployment name is empty")
	}
	klog.V(3).Infof("Deployment %s started.", deploymentName)
	deploymentRequest := appsv1.DeploymentRequest{
		Name: deploymentName,
		// latest is set to true to prevent image name resolution issue
		// inspired from https://github.com/openshift/origin/blob/882ed02142fbf7ba16da9f8efeb31dab8cfa8889/pkg/oc/cli/rollout/latest.go#L194
		Latest: true,
		Force:  true,
	}
	result, err := c.appsClient.DeploymentConfigs(c.Namespace).Instantiate(deploymentName, &deploymentRequest)
	if err != nil {
		return "", errors.Wrapf(err, "unable to instantiate Deployment for %s", deploymentName)
	}
	klog.V(3).Infof("Deployment %s for DeploymentConfig %s triggered.", deploymentName, result.Name)

	return result.Name, nil
}

// GetPodUsingDeploymentConfig gets the pod using deployment config name
func (c *Client) GetPodUsingDeploymentConfig(componentName, appName string) (*corev1.Pod, error) {
	deploymentConfigName, err := util.NamespaceOpenShiftObject(componentName, appName)
	if err != nil {
		return nil, err
	}

	// Find Pod for component
	podSelector := fmt.Sprintf("deploymentconfig=%s", deploymentConfigName)
	return c.GetKubeClient().GetOnePodFromSelector(podSelector)
}
