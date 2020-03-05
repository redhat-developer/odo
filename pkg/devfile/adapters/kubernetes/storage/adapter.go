package storage

import (
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/kclient"
)

// New instantiantes a storage adapter
func New(adapterContext common.AdapterContext, client kclient.Client) common.StorageAdapter {
	return &Adapter{
		Client:         client,
		AdapterContext: adapterContext,
	}
}

// Adapter is a storage adapter implementation for Kubernetes
type Adapter struct {
	Client kclient.Client
	common.AdapterContext
	VolumeNameToPVCName map[string]string
}

// GetVolumeNameToPVCName returns the map of volume name to its corresponding pvc name
func (a *Adapter) GetVolumeNameToPVCName() map[string]string {
	return a.VolumeNameToPVCName
}

// Create creates the component pvc storage if it does not exist and adds them to the storage adapter struct
func (a *Adapter) Create(volumes []common.Volume) (err error) {

	// createComponentStorage creates PVC from the unique Devfile volumes if it does not exist and returns a map of volume name to the PVC created
	a.VolumeNameToPVCName, err = CreateComponentStorage(&a.Client, volumes, a.ComponentName)
	if err != nil {
		return err
	}

	return
}
