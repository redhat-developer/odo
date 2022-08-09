package genericclioptions

import (
	"errors"
	"fmt"

	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
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
		exitMessage := `The current directory does not represent an odo component. 

To get started,%s
  * Open this folder in your favorite IDE and start editing, your changes will be reflected directly on the cluster.

Visit https://odo.dev for more information.`

		if isEmpty, _ := location.DirIsEmpty(filesystem.DefaultFs{}, componentContext); isEmpty {
			exitMessage = fmt.Sprintf(exitMessage, `
  * Create and move to a new directory
  * Use "odo init" to initialize an odo component in the folder.
  * Use "odo dev" to deploy it on cluster.`)
		} else {
			exitMessage = fmt.Sprintf(exitMessage, `
  * Use "odo dev" to initialize an odo component for this folder and deploy it on cluster.`)
		}
		//revive:disable:error-strings This is a top-level error message displayed as is to the end user
		return nil, errors.New(exitMessage)
		//revive:enable:error-strings
	}

	return envInfo, nil
}
