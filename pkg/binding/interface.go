package binding

import (
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/api"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Client interface {
	// GetFlags gets the necessary flags for binding
	GetFlags(flags map[string]string) map[string]string
	// Validate returns error if the backend failed to validate; mainly useful for flags backend
	Validate(flags map[string]string) error
	// SelectServiceInstance returns the service to bind to the component
	SelectServiceInstance(flags map[string]string, serviceMap map[string]unstructured.Unstructured) (string, error)
	// AskBindingName returns the name to be set for the binding
	AskBindingName(serviceName, componentName string, flags map[string]string) (string, error)
	// AskBindAsFiles asks if the service should be bound as files
	AskBindAsFiles(flags map[string]string) (bool, error)

	// AddBinding adds the ServiceBinding manifest to the devfile
	AddBinding(bindingName string, bindAsFiles bool, unstructuredService unstructured.Unstructured, obj parser.DevfileObj, componentContext string) (parser.DevfileObj, error)
	// GetServiceInstances returns a map of bindable instance name with its unstructured.Unstructured object, and an error
	GetServiceInstances() (map[string]unstructured.Unstructured, error)

	GetBindingsFromDevfile(devfileObj parser.DevfileObj, context string) ([]api.ServiceBinding, error)
}
