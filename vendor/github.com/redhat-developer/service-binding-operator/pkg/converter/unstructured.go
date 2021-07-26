package converter

import (
	"errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ToUnstructured converts a runtime object into Unstructured, and can return errors related to it.
func ToUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: data}, nil
}

// ToUnstructuredAsGVK converts a runtime object into Unstructured, and set as given GVK. It can
// return errors related to conversion.
func ToUnstructuredAsGVK(
	obj interface{},
	gvk schema.GroupVersionKind,
) (*unstructured.Unstructured, error) {
	u, err := ToUnstructured(obj)
	if err != nil {
		return nil, err
	}
	u.SetGroupVersionKind(gvk)
	return u, nil
}

// NestedResources returns slice of resources of the type specified by obj arg on the given path inside the given resource represented by a map
// Additionally the function gives an indication if specified resource is found or error if the found slice does not contain resources of the given type
func NestedResources(obj interface{}, resource map[string]interface{}, path ...string) ([]map[string]interface{}, bool, error) {
	val, found, err := unstructured.NestedFieldNoCopy(resource, path...)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, found, nil
	}
	valSlice, ok := val.([]interface{})
	if !ok {
		return nil, true, errors.New("not a slice")
	}
	var containers []map[string]interface{}
	for _, item := range valSlice {
		u, ok := item.(map[string]interface{})
		if !ok {
			return nil, true, errors.New("not a map")
		}
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(u, obj)
		if err != nil {
			return nil, true, err
		}
		containers = append(containers, u)
	}
	return containers, true, nil
}
