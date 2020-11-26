package kclient

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

// constants for deployments
const (
	DeploymentKind       = "Deployment"
	DeploymentAPIVersion = "apps/v1"

	// TimedOutReason is added in a deployment when its newest replica set fails to show any progress
	// within the given deadline (progressDeadlineSeconds).
	timedOutReason = "ProgressDeadlineExceeded"
)

// GetDeploymentByName gets a deployment by querying by name
func (c *Client) GetDeploymentByName(name string) (*appsv1.Deployment, error) {
	deployment, err := c.KubeClient.AppsV1().Deployments(c.Namespace).Get(name, metav1.GetOptions{})
	return deployment, err
}

// GetOneDeploymentFromSelector returns the Deployment object associated
// with the given selector.
// An error is thrown when exactly one Deployment is not found for the
// selector.
func (c *Client) GetOneDeploymentFromSelector(selector string) (*appsv1.Deployment, error) {
	deployments, err := c.GetDeploymentFromSelector(selector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get Deployments for the selector: %v", selector)
	}

	num := len(deployments)
	if num == 0 {
		return nil, fmt.Errorf("no Deployment was found for the selector: %v", selector)
	} else if num > 1 {
		return nil, fmt.Errorf("multiple Deployments exist for the selector: %v. Only one must be present", selector)
	}

	return &deployments[0], nil
}

// GetDeploymentsFromSelector returns an array of Deployment resources which
// match the given selector
func (c *Client) GetDeploymentFromSelector(selector string) ([]appsv1.Deployment, error) {
	var deploymentList *appsv1.DeploymentList
	var err error

	if selector != "" {
		deploymentList, err = c.KubeClient.AppsV1().Deployments(c.Namespace).List(metav1.ListOptions{
			LabelSelector: selector,
		})
	} else {
		deploymentList, err = c.KubeClient.AppsV1().Deployments(c.Namespace).List(metav1.ListOptions{
			FieldSelector: fields.Set{"metadata.namespace": c.Namespace}.AsSelector().String(),
		})
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Deployments")
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

	return c.KubeClient.AppsV1().Deployments(c.Namespace).List(metav1.ListOptions{
		LabelSelector: selector,
	})
}

// WaitForDeploymentRollout waits for deployment to finish rollout. Returns the state of the deployment after rollout.
func (c *Client) WaitForDeploymentRollout(deploymentName string) (*appsv1.Deployment, error) {
	klog.V(3).Infof("Waiting for %s deployment rollout", deploymentName)
	s := log.Spinner("Waiting for component to start")
	defer s.End(false)

	w, err := c.KubeClient.AppsV1().Deployments(c.Namespace).Watch(metav1.ListOptions{FieldSelector: "metadata.name=" + deploymentName})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to watch deployment")
	}
	defer w.Stop()

	success := make(chan *appsv1.Deployment)
	failure := make(chan error)

	go func() {
		defer close(success)
		defer close(failure)

		for {
			val, ok := <-w.ResultChan()
			if !ok {
				failure <- errors.New("watch channel was closed")
				return
			}
			//based on https://github.com/kubernetes/kubectl/blob/9a3954bf653c874c8af6f855f2c754a8e1a44b9e/pkg/polymorphichelpers/rollout_status.go#L66-L91
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
						s.End(true)
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
		return nil, errors.Errorf("timeout while waiting for %s deployment roll out", deploymentName)
	}
}

// CreateDeployment creates a deployment based on the given deployment spec
func (c *Client) CreateDeployment(deploymentSpec appsv1.DeploymentSpec) (*appsv1.Deployment, error) {
	// inherit ObjectMeta from deployment spec so that namespace, labels, owner references etc will be the same
	objectMeta := deploymentSpec.Template.ObjectMeta

	deployment := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       DeploymentKind,
			APIVersion: DeploymentAPIVersion,
		},
		ObjectMeta: objectMeta,
		Spec:       deploymentSpec,
	}

	deploy, err := c.KubeClient.AppsV1().Deployments(c.Namespace).Create(&deployment)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create Deployment %s", objectMeta.Name)
	}
	return deploy, nil
}

