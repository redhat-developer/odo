package genericclioptions

import (
	"errors"

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
	configIsNeeded, err := cmdline.CheckIfConfigurationNeeded()
	if err != nil {
		return nil, err
	}
	if !configIsNeeded {
		return envInfo, nil
	}

	// Check to see if the environment file exists
	if !envInfo.Exists() {
		//revive:disable:error-strings This is a top-level error message displayed as is to the end user
		return nil, errors.New(`The current directory does not represent an odo component.
If you are working from an empty directory, use "odo init" to get started, else start editing your component, use "odo dev" and open this folder in your favorite IDE. Changes will be directly reflected on the cluster.
To deploy your component to a cluster use "odo deploy".
Or switch to directory with a component.`)
		//revive:enable:error-strings
	}

	return envInfo, nil
}
