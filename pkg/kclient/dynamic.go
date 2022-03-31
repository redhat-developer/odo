package kclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog"
)

// CreateDynamicResource creates a dynamic custom resource
func (c *Client) CreateDynamicResource(resource unstructured.Unstructured) error {
	klog.V(5).Infoln("Applying resource via server-side apply:")
	klog.V(5).Infoln(resourceAsJson(resource.Object))
	data, err := json.Marshal(resource.Object)
	if err != nil {
		return fmt.Errorf("unable to marshal resource: %w", err)
	}

	gvr, err := c.GetRestMappingFromUnstructured(resource)
	if err != nil {
		return err
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
func (c *Client) ListDynamicResources(gvr schema.GroupVersionResource) (*unstructured.UnstructuredList, error) {

	if c.DynamicClient == nil {
		return nil, nil
	}

	list, err := c.DynamicClient.Resource(gvr).Namespace(c.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list, nil
}

// GetDynamicResource returns an unstructured instance of a Custom Resource currently deployed in the active namespace
func (c *Client) GetDynamicResource(gvr schema.GroupVersionResource, name string) (*unstructured.Unstructured, error) {
	res, err := c.DynamicClient.Resource(gvr).Namespace(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// UpdateDynamicResource updates a dynamic resource
func (c *Client) UpdateDynamicResource(gvr schema.GroupVersionResource, name string, u *unstructured.Unstructured) error {
	_, err := c.DynamicClient.Resource(gvr).Namespace(c.Namespace).Update(context.TODO(), u, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

type GVRN struct {
	gvr  schema.GroupVersionResource
	name string
}

// DeleteDynamicResource deletes an instance, specified by name, of a Custom Resource
// if wait is true, it will set the PropagationPolicy to DeletePropagationForeground
// to wait for owned resources to be deleted (only for resources with a BlockOwnerDeletion set to true)
func (c *Client) DeleteDynamicResource(name string, gvr schema.GroupVersionResource, wait bool) error {

	doDeleteResource := func() error {
		return c.DynamicClient.Resource(gvr).Namespace(c.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{
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

	// Search resources referencing this resource without BlockOwnerDeletion, to handle waiting their deletion here
	thisRes, err := c.GetDynamicResource(gvr, name)
	if err != nil {
		return err
	}
	all, err := c.GetAllResourcesFromSelector("", c.Namespace)
	if err != nil {
		return err
	}

	var toWait []GVRN
	for _, res := range all {
		ownerRefs := res.GetOwnerReferences()
		for _, ownerRef := range ownerRefs {
			if ownerRef.UID == thisRes.GetUID() {
				if ownerRef.BlockOwnerDeletion == nil || !*ownerRef.BlockOwnerDeletion {
					mapping, err2 := c.GetRestMappingFromUnstructured(res)
					if err2 != nil {
						return err2
					}
					toWait = append(toWait, GVRN{
						gvr:  mapping.Resource,
						name: res.GetName(),
					})
				}
			}
		}
	}

	err = doDeleteResource()
	if err != nil {
		return err
	}
	err = c.WaitDynamicResourceDeleted(gvr, name)
	if err != nil {
		return err
	}
	for _, wait := range toWait {
		err = c.WaitDynamicResourceDeleted(wait.gvr, wait.name)
		if err != nil {
			return err
		}
	}
	return nil
}

// WaitDynamicResourceDeleted waits for the given resource to be deleted, with a timeout
func (c *Client) WaitDynamicResourceDeleted(gvr schema.GroupVersionResource, name string) error {

	watcher, err := c.DynamicClient.Resource(gvr).Namespace(c.Namespace).Watch(context.TODO(), metav1.ListOptions{FieldSelector: "metadata.name=" + name})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	_, err = c.GetDynamicResource(gvr, name)
	if err != nil {
		// deletion is done if the resource does not exist
		if kerrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	for {
		select {
		case <-time.After(time.Minute):
			return fmt.Errorf("timeout while waiting for %q resource to be deleted", name)

		case val, ok := <-watcher.ResultChan():
			if !ok {
				return errors.New("error getting value from resultchan")
			}
			if val.Type == watch.Deleted {
				return nil
			}
		}
	}
}
