package component

import (
	"fmt"
	"os"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"

	"github.com/pkg/errors"
	"k8s.io/klog"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/docker/storage"
	"github.com/openshift/odo/pkg/devfile/adapters/docker/utils"
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/kclient/generator"
	"github.com/openshift/odo/pkg/lclient"
	"github.com/openshift/odo/pkg/log"
)

const (
	LocalhostIP = "127.0.0.1"
)

func (a Adapter) createComponent() (err error) {
	componentName := a.ComponentName

	log.Infof("\nCreating Docker resources for component %s", a.ComponentName)

	containerComponents := generator.GetDevfileContainerComponents(a.Devfile.Data)
	if len(containerComponents) == 0 {
		return fmt.Errorf("no valid components found in the devfile")
	}

	// Get the storage adapter and create the volumes if it does not exist
	stoAdapter := storage.New(a.AdapterContext, a.Client)
	err = stoAdapter.Create(a.uniqueStorage)
	if err != nil {
		return errors.Wrapf(err, "unable to create Docker storage adapter for component %s", componentName)
	}

	// Loop over each container component and start a container for it
	for _, comp := range containerComponents {
		var dockerVolumeMounts []mount.Mount
		for _, vol := range a.containerNameToVolumes[comp.Name] {

			volMount := mount.Mount{
				Type:   mount.TypeVolume,
				Source: a.volumeNameToDockerVolName[vol.Name],
				Target: vol.ContainerPath,
			}
			dockerVolumeMounts = append(dockerVolumeMounts, volMount)
		}
		err = a.pullAndStartContainer(dockerVolumeMounts, comp)
		if err != nil {
			return errors.Wrapf(err, "unable to pull and start container %s for component %s", comp.Name, componentName)
		}
	}
	klog.V(2).Infof("Successfully created all containers for component %s", componentName)

	return nil
}

func (a Adapter) updateComponent() (componentExists bool, err error) {
	klog.V(2).Info("The component already exists, attempting to update it")
	componentExists = true
	componentName := a.ComponentName

	// Get the storage adapter and create the volumes if it does not exist
	stoAdapter := storage.New(a.AdapterContext, a.Client)
	err = stoAdapter.Create(a.uniqueStorage)

	containerComponents := generator.GetDevfileContainerComponents(a.Devfile.Data)
	if len(containerComponents) == 0 {
		return componentExists, fmt.Errorf("no valid components found in the devfile")
	}

	for _, comp := range containerComponents {
		// Check to see if this component is already running and if so, update it
		// If component isn't running, re-create it, as it either may be new, or crashed.
		containers, err := a.Client.GetContainersByComponentAndAlias(componentName, comp.Name)
		if err != nil {
			return false, errors.Wrapf(err, "unable to list containers for component %s", componentName)
		}

		var dockerVolumeMounts []mount.Mount
		for _, vol := range a.containerNameToVolumes[comp.Name] {
			volMount := mount.Mount{
				Type:   mount.TypeVolume,
				Source: a.volumeNameToDockerVolName[vol.Name],
				Target: vol.ContainerPath,
			}
			dockerVolumeMounts = append(dockerVolumeMounts, volMount)
		}
		if len(containers) == 0 {
			log.Infof("\nCreating Docker resources for component %s", a.ComponentName)

			// Container doesn't exist, so need to pull its image (to be safe) and start a new container
			err = a.pullAndStartContainer(dockerVolumeMounts, comp)
			if err != nil {
				return false, errors.Wrapf(err, "unable to pull and start container %s for component %s", comp.Name, componentName)
			}

			// Update componentExists so that we re-sync project and initialize supervisord if required
			componentExists = false
		} else if len(containers) == 1 {
			// Container already exists
			containerID := containers[0].ID

			// Get the associated container config, host config and mounts from the container
			containerConfig, hostConfig, mounts, err := a.Client.GetContainerConfigHostConfigAndMounts(containerID)
			if err != nil {
				return componentExists, errors.Wrapf(err, "unable to get the container config for component %s", componentName)
			}

			portMap, namePortMapping, err := getPortMap(a.Context, comp.Container.Endpoints, false)
			if err != nil {
				return componentExists, errors.Wrapf(err, "unable to get the port map from env.yaml file for component %s", componentName)
			}
			for port, urlName := range namePortMapping {
				containerConfig.Labels[port.Port()] = urlName
			}

			// See if the container needs to be updated
			if utils.DoesContainerNeedUpdating(comp, containerConfig, hostConfig, dockerVolumeMounts, mounts, portMap) {
				log.Infof("\nCreating Docker resources for component %s", a.ComponentName)

				s := log.SpinnerNoSpin("Updating the component " + comp.Name)
				defer s.End(false)

				// Remove the container
				err := a.Client.RemoveContainer(containerID)
				if err != nil {
					return componentExists, errors.Wrapf(err, "unable to remove container %s for component %s", containerID, comp.Name)
				}

				// Start the container
				err = a.startComponent(dockerVolumeMounts, comp)
				if err != nil {
					return false, errors.Wrapf(err, "unable to start container for devfile component %s", comp.Name)
				}

				klog.V(2).Infof("Successfully created container %s for component %s", comp.Container.Image, componentName)
				s.End(true)

				// Update componentExists so that we re-sync project and initialize supervisord if required
				componentExists = false
			}
		} else {
			// Multiple containers were returned with the specified label (which should be unique)
			// Error out, as this isn't expected
			return true, fmt.Errorf("found multiple running containers for devfile component %s and cannot push changes", comp.Name)
		}
	}

	return
}

