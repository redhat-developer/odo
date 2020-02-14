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
	Client           kclient.Client
	common.AdapterMetadata
}

// New instantiates a kubernetes adapter
func New(adapterMetadata common.AdapterMetadata, client kclient.Client) Adapter {

	compAdapter := component.New(adapterMetadata, client)

	return Adapter{
		componentAdapter: compAdapter,
		AdapterMetadata:  adapterMetadata,
		Client:           client,
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
