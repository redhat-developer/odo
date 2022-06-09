package binding

import (
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/api"
)

func (o *BindingClient) ListAllBindings(devfileObj parser.DevfileObj, context string) ([]api.ServiceBinding, []string, error) {
	bindings, err := o.GetBindingsFromDevfile(devfileObj, context)
	inDevfile := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		inDevfile = append(inDevfile, binding.Name)
	}
	return bindings, inDevfile, err
}
