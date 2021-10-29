package component

import (
	"path/filepath"

	"github.com/openshift/odo/pkg/odo/genericclioptions"

	"github.com/openshift/odo/pkg/devfile/location"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (co *CreateOptions) Complete2(name string, cmd *cobra.Command, args []string) (err error) {
	var devfileData []byte
	// GETTERS
	// Get context
	co.Context, err = getContext(co.now, cmd)
	if err != nil {
		return err
	}
	// Get the app name
	co.appName = genericclioptions.ResolveAppFlag(cmd)

	// Get DevfilePath
	co.DevfilePath = location.DevfileLocation(co.componentContext)
	//Check whether the directory already contains a devfile, this check should happen early
	co.devfileMetadata.userCreatedDevfile = util.CheckPathExists(co.DevfilePath)

	// EnvFilePath is the path of env file for devfile component
	envFilePath := filepath.Join(LocalDirectoryDefaultLocation, EnvYAMLFilePath)
	if co.componentContext != "" {
		envFilePath = filepath.Join(co.componentContext, EnvYAMLFilePath)
	}

	// Use Interactive mode if: 1) no args are passed || 2) the devfile exists || 3) --devfile is used
	if len(args) == 0 && !util.CheckPathExists(co.DevfilePath) && co.devfileMetadata.devfilePath.value == "" {
		co.interactive = true
	}

	// CONFLICT CHECK
	// Check if a component exists
	if util.CheckPathExists(envFilePath) && util.CheckPathExists(co.DevfilePath) {
		return errors.New("this directory already contains a component")
	}
	// Check if there is a dangling env file; delete the env file if found
	if util.CheckPathExists(envFilePath) && !util.CheckPathExists(co.DevfilePath) {
		log.Warningf("Found a dangling env file without a devfile, overwriting it")
		// Note: if the IF condition seems to have a side-effect, it is better to do the condition check separately, like below
		err := util.DeletePath(envFilePath)
		if err != nil {
			return err
		}
	}
	//Check if the directory already contains a devfile when --devfile flag is passed
	if util.CheckPathExists(co.DevfilePath) && co.devfileMetadata.devfilePath.value != "" && !util.PathEqual(co.DevfilePath, co.devfileMetadata.devfilePath.value) {
		return errors.New("this directory already contains a devfile, you can't specify devfile via --devfile")
	}
	// Check if both --devfile and --registry flag are used, in which case raise an error
	if co.devfileMetadata.devfileRegistry.Name != "" && co.devfileMetadata.devfilePath.value != "" {
		return errors.New("you can't specify registry via --registry if you want to use the devfile that is specified via --devfile")
	}

	//	Interactive Mode
	//  Normal Mode
	//	Existing devfile Mode
	//	--devfile Mode

	return nil
}