func (a Adapter) pullAndStartContainer(mounts []mount.Mount, comp versionsCommon.DevfileComponent) error {
	// Container doesn't exist, so need to pull its image (to be safe) and start a new container
	s := log.Spinnerf("Pulling image %s", comp.Container.Image)

	err := a.Client.PullImage(comp.Container.Image)
	if err != nil {
		s.End(false)
		return errors.Wrapf(err, "Unable to pull %s image", comp.Container.Image)
	}
	s.End(true)

	// Start the component container
	err = a.startComponent(mounts, comp)
	if err != nil {
		return errors.Wrapf(err, "unable to start container for devfile component %s", comp.Name)
	}

	klog.V(2).Infof("Successfully created container %s for component %s", comp.Container.Image, a.ComponentName)
	return nil
}

func (a Adapter) startComponent(mounts []mount.Mount, comp versionsCommon.DevfileComponent) error {
	hostConfig, namePortMapping, err := a.generateAndGetHostConfig(comp.Container.Endpoints)
	hostConfig.Mounts = mounts
	if err != nil {
		return err
	}

	// Get the run command and update it's component entrypoint with supervisord if required and component's env command & workdir
	runCommand, err := common.GetRunCommand(a.Devfile.Data, a.devfileRunCmd)
	if err != nil {
		return err
	}
	updateComponentWithSupervisord(&comp, runCommand, a.supervisordVolumeName, &hostConfig)

	// If the component set `mountSources` to true, add the source volume and env PROJECTS_ROOT to it
	if comp.Container.MountSources {
		var syncFolder, projectsRoot string
		if comp.Container.SourceMapping != "" {
			syncFolder = comp.Container.SourceMapping
		} else if projectsRoot = common.GetComponentEnvVar(common.EnvProjectsRoot, comp.Container.Env); projectsRoot != "" {
			syncFolder = projectsRoot
		} else {
			syncFolder = lclient.OdoSourceVolumeMount
		}
		utils.AddVolumeToContainer(a.projectVolumeName, syncFolder, &hostConfig)

		// Set PROJECTS_ROOT as an env var in the container if not already set
		if projectsRoot == "" {
			envName := common.EnvProjectsRoot
			envValue := syncFolder
			comp.Container.Env = append(comp.Container.Env, versionsCommon.Env{
				Name:  envName,
				Value: envValue,
			})
		}
	}

	// Generate the container config after updating the component with the necessary data
	containerConfig := a.generateAndGetContainerConfig(a.ComponentName, comp)
	for port, urlName := range namePortMapping {
		containerConfig.Labels[port.Port()] = urlName
	}

	// Create the docker container
	s := log.Spinner("Starting container for " + comp.Container.Image)
	defer s.End(false)
	_, err = a.Client.StartContainer(&containerConfig, &hostConfig, nil)
	if err != nil {
		return err
	}
	s.End(true)

	return nil
}

func (a Adapter) generateAndGetContainerConfig(componentName string, comp versionsCommon.DevfileComponent) container.Config {
	// Convert the env vars in the Devfile to the format expected by Docker
	envVars := utils.ConvertEnvs(comp.Container.Env)
	ports := utils.ConvertPorts(comp.Container.Endpoints)
	containerLabels := utils.GetContainerLabels(componentName, comp.Name)
	containerConfig := a.Client.GenerateContainerConfig(comp.Container.Image, comp.Container.Command, comp.Container.Args, envVars, containerLabels, ports)

	return containerConfig
}

