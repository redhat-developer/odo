package genericclioptions

import (
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

	return envInfo, nil
}
