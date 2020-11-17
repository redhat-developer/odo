package genericclioptions

import (
	"fmt"

	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/odo/util"
	pkgUtil "github.com/openshift/odo/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

// GetValidEnvInfo is just a wrapper for getValidEnvInfo
func GetValidEnvInfo(command *cobra.Command) (*envinfo.EnvSpecificInfo, error) {
	return getValidEnvInfo(command)
}

// getValidEnvInfo accesses the environment file
func getValidEnvInfo(command *cobra.Command) (*envinfo.EnvSpecificInfo, error) {

	// Get details from the env file
	componentContext := FlagValueIfSet(command, ContextFlagName)

	// Grab the absolute path of the env file
	if componentContext != "" {
		fAbs, err := pkgUtil.GetAbsPath(componentContext)
		util.LogErrorAndExit(err, "")
		componentContext = fAbs
	} else {
		fAbs, err := pkgUtil.GetAbsPath(".")
		util.LogErrorAndExit(err, "")
		componentContext = fAbs
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
		return nil, fmt.Errorf("The current directory does not represent an odo component. Use 'odo create' to create component here or switch to directory with a component")
	}

	return envInfo, nil
}

func getValidConfig(command *cobra.Command, ignoreMissingConfiguration bool) (*config.LocalConfigInfo, error) {

	// Get details from the local config file
	contextDir := FlagValueIfSet(command, ContextFlagName)

	// Grab the absolute path of the configuration
	if contextDir != "" {
		fAbs, err := pkgUtil.GetAbsPath(contextDir)
		util.LogErrorAndExit(err, "")
		contextDir = fAbs
	}
	// Access the local configuration
	localConfiguration, err := config.NewLocalConfigInfo(contextDir)
	if err != nil {
		return nil, err
	}

	// Now we check to see if we can skip gathering the information.
	// If true, we just return.
	canWeSkip, err := checkIfConfigurationNeeded(command)
	if err != nil {
		return nil, err
	}
	if canWeSkip {
		return localConfiguration, nil
	}

	// If file does not exist at this point, raise an error
	// HOWEVER..
	// When using auto-completion, we should NOT error out, just ignore the fact that there is no configuration
	if !localConfiguration.Exists() && ignoreMissingConfiguration {
		klog.V(4).Info("There is NO config file that exists, we are however ignoring this as the ignoreMissingConfiguration flag has been passed in as true")
	} else if !localConfiguration.Exists() {
		return nil, fmt.Errorf("The current directory does not represent an odo component. Use 'odo create' to create component here or switch to directory with a component")
	}

	// else simply return the local config info
	return localConfiguration, nil
}