func (a Adapter) generateAndGetHostConfig(endpoints []versionsCommon.Endpoint) (container.HostConfig, map[nat.Port]string, error) {
	// Convert the port bindings from env.yaml and generate docker host config
	portMap, namePortMapping, err := getPortMap(a.Context, endpoints, true)
	if err != nil {
		return container.HostConfig{}, map[nat.Port]string{}, err
	}

	hostConfig := container.HostConfig{}
	if len(portMap) > 0 {
		hostConfig = a.Client.GenerateHostConfig(false, false, portMap)
	}

	return hostConfig, namePortMapping, nil
}

func getPortMap(context string, endpoints []versionsCommon.Endpoint, show bool) (nat.PortMap, map[nat.Port]string, error) {
	// Convert the exposed and internal port pairs saved in env.yaml file to PortMap
	// Todo: Use context to get the appropriate envinfo after context is supported in experimental mode
	portmap := nat.PortMap{}
	namePortMapping := make(map[nat.Port]string)

	var dir string
	var err error
	if context == "" {
		dir, err = os.Getwd()
		if err != nil {
			return nil, nil, err
		}
	} else {
		dir = context
	}

	envInfo, err := envinfo.NewEnvSpecificInfo(dir)
	if err != nil {
		return nil, nil, err
	}

	urlArr := envInfo.GetURL()

	for _, url := range urlArr {
		if url.ExposedPort > 0 && common.IsPortPresent(endpoints, url.Port) {
			port, err := nat.NewPort("tcp", strconv.Itoa(url.Port))
			if err != nil {
				return nil, nil, err
			}
			portmap[port] = []nat.PortBinding{
				{
					HostIP:   LocalhostIP,
					HostPort: strconv.Itoa(url.ExposedPort),
				},
			}
			namePortMapping[port] = url.Name
			if show {
				log.Successf("URL %v:%v created", LocalhostIP, url.ExposedPort)
			}
		} else if url.ExposedPort > 0 && len(endpoints) > 0 && !common.IsPortPresent(endpoints, url.Port) {
			return nil, nil, fmt.Errorf("error creating url: odo url config's port is not present in the devfile. Please re-create odo url with the new devfile port")
		}
	}

	return portmap, namePortMapping, nil
}

// createProjectVolumeIfReqd creates a project volume if absent and returns the
// name of the created project volume
func (a Adapter) createProjectVolumeIfReqd() (string, error) {
	var projectVolumeName string
	componentName := a.ComponentName

	// Get the project source volume
	projectVolumeLabels := utils.GetProjectVolumeLabels(componentName)
	projectVols, err := a.Client.GetVolumesByLabel(projectVolumeLabels)
	if err != nil {
		return "", errors.Wrapf(err, "unable to retrieve source volume for component "+componentName)
	}

	if len(projectVols) == 0 {
		// A source volume needs to be created
		projectVolumeName, err = storage.GenerateVolName(lclient.ProjectSourceVolumeName, componentName)
		if err != nil {
			return "", errors.Wrapf(err, "unable to generate project source volume name for component %s", componentName)
		}
		_, err := a.Client.CreateVolume(projectVolumeName, projectVolumeLabels)
		if err != nil {
			return "", errors.Wrapf(err, "unable to create project source volume for component %s", componentName)
		}
	} else if len(projectVols) == 1 {
		projectVolumeName = projectVols[0].Name
	} else if len(projectVols) > 1 {
		return "", errors.New(fmt.Sprintf("multiple source volumes found for component %s", componentName))
	}

	return projectVolumeName, nil
}

// createAndInitSupervisordVolumeIfReqd creates the supervisord volume and initializes
// it with supervisord bootstrap image - assembly files and supervisord binary
// returns the name of the supervisord volume and an error if present
func (a Adapter) createAndInitSupervisordVolumeIfReqd(componentExists bool) (string, error) {
	var supervisordVolumeName string
	componentName := a.ComponentName

	supervisordLabels := utils.GetSupervisordVolumeLabels(componentName)
	supervisordVolumes, err := a.Client.GetVolumesByLabel(supervisordLabels)
	if err != nil {
		return "", errors.Wrapf(err, "unable to retrieve supervisord volume for component")
	}

	if len(supervisordVolumes) == 0 {
		supervisordVolumeName, err = storage.GenerateVolName(common.SupervisordVolumeName, componentName)
		if err != nil {
			return "", errors.Wrapf(err, "unable to generate volume name for supervisord")
		}
		_, err := a.Client.CreateVolume(supervisordVolumeName, supervisordLabels)
		if err != nil {
			return "", errors.Wrapf(err, "unable to create supervisord volume for component")
		}
	} else {
		supervisordVolumeName = supervisordVolumes[0].Name
	}

	if !componentExists {
		log.Info("\nInitialization")
		s := log.Spinner("Initializing the component")
		defer s.End(false)

		err = a.startBootstrapSupervisordInitContainer(supervisordVolumeName)
		if err != nil {
			return "", errors.Wrapf(err, "unable to start supervisord container for component")
		}

		s.End(true)
	}

	return supervisordVolumeName, nil
}

