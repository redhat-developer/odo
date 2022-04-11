package binding

import (
	"github.com/devfile/library/pkg/devfile/parser"
)

type Client interface {
	GetFlags(flags map[string]string) map[string]string
	Validate(flags map[string]string) error
	SelectServiceInstance(flags map[string]string) (string, error)
	AskBindingName(componentName string, flags map[string]string) (string, error)
	AskBindAsFiles(flags map[string]string) (bool, error)

	CreateBinding(service string, bindingName string, bindAsFiles bool, obj parser.DevfileObj) error
	GetServiceInstances() ([]string, error)
}
