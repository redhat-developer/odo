package annotations

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const ConfigMapValue = "binding:env:object:configmap"

// IsConfigMap returns true if the annotation value should trigger config map handler.
func IsConfigMap(s string) bool {
	return ConfigMapValue == s
}

// NewConfigMapHandler constructs an annotation handler that can extract related data from config
// maps.
func NewConfigMapHandler(
	client dynamic.Interface,
	bindingInfo *BindingInfo,
	resource unstructured.Unstructured,
	restMapper meta.RESTMapper,
) (Handler, error) {
	return NewResourceHandler(
		client,
		bindingInfo,
		resource,
		schema.GroupVersionResource{
			Version:  "v1",
			Resource: "configmaps",
		},
		&dataPath,
		restMapper,
	)
}
