package binding

import (
	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/binding/asker"
	"github.com/redhat-developer/odo/pkg/binding/backend"
	"github.com/redhat-developer/odo/pkg/kclient"
)

type BindingClient struct {
	// Backends
	flagsBackend       *backend.FlagsBackend
	interactiveBackend *backend.InteractiveBackend

	// Clients
	kubernetesClient kclient.ClientInterface
}

func NewBindingClient(kubernetesClient kclient.ClientInterface) *BindingClient {
	// We create the asker client and the backends here and not at the CLI level, as we want to hide these details to the CLI
	askerClient := asker.NewSurveyAsker()
	return &BindingClient{
		flagsBackend:       backend.NewFlagsBackend(),
		interactiveBackend: backend.NewInteractiveBackend(askerClient),
		kubernetesClient:   kubernetesClient,
	}
}

// GetFlags gets the flag specific to init operation so that it can correctly decide on the backend to be used
// It ignores all the flags except the ones specific to init operation, for e.g. verbosity flag
func (o *BindingClient) GetFlags(flags map[string]string) map[string]string {
	initFlags := map[string]string{}
	for flag, value := range flags {
		if flag == backend.FLAG_NAME || flag == backend.FLAG_SERVICE || flag == backend.FLAG_BIND_AS_FILES {
			initFlags[flag] = value
		}
	}
	return initFlags
}

// Validate calls Validate method of the adequate backend
func (o *BindingClient) Validate(flags map[string]string) error {
	var backend backend.CreateBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.Validate(flags)
}

func (o *BindingClient) SelectServiceInstance(flags map[string]string) (string, error) {
	var backend backend.CreateBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}

	return backend.SelectServiceInstance(flags)
}

func (o *BindingClient) AskBindingName(componentName string, flags map[string]string) (string, error) {
	var backend backend.CreateBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.AskBindingName(componentName, flags)
}

func (o *BindingClient) AskBindAsFiles(flags map[string]string) (bool, error) {
	var backend backend.CreateBindingBackend
	if len(flags) == 0 {
		backend = o.interactiveBackend
	} else {
		backend = o.flagsBackend
	}
	return backend.AskBindAsFiles(flags)
}

func (o *BindingClient) CreateBinding(service string, bindingName string, bindAsFiles bool, obj parser.DevfileObj) error {
	return nil
}

func (o *BindingClient) GetServiceInstances() ([]string, error) {
	o.kubernetesClient.GetBindableKinds()
}
