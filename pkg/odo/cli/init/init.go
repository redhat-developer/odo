package init

import (
	"context"
	"errors"
	"fmt"
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"

	"github.com/spf13/cobra"

	"github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/init/backend"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"

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

	// Clients
	clientset *clientset.Clientset

	// Flags passed to the command
	flags map[string]string

	// Destination directory
	contextDir string
}

// NewInitOptions creates a new InitOptions instance
func NewInitOptions() *InitOptions {
	return &InitOptions{}
}

func (o *InitOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete will build the parameters for init, using different backends based on the flags set,
// either by using flags or interactively if no flag is passed
// Complete will return an error immediately if the current working directory is not empty
func (o *InitOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {

	o.ctx = cmdline.Context()

	o.contextDir, err = o.clientset.FS.Getwd()
	if err != nil {
		return err
	}

	o.flags = cmdline.GetFlags()

	return nil
}

// Validate validates the InitOptions based on completed values
func (o *InitOptions) Validate() error {

	devfilePresent, err := location.DirectoryContainsDevfile(o.clientset.FS, o.contextDir)
	if err != nil {
		return err
	}
	if devfilePresent {
		return errors.New("a devfile already exists in the current directory")
	}

	err = o.clientset.InitClient.Validate(o.flags, o.clientset.FS, o.contextDir)
	if err != nil {
		return err
	}
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
			err = fmt.Errorf("%w\nthe command failed after downloading the starter project. By security, the directory is not cleaned up", err)
		} else {
			_ = o.clientset.FS.Remove("devfile.yaml")
			err = fmt.Errorf("%w\nthe command failed, the devfile has been removed from current directory", err)
		}
	}()

	isEmptyDir, err := location.DirIsEmpty(o.clientset.FS, o.contextDir)
	if err != nil {
		return err
	}
	if isEmptyDir && len(o.flags) == 0 {
		fmt.Println("The current directory is empty. odo will help you start a new project.")
	} else if len(o.flags) == 0 {
		fmt.Println("The current directory already contains source code. " +
			"odo will try to autodetect the language and project type in order to select the best suited Devfile for your project.")
	}

	devfileObj, devfilePath, err := o.clientset.InitClient.SelectAndPersonalizeDevfile(o.flags, o.contextDir)

	scontext.SetComponentType(o.ctx, component.GetComponentTypeFromDevfileMetadata(devfileObj.Data.GetMetadata()))

	starterInfo, err := o.clientset.InitClient.SelectStarterProject(devfileObj, o.flags, o.clientset.FS, o.contextDir)
	if err != nil {
		return err
	}
	if starterInfo != nil {
		// WARNING: this will remove all the content of the destination directory, ie the devfile.yaml file
		err = o.clientset.InitClient.DownloadStarterProject(starterInfo, o.contextDir)
		if err != nil {
			return fmt.Errorf("unable to download starter project %q: %w", starterInfo.Name, err)
		}
		starterDownloaded = true

		// in case the starter project contains a devfile, read it again
		if _, err = o.clientset.FS.Stat(devfilePath); err == nil {
			devfileObj, _, err = devfile.ParseDevfileAndValidate(parser.ParserArgs{Path: devfilePath, FlattenedDevfile: pointer.BoolPtr(false)})
			if err != nil {
				return err
			}
		}
	}

	// Set the name in the devfile *AND* writes the devfile back to the disk in case
	// it has been removed and not replaced by the starter project
	err = o.clientset.InitClient.PersonalizeName(devfileObj, o.flags)
	if err != nil {
		return fmt.Errorf("Failed to update the devfile's name: %w", err)
	}

	exitMessage := fmt.Sprintf(`
Your new component %q is ready in the current directory.
To start editing your component, use "odo dev" and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.`, devfileObj.Data.GetMetadata().Name)

	// show information about `odo deploy` command only if the devfile has kind: deploy command defined.
	commands, _ := devfileObj.Data.GetCommands(common.DevfileOptions{
		CommandOptions: common.CommandOptions{
			CommandGroupKind: v1alpha2.DeployCommandGroupKind,
		},
	})
	if len(commands) != 0 {
		exitMessage += "\nTo deploy your component to a cluster use \"odo deploy\"."
	}
	log.Italic(exitMessage)

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
	clientset.Add(initCmd, clientset.PREFERENCE, clientset.FILESYSTEM, clientset.REGISTRY, clientset.INIT)

	initCmd.Flags().String(backend.FLAG_NAME, "", "name of the component to create")
	initCmd.Flags().String(backend.FLAG_DEVFILE, "", "name of the devfile in devfile registry")
	initCmd.Flags().String(backend.FLAG_DEVFILE_REGISTRY, "", "name of the devfile registry (as configured in \"odo registry list\"). It can be used in combination with --devfile, but not with --devfile-path")
	initCmd.Flags().String(backend.FLAG_STARTER, "", "name of the starter project. Available starter projects can be found with \"odo catalog describe component <devfile>\"")
	initCmd.Flags().String(backend.FLAG_DEVFILE_PATH, "", "path to a devfile. This is an alternative to using devfile from Devfile registry. It can be local filesystem path or http(s) URL")

	// Add a defined annotation in order to appear in the help menu
	initCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return initCmd
}