// startBootstrapSupervisordInitContainer pulls the supervisord bootstrap image, mounts the supervisord
// volume, starts the bootstrap container and initializes the supervisord volume via its entrypoint
func (a Adapter) startBootstrapSupervisordInitContainer(supervisordVolumeName string) error {
	componentName := a.ComponentName
	supervisordLabels := utils.GetSupervisordVolumeLabels(componentName)
	image := common.GetBootstrapperImage()
	command := []string{"/usr/bin/cp"}
	args := []string{
		"-r",
		common.OdoInitImageContents,
		common.SupervisordMountPath,
	}

	var s *log.Status
	if log.IsDebug() {
		s = log.Spinnerf("Pulling image %s", image)
		defer s.End(false)
	}

	err := a.Client.PullImage(image)
	if err != nil {
		return errors.Wrapf(err, "unable to pull %s image", image)
	}
	if log.IsDebug() {
		s.End(true)
	}

	containerConfig := a.Client.GenerateContainerConfig(image, command, args, nil, supervisordLabels, nil)
	hostConfig := container.HostConfig{}

	utils.AddVolumeToContainer(supervisordVolumeName, common.SupervisordMountPath, &hostConfig)

	// Create the docker container
	if log.IsDebug() {
		s = log.Spinnerf("Starting container for %s", image)
		defer s.End(false)
	}
	containerID, err := a.Client.StartContainer(&containerConfig, &hostConfig, nil)
	if err != nil {
		return err
	}
	if log.IsDebug() {
		s.End(true)
	}

	// Wait for the container to exit before removing it
	err = a.Client.WaitForContainer(containerID, container.WaitConditionNotRunning)
	if err != nil {
		return errors.Wrapf(err, "supervisord init container %s failed to complete", containerID)
	}

	err = a.Client.RemoveContainer(containerID)
	if err != nil {
		return errors.Wrapf(err, "unable to remove supervisord init container %s", containerID)
	}

	return nil
}

// UpdateComponentWithSupervisord updates the devfile component's
// 1. command and args with supervisord, if absent
// 2. env with ODO_COMMAND_RUN and ODO_COMMAND_RUN_WORKING_DIR, if absent
func updateComponentWithSupervisord(comp *versionsCommon.DevfileComponent, runCommand versionsCommon.DevfileCommand, supervisordVolumeName string, hostConfig *container.HostConfig) {

	// Mount the supervisord volume for the run command container
	if runCommand.Exec.Component == comp.Name {
		utils.AddVolumeToContainer(supervisordVolumeName, common.SupervisordMountPath, hostConfig)

		if len(comp.Container.Command) == 0 && len(comp.Container.Args) == 0 {
			klog.V(2).Infof("Updating container %v entrypoint with supervisord", comp.Name)
			comp.Container.Command = append(comp.Container.Command, common.SupervisordBinaryPath)
			comp.Container.Args = append(comp.Container.Args, "-c", common.SupervisordConfFile)
		}

		if !common.IsEnvPresent(comp.Container.Env, common.EnvOdoCommandRun) {
			envName := common.EnvOdoCommandRun
			envValue := runCommand.Exec.CommandLine
			comp.Container.Env = append(comp.Container.Env, versionsCommon.Env{
				Name:  envName,
				Value: envValue,
			})
		}

		if !common.IsEnvPresent(comp.Container.Env, common.EnvOdoCommandRunWorkingDir) && runCommand.Exec.WorkingDir != "" {
			envName := common.EnvOdoCommandRunWorkingDir
			envValue := runCommand.Exec.WorkingDir
			comp.Container.Env = append(comp.Container.Env, versionsCommon.Env{
				Name:  envName,
				Value: envValue,
			})
		}
	}
}
