package kclient

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/api"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog"
)

func (c *Client) GetRegistryList() ([]api.Registry, error) {
	namespacedList, err := c.ListDynamicResources("", schema.GroupVersionResource{
		Group:    "registry.devfile.io",
		Version:  "v1alpha1",
		Resource: "devfileregistrieslists",
	}, "")
	if err != nil {
		if !kerrors.IsForbidden(err) && !kerrors.IsUnauthorized(err) {
			return nil, err
		} else {
			klog.V(4).Infof("accessing %q is forbidden or unauthorized", "devfileregistrieslists")
		}
	}

	clusterList, err := c.ListClusterWideDynamicResources(schema.GroupVersionResource{
		Group:    "registry.devfile.io",
		Version:  "v1alpha1",
		Resource: "clusterdevfileregistrieslists",
	}, "")
	if err != nil {
		if !kerrors.IsForbidden(err) && !kerrors.IsUnauthorized(err) {
			return nil, err
		} else {
			klog.V(4).Infof("accessing %q is forbidden or unauthorized", "clusterdevfileregistrieslists")
		}
	}

	result, err := addDevfileRegistries(namespacedList)
	if err != nil {
		return nil, err
	}

	clusterResult, err := addDevfileRegistries(clusterList)
	if err != nil {
		return nil, err
	}

	return append(result, clusterResult...), nil
}

func addDevfileRegistries(list *unstructured.UnstructuredList) ([]api.Registry, error) {
	if list == nil {
		return nil, nil
	}

	result := []api.Registry{}
	for _, item := range list.Items {
		vals, found, err := unstructured.NestedSlice(item.Object, "spec", "devfileRegistries")
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		for _, val := range vals {
			castedVal, ok := val.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("unable to read resource %s/%s", item.GetKind(), item.GetName())
			}
			name, ok := castedVal["name"].(string)
			if !ok {
				return nil, fmt.Errorf("unable to read name in resource %s/%s", item.GetKind(), item.GetName())
			}
			url, ok := castedVal["url"].(string)
			if !ok {
				return nil, fmt.Errorf("unable to read url in resource %s/%s", item.GetKind(), item.GetName())
			}

			skipTLSVerify, ok := castedVal["skipTLSVerify"].(bool)
			if !ok {
				return nil, fmt.Errorf("unable to read skipTLSVerify in resource %s/%s", item.GetKind(), item.GetName())
			}

			result = append(result, api.Registry{
				Name:   name,
				URL:    url,
				Secure: !skipTLSVerify,
			})
		}
	}
	return result, nil
}
