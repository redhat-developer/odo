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
	"github.com/redhat-developer/odo/pkg/init/backend"
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

	// Flags passed to the command
	flags map[string]string

	// devfileLocation is the information needed to pull a devfile
	devfileLocation *backend.DevfileLocation

	// Destination directory
	contextDir string
}

// NewInitOptions creates a new InitOptions instance
func NewInitOptions(fsys filesystem.Filesystem, initClient _init.Client, prefClient preference.Client) *InitOptions {
	return &InitOptions{
		fsys:             fsys,
		initClient:       initClient,
		preferenceClient: prefClient,
		devfileLocation:  &backend.DevfileLocation{},
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

	o.flags = cmdline.GetFlags()

	err = o.initClient.Validate(o.flags)
	if err != nil {
		return err
	}
	o.devfileLocation, err = o.initClient.SelectDevfile(o.flags)
	if err != nil {
		return err
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
	return nil
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

	starterInfo, err := o.initClient.SelectStarterProject(devfileObj, o.flags)
	if starterInfo != nil {
		// WARNING: this will remove all the content of the destination directory, ie the devfile.yaml file
		err = o.initClient.DownloadStarterProject(starterInfo, o.contextDir)
		if err != nil {
			return fmt.Errorf("unable to download starter project %q: %w", starterInfo.Name, err)
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
	err = o.initClient.PersonalizeName(devfileObj, o.flags)
	if err != nil {
		return fmt.Errorf("Failed to update the devfile's name: %w", err)
	}

	log.Italicf(`
Your new component %q is ready in the current directory.
To start editing your component, use "odo dev" and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
To deploy your component to a cluster use "odo deploy".`, devfileObj.Data.GetMetadata().Name)

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

	backends := []backend.InitBackend{
		backend.NewFlagsBackend(prefClient),
		backend.NewInteractiveBackend(asker.NewSurveyAsker(), catalog.NewCatalogClient(filesystem.DefaultFs{}, prefClient)),
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

	initCmd.Flags().String(backend.FLAG_NAME, "", "name of the component to create")
	initCmd.Flags().String(backend.FLAG_DEVFILE, "", "name of the devfile in devfile registry")
	initCmd.Flags().String(backend.FLAG_DEVFILE_REGISTRY, "", "name of the devfile registry (as configured in \"odo registry list\"). It can be used in combination with --devfile, but not with --devfile-path")
	initCmd.Flags().String(backend.FLAG_STARTER, "", "name of the starter project. Available starter projects can be found with \"odo catalog describe component <devfile>\"")
	initCmd.Flags().String(backend.FLAG_DEVFILE_PATH, "", "path to a devfile. This is an alternative to using devfile from Devfile registry. It can be local filesystem path or http(s) URL")

	// Add a defined annotation in order to appear in the help menu
	initCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return initCmd
}
