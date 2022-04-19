package backend

import (
	"fmt"

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

func (o *InteractiveBackend) SelectServiceInstance(_ map[string]string, options []string) (string, error) {
	return o.askerClient.AskServiceInstance(options)
}

func (o *InteractiveBackend) AskBindingName(componentName string, _ map[string]string) (string, error) {
	return o.askerClient.AskServiceBindingName(fmt.Sprintf("my-%s-app", componentName))
}

func (o *InteractiveBackend) AskBindAsFiles(_ map[string]string) (bool, error) {
	return o.askerClient.AskBindAsFiles()
}
