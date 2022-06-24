package backend

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/redhat-developer/odo/pkg/binding/asker"
	"github.com/redhat-developer/odo/pkg/kclient"
)

type selectWorkloadStep int

const (
	step_select_kind selectWorkloadStep = iota
	step_select_name
	step_selected
)

// InteractiveBackend is a backend that will ask information interactively using the `asker` package
type InteractiveBackend struct {
	askerClient      asker.Asker
	kubernetesClient kclient.ClientInterface
}

func NewInteractiveBackend(askerClient asker.Asker, kubernetesClient kclient.ClientInterface) *InteractiveBackend {
	return &InteractiveBackend{
		askerClient:      askerClient,
		kubernetesClient: kubernetesClient,
	}
}

func (o *InteractiveBackend) Validate(_ map[string]string, _ bool) error {
	return nil
}

func (o *InteractiveBackend) SelectWorkloadInstance(_ string) (string, schema.GroupVersionKind, error) {

	step := step_select_kind
	var selectedGVK schema.GroupVersionKind
	var selectedName string
loop:
	for {
		switch step {
		case step_select_kind:
			options, allWorkloadsKinds, err := o.kubernetesClient.GetWorkloadKinds()
			if err != nil {
				return "", schema.GroupVersionKind{}, err
			}
			i, err := o.askerClient.SelectWorkloadResource(options)
			if err != nil {
				return "", schema.GroupVersionKind{}, err
			}
			selectedGVK = allWorkloadsKinds[i]
			step++

		case step_select_name:
			// Get the resources of this kind
			gvr, err := o.kubernetesClient.GetGVRFromGVK(selectedGVK)
			if err != nil {
				return "", schema.GroupVersionKind{}, err
			}
			resourceList, err := o.kubernetesClient.ListDynamicResources(gvr)
			if err != nil {
				return "", schema.GroupVersionKind{}, err
			}

			// Ask to select the name of the resource
			names := make([]string, 0, len(resourceList.Items))
			for _, resource := range resourceList.Items {
				names = append(names, resource.GetName())
			}
			var back bool
			back, selectedName, err = o.askerClient.SelectWorkloadResourceName(names)
			if err != nil {
				return "", schema.GroupVersionKind{}, err
			}
			if back {
				step--
			} else {
				step++
			}

		case step_selected:
			break loop
		}
	}

	// Ask the name if DOES NOT EXIST is selected
	var err error
	if selectedName == "" {
		selectedName, err = o.askerClient.AskWorkloadResourceName()
		if err != nil {
			return "", schema.GroupVersionKind{}, err
		}
	}
	return selectedName, selectedGVK, nil
}

func (o *InteractiveBackend) SelectServiceInstance(_ string, serviceMap map[string]unstructured.Unstructured) (string, error) {
	var options []string
	for name := range serviceMap {
		options = append(options, name)
	}
	return o.askerClient.AskServiceInstance(options)
}

func (o *InteractiveBackend) AskBindingName(defaultName string, _ map[string]string) (string, error) {
	return o.askerClient.AskServiceBindingName(defaultName)
}

func (o *InteractiveBackend) AskBindAsFiles(_ map[string]string) (bool, error) {
	return o.askerClient.AskBindAsFiles()
}

func (o *InteractiveBackend) SelectCreationOptions(flags map[string]string) ([]asker.CreationOption, error) {
	return o.askerClient.SelectCreationOptions()
}

func (o *InteractiveBackend) AskOutputFilePath(flags map[string]string, defaultValue string) (string, error) {
	return o.askerClient.AskOutputFilePath(defaultValue)
}
