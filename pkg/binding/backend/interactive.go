package backend

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/redhat-developer/odo/pkg/binding/asker"
	"github.com/redhat-developer/odo/pkg/kclient"
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

func (o *InteractiveBackend) Validate(_ map[string]string) error {
	return nil
}

func (o *InteractiveBackend) SelectWorkloadInstance(workloadName string) (string, schema.GroupVersionKind, error) {

	// Ask to select the kind
	options, allWorkloadsKinds, err := o.kubernetesClient.GetWorkloadKinds()
	if err != nil {
		return "", schema.GroupVersionKind{}, err
	}
	i, err := o.askerClient.SelectWorkloadResource(options)
	if err != nil {
		return "", schema.GroupVersionKind{}, err
	}
	selectedGVK := allWorkloadsKinds[i]

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
	name, err := o.askerClient.SelectWorkloadResourceName(names)
	if err != nil {
		return "", schema.GroupVersionKind{}, err
	}

	// Ask the name if DOES NOT EXIST is selected
	if name == "" {
		name, err = o.askerClient.AskWorkloadResourceName()
		if err != nil {
			return "", schema.GroupVersionKind{}, err
		}
	}
	return name, selectedGVK, nil
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
	return o.askerClient.SelectCreationOption()
}

func (o *InteractiveBackend) AskOutputFilePath(flags map[string]string, defaultValue string) (string, error) {
	return o.askerClient.AskOutputFilePath(defaultValue)
}
