package genericclioptions

import (
	"fmt"

	"github.com/openshift/odo/pkg/envinfo"
	"github.com/spf13/cobra"
)

// GetValidEnvInfo is just a wrapper for getValidEnvInfo
func GetValidEnvInfo(command *cobra.Command) (*envinfo.EnvSpecificInfo, error) {
	return getValidEnvInfo(command)
}

// getValidEnvInfo accesses the environment file
func getValidEnvInfo(command *cobra.Command) (*envinfo.EnvSpecificInfo, error) {

	componentContext, err := GetContextFlagValue(command)

	if err != nil {
		return nil, err
	}

	// Access the env file
	envInfo, err := envinfo.NewEnvSpecificInfo(componentContext)
	if err != nil {
		return nil, err
	}

	// Now we check to see if we can skip gathering the information.
	// Return if we can skip gathering configuration information
	canWeSkip, err := checkIfConfigurationNeeded(command)
	if err != nil {
		return nil, err
	}
	if canWeSkip {
		return envInfo, nil
	}

	// Check to see if the environment file exists
	if !envInfo.Exists() {
		return nil, fmt.Errorf("the current directory does not represent an odo component. Use 'odo create' to create component here or switch to directory with a component")
	}

	return envInfo, nil
}
