package binding

import (
	"sort"
	"strings"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/labels"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BindingSet represents a Set of Bindings, indexed by their names
type BindingSet struct {
	m map[string]api.ServiceBinding
}

// newBindingSet creates a new empty Set of Bindings
func newBindingSet() BindingSet {
	return BindingSet{
		m: map[string]api.ServiceBinding{},
	}
}

// add a new Binding to the set, overriding the value of a previous
// binding with the same name
func (o *BindingSet) add(binding api.ServiceBinding) {
	o.m[binding.Name] = binding
}

// toArray returns the list of bindings in the Set as an array, ordered by their names
func (o *BindingSet) toArray() []api.ServiceBinding {
	var result []api.ServiceBinding
	for _, v := range o.m {
		result = append(result, v)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// ListAllBindings returns the list of Service Binding resources either defined in local Devfile
// or deployed in the current namespace
func (o *BindingClient) ListAllBindings(devfileObj parser.DevfileObj, context string) ([]api.ServiceBinding, []string, error) {

	bindingList := newBindingSet()
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
// and adds the running Mode according to its labels
// then adds it to the bindingList if not already in the list
func (o *BindingClient) process(bindingList BindingSet, sb metav1.Object) (BindingSet, error) {
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
	if binding.Status == nil {
		binding.Status = &api.ServiceBindingStatus{}
	}
	binding.Status.RunningIn = api.NewRunningModes()
	binding.Status.RunningIn.AddRunningMode(api.RunningMode(strings.ToLower(mode)))
	return binding
}
