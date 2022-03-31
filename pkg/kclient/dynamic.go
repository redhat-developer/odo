package kclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	apiMachineryWatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog"
)

// CreateDynamicResource creates a dynamic custom resource
func (c *Client) CreateDynamicResource(resource unstructured.Unstructured, gvr *meta.RESTMapping) error {
	klog.V(5).Infoln("Applying resource via server-side apply:")
	klog.V(5).Infoln(resourceAsJson(resource.Object))
	data, err := json.Marshal(resource.Object)
	if err != nil {
		return fmt.Errorf("unable to marshal resource: %w", err)
	}

	// Patch the dynamic resource
	_, err = c.DynamicClient.Resource(gvr.Resource).Namespace(c.Namespace).Patch(context.TODO(), resource.GetName(), types.ApplyPatchType, data, metav1.PatchOptions{FieldManager: FieldManager, Force: boolPtr(true)})
	if err != nil {
		return err
	}

	return nil
}

// ListDynamicResource returns an unstructured list of instances of a Custom
// Resource currently deployed in the active namespace of the cluster
func (c *Client) ListDynamicResource(group, version, resource string) (*unstructured.UnstructuredList, error) {

	if c.DynamicClient == nil {
		return nil, nil
	}

	deploymentRes := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}

	list, err := c.DynamicClient.Resource(deploymentRes).Namespace(c.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list, nil
}

// GetDynamicResource returns an unstructured instance of a Custom Resource currently deployed in the active namespace
func (c *Client) GetDynamicResource(group, version, resource, name string) (*unstructured.Unstructured, error) {
	deploymentRes := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}

	res, err := c.DynamicClient.Resource(deploymentRes).Namespace(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// UpdateDynamicResource updates a dynamic resource
func (c *Client) UpdateDynamicResource(group, version, resource, name string, u *unstructured.Unstructured) error {
	deploymentRes := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}

	_, err := c.DynamicClient.Resource(deploymentRes).Namespace(c.Namespace).Update(context.TODO(), u, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// DeleteDynamicResource deletes an instance, specified by name, of a Custom Resource
// if wait is true, it will set the PropagationPolicy to DeletePropagationForeground
// to wait for owned resources to be deleted (only for resources with a BlockOwnerDeletion set to true)
func (c *Client) DeleteDynamicResource(name, group, version, resourceName string, wait bool) error {

	resource := schema.GroupVersionResource{Group: group, Version: version, Resource: resourceName}

	doDeleteResource := func() error {
		return c.DynamicClient.Resource(resource).Namespace(c.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{
			PropagationPolicy: func(f metav1.DeletionPropagation) *metav1.DeletionPropagation {
				if wait {
					return &f
				}
				return nil
			}(metav1.DeletePropagationForeground),
		})
	}

	if !wait {
		return doDeleteResource()
	}

	var err error
	var watch apiMachineryWatch.Interface
	watch, err = c.DynamicClient.Resource(resource).Namespace(c.Namespace).Watch(context.TODO(), metav1.ListOptions{FieldSelector: "metadata.name=" + name})
	if err != nil {
		return err
	}
	defer watch.Stop()

	err = doDeleteResource()
	if err != nil {
		return err
	}

	for {
		select {
		case <-time.After(time.Minute):
			return fmt.Errorf("timeout while waiting for %q resource to be deleted", name)

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
