package component

import (
	"fmt"

	"github.com/docker/docker/api/types/mount"
	"github.com/golang/glog"
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
func (a Adapter) Delete(labels map[string]string) error {

	componentName, exists := labels["component"]
	if !exists {
		return errors.New("unable to delete component without a component label")
	}

	list, err := a.Client.GetContainerList()
	if err != nil {
		return errors.Wrap(err, "unable to retrieve container list for delete operation")
	}

	componentContainer := a.Client.GetContainersByComponent(componentName, list)

	if len(componentContainer) == 0 {
		return errors.Errorf("the component %s doesn't exist", a.ComponentName)
	}

	// Get all volumes that match our component label
	volumeLabels := utils.GetProjectVolumeLabels(componentName)
	vols, err := a.Client.GetVolumesByLabel(volumeLabels)
	if err != nil {
		return errors.Wrapf(err, "unable to retrieve source volume for component "+componentName)
	}
	if len(vols) == 0 {
		return fmt.Errorf("unable to find source volume for component %s", componentName)
	}

	volumesToDelete := map[string] /* volume name -> volume name (value unused)*/ string{}

	for _, container := range componentContainer {

		glog.V(3).Infof("Deleting container %s for component %s", container.ID, componentName)
		err = a.Client.RemoveContainer(container.ID)
		if err != nil {
			return errors.Wrapf(err, "unable to remove container ID %s of component %s", container.ID, componentName)
		}

		// Generate a list of mounted volumes of the odo-managed container
		volumeNames := map[string]string{}
		for _, m := range container.Mounts {

			if m.Type != mount.TypeVolume {
				continue
			}
			volumeNames[m.Name] = m.Name
		}

		for _, vol := range vols {
			// If the volume was found to be attached to the component's container, then add the volume
			// to the deletion list.
			if _, ok := volumeNames[vol.Name]; ok {
				volumesToDelete[vol.Name] = vol.Name
			}
		}
	}

	// Finally, delete the volumes we discovered during container deletion.
	for name := range volumesToDelete {
		glog.V(3).Infof("Deleting the volume %s for component %s", name, componentName)
		err := a.Client.RemoveVolume(name)
		if err != nil {
			return errors.Wrapf(err, "Unable to remove volume %s of component %s", name, componentName)
		}
	}

	return nil

}
