package component

import (
	"fmt"
	"os"
	"strconv"

	"reflect"

	"github.com/docker/go-connections/nat"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/golang/glog"
	"github.com/pkg/errors"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/adapters/docker/storage"
	"github.com/openshift/odo/pkg/devfile/adapters/docker/utils"
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/exec"
	"github.com/openshift/odo/pkg/lclient"
	"github.com/openshift/odo/pkg/log"
)

var LocalhostIP = "127.0.0.1"

const (
	envCheProjectsRoot         = "CHE_PROJECTS_ROOT"
	envOdoCommandRunWorkingDir = "ODO_COMMAND_RUN_WORKING_DIR"
	envOdoCommandRun           = "ODO_COMMAND_RUN"
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

	// Get the supervisord volume
	var supervisordVolumeName string
	supervisordLabels := utils.GetSupervisordVolumeLabels()
	supervisordVols, err := a.Client.GetVolumesByLabel(supervisordLabels)
	if err != nil {
		return errors.Wrapf(err, "Unable to retrieve supervisord volume for component "+componentName)
	}
	if len(supervisordVols) == 0 {
		supervisordVolumeName, err = utils.CreateAndInitSupervisordVolume(a.Client)
		if err != nil {
			return errors.Wrapf(err, "Unable to create supervisord volume for component %s", componentName)
		}
	} else {
		supervisordVolumeName = supervisordVols[0].Name
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
		err = a.pullAndStartContainer(dockerVolumeMounts, projectVolumeName, supervisordVolumeName, comp)
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
	projectVols, err := a.Client.GetVolumesByLabel(volumeLabels)
	if err != nil {
		return errors.Wrapf(err, "Unable to retrieve source volume for component "+componentName)
	}
	if len(projectVols) == 0 {
		return fmt.Errorf("Unable to find source volume for component %s", componentName)
	} else if len(projectVols) > 1 {
		return errors.Wrapf(err, "Error, multiple source volumes found for component %s", componentName)
	}
	projectVolumeName := projectVols[0].Name

	// Get the supervisord volume
	var supervisordVolumeName string
	supervisordLabels := utils.GetSupervisordVolumeLabels()
	supervisordVols, err := a.Client.GetVolumesByLabel(supervisordLabels)
	if err != nil {
		return errors.Wrapf(err, "Unable to retrieve supervisord volume for component "+componentName)
	}
	if len(supervisordVols) == 0 {
		supervisordVolumeName, err = utils.CreateAndInitSupervisordVolume(a.Client)
		if err != nil {
			return errors.Wrapf(err, "Unable to create supervisord volume for component %s", componentName)
		}
	} else {
		supervisordVolumeName = supervisordVols[0].Name
	}

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
			err = a.pullAndStartContainer(dockerVolumeMounts, projectVolumeName, supervisordVolumeName, comp)
			if err != nil {
				return errors.Wrapf(err, "unable to pull and start container %s for component %s", *comp.Alias, componentName)
			}
		} else if len(containers) == 1 {
			// Container already exists
			containerID := containers[0].ID

			// Get the associated container config from the container ID
			containerConfig, hostConfig, mounts, err := a.Client.GetContainerConfigHostConfigAndMounts(containerID)
			if err != nil {
				return errors.Wrapf(err, "unable to get the container config for component %s", componentName)
			}
			portMap, err := getPortMap()
			if err != nil {
				return errors.Wrapf(err, "unable to get the port map from env.yaml file for component %s", componentName)
			}
			// See if the container needs to be updated
			if utils.DoesContainerNeedUpdating(comp, containerConfig, hostConfig, dockerVolumeMounts, mounts, portMap) {
				s := log.Spinner("Updating the component " + *comp.Alias)
				defer s.End(false)
				// Remove the container
				err := a.Client.RemoveContainer(containerID)
				if err != nil {
					return errors.Wrapf(err, "Unable to remove container %s for component %s", containerID, *comp.Alias)
				}

				// Start the container
				err = a.startContainer(dockerVolumeMounts, projectVolumeName, supervisordVolumeName, comp)
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

func (a Adapter) pullAndStartContainer(mounts []mount.Mount, projectVolumeName, supervisordVolumeName string, comp versionsCommon.DevfileComponent) error {
	// Container doesn't exist, so need to pull its image (to be safe) and start a new container
	s := log.Spinner("Pulling image " + *comp.Image)

	err := a.Client.PullImage(*comp.Image)
	if err != nil {
		s.End(false)
		return errors.Wrapf(err, "Unable to pull %s image", *comp.Image)
	}
	s.End(true)

	// Start the container
	err = a.startContainer(mounts, projectVolumeName, supervisordVolumeName, comp)
	if err != nil {
		return errors.Wrapf(err, "Unable to start container for devfile component %s", *comp.Alias)
	}
	glog.V(3).Infof("Successfully created container %s for component %s", *comp.Image, a.ComponentName)
	return nil
}

func (a Adapter) startContainer(mounts []mount.Mount, projectVolumeName, supervisordVolumeName string, comp versionsCommon.DevfileComponent) error {
	hostConfig, err := a.generateAndGetHostConfig()
	hostConfig.Mounts = mounts
	if err != nil {
		return err
	}

	runCommand, err := adaptersCommon.GetRunCommand(a.Devfile.Data, a.devfileRunCmd)
	if err != nil {
		return err
	}

	// Mount the supervisord volume for the run command container
	for _, action := range runCommand.Actions {
		if *action.Component == *comp.Alias {
			utils.AddVolumeToComp(supervisordVolumeName, adaptersCommon.SupervisordMountPath, &hostConfig)
		}

		if len(comp.Command) == 0 && len(comp.Args) == 0 {
			glog.V(3).Infof("Updating container %v entrypoint with supervisord", comp.Alias)
			comp.Command = append(comp.Command, adaptersCommon.SupervisordBinaryPath)
			comp.Args = append(comp.Args, "-c", adaptersCommon.SupervisordConfFile)
		}

		if !adaptersCommon.IsEnvPresent(comp.Env, envOdoCommandRun) {
			envName := envOdoCommandRun
			envValue := *action.Command
			comp.Env = append(comp.Env, versionsCommon.DockerimageEnv{
				Name:  &envName,
				Value: &envValue,
			})
		}

		if !adaptersCommon.IsEnvPresent(comp.Env, envOdoCommandRunWorkingDir) && action.Workdir != nil {
			envName := envOdoCommandRunWorkingDir
			envValue := *action.Workdir
			comp.Env = append(comp.Env, versionsCommon.DockerimageEnv{
				Name:  &envName,
				Value: &envValue,
			})
		}
	}

	// If the component set `mountSources` to true, add the source volume to it
	if comp.MountSources {
		utils.AddVolumeToComp(projectVolumeName, lclient.OdoSourceVolumeMount, &hostConfig)

		if !adaptersCommon.IsEnvPresent(comp.Env, envCheProjectsRoot) {
			envName := envCheProjectsRoot
			envValue := lclient.OdoSourceVolumeMount
			comp.Env = append(comp.Env, versionsCommon.DockerimageEnv{
				Name:  &envName,
				Value: &envValue,
			})
		}
	}

	containerConfig := a.generateAndGetContainerConfig(a.ComponentName, comp)

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
	ports := utils.ConvertPorts(comp.Endpoints)
	containerLabels := map[string]string{
		"component": componentName,
		"alias":     *comp.Alias,
	}
	containerConfig := a.Client.GenerateContainerConfig(*comp.Image, comp.Command, comp.Args, envVars, containerLabels, ports)
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

// Push syncs source code from the user's disk to the component
func (a Adapter) execDevfile(pushDevfileCommands []versionsCommon.DevfileCommand, componentExists, show bool, podName string, containers []types.Container) (err error) {
	var buildRequired bool
	var s *log.Status

	if len(pushDevfileCommands) == 1 {
		// if there is one command, it is the mandatory run command. No need to build.
		buildRequired = false
	} else if len(pushDevfileCommands) == 2 {
		// if there are two commands, it is the optional build command and the mandatory run command, set buildRequired to true
		buildRequired = true
	} else {
		return fmt.Errorf("error executing devfile commands - there should be at least 1 command or at most 2 commands, currently there are %v commands", len(pushDevfileCommands))
	}

	for i := 0; i < len(pushDevfileCommands); i++ {
		command := pushDevfileCommands[i]

		// Exec the devBuild command if buildRequired is true
		if (command.Name == string(common.DefaultDevfileBuildCommand) || command.Name == a.devfileBuildCmd) && buildRequired {
			glog.V(3).Infof("Executing devfile command %v", command.Name)

			for _, action := range command.Actions {
				// Change to the workdir and execute the command
				var cmdArr []string
				if action.Workdir != nil {
					cmdArr = []string{"/bin/sh", "-c", "cd " + *action.Workdir + " && " + *action.Command}
				} else {
					cmdArr = []string{"/bin/sh", "-c", *action.Command}
				}

				// Get the containerID
				containerID := ""
				for _, container := range containers {
					if container.Labels["alias"] == *action.Component {
						containerID = container.ID
					}
				}

				if show {
					s = log.SpinnerNoSpin("Executing " + command.Name + " command " + fmt.Sprintf("%q", *action.Command))
				} else {
					s = log.Spinner("Executing " + command.Name + " command " + fmt.Sprintf("%q", *action.Command))
				}

				defer s.End(false)

				err = exec.ExecuteCommand(&a.Client, podName, containerID, cmdArr, show)
				if err != nil {
					s.End(false)
					return err
				}
				s.End(true)
			}

			// Reset the for loop counter and iterate through all the devfile commands again for others
			i = -1
			// Set the buildRequired to false since we already executed the build command
			buildRequired = false
		} else if (command.Name == string(common.DefaultDevfileRunCommand) || command.Name == a.devfileRunCmd) && !buildRequired {
			// Always check for buildRequired is false, since the command may be iterated out of order and we always want to execute devBuild first if buildRequired is true. If buildRequired is false, then we don't need to build and we can execute the devRun command
			glog.V(3).Infof("Executing devfile command %v", command.Name)

			for _, action := range command.Actions {

				// Get the containerID
				containerID := ""
				for _, container := range containers {
					if container.Labels["alias"] == *action.Component {
						containerID = container.ID
					}
				}

				// Check if the devfile run component containers have supervisord as the entrypoint.
				// Start the supervisord if the odo component does not exist
				if !componentExists {
					err = a.InitRunContainerSupervisord(*action.Component, podName, containers)
					if err != nil {
						return
					}
				}

				// Exec the supervisord ctl stop and start for the devrun program
				type devRunExecutable struct {
					command []string
				}
				devRunExecs := []devRunExecutable{
					{
						command: []string{common.SupervisordBinaryPath, "ctl", "stop", "all"},
					},
					{
						command: []string{common.SupervisordBinaryPath, "ctl", "start", string(common.DefaultDevfileRunCommand)},
					},
				}

				s = log.Spinner("Executing " + command.Name + " command " + fmt.Sprintf("%q", *action.Command))
				defer s.End(false)

				for _, devRunExec := range devRunExecs {

					err = exec.ExecuteCommand(&a.Client, podName, containerID, devRunExec.command, show)
					if err != nil {
						s.End(false)
						return
					}
				}
				s.End(true)
			}
		}
	}

	return
}

// InitRunContainerSupervisord initializes the supervisord in the container if
// the container has entrypoint that is not supervisord
func (a Adapter) InitRunContainerSupervisord(containerName, podName string, containers []types.Container) (err error) {
	for _, container := range containers {
		glog.V(3).Infof("MJF container.Labels[alias] %v", container.Labels["alias"])
		glog.V(3).Infof("MJF container.Command %v", container.Command)
		if container.Labels["alias"] == containerName && !reflect.DeepEqual(container.Command, []string{common.SupervisordBinaryPath}) {
			command := []string{common.SupervisordBinaryPath, "-c", common.SupervisordConfFile, "-d"}
			err = exec.ExecuteCommand(&a.Client, podName, container.ID, command, true)
		}
	}

	return
}
