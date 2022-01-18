package init

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
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
	// Backends to build init parameters
	backends []ParamsBuilder

	// filesystem on which command is running
	fsys filesystem.Filesystem

	// the parameters needed to run the init procedure
	initParams
}

// NewInitOptions creates a new InitOptions instance
func NewInitOptions(backends []ParamsBuilder, fsys filesystem.Filesystem) *InitOptions {
	return &InitOptions{
		backends: backends,
		fsys:     fsys,
	}
}

// Complete will build the parameters for init, using different backends based on the flags set,
// either by using flags or interactively is no flag is passed
// Complete will return an error immediately if the current working directory is not empty
func (o *InitOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {

	empty, err := isEmpty(o.fsys, ".")
	if err != nil {
		return err
	}
	if !empty {
		return errors.New("The current directory is not empty. You can bootstrap new component only in empty directory.\nIf you have existing code that you want to deploy use `odo deploy` or use `odo dev` command to quickly iterate on your component.")
	}

	flags := cmdline.GetFlags()
	done := false
	for _, backend := range o.backends {
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

func isEmpty(fsys filesystem.Filesystem, path string) (bool, error) {
	files, err := fsys.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(files) == 0, nil
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
	backends := []ParamsBuilder{
		&FlagsBuilder{},
		NewInteractiveBuilder(NewSurveyAsker(), catalog.NewCatalogClient()),
	}

	o := NewInitOptions(backends, filesystem.DefaultFs{})
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
