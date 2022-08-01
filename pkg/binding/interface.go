package binding

import (
	"github.com/devfile/library/pkg/devfile/parser"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/binding/asker"
)

type Client interface {
	// GetFlags gets the necessary flags for binding
	GetFlags(flags map[string]string) map[string]string
	// GetServiceInstances returns a map of bindable instance name with its unstructured.Unstructured object from the specified namespace, and an error
	GetServiceInstances(namespace string) (map[string]unstructured.Unstructured, error)
	// GetBindingsFromDevfile returns the bindings defined in the devfile with the status extracted from cluster
	GetBindingsFromDevfile(devfileObj parser.DevfileObj, context string) ([]api.ServiceBinding, error)
	// GetBindingFromCluster returns information about a binding in the cluster (either from group binding.operators.coreos.com or servicebinding.io)
	GetBindingFromCluster(name string) (api.ServiceBinding, error)

	// add.go

	// SelectNamespace returns the namespace which services instances should be listed from.
	// An empty return value means that service instances will be listed from the current namespace.
	SelectNamespace(flags map[string]string) (string, error)
	// ValidateAddBinding returns error if the backend failed to validate; mainly useful for flags backend
	// withDevfile indicates if a Devfile is present in the current directory
	ValidateAddBinding(flags map[string]string, withDevfile bool) error
	// SelectServiceInstance returns the service to bind to the component
	SelectServiceInstance(flags map[string]string, serviceMap map[string]unstructured.Unstructured) (string, error)
	// SelectWorkloadInstance returns the workload to bind, when a devfile is not in use
	SelectWorkloadInstance(flags map[string]string) (string, schema.GroupVersionKind, error)
	// AskBindingName returns the name to be set for the binding
	AskBindingName(serviceName, componentName string, flags map[string]string) (string, error)
	// AskBindAsFiles asks if the service should be bound as files
	AskBindAsFiles(flags map[string]string) (bool, error)
	// AskNamingStrategy asks the naming strategy to be used for the binding
	AskNamingStrategy(flags map[string]string) (string, error)
	// AddBindingToDevfile adds the ServiceBinding manifest to the devfile
	AddBindingToDevfile(
		bindingName string,
		bindAsFiles bool,
		serviceNs string,
		namingStrategy string,
		unstructuredService unstructured.Unstructured,
		obj parser.DevfileObj,
	) (parser.DevfileObj, error)
	// AddBinding creates a binding in file and cluster (if options selected)
	// and returns the selected options, the binding definition as string (if option selected)
	// and the filename where definition is written (if options selected)
	AddBinding(
		flags map[string]string,
		bindingName string,
		bindAsFiles bool,
		serviceNs string,
		namingStrategy string,
		unstructuredService unstructured.Unstructured,
		workloadName string,
		workloadGVK schema.GroupVersionKind,
	) (selectedOptions []asker.CreationOption, bindingDef string, filename string, err error)

	// list.go

	// ListAllBindings returns all bindings either defined in the Devfile and/or deployed to the cluster
	// inDevfile contains the names of the bindings at least defined in the devfile
	ListAllBindings(devfileObj parser.DevfileObj, context string) (bindings []api.ServiceBinding, inDevfile []string, err error)

	// remove.go

	// ValidateRemoveBinding validates if the command has adequate arguments/flags
	ValidateRemoveBinding(flags map[string]string) error
	// RemoveBinding removes the binding from devfile
	RemoveBinding(bindingName string, obj parser.DevfileObj) (parser.DevfileObj, error)

	// CheckServiceBindingsInjectionDone checks that all service bindings pointing to component have InjectionReady condition
	CheckServiceBindingsInjectionDone(componentName string, appName string) (bool, error)
}
