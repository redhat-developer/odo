package backend

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type AddBindingBackend interface {
	// Validate returns error if the backend failed to validate; mainly useful for flags backend
	Validate(flags map[string]string) error
	// SelectServiceInstance asks user to select the service to be bound to the component;
	// it returns the service name in the form of '<name> (<kind>.<apigroup>)'
	SelectServiceInstance(serviceName string, serviceMap map[string]unstructured.Unstructured) (string, error)
	// AskBindingName asks for the service name to be set
	AskBindingName(defaultName string, flags map[string]string) (string, error)
	// AskBindAsFiles asks if service should be binded as files
	AskBindAsFiles(flags map[string]string) (bool, error)
}
