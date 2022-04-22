package kclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/util"

	apiMachineryWatch "k8s.io/apimachinery/pkg/watch"

	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
)

func boolPtr(b bool) *bool {
	return &b
}

const (
	DeploymentKind       = "Deployment"
	DeploymentAPIVersion = "apps/v1"

	// TimedOutReason is added in a deployment when its newest replica set fails to show any progress
	// within the given deadline (progressDeadlineSeconds).
	timedOutReason = "ProgressDeadlineExceeded"
)

// GetDeploymentByName gets a deployment by querying by name
func (c *Client) GetDeploymentByName(name string) (*appsv1.Deployment, error) {
	deployment, err := c.KubeClient.AppsV1().Deployments(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	// TODO(pvala): Figure out why Kind and APIVersion are not added to the deployment object
	deployment.APIVersion = DeploymentAPIVersion
	deployment.Kind = DeploymentKind
	return deployment, err
}

// GetOneDeployment returns the Deployment object associated with the given component and app
func (c *Client) GetOneDeployment(componentName, appName string) (*appsv1.Deployment, error) {
	labels := componentlabels.GetLabels(componentName, appName, false)
	labels[componentlabels.OdoModeLabel] = componentlabels.ComponentDevName
	selector := util.ConvertLabelsToSelector(labels)
	return c.GetOneDeploymentFromSelector(selector)
}

// GetOneDeploymentFromSelector returns the Deployment object associated
// with the given selector.
// An error is thrown when exactly one Deployment is not found for the
// selector.
func (c *Client) GetOneDeploymentFromSelector(selector string) (*appsv1.Deployment, error) {
	deployments, err := c.GetDeploymentFromSelector(selector)
	if err != nil {
		return nil, fmt.Errorf("unable to get Deployments for the selector: %v: %w", selector, err)
	}

	num := len(deployments)
	if num == 0 {
		return nil, &DeploymentNotFoundError{Selector: selector}
	} else if num > 1 {
		return nil, fmt.Errorf("multiple Deployments exist for the selector: %v. Only one must be present", selector)
	}

	return &deployments[0], nil
}

// GetDeploymentFromSelector returns an array of Deployment resources which
// match the given selector
func (c *Client) GetDeploymentFromSelector(selector string) ([]appsv1.Deployment, error) {
	var deploymentList *appsv1.DeploymentList
	var err error

	if selector != "" {
		deploymentList, err = c.KubeClient.AppsV1().Deployments(c.Namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: selector,
		})
	} else {
		deploymentList, err = c.KubeClient.AppsV1().Deployments(c.Namespace).List(context.TODO(), metav1.ListOptions{
			FieldSelector: fields.Set{"metadata.namespace": c.Namespace}.AsSelector().String(),
		})
	}
	if err != nil {
		return nil, fmt.Errorf("unable to list Deployments: %w", err)
	}
	return deploymentList.Items, nil
}

// getDeploymentCondition returns the condition with the provided type
// from https://github.com/kubernetes/kubectl/blob/8bc20f428d7d5aed031de5fa160081de7b5af2b0/pkg/util/deployment/deployment.go#L58
func getDeploymentCondition(status appsv1.DeploymentStatus, condType appsv1.DeploymentConditionType) *appsv1.DeploymentCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

