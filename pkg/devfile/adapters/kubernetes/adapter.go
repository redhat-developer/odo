package kubernetes

import (
	"fmt"
	"io"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/portForward"
	"github.com/redhat-developer/odo/pkg/preference"

	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/kubernetes/component"
)

// Adapter maps Devfiles to Kubernetes resources and actions
type Adapter struct {
	componentAdapter common.ComponentAdapter
}

var _ common.ComponentAdapter = (*Adapter)(nil)

// New instantiates a kubernetes adapter
func New(
	adapterContext common.AdapterContext,
	client kclient.ClientInterface,
	prefClient preference.Client,
	portForwardClient portForward.Client,
	randomPorts bool,
	errOut io.Writer,
) Adapter {

	compAdapter := component.New(adapterContext, client, prefClient, portForwardClient, randomPorts, errOut)

	return Adapter{
		componentAdapter: &compAdapter,
	}
}

// Push creates Kubernetes resources that correspond to the devfile if they don't already exist
func (k Adapter) Push(parameters common.PushParameters) error {

	err := k.componentAdapter.Push(parameters)
	if err != nil {
		return fmt.Errorf("failed to create the component: %w", err)
	}

	return nil
}
