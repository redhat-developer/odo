package backend

import (
	servicebinding "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
)

type CreateBindingBackend interface {
	// Validate returns error if the backend failed to validate; mainly useful for flags backend
	Validate(flags map[string]string) error
	// SelectServiceInstance asks user to select the service to be bound to the component;
	// it returns the service name in the form of '<name> (<kind>.<apigroup>)'
	SelectServiceInstance(flags map[string]string, options []string, serviceMap map[string]servicebinding.Ref) (string, error)
	// AskBindingName asks for the service name to be set
	AskBindingName(componentName string, flags map[string]string) (string, error)
	// AskBindAsFiles asks if service should be binded as files
	AskBindAsFiles(flags map[string]string) (bool, error)
}
