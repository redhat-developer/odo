package init

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "init"

const FLAG_NAME = "name"
const FLAG_DEVFILE = "devfile"
const FLAG_DEVFILE_REGISTRY = "devfile-registry"
const FLAG_STARTER = "starter"
const FLAG_DEVFILE_PATH = "devfile-path"

var initExample = templates.Examples(`
  # Boostrap a new project in interactive mode
  %[1]s
`)

type InitOptions struct {
	// Context
	*genericclioptions.Context

	// the parameters needed to run the init procedure
	initParams
}

// NewInitOptions creates a new InitOptions instance
func NewInitOptions() *InitOptions {
	return &InitOptions{}
}

// Complete will build the parameters for init, using different backends based on the flags set,
// either by using flags or interactively is no flag is passed
func (o *InitOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	flags := cmdline.GetFlags()
	backends := []ParamsBuilder{
		&FlagsBuilder{},
		&InteractiveBuilder{},
	}

	done := false
	for _, backend := range backends {
		if backend.IsAdequate(flags) {
			o.initParams, err = backend.ParamsBuild()
			if err != nil {
				return err
			}
			done = true
			break
		}
	}

	if !done {
		util.LogErrorAndExit(nil, "no backend found to build init parameters. This should not happen")
	}
	return nil
}

// Validate validates the InitOptions based on completed values
func (o *InitOptions) Validate() error {
	return o.initParams.validate()
}

// Run contains the logic for the odo command
func (o *InitOptions) Run() error {
	return nil
}

// NewCmdInit implements the odo command
func NewCmdInit(name, fullName string) *cobra.Command {
	o := NewInitOptions()
	initCmd := &cobra.Command{
		Use:     name,
		Short:   "Init bootstraps a new project",
		Long:    "Bootstraps a new project",
		Example: fmt.Sprintf(initExample, fullName),
		Args:    cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	initCmd.Flags().StringVar(&o.name, FLAG_NAME, "", "name of the component to create")
	initCmd.Flags().StringVar(&o.devfile, FLAG_DEVFILE, "", "name of the devfile in devfile registry")
	initCmd.Flags().StringVar(&o.devfileRegistry, FLAG_DEVFILE_REGISTRY, "", "name of the devfile registry (as configured in odo registry). It can be used in combination with --devfile, but not with --devfile-path")
	initCmd.Flags().StringVar(&o.starter, FLAG_STARTER, "", "name of the starter project")
	initCmd.Flags().StringVar(&o.devfilePath, FLAG_DEVFILE_PATH, "", "path to a devfile. This is alternative to using devfile from Devfile registry. It can be local filesystem path or http(s) URL")

	// Add a defined annotation in order to appear in the help menu
	initCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return initCmd
}
