package component

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/golang/glog"
	"github.com/pkg/errors"

	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/docker/storage"
	"github.com/openshift/odo/pkg/devfile/adapters/docker/utils"
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/log"
)

func (a Adapter) createComponent() (err error) {
	componentName := a.ComponentName

	// Get or create the project source volume
	var projectVolumeName string
	projectVolumeLabels := utils.GetProjectVolumeLabels(componentName)
	vols, err := a.Client.GetVolumesByLabel(projectVolumeLabels)
	if err != nil {
		return errors.Wrapf(err, "Unable to retrieve source volume for component "+componentName)
	}
	if len(vols) == 0 {
		// A source volume needs to be created
		volume, err := a.Client.CreateVolume("", projectVolumeLabels)
		if err != nil {
			return errors.Wrapf(err, "Unable to create project source volume for component %s", componentName)
		}
		projectVolumeName = volume.Name
	} else if len(vols) == 1 {
		projectVolumeName = vols[0].Name
	} else if len(vols) > 1 {
		return errors.Wrapf(err, "Error, multiple source volumes found for component %s", componentName)
	}

	supportedComponents := adaptersCommon.GetSupportedComponents(a.Devfile.Data)
	if len(supportedComponents) == 0 {
		return fmt.Errorf("No valid components found in the devfile")
	}

	// Process the volumes defined in the devfile
	componentAliasToVolumes := adaptersCommon.GetVolumes(a.Devfile)
	uniqueStorage, volumeMapping, err := storage.ProcessVolumes(&a.Client, componentName, componentAliasToVolumes)
	if err != nil {
		return errors.Wrapf(err, "Unable to process volumes for component %s", componentName)
	}

	// Get the storage adapter and create the volumes if it does not exist
	stoAdapter := storage.New(a.AdapterContext, a.Client)
	err = stoAdapter.Create(uniqueStorage)
	if err != nil {
		return errors.Wrapf(err, "Unable to create Docker storage adapter for component %s", componentName)
	}

	// Loop over each component and start a container for it
	for _, comp := range supportedComponents {
		var dockerVolumeMounts []mount.Mount
		for _, vol := range componentAliasToVolumes[*comp.Alias] {
			volMount := mount.Mount{
				Type:   mount.TypeVolume,
				Source: volumeMapping[*vol.Name],
				Target: *vol.ContainerPath,
			}
			dockerVolumeMounts = append(dockerVolumeMounts, volMount)
		}
		err = a.pullAndStartContainer(dockerVolumeMounts, projectVolumeName, comp)
		if err != nil {
			return errors.Wrapf(err, "unable to pull and start container %s for component %s", *comp.Alias, componentName)
		}
	}
	glog.V(3).Infof("Successfully created all containers for component %s", componentName)

	return nil
}

