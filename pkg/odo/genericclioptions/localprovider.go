package genericclioptions

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
)

// GetValidEnvInfo accesses the environment file
func GetValidEnvInfo(cmdline cmdline.Cmdline) (*envinfo.EnvSpecificInfo, error) {
	componentContext, err := cmdline.GetWorkingDirectory()

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
	canWeSkip, err := cmdline.CheckIfConfigurationNeeded()
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