// ListDeployments lists all deployments by selector
func (c *Client) ListDeployments(selector string) (*appsv1.DeploymentList, error) {

	return c.KubeClient.AppsV1().Deployments(c.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
}

// WaitForPodDeletion waits for the given pod to be deleted
func (c *Client) WaitForPodDeletion(name string) error {
	watch, err := c.KubeClient.CoreV1().Pods(c.Namespace).Watch(context.TODO(), metav1.ListOptions{FieldSelector: "metadata.name=" + name})
	if err != nil {
		return err
	}
	defer watch.Stop()

	if _, err = c.KubeClient.CoreV1().Pods(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{}); kerrors.IsNotFound(err) {
		return nil
	}

	for {
		select {
		case <-time.After(time.Minute):
			return fmt.Errorf("timeout while waiting for %q pod to be deleted", name)

		case val, ok := <-watch.ResultChan():
			if !ok {
				return errors.New("error getting value from resultchan")
			}
			if val.Type == apiMachineryWatch.Deleted {
				return nil
			}
		}
	}
}

// WaitForDeploymentRollout waits for deployment to finish rollout. Returns the state of the deployment after rollout.
func (c *Client) WaitForDeploymentRollout(deploymentName string) (*appsv1.Deployment, error) {
	klog.V(3).Infof("Waiting for %s deployment rollout", deploymentName)

	w, err := c.KubeClient.AppsV1().Deployments(c.Namespace).Watch(context.TODO(), metav1.ListOptions{FieldSelector: "metadata.name=" + deploymentName})
	if err != nil {
		return nil, fmt.Errorf("unable to watch deployment: %w", err)
	}
	defer w.Stop()

	success := make(chan *appsv1.Deployment)
	failure := make(chan error)

	// Collect all the events in a separate go routine
	failedEvents := make(map[string]corev1.Event)
	quit := make(chan int)
	go c.CollectEvents("", failedEvents, quit)

	go func() {
		defer close(success)
		defer close(failure)

		for {
			val, ok := <-w.ResultChan()
			if !ok {
				failure <- errors.New("watch channel was closed")
				return
			}
			// based on https://github.com/kubernetes/kubectl/blob/9a3954bf653c874c8af6f855f2c754a8e1a44b9e/pkg/polymorphichelpers/rollout_status.go#L66-L91
			if deployment, ok := val.Object.(*appsv1.Deployment); ok {
				for _, cond := range deployment.Status.Conditions {
					// using this just for debugging message, so ignoring error on purpose
					jsonCond, _ := json.Marshal(cond)
					klog.V(3).Infof("Deployment Condition: %s", string(jsonCond))
				}
				if deployment.Generation <= deployment.Status.ObservedGeneration {
					cond := getDeploymentCondition(deployment.Status, appsv1.DeploymentProgressing)
					if cond != nil && cond.Reason == timedOutReason {
						failure <- fmt.Errorf("deployment %q exceeded its progress deadline", deployment.Name)
					} else if deployment.Spec.Replicas != nil && deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas {
						klog.V(3).Infof("Waiting for deployment %q rollout to finish: %d out of %d new replicas have been updated...\n", deployment.Name, deployment.Status.UpdatedReplicas, *deployment.Spec.Replicas)
					} else if deployment.Status.Replicas > deployment.Status.UpdatedReplicas {
						klog.V(3).Infof("Waiting for deployment %q rollout to finish: %d old replicas are pending termination...\n", deployment.Name, deployment.Status.Replicas-deployment.Status.UpdatedReplicas)
					} else if deployment.Status.AvailableReplicas < deployment.Status.UpdatedReplicas {
						klog.V(3).Infof("Waiting for deployment %q rollout to finish: %d of %d updated replicas are available...\n", deployment.Name, deployment.Status.AvailableReplicas, deployment.Status.UpdatedReplicas)
					} else {
						klog.V(3).Infof("Deployment %q successfully rolled out\n", deployment.Name)
						success <- deployment
					}
				}
				klog.V(3).Infof("Waiting for deployment spec update to be observed...\n")

			} else {
				failure <- errors.New("unable to convert event object to Pod")
			}
		}
	}()

	select {
	case val := <-success:
		return val, nil
	case err := <-failure:
		return nil, err
	case <-time.After(5 * time.Minute):
		errorMessage := fmt.Sprintf("timeout while waiting for %s deployment roll out", deploymentName)
		if len(failedEvents) != 0 {
			tableString := getErrorMessageFromEvents(failedEvents)

			errorMessage = errorMessage + fmt.Sprintf(`\nFor more information to help determine the cause of the error, re-run with '-v'.
See below for a list of failed events that occured more than %d times during deployment:
%s`, failedEventCount, tableString.String())

			return nil, fmt.Errorf(errorMessage)
		}

		return nil, fmt.Errorf("timeout while waiting for %s deployment roll out", deploymentName)
	}
}

func resourceAsJson(resource interface{}) string {
	data, _ := json.MarshalIndent(resource, " ", " ")
	return string(data)
}

// CreateDeployment creates a deployment based on the given deployment spec
func (c *Client) CreateDeployment(deploy appsv1.Deployment) (*appsv1.Deployment, error) {
	deployment, err := c.KubeClient.AppsV1().Deployments(c.Namespace).Create(context.TODO(), &deploy, metav1.CreateOptions{FieldManager: FieldManager})
	if err != nil {
		return nil, fmt.Errorf("unable to create Deployment %s: %w", deploy.Name, err)
	}
	return deployment, nil
}

// UpdateDeployment updates a deployment based on the given deployment spec
func (c *Client) UpdateDeployment(deploy appsv1.Deployment) (*appsv1.Deployment, error) {
	deployment, err := c.KubeClient.AppsV1().Deployments(c.Namespace).Update(context.TODO(), &deploy, metav1.UpdateOptions{FieldManager: FieldManager})
	if err != nil {
		return nil, fmt.Errorf("unable to update Deployment %s: %w", deploy.Name, err)
	}
	return deployment, nil
}

// ApplyDeployment creates or updates a deployment based on the given deployment spec
// It is using force:true to make sure that if someone changed one of the values that odo manages,
// odo overrides it with the value it expects instead of failing due to conflict.
func (c *Client) ApplyDeployment(deploy appsv1.Deployment) (*appsv1.Deployment, error) {
	data, err := json.Marshal(deploy)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal deployment: %w", err)
	}
	klog.V(5).Infoln("Applying Deployment via server-side apply:")
	klog.V(5).Infoln(resourceAsJson(deploy))

	err = c.removeDuplicateEnv(deploy.Name)
	if err != nil {
		return nil, err
	}

	deployment, err := c.KubeClient.AppsV1().Deployments(c.Namespace).Patch(context.TODO(), deploy.Name, types.ApplyPatchType, data, metav1.PatchOptions{FieldManager: FieldManager, Force: boolPtr(true)})
	if err != nil {
		return nil, fmt.Errorf("unable to update Deployment %s: %w", deploy.Name, err)
	}
	return deployment, nil
}

