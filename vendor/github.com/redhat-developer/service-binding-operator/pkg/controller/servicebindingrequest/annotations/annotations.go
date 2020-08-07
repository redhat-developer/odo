package annotations

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// bindingType encodes the medium the binding should deliver the configuration value.
type bindingType string

const (
	// BindingTypeVolumeMount indicates the binding should happen through a volume mount.
	BindingTypeVolumeMount bindingType = "volumemount"
	// BindingTypeEnvVar indicates the binding should happen through environment variables.
	BindingTypeEnvVar bindingType = "env"
)

// supportedBindingTypes contains all currently supported binding types.
var supportedBindingTypes = map[bindingType]bool{
	BindingTypeVolumeMount: true,
	BindingTypeEnvVar:      true,
}

// dataPath is the path ConfigMap and Secret resources use to scope their data.
//
// note: it is currently used to provide a pointer to the "data" string, which is the location
// ConfigMap and Secret resources keep user data.
var dataPath = "data"

// Result contains data that has been collected by an annotation handler.
type Result struct {
	// Data contains the annotation data collected by an annotation handler inside a deep structure
	// with its root being the value specified in the Path field.
	Data map[string]interface{}
	// Type indicates where the Object field should be injected in the application; can be either
	// "env" or "volumemount".
	Type bindingType
	// Path is the nested location the collected data can be found in the Data field.
	Path string
}

// Handler should be implemented by types that want to offer a mechanism to provide binding data to
// the system.
type Handler interface {
	// Handle returns binding data.
	Handle() (Result, error)
}

type ErrHandlerNotFound string

func (e ErrHandlerNotFound) Error() string {
	return fmt.Sprintf("could not find handler for annotation value %q", string(e))
}

func IsErrHandlerNotFound(err error) bool {
	_, ok := err.(ErrHandlerNotFound)
	return ok
}

// BuildHandler attempts to create an annotation handler for the given annotationKey and
// annotationValue. kubeClient is required by some annotation handlers, and an error is returned in
// the case it is required by an annotation handler but is not defined.
func BuildHandler(
	kubeClient dynamic.Interface,
	obj *unstructured.Unstructured,
	annotationKey string,
	annotationValue string,
	restMapper meta.RESTMapper,
) (Handler, error) {
	bindingInfo, err := NewBindingInfo(annotationKey, annotationValue)
	if err != nil {
		return nil, err
	}

	val := bindingInfo.Value

	switch {
	case IsAttribute(val):
		return NewAttributeHandler(bindingInfo, *obj), nil
	case IsSecret(val):
		return NewSecretHandler(kubeClient, bindingInfo, *obj, restMapper)
	case IsConfigMap(val):
		return NewConfigMapHandler(kubeClient, bindingInfo, *obj, restMapper)
	default:
		return nil, ErrHandlerNotFound(val)
	}
}
