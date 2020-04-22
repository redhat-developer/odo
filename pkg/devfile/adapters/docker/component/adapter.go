package component

import (
	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/docker/storage"
	"github.com/openshift/odo/pkg/devfile/adapters/docker/utils"
	"github.com/openshift/odo/pkg/lclient"
)

// New instantiantes a component adapter
func New(adapterContext common.AdapterContext, client lclient.Client) Adapter {
	return Adapter{
		Client:         client,
		AdapterContext: adapterContext,
	}
}

// Adapter is a component adapter implementation for Kubernetes
type Adapter struct {
	Client lclient.Client
	common.AdapterContext

	componentAliasToVolumes   map[string][]adaptersCommon.DevfileVolume
	uniqueStorage             []adaptersCommon.Storage
	volumeNameToDockerVolName map[string]string
}

// Push updates the component if a matching component exists or creates one if it doesn't exist
func (a Adapter) Push(parameters common.PushParameters) (err error) {
	componentExists := utils.ComponentExists(a.Client, a.ComponentName)

	// Process the volumes defined in the devfile
	a.componentAliasToVolumes = adaptersCommon.GetVolumes(a.Devfile)
	a.uniqueStorage, a.volumeNameToDockerVolName, err = storage.ProcessVolumes(&a.Client, a.ComponentName, a.componentAliasToVolumes)
	if err != nil {
		return errors.Wrapf(err, "Unable to process volumes for component %s", a.ComponentName)
	}

	if componentExists {
		err = a.updateComponent()
	} else {
		err = a.createComponent()
	}

	if err != nil {
		return errors.Wrap(err, "unable to create or update component")
	}

	return nil
}

// DoesComponentExist returns true if a component with the specified name exists, false otherwise
func (a Adapter) DoesComponentExist(cmpName string) bool {
	return utils.ComponentExists(a.Client, cmpName)
}

// Delete attempts to delete the component with the specified labels, returning an error if it fails
// Stub function until the functionality is added as part of https://github.com/openshift/odo/issues/2581
func (d Adapter) Delete(labels map[string]string) error {
	return nil
}
