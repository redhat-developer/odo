package binding

import (
	"github.com/devfile/library/pkg/devfile/parser"
	servicebinding "github.com/redhat-developer/service-binding-operator/apis/binding/v1alpha1"
)

type Client interface {
	// GetFlags gets the necessary flags for binding
	GetFlags(flags map[string]string) map[string]string
	// Validate returns error if the backend failed to validate; mainly useful for flags backend
	Validate(flags map[string]string) error
	// SelectServiceInstance returns the service to bind to the component
	SelectServiceInstance(flags map[string]string, options []string, serviceMap map[string]servicebinding.Ref) (string, error)
	// AskBindingName returns the name to be set for the binding
	AskBindingName(serviceName, componentName string, flags map[string]string) (string, error)
	// AskBindAsFiles asks if the service should be binded as files
	AskBindAsFiles(flags map[string]string) (bool, error)

	// CreateBinding adds the ServiceBinding manifest to the devfile
	CreateBinding(service string, bindingName string, bindAsFiles bool, obj parser.DevfileObj, serviceMap map[string]servicebinding.Ref, componentContext string) error
	// GetServiceInstances gets a list of all the bindable instance names, an error, and a map of bindable instance name with it's servicebinding.Ref;
	// the map will be passed onto CreateBinding, and SelectServiceInstance
	GetServiceInstances() ([]string, map[string]servicebinding.Ref, error)
}
