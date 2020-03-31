package utils

import (
	"fmt"

	"github.com/openshift/odo/pkg/devfile/versions/common"
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
