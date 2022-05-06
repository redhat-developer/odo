package backend

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/redhat-developer/odo/pkg/binding/asker"
)

// InteractiveBackend is a backend that will ask information interactively using the `asker` package
type InteractiveBackend struct {
	askerClient asker.Asker
}

func NewInteractiveBackend(askerClient asker.Asker) *InteractiveBackend {
	return &InteractiveBackend{
		askerClient: askerClient,
	}
}

func (o *InteractiveBackend) Validate(_ map[string]string) error {
	return nil
}

func (o *InteractiveBackend) SelectServiceInstance(_ map[string]string, serviceMap map[string]unstructured.Unstructured) (string, error) {
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