func (a Adapter) updateComponent() (err error) {
	glog.V(3).Info("The component already exists, attempting to update it")
	componentName := a.ComponentName

	// Get the project source volume
	volumeLabels := utils.GetProjectVolumeLabels(componentName)
	vols, err := a.Client.GetVolumesByLabel(volumeLabels)
	if err != nil {
		return errors.Wrapf(err, "Unable to retrieve source volume for component "+componentName)
	}
	if len(vols) == 0 {
		return fmt.Errorf("Unable to find source volume for component %s", componentName)
	}
	projectVolumeName := vols[0].Name

	// Process the volumes defined in the devfile
	componentAliasToVolumes := adaptersCommon.GetVolumes(a.Devfile)
	uniqueStorage, volumeMapping, err := storage.ProcessVolumes(&a.Client, componentName, componentAliasToVolumes)
	if err != nil {
		return errors.Wrapf(err, "Unable to process volumes for component %s", componentName)
	}

	// Get the storage adapter and create the volumes if it does not exist
	stoAdapter := storage.New(a.AdapterContext, a.Client)
	err = stoAdapter.Create(uniqueStorage)

	supportedComponents := adaptersCommon.GetSupportedComponents(a.Devfile.Data)
	if len(supportedComponents) == 0 {
		return fmt.Errorf("No valid components found in the devfile")
	}

	for _, comp := range supportedComponents {
		// Check to see if this component is already running and if so, update it
		// If component isn't running, re-create it, as it either may be new, or crashed.
		containers, err := a.Client.GetContainersByComponentAndAlias(componentName, *comp.Alias)
		if err != nil {
			return errors.Wrapf(err, "unable to list containers for component %s", componentName)
		}

		var dockerVolumeMounts []mount.Mount
		for _, vol := range componentAliasToVolumes[*comp.Alias] {
			volMount := mount.Mount{
				Type:   mount.TypeVolume,
				Source: volumeMapping[*vol.Name],
				Target: *vol.ContainerPath,
			}
			dockerVolumeMounts = append(dockerVolumeMounts, volMount)
		}

		if len(containers) == 0 {
			// Container doesn't exist, so need to pull its image (to be safe) and start a new container
			err = a.pullAndStartContainer(dockerVolumeMounts, projectVolumeName, comp)
			if err != nil {
				return errors.Wrapf(err, "unable to pull and start container %s for component %s", *comp.Alias, componentName)
			}
		} else if len(containers) == 1 {
			// Container already exists
			containerID := containers[0].ID

			// Get the associated container config from the container ID
			containerConfig, mounts, err := a.Client.GetContainerConfigAndMounts(containerID)
			if err != nil {
				return errors.Wrapf(err, "unable to get the container config for component %s", componentName)
			}

			// See if the container needs to be updated
			if utils.DoesContainerNeedUpdating(comp, containerConfig, dockerVolumeMounts, mounts) {
				s := log.Spinner("Updating the component " + *comp.Alias)
				defer s.End(false)
				// Remove the container
				err := a.Client.RemoveContainer(containerID)
				if err != nil {
					return errors.Wrapf(err, "Unable to remove container %s for component %s", containerID, *comp.Alias)
				}

				// Start the container
				err = a.startContainer(dockerVolumeMounts, projectVolumeName, comp)
				if err != nil {
					return errors.Wrapf(err, "Unable to start container for devfile component %s", *comp.Alias)
				}
				glog.V(3).Infof("Successfully created container %s for component %s", *comp.Image, componentName)
				s.End(true)
			}
		} else {
			// Multiple containers were returned with the specified label (which should be unique)
			// Error out, as this isn't expected
			return fmt.Errorf("Found multiple running containers for devfile component %s and cannot push changes", *comp.Alias)
		}
	}
	return nil
}

func (a Adapter) pullAndStartContainer(mounts []mount.Mount, projectVolumeName string, comp versionsCommon.DevfileComponent) error {
	// Container doesn't exist, so need to pull its image (to be safe) and start a new container
	s := log.Spinner("Pulling image " + *comp.Image)

	err := a.Client.PullImage(*comp.Image)
	if err != nil {
		s.End(false)
		return errors.Wrapf(err, "Unable to pull %s image", *comp.Image)
	}
	s.End(true)

	// Start the container
	err = a.startContainer(mounts, projectVolumeName, comp)
	if err != nil {
		return errors.Wrapf(err, "Unable to start container for devfile component %s", *comp.Alias)
	}
	glog.V(3).Infof("Successfully created container %s for component %s", *comp.Image, a.ComponentName)
	return nil
}

func (a Adapter) startContainer(mounts []mount.Mount, projectVolumeName string, comp versionsCommon.DevfileComponent) error {
	containerConfig := a.generateAndGetContainerConfig(a.ComponentName, comp)
	hostConfig := container.HostConfig{
		Mounts: mounts,
	}

	// If the component set `mountSources` to true, add the source volume to it
	if comp.MountSources {
		utils.AddProjectVolumeToComp(projectVolumeName, &hostConfig)
	}

	// Create the docker container
	s := log.Spinner("Starting container for " + *comp.Image)
	defer s.End(false)
	err := a.Client.StartContainer(&containerConfig, &hostConfig, nil)
	if err != nil {
		return err
	}
	s.End(true)
	return nil
}

func (a Adapter) generateAndGetContainerConfig(componentName string, comp versionsCommon.DevfileComponent) container.Config {
	// Convert the env vars in the Devfile to the format expected by Docker
	envVars := utils.ConvertEnvs(comp.Env)

	containerLabels := map[string]string{
		"component": componentName,
		"alias":     *comp.Alias,
	}

	containerConfig := a.Client.GenerateContainerConfig(*comp.Image, comp.Command, comp.Args, envVars, containerLabels)
	return containerConfig
}
