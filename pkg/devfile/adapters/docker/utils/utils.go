package utils

import (
	"fmt"
	"strconv"

	"github.com/docker/go-connections/nat"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/openshift/odo/pkg/kclient/generator"

	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/parser/data"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/lclient"
	"github.com/openshift/odo/pkg/util"

	"github.com/pkg/errors"
)

const (
	// SupervisordVolume is supervisord volume type
	SupervisordVolume = "supervisord"

	// ProjectsVolume is project source volume type
	ProjectsVolume = "projects"
)

// ComponentExists checks if a component exist
// returns true, if the number of containers equals the number of unique devfile container components
// returns false, if number of containers is zero
// returns an error, if number of containers is more than zero but does not equal the number of unique devfile container components
func ComponentExists(client lclient.Client, data data.DevfileData, name string) (bool, error) {
	containers, err := GetComponentContainers(client, name)
	if err != nil {
		return false, errors.Wrapf(err, "unable to get the containers for component %s", name)
	}

	containerComponents := generator.GetDevfileContainerComponents(data)

	var componentExists bool
	if len(containers) == 0 {
		componentExists = false
	} else if len(containers) == len(containerComponents) {
		componentExists = true
	} else if len(containers) > 0 && len(containers) != len(containerComponents) {
		return true, errors.New(fmt.Sprintf("component %s is in an invalid state, please execute odo delete and retry odo push", name))
	}

	return componentExists, nil
}

// GetComponentContainers returns a list of the running component containers
func GetComponentContainers(client lclient.Client, componentName string) (containers []types.Container, err error) {
	containerList, err := client.GetContainerList(false)
	if err != nil {
		return nil, err
	}
	containers = client.GetContainersByComponent(componentName, containerList)

	return containers, nil
}

// GetContainerIDForAlias returns the container ID for the devfile alias from a list of containers
func GetContainerIDForAlias(containers []types.Container, alias string) string {
	containerID := ""
	for _, container := range containers {
		if container.Labels["alias"] == alias {
			containerID = container.ID
		}
	}
	return containerID
}

// ConvertEnvs converts environment variables from the devfile structure to an array of strings, as expected by Docker
func ConvertEnvs(vars []common.Env) []string {
	dockerVars := []string{}
	for _, env := range vars {
		envString := fmt.Sprintf("%s=%s", env.Name, env.Value)
		dockerVars = append(dockerVars, envString)
	}
	return dockerVars
}

// ConvertPorts converts endpoints from the devfile structure to PortSet, which is expected by Docker
func ConvertPorts(endpoints []common.Endpoint) nat.PortSet {
	portSet := nat.PortSet{}
	for _, endpoint := range endpoints {
		port := nat.Port(strconv.Itoa(int(endpoint.TargetPort)) + "/tcp")
		portSet[port] = struct{}{}
	}
	return portSet
}

// DoesContainerNeedUpdating returns true if a given container needs to be removed and recreated
// This function compares values in the container vs corresponding values in the devfile component.
// If any of the values between the two differ, a restart is required (and this function returns true)
// Unlike Kube, Docker doesn't provide a mechanism to update a container in place only when necesary
// so this function is necessary to prevent having to restart the container on every odo pushs
func DoesContainerNeedUpdating(component common.DevfileComponent, containerConfig *container.Config, hostConfig *container.HostConfig, devfileMounts []mount.Mount, containerMounts []types.MountPoint, portMap nat.PortMap) bool {
	// If the image was changed in the devfile, the container needs to be updated
	if component.Container.Image != containerConfig.Image {
		return true
	}

	// Update the container if the volumes were updated in the devfile
	for _, devfileMount := range devfileMounts {
		if !containerHasMount(devfileMount, containerMounts) {
			return true
		}
	}

	// Update the container if the env vars were updated in the devfile
	// Need to convert the devfile envvars to the format expected by Docker
	devfileEnvVars := ConvertEnvs(component.Container.Env)
	for _, envVar := range devfileEnvVars {
		if !containerHasEnvVar(envVar, containerConfig.Env) {
			return true
		}
	}

	devfilePorts := ConvertPorts(component.Container.Endpoints)
	for port := range devfilePorts {
		if !containerHasPort(port, containerConfig.ExposedPorts) {
			return true
		}
	}

	for localInternalPort, localPortbinding := range portMap {
		if hostConfig.PortBindings[localInternalPort] == nil || hostConfig.PortBindings[localInternalPort][0].HostPort != localPortbinding[0].HostPort {
			// if there is no exposed port assigned to the internal port for the container, or if the exposed port has changed
			return true
		}
	}

	for containerInternalPort := range hostConfig.PortBindings {
		if portMap[containerInternalPort] == nil {
			// if the url is locally deleted
			return true
		}
	}

	return false
}

// AddVolumeToContainer adds the volume name and mount to the container host config
func AddVolumeToContainer(volumeName, volumeMount string, hostConfig *container.HostConfig) *container.HostConfig {
	mount := mount.Mount{
		Type:   mount.TypeVolume,
		Source: volumeName,
		Target: volumeMount,
	}
	hostConfig.Mounts = append(hostConfig.Mounts, mount)

	return hostConfig
}

// GetProjectVolumeLabels returns the label selectors used to retrieve the project/source volume for a given component
func GetProjectVolumeLabels(componentName string) map[string]string {
	volumeLabels := map[string]string{
		"component": componentName,
		"type":      ProjectsVolume,
	}
	return volumeLabels
}

// GetContainerLabels returns the label selectors used to retrieve/create the component container
func GetContainerLabels(componentName, alias string) map[string]string {
	containerLabels := map[string]string{
		"component": componentName,
		"alias":     alias,
	}
	return containerLabels
}

// GetSupervisordVolumeLabels returns the label selectors used to retrieve the supervisord volume
func GetSupervisordVolumeLabels(componentName string) map[string]string {
	image := adaptersCommon.GetBootstrapperImage()
	_, imageWithoutTag, _, imageTag := util.ParseComponentImageName(image)

	supervisordLabels := map[string]string{
		"component": componentName,
		"type":      SupervisordVolume,
		"image":     imageWithoutTag,
		"version":   imageTag,
	}
	return supervisordLabels
}

// containerHasEnvVar returns true if the specified env var (and value) exist in the list of container env vars
func containerHasEnvVar(envVar string, containerEnv []string) bool {
	for _, env := range containerEnv {
		if envVar == env {
			return true
		}
	}
	return false
}

// containerHasMount returns true if the specified volume is mounted in the given container
func containerHasMount(devfileMount mount.Mount, containerMounts []types.MountPoint) bool {
	for _, mount := range containerMounts {
		if devfileMount.Source == mount.Name && devfileMount.Target == mount.Destination {
			return true
		}
	}
	return false
}

func containerHasPort(devfilePort nat.Port, exposedPorts nat.PortSet) bool {
	for port := range exposedPorts {
		if devfilePort.Port() == port.Port() {
			return true
		}
	}
	return false
}
