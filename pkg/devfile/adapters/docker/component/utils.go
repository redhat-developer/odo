package component

import (
	"fmt"
	"os"
	"strconv"

	"github.com/docker/go-connections/nat"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/golang/glog"
	"github.com/pkg/errors"

	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/docker/storage"
	"github.com/openshift/odo/pkg/devfile/adapters/docker/utils"
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/log"
)

var LocalhostIP = "127.0.0.1"

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
		projectVolumeName, err = storage.GenerateVolNameFromDevfileVol("odo-project-source", a.ComponentName)
		if err != nil {
			return errors.Wrapf(err, "Unable to generate project source volume name for component %s", componentName)
		}
		_, err := a.Client.CreateVolume(projectVolumeName, projectVolumeLabels)
		if err != nil {
			return errors.Wrapf(err, "Unable to create project source volume for component %s", componentName)
		}
	} else if len(vols) == 1 {
		projectVolumeName = vols[0].Name
	} else if len(vols) > 1 {
		return errors.Wrapf(err, "Error, multiple source volumes found for component %s", componentName)
	}

	supportedComponents := adaptersCommon.GetSupportedComponents(a.Devfile.Data)
	if len(supportedComponents) == 0 {
		return fmt.Errorf("No valid components found in the devfile")
	}

	// Get the storage adapter and create the volumes if it does not exist
	stoAdapter := storage.New(a.AdapterContext, a.Client)
	err = stoAdapter.Create(a.uniqueStorage)
	if err != nil {
		return errors.Wrapf(err, "Unable to create Docker storage adapter for component %s", componentName)
	}

	// Loop over each component and start a container for it
	for _, comp := range supportedComponents {
		var dockerVolumeMounts []mount.Mount
		for _, vol := range a.componentAliasToVolumes[*comp.Alias] {
			volMount := mount.Mount{
				Type:   mount.TypeVolume,
				Source: a.volumeNameToDockerVolName[*vol.Name],
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
	} else if len(vols) > 1 {
		return errors.Wrapf(err, "Error, multiple source volumes found for component %s", componentName)
	}
	projectVolumeName := vols[0].Name

	// Get the storage adapter and create the volumes if it does not exist
	stoAdapter := storage.New(a.AdapterContext, a.Client)
	err = stoAdapter.Create(a.uniqueStorage)

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
		for _, vol := range a.componentAliasToVolumes[*comp.Alias] {
			volMount := mount.Mount{
				Type:   mount.TypeVolume,
				Source: a.volumeNameToDockerVolName[*vol.Name],
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

	hostConfig, err := a.generateAndGetHostConfig()
	hostConfig.Mounts = mounts
	if err != nil {
		return err
	}
	// If the component set `mountSources` to true, add the source volume to it
	if comp.MountSources {
		utils.AddProjectVolumeToComp(projectVolumeName, &hostConfig)
	}

	// Create the docker container
	s := log.Spinner("Starting container for " + *comp.Image)
	defer s.End(false)
	err = a.Client.StartContainer(&containerConfig, &hostConfig, nil)
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

	// For each endpoint defined in the component in the devfile, add it to the portset for the container
	portSet := nat.PortSet{}
	for _, endpoint := range comp.Endpoints {
		port := fmt.Sprint(*endpoint.Port) + "/tcp"
		portSet[nat.Port(port)] = struct{}{}
	}

	containerConfig := a.Client.GenerateContainerConfig(*comp.Image, comp.Command, comp.Args, envVars, containerLabels, portSet)
	return containerConfig
}

func (a Adapter) generateAndGetHostConfig() (container.HostConfig, error) {
	// Convert the port bindings from env.yaml and generate docker host config
	portMap, err := getPortMap()
	if err != nil {
		return container.HostConfig{}, err
	}
	hostConfig := container.HostConfig{}
	if len(portMap) > 0 {
		hostConfig = a.Client.GenerateHostConfig(false, false, portMap)
	}
	return hostConfig, nil
}

func getPortMap() (nat.PortMap, error) {
	// Convert the exposed and internal port pairs saved in env.yaml file to PortMap
	// Todo: Use context to get the approraite envinfo after context is supported in experimental mode
	dir, err := os.Getwd()
	if err != nil {
		return nat.PortMap{}, err
	}
	envInfo, err := envinfo.NewEnvSpecificInfo(dir)
	if err != nil {
		return nat.PortMap{}, err
	}
	urlArr := envInfo.GetURL()
	if len(urlArr) > 0 {
		portmap := nat.PortMap{}
		for _, element := range urlArr {
			if element.ExposedPort > 0 {
				port, err := nat.NewPort("tcp", strconv.Itoa(element.Port))
				if err != nil {
					return nat.PortMap{}, err
				}
				portmap[port] = []nat.PortBinding{
					nat.PortBinding{
						HostIP:   LocalhostIP,
						HostPort: strconv.Itoa(element.ExposedPort),
					},
				}
				log.Info("\nApplying URL configuration")
				log.Successf("URL %v:%v created", LocalhostIP, element.ExposedPort)
			}
		}
		return portmap, nil
	}

	return nat.PortMap{}, nil
}
