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

To get started,
  * Create and move to a new directory, or use an existing one.
  * Run "odo init" from the directory to initialize an odo component.
  * Start editing the component in an IDE and run "odo dev" to see your changes get reflected on the cluster.

Visit https://odo.dev for more information.`)
		//revive:enable:error-strings
	}

	return envInfo, nil
}
