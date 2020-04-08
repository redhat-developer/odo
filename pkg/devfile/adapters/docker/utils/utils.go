package utils

import (
	"fmt"
	"reflect"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/lclient"
)

// ComponentExists checks if Docker containers labeled with the specified component name exists
func ComponentExists(client lclient.Client, name string) bool {
	containerList, err := client.GetContainerList()
	if err != nil {
		return false
	}
	containers := client.GetContainersByComponent(name, containerList)
	return len(containers) != 0
}

// ConvertEnvs converts environment variables from the devfile structure to an array of strings, as expected by Docker
func ConvertEnvs(vars []common.DockerimageEnv) []string {
	dockerVars := []string{}
	for _, env := range vars {
		envString := fmt.Sprintf("%s=%s", *env.Name, *env.Value)
		dockerVars = append(dockerVars, envString)
	}
	return dockerVars
}

// DoesContainerNeedUpdating returns true if a given container needs to be removed and recreated
// This function compares values in the container vs corresponding values in the devfile component.
// If any of the values between the two differ, a restart is required (and this function returns true)
// Unlike Kube, Docker doesn't provide a mechanism to update a container in place only when necesary
// so this function is necessary to prevent having to restart the container on every odo pushs
func DoesContainerNeedUpdating(component common.DevfileComponent, containerConfig *container.Config) bool {
	// If the image was changed in the devfile, the container needs to be updated
	if *component.Image != containerConfig.Image {
		return true
	}

	// Update the container if the env vars were updated in the devfile
	// Need to convert the devfile envvars to the format expected by Docker
	devfileEnvVars := ConvertEnvs(component.Env)
	return !reflect.DeepEqual(devfileEnvVars, containerConfig.Env)
}

func AddProjectVolumeToComp(projectVolumeName string, hostConfig *container.HostConfig) *container.HostConfig {
	mount := mount.Mount{
		Type:   mount.TypeVolume,
		Source: projectVolumeName,
		Target: lclient.OdoSourceVolumeMount,
	}
	hostConfig.Mounts = append(hostConfig.Mounts, mount)

	return hostConfig
}

// GetProjectVolumeLabels returns the label selectors used to retrieve the project/source volume for a given component
func GetProjectVolumeLabels(componentName string) map[string]string {
	volumeLabels := map[string]string{
		"component": componentName,
		"type":      "projects",
	}
	return volumeLabels
}
