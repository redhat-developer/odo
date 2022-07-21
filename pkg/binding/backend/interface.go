package backend

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/redhat-developer/odo/pkg/binding/asker"
)

type AddBindingBackend interface {
	// SelectNamespace returns the namespace which services instances should be listed from.
	// An empty return value means that service instances will be listed from the current namespace.
	SelectNamespace(flags map[string]string) (string, error)
	// Validate returns error if the backend failed to validate; mainly useful for flags backend
	Validate(flags map[string]string, withDevfile bool) error
	// SelectWorkloadInstance asks user to select the workload to be bind;
	// it returns the workload name in the form of '<name> (<kind>.<apigroup>)'
	SelectWorkloadInstance(workloadName string) (string, schema.GroupVersionKind, error)
	// SelectServiceInstance asks user to select the service to be bound to the component;
	// it returns the service name in the form of '<name> (<kind>.<apigroup>)'
	SelectServiceInstance(serviceName string, serviceMap map[string]unstructured.Unstructured) (string, error)
	// AskBindingName asks for the service name to be set
	AskBindingName(defaultName string, flags map[string]string) (string, error)
	// AskBindAsFiles asks if service should be binded as files
	AskBindAsFiles(flags map[string]string) (bool, error)
	// AskNamingStrategy asks for the naming strategy to be used
	AskNamingStrategy(flags map[string]string) (string, error)
	// SelectCreationOption asks to select how to output the created servicebinding
	SelectCreationOptions(flags map[string]string) ([]asker.CreationOption, error)
	// AskOutputFilePath asks for the path of the file to output service binding
	AskOutputFilePath(flags map[string]string, defaultValue string) (string, error)
}
