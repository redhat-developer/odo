package kclient

import (
	"fmt"
	"time"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

// WaitForDeploymentRollout waits for deployment to finish rollout. Returns the state of the deployment after rollout.
func (c *Client) WaitForDeploymentRollout(deploymentName string) (*appsv1.Deployment, error) {
	klog.V(4).Infof("Waiting for %s deployment rollout", deploymentName)
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
				if deployment.Generation <= deployment.Status.ObservedGeneration {
					cond := getDeploymentCondition(deployment.Status, appsv1.DeploymentProgressing)
					if cond != nil && cond.Reason == timedOutReason {
						failure <- fmt.Errorf("deployment %q exceeded its progress deadline", deployment.Name)
					} else if deployment.Spec.Replicas != nil && deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas {
						klog.V(4).Infof("Waiting for deployment %q rollout to finish: %d out of %d new replicas have been updated...\n", deployment.Name, deployment.Status.UpdatedReplicas, *deployment.Spec.Replicas)
					} else if deployment.Status.Replicas > deployment.Status.UpdatedReplicas {
						klog.V(4).Infof("Waiting for deployment %q rollout to finish: %d old replicas are pending termination...\n", deployment.Name, deployment.Status.Replicas-deployment.Status.UpdatedReplicas)
					} else if deployment.Status.AvailableReplicas < deployment.Status.UpdatedReplicas {
						klog.V(4).Infof("Waiting for deployment %q rollout to finish: %d of %d updated replicas are available...\n", deployment.Name, deployment.Status.AvailableReplicas, deployment.Status.UpdatedReplicas)
					} else {
						s.End(true)
						klog.V(4).Infof("Deployment %q successfully rolled out\n", deployment.Name)
						success <- deployment
					}
				}
				klog.V(4).Infof("Waiting for deployment spec update to be observed...\n")

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
	klog.V(4).Infof("Selectors used for deletion: %s", selector)

	// Delete Deployment
	klog.V(4).Info("Deleting Deployment")

	return c.KubeClient.AppsV1().Deployments(c.Namespace).DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: selector})
}

// CreateDynamicDeployment creates a dynamic deployment for Operator backed service
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
