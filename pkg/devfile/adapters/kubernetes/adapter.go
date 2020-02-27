package kubernetes

import (
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/kubernetes/component"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/pkg/errors"
)

// Adapter maps Devfiles to Kubernetes resources and actions
type Adapter struct {
	componentAdapter common.ComponentAdapter
}

// New instantiates a kubernetes adapter
func New(adapterContext common.AdapterContext, client kclient.Client) Adapter {

	compAdapter := component.New(adapterContext, client)

	return Adapter{
		componentAdapter: compAdapter,
	}
}

// Start creates Kubernetes resources that correspond to the devfile if they don't already exist
func (k Adapter) Start() error {

	err := k.componentAdapter.Start()
	if err != nil {
		return errors.Wrap(err, "Failed to start the component")
	}

	return nil
}