// UpdateDeployment updates a deployment based on the given deployment spec
func (c *Client) UpdateDeployment(deploymentSpec appsv1.DeploymentSpec) (*appsv1.Deployment, error) {
	// inherit ObjectMeta from deployment spec so that namespace, labels, owner references etc will be the same
	objectMeta := deploymentSpec.Template.ObjectMeta

	deployment := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       DeploymentKind,
			APIVersion: DeploymentAPIVersion,
		},
		ObjectMeta: objectMeta,
		Spec:       deploymentSpec,
	}

	deploy, err := c.KubeClient.AppsV1().Deployments(c.Namespace).Update(&deployment)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to update Deployment %s", objectMeta.Name)
	}
	return deploy, nil
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

	return c.KubeClient.AppsV1().Deployments(c.Namespace).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: selector})
}

// CreateDynamicResource creates a dynamic custom resource
func (c *Client) CreateDynamicResource(exampleCustomResource map[string]interface{}, group, version, resource string) error {
	deploymentRes := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}

	deployment := &unstructured.Unstructured{
		Object: exampleCustomResource,
	}

	// Create the dynamic resource based on the alm-example for the CRD
	_, err := c.DynamicClient.Resource(deploymentRes).Namespace(c.Namespace).Create(deployment, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

// ListDynamicResource returns an unstructured list of instances of a Custom
// Resource currently deployed in the active namespace of the cluster
func (c *Client) ListDynamicResource(group, version, resource string) (*unstructured.UnstructuredList, error) {
	deploymentRes := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}

	list, err := c.DynamicClient.Resource(deploymentRes).Namespace(c.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list, nil
}

// DeleteDynamicResource deletes an instance, specified by name, of a Custom Resource
func (c *Client) DeleteDynamicResource(name, group, version, resource string) error {
	deploymentRes := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}

	return c.DynamicClient.Resource(deploymentRes).Namespace(c.Namespace).Delete(name, &metav1.DeleteOptions{})
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

		//in this case we need to add the full envFrom value
		return fmt.Sprintf(`[{ "op": "add", "path": "/spec/template/spec/containers/0/envFrom", "value": [{"secretRef": {"name": "%s"}}] }]`, secretName), nil
	}

	return c.patchDeployment(componentName, deploymentPatchProvider)
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
			return "", fmt.Errorf("Deployment does not contain a link to %s", secretName)
		}

		return fmt.Sprintf(`[{"op": "remove", "path": "/spec/template/spec/containers/0/envFrom/%d"}]`, indexForRemoval), nil
	}

	return c.patchDeployment(componentName, deploymentPatchProvider)
}

// this function will look up the appropriate DC, and execute the specified patch
// the whole point of using patch is to avoid race conditions where we try to update
// deployment while it's being simultaneously updated from another source (for example Kubernetes itself)
// this will result in the triggering of a redeployment
func (c *Client) patchDeployment(deploymentName string, deploymentPatchProvider deploymentPatchProvider) error {

	deployment, err := c.GetDeploymentByName(deploymentName)
	if err != nil {
		return errors.Wrapf(err, "Unable to locate Deployment %s", deploymentName)
	}

	if deploymentPatchProvider != nil {
		patch, err := deploymentPatchProvider(deployment)
		if err != nil {
			return errors.Wrap(err, "Unable to create a patch for the Deployment")
		}

		// patch the Deployment with the secret
		_, err = c.KubeClient.AppsV1().Deployments(c.Namespace).Patch(deploymentName, types.JSONPatchType, []byte(patch))
		if err != nil {
			return errors.Wrapf(err, "Deployment not patched %s", deployment.Name)
		}
	} else {
		return errors.Wrapf(err, "deploymentPatch was not properly set")
	}

	return nil
}

// GetDeploymentLabelValues get label values of given label from objects in project that are matching selector
// returns slice of unique label values
func (c *Client) GetDeploymentLabelValues(label string, selector string) ([]string, error) {

	// List DeploymentConfig according to selectors
	dcList, err := c.appsClient.Deployments(c.Namespace).List(metav1.ListOptions{LabelSelector: selector})
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

// GetDeploymentConfigsFromSelector returns an array of Deployment Config
// resources which match the given selector
func (c *Client) GetDeploymentConfigsFromSelector(selector string) ([]appsv1.Deployment, error) {
	var dcList *appsv1.DeploymentList
	var err error

	if selector != "" {
		dcList, err = c.appsClient.Deployments(c.Namespace).List(metav1.ListOptions{
			LabelSelector: selector,
		})
	} else {
		dcList, err = c.appsClient.Deployments(c.Namespace).List(metav1.ListOptions{
			FieldSelector: fields.Set{"metadata.namespace": c.Namespace}.AsSelector().String(),
		})
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to list DeploymentConfigs")
	}
	return dcList.Items, nil
}
