package binding

import (
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/api"
)

func (o *BindingClient) ListAllBindings(devfileObj parser.DevfileObj, context string) ([]api.ServiceBinding, []string, error) {

	bindingsInDevfile, err := o.GetBindingsFromDevfile(devfileObj, context)
	if err != nil {
		return nil, nil, err
	}

	inDevfile := make([]string, 0, len(bindingsInDevfile))
	for _, binding := range bindingsInDevfile {
		inDevfile = append(inDevfile, binding.Name)
	}

	bindingsMap := map[string]api.ServiceBinding{}
	for _, binding := range bindingsInDevfile {
		bindingsMap[binding.Name] = binding
	}

	specs, bindings, err := o.kubernetesClient.ListServiceBindingsFromAllGroups()
	if err != nil {
		return nil, nil, err
	}

	allNames := make([]string, 0, len(specs)+len(bindings))
	for _, spec := range specs {
		name := spec.GetName()
		if _, found := bindingsMap[name]; !found {
			allNames = append(allNames, name)
		}
	}

	for _, binding := range bindings {
		name := binding.GetName()
		if _, found := bindingsMap[name]; !found {
			allNames = append(allNames, name)
		}
	}

	for _, name := range allNames {
		var info api.ServiceBinding
		info, err = o.GetBindingFromCluster(name)
		if err != nil {
			return nil, nil, err
		}
		bindingsMap[name] = info
	}

	result := make([]api.ServiceBinding, 0, len(bindingsMap))
	for _, v := range bindingsMap {
		result = append(result, v)
	}

	return result, inDevfile, err
}
