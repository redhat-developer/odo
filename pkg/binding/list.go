package binding

import (
	"sort"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/labels"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BindingList struct {
	m map[string]api.ServiceBinding
}

func newBindingList() BindingList {
	return BindingList{
		m: map[string]api.ServiceBinding{},
	}
}

func (o *BindingList) add(binding api.ServiceBinding) {
	o.m[binding.Name] = binding
}

func (o *BindingList) toArray() []api.ServiceBinding {
	var result []api.ServiceBinding
	for _, v := range o.m {
		result = append(result, v)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func (o *BindingClient) ListAllBindings(devfileObj parser.DevfileObj, context string) ([]api.ServiceBinding, []string, error) {

	bindingList := newBindingList()
	var namesInDevfile []string

	if devfileObj.Data != nil {
		var err error
		var bindingsInDevfile []api.ServiceBinding
		bindingsInDevfile, err = o.GetBindingsFromDevfile(devfileObj, context)
		if err != nil {
			return nil, nil, err
		}
		for _, binding := range bindingsInDevfile {
			bindingList.add(binding)
			namesInDevfile = append(namesInDevfile, binding.Name)
		}
	}

	specs, bindings, err := o.kubernetesClient.ListServiceBindingsFromAllGroups()
	if err != nil {
		return nil, nil, err
	}

	for i := range specs {
		bindingList, err = o.process(bindingList, &specs[i])
		if err != nil {
			return nil, nil, err
		}
	}

	for i := range bindings {
		bindingList, err = o.process(bindingList, &bindings[i])
		if err != nil {
			return nil, nil, err
		}
	}

	return bindingList.toArray(), namesInDevfile, nil
}

// process gets information about the sb from the cluster
// and adds the running Mode accoring to its labels
// then adds it to the bindingList if not already in the list
func (o *BindingClient) process(bindingList BindingList, sb metav1.Object) (BindingList, error) {
	name := sb.GetName()
	var info api.ServiceBinding
	info, err := o.GetBindingFromCluster(name)
	if err != nil {
		return bindingList, err
	}
	setRunningMode(info, labels.GetMode(sb.GetLabels()))
	bindingList.add(info)
	return bindingList, nil
}

// setRunningMode sets the running mode in the status of the servicebinding,
// initializing the status structure if necessary
func setRunningMode(binding api.ServiceBinding, mode string) api.ServiceBinding {
	runningMode := api.RunningMode(mode)
	if binding.Status == nil {
		binding.Status = &api.ServiceBindingStatus{}
	}
	binding.Status.RunningIn = []api.RunningMode{runningMode}
	return binding
}
