package kclient

import (
	"context"
	"encoding/json"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog"

	odolabels "github.com/redhat-developer/odo/pkg/labels"
)

func boolPtr(b bool) *bool {
	return &b
}

const (
	DeploymentKind       = "Deployment"
	DeploymentAPIVersion = "apps/v1"
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
func (c *Client) GetOneDeployment(componentName, appName string, isPartOfComponent bool) (*appsv1.Deployment, error) {
	selector := odolabels.GetSelector(componentName, appName, odolabels.ComponentDevMode, isPartOfComponent)
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

// GetDeploymentFromSelector returns an array of Deployment resources which match the given selector
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

// GetDeploymentAPIVersion returns a map with Group, Version, Resource information of Deployment objects
// depending on the GVR supported by the cluster
func (c *Client) GetDeploymentAPIVersion() (schema.GroupVersionKind, error) {
	extV1Beta1, err := c.IsDeploymentExtensionsV1Beta1()
	if err != nil {
		return schema.GroupVersionKind{}, err
	}

	if extV1Beta1 {
		// this indicates we're running on OCP 3.11 cluster
		return schema.GroupVersionKind{
			Group:   "extensions",
			Version: "v1beta1",
			Kind:    "Deployment",
		}, nil
	}

	return schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	}, nil
}

func (c *Client) IsDeploymentExtensionsV1Beta1() (bool, error) {
	return c.IsResourceSupported("extensions", "v1beta1", "deployments")
}

// DeploymentWatcher returns a watcher on Deployments into the current namespace
// with the given label selector
func (c *Client) DeploymentWatcher(ctx context.Context, selector string) (watch.Interface, error) {
	ns := c.GetCurrentNamespace()
	return c.GetClient().AppsV1().Deployments(ns).
		Watch(ctx, metav1.ListOptions{
			LabelSelector: selector,
		})
}
