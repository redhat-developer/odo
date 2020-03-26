package storage

import (
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/lclient"
)

// New instantiantes a storage adapter
func New(adapterContext common.AdapterContext, client lclient.Client) common.StorageAdapter {
	return &Adapter{
		Client:         client,
		AdapterContext: adapterContext,
	}
}

// Adapter is a storage adapter implementation for Kubernetes
type Adapter struct {
	Client lclient.Client
	common.AdapterContext
}

// Create creates the component pvc storage if it does not exist
func (a *Adapter) Create(storages []common.Storage) (err error) {

	// createComponentStorage creates PVC from the unique Devfile volumes if it does not exist
	err = CreateComponentStorage(&a.Client, storages, a.ComponentName)
	if err != nil {
		return err
	}

	return
}