// removeDuplicateEnv removes duplicate environment variables from containers, due to a bug in Service Binding Operator:
// https://github.com/redhat-developer/service-binding-operator/issues/983
func (c *Client) removeDuplicateEnv(deploymentName string) error {
	deployment, err := c.KubeClient.AppsV1().Deployments(c.Namespace).Get(context.Background(), deploymentName, metav1.GetOptions{})
	if kerrors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	changes := false
	containers := deployment.Spec.Template.Spec.Containers
	for i := range containers {
		found := map[string]bool{}
		var newEnv []corev1.EnvVar
		for _, env := range containers[i].Env {
			if _, ok := found[env.Name]; !ok {
				found[env.Name] = true
				newEnv = append(newEnv, env)
			} else {
				changes = true
			}
		}
		containers[i].Env = newEnv
	}
	if changes {
		_, err = c.KubeClient.AppsV1().Deployments(c.Namespace).Update(context.Background(), deployment, metav1.UpdateOptions{})
		if kerrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

// DeleteDeployment deletes the deployments with the given selector
func (c *Client) DeleteDeployment(labels map[string]string) error {
	if labels == nil {
		return errors.New("labels for deletion are empty")
	}
	// convert labels to selector
	selector := util.ConvertLabelsToSelector(labels)
	klog.V(3).Infof("Selectors used for deletion: %s", selector)

	// Delete Deployment
	klog.V(3).Info("Deleting Deployment")

	return c.KubeClient.AppsV1().Deployments(c.Namespace).DeleteCollection(context.TODO(), metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: selector})
}

// Define a function that is meant to create patch based on the contents of the deployment
type deploymentPatchProvider func(deployment *appsv1.Deployment) (string, error)

// LinkSecret links a secret to the Deployment of a component
func (c *Client) LinkSecret(secretName, componentName, applicationName string) error {

	var deploymentPatchProvider = func(d *appsv1.Deployment) (string, error) {
		if len(d.Spec.Template.Spec.Containers[0].EnvFrom) > 0 {
			// we always add the link as the first value in the envFrom array. That way we don't need to know the existing value
			return fmt.Sprintf(`[{ "op": "add", "path": "/spec/template/spec/containers/0/envFrom/0", "value": {"secretRef": {"name": "%s"}} }]`, secretName), nil
		}

		// in this case we need to add the full envFrom value
		return fmt.Sprintf(`[{ "op": "add", "path": "/spec/template/spec/containers/0/envFrom", "value": [{"secretRef": {"name": "%s"}}] }]`, secretName), nil
	}

	return c.jsonPatchDeployment(componentlabels.GetSelector(componentName, applicationName), deploymentPatchProvider)
}

// UnlinkSecret unlinks a secret to the Deployment of a component
func (c *Client) UnlinkSecret(secretName, componentName, applicationName string) error {
	// Remove the Secret from the container
	var deploymentPatchProvider = func(d *appsv1.Deployment) (string, error) {
		indexForRemoval := -1
		for i, env := range d.Spec.Template.Spec.Containers[0].EnvFrom {
			if env.SecretRef.Name == secretName {
				indexForRemoval = i
				break
			}
		}

		if indexForRemoval == -1 {
			return "", fmt.Errorf("deployment does not contain a link to %s", secretName)
		}

		return fmt.Sprintf(`[{"op": "remove", "path": "/spec/template/spec/containers/0/envFrom/%d"}]`, indexForRemoval), nil
	}

	return c.jsonPatchDeployment(componentlabels.GetSelector(componentName, applicationName), deploymentPatchProvider)
}

// jsonPatchDeployment will look up the appropriate Deployment, and execute the specified patch
// the whole point of using patch is to avoid race conditions where we try to update
// deployment while it's being simultaneously updated from another source (for example Kubernetes itself)
// this will result in the triggering of a redeployment
func (c *Client) jsonPatchDeployment(deploymentSelector string, deploymentPatchProvider deploymentPatchProvider) error {

	deployment, err := c.GetOneDeploymentFromSelector(deploymentSelector)
	if err != nil {
		return fmt.Errorf("unable to locate Deployment with selector %q: %w", deploymentSelector, err)
	}

	if deploymentPatchProvider != nil {
		patch, e := deploymentPatchProvider(deployment)
		if e != nil {
			return fmt.Errorf("unable to create a patch for the Deployment: %w", e)
		}

		// patch the Deployment with the secret
		_, e = c.KubeClient.AppsV1().Deployments(c.Namespace).Patch(context.TODO(), deployment.Name, types.JSONPatchType, []byte(patch), metav1.PatchOptions{FieldManager: FieldManager})
		if e != nil {
			return fmt.Errorf("deployment not patched %s: %w", deployment.Name, e)
		}
	} else {
		return errors.New("deploymentPatch was not properly set")
	}

	return nil
}

// GetDeploymentLabelValues get label values of given label from objects in project that are matching selector
// returns slice of unique label values
func (c *Client) GetDeploymentLabelValues(label string, selector string) ([]string, error) {

	// List Deployment according to selectors
	deploymentList, err := c.appsClient.Deployments(c.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, fmt.Errorf("unable to list DeploymentConfigs: %w", err)
	}

	// Grab all the matched strings
	var values []string
	for _, elem := range deploymentList.Items {
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

// GetDeploymentAPIVersion returns a map with Group, Version, Resource information of Deployment objects
// depending on the GVR supported by the cluster
func (c *Client) GetDeploymentAPIVersion() (metav1.GroupVersionResource, error) {
	extV1Beta1, err := c.IsDeploymentExtensionsV1Beta1()
	if err != nil {
		return metav1.GroupVersionResource{}, err
	}

	if extV1Beta1 {
		// this indicates we're running on OCP 3.11 cluster
		return metav1.GroupVersionResource{
			Group:    "extensions",
			Version:  "v1beta1",
			Resource: "deployments",
		}, nil
	}

	return metav1.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}, nil
}

func (c *Client) IsDeploymentExtensionsV1Beta1() (bool, error) {
	return c.IsResourceSupported("extensions", "v1beta1", "deployments")
}
