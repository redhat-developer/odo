package init

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/component"
	_init "github.com/redhat-developer/odo/pkg/init"
	"github.com/redhat-developer/odo/pkg/init/asker"
	"github.com/redhat-developer/odo/pkg/init/params"
	"github.com/redhat-developer/odo/pkg/init/registry"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/preference"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"

	"k8s.io/kubectl/pkg/util/templates"
	"k8s.io/utils/pointer"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "init"

var initExample = templates.Examples(`
  # Boostrap a new component in interactive mode
  %[1]s

  # Bootstrap a new component with a specific devfile from registry
  %[1]s --name my-app --devfile nodejs
  
  # Bootstrap a new component with a specific devfile from a specific registry
  %[1]s --name my-app --devfile nodejs --devfile-registry MyRegistry
  
  # Bootstrap a new component with a specific devfile from the local filesystem
  %[1]s --name my-app --devfile-path $HOME/devfiles/nodejs/devfile.yaml
  
  # Bootstrap a new component with a specific devfile from the web
  %[1]s --name my-app --devfile-path https://devfiles.example.com/nodejs/devfile.yaml

  # Bootstrap a new component and download a starter project
  %[1]s --name my-app --devfile nodejs --starter nodejs-starter
  `)

type InitOptions struct {
	// CMD context
	ctx context.Context

	// filesystem on which command is running
	fsys filesystem.Filesystem

	// Clients
	initClient       _init.Client
	preferenceClient preference.Client

	// devfileLocation is the information needed to pull a devfile
	devfileLocation *params.DevfileLocation

	// starter is the name of the starter project to download
	starter string

	// componentName is the name of component to set in devfile
	componentName string

	// Destination directory
	contextDir string
}

// NewInitOptions creates a new InitOptions instance
func NewInitOptions(fsys filesystem.Filesystem, initClient _init.Client, prefClient preference.Client) *InitOptions {
	return &InitOptions{
		fsys:             fsys,
		initClient:       initClient,
		preferenceClient: prefClient,
		devfileLocation:  &params.DevfileLocation{},
	}
}

// Complete will build the parameters for init, using different backends based on the flags set,
// either by using flags or interactively if no flag is passed
// Complete will return an error immediately if the current working directory is not empty
func (o *InitOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {

	o.ctx = cmdline.Context()

	o.contextDir, err = o.fsys.Getwd()
	if err != nil {
		return err
	}

	empty, err := isEmpty(o.fsys, o.contextDir)
	if err != nil {
		return err
	}
	if !empty {
		return errors.New("The current directory is not empty. You can bootstrap new component only in empty directory.\nIf you have existing code that you want to deploy use `odo deploy` or use `odo dev` command to quickly iterate on your component.")
	}

	flags := cmdline.GetFlags()
	o.devfileLocation, err = o.initClient.SelectDevfile(flags)
	if err != nil {
		odoutil.LogErrorAndExit(err, "")
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
	return o.devfileLocation.Validate(o.preferenceClient)
}

// Run contains the logic for the odo command
func (o *InitOptions) Run() (err error) {

	var starterDownloaded bool

	defer func() {
		if err == nil {
			return
		}
		if starterDownloaded {
			err = fmt.Errorf("%w\nThe command failed after downloading the starter project. By security, the directory is not cleaned up.", err)
		} else {
			_ = o.fsys.Remove("devfile.yaml")
			err = fmt.Errorf("%w\nThe command failed, the devfile has been removed from current directory.", err)
		}
	}()

	devfilePath, err := o.initClient.DownloadDevfile(o.devfileLocation, o.contextDir)
	if err != nil {
		return fmt.Errorf("Unable to download devfile: %w", err)
	}

	devfileObj, _, err := devfile.ParseDevfileAndValidate(parser.ParserArgs{Path: devfilePath, FlattenedDevfile: pointer.BoolPtr(false)})
	if err != nil {
		return err
	}

	scontext.SetComponentType(o.ctx, component.GetComponentTypeFromDevfileMetadata(devfileObj.Data.GetMetadata()))

	if o.starter != "" {
		// WARNING: this will remove all the content of the destination directory, ie the devfile.yaml file
		err = o.initClient.DownloadStarterProject(devfileObj, o.starter, o.contextDir)
		if err != nil {
			return fmt.Errorf("unable to download starter project %q: %w", o.starter, err)
		}
		starterDownloaded = true

		// in case the starter project contains a devfile, read it again
		if _, err = o.fsys.Stat(devfilePath); err == nil {
			devfileObj, _, err = devfile.ParseDevfileAndValidate(parser.ParserArgs{Path: devfilePath, FlattenedDevfile: pointer.BoolPtr(false)})
			if err != nil {
				return err
			}
		}
	}

	// Set the name in the devfile *AND* writes the devfile back to the disk in case
	// it has been removed and not replaced by the starter project
	err = devfileObj.SetMetadataName(o.componentName)
	if err != nil {
		return fmt.Errorf("Failed to update the devfile's name: %w", err)
	}

	log.Italicf(`
Your new component %q is ready in the current directory.
To start editing your component, use "odo dev" and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
To deploy your component to a cluster use "odo deploy".`, o.componentName)

	return nil
}

// NewCmdInit implements the odo command
func NewCmdInit(name, fullName string) *cobra.Command {
	fsys := filesystem.DefaultFs{}
	prefClient, err := preference.NewClient()
	registryClient := registry.NewRegistryClient()
	if err != nil {
		odoutil.LogErrorAndExit(err, "unable to set preference, something is wrong with odo, kindly raise an issue at https://github.com/redhat-developer/odo/issues/new?template=Bug.md")
	}

	backends := []params.ParamsBuilder{
		params.NewFlagsBuilder(),
		params.NewInteractiveBuilder(asker.NewSurveyAsker(), catalog.NewCatalogClient(filesystem.DefaultFs{}, prefClient)),
	}
	o := NewInitOptions(fsys, _init.NewInitClient(backends, fsys, prefClient, registryClient), prefClient)
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

	initCmd.Flags().StringVar(&o.componentName, params.FLAG_NAME, "", "name of the component to create")
	initCmd.Flags().StringVar(&o.devfileLocation.Devfile, params.FLAG_DEVFILE, "", "name of the devfile in devfile registry")
	initCmd.Flags().StringVar(&o.devfileLocation.DevfileRegistry, params.FLAG_DEVFILE_REGISTRY, "", "name of the devfile registry (as configured in \"odo registry list\"). It can be used in combination with --devfile, but not with --devfile-path")
	initCmd.Flags().StringVar(&o.starter, params.FLAG_STARTER, "", "name of the starter project. Available starter projects can be found with \"odo catalog describe component <devfile>\"")
	initCmd.Flags().StringVar(&o.devfileLocation.DevfilePath, params.FLAG_DEVFILE_PATH, "", "path to a devfile. This is an alternative to using devfile from Devfile registry. It can be local filesystem path or http(s) URL")

	// Add a defined annotation in order to appear in the help menu
	initCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return initCmd
}
