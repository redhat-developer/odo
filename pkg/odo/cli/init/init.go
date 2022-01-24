package init

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/init/asker"
	"github.com/redhat-developer/odo/pkg/odo/cli/init/params"
	"github.com/redhat-developer/odo/pkg/odo/cli/init/registry"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/segment"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/util"

	"k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "init"

var initExample = templates.Examples(`
  # Boostrap a new component in interactive mode
  %[1]s
`)

type InitOptions struct {
	// Backends to build init parameters
	backends []params.ParamsBuilder

	// filesystem on which command is running
	fsys filesystem.Filesystem

	// Clients
	preferenceClient preference.Client
	registryClient   registry.Client

	// the parameters needed to run the init procedure
	params.InitParams

	// Destination directory
	contextDir string
}

// NewInitOptions creates a new InitOptions instance
func NewInitOptions(backends []params.ParamsBuilder, fsys filesystem.Filesystem, prefClient preference.Client, registryClient registry.Client) *InitOptions {
	return &InitOptions{
		backends:         backends,
		fsys:             fsys,
		preferenceClient: prefClient,
		registryClient:   registryClient,
	}
}

// Complete will build the parameters for init, using different backends based on the flags set,
// either by using flags or interactively if no flag is passed
// Complete will return an error immediately if the current working directory is not empty
func (o *InitOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {

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
	done := false
	for _, backend := range o.backends {
		if backend.IsAdequate(flags) {
			o.InitParams, err = backend.ParamsBuild()
			if err != nil {
				return err
			}
			done = true
			break
		}
	}

	if !done {
		odoutil.LogErrorAndExit(nil, "no backend found to build init parameters. This should not happen")
	}

	scontext.SetComponentType(cmdline.Context(), o.InitParams.Devfile)

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
	return o.InitParams.Validate(o.preferenceClient)
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

	destDevfile := filepath.Join(o.contextDir, "devfile.yaml")
	if o.InitParams.DevfilePath != "" {
		err = o.downloadDirect(o.InitParams.DevfilePath, destDevfile)
	} else {
		err = o.downloadFromRegistry(o.InitParams.DevfileRegistry, o.InitParams.Devfile, o.contextDir)
	}
	if err != nil {
		return fmt.Errorf("Unable to download devfile: %w", err)
	}

	resolved := false
	devfileObj, _, err := devfile.ParseDevfileAndValidate(parser.ParserArgs{Path: destDevfile, FlattenedDevfile: &resolved})
	if err != nil {
		return err
	}

	if o.InitParams.Starter != "" {
		// WARNING: this will remove all the content of the destination directory, ie the devfile.yaml file
		err = o.downloadStarterProject(devfileObj, o.InitParams.Starter, o.contextDir)
		if err != nil {
			return fmt.Errorf("unable to download starter project %q: %w", o.InitParams.Starter, err)
		}
		starterDownloaded = true

		// in case the starter project contains a devfile, read it again
		if _, err = o.fsys.Stat(destDevfile); err == nil {
			resolved := false
			devfileObj, _, err = devfile.ParseDevfileAndValidate(parser.ParserArgs{Path: destDevfile, FlattenedDevfile: &resolved})
			if err != nil {
				return err
			}
		}
	}

	// Set the name in the devfile *AND* writes the devfile back to the disk in case
	// it has been removed and not replaced by the starter project
	err = devfileObj.SetMetadataName(o.InitParams.Name)
	if err != nil {
		return fmt.Errorf("Failed to update the devfile's name: %w", err)
	}

	log.Italicf(`
Your new component %q is ready in the current directory.
To start editing your component, use "odo dev" and open this folder in your favorite IDE.
Changes will be directly reflected on the cluster.
To deploy your component to a cluster use "odo deploy".`, o.InitParams.Name)

	return nil
}

// downloadDirect downloads a devfile at the provided URL and saves it in dest
func (o *InitOptions) downloadDirect(URL string, dest string) error {
	parsedURL, err := url.Parse(URL)
	if err != nil {
		return err
	}
	if strings.HasPrefix(parsedURL.Scheme, "http") {
		downloadSpinner := log.Spinnerf("Downloading devfile from %q", URL)
		defer downloadSpinner.End(false)
		params := util.HTTPRequestParams{
			URL: URL,
		}
		devfileData, err := o.registryClient.DownloadFileInMemory(params)
		if err != nil {
			return err
		}
		err = o.fsys.WriteFile(dest, devfileData, 0644)
		if err != nil {
			return err
		}
		downloadSpinner.End(true)
	} else {
		downloadSpinner := log.Spinnerf("Copying devfile from %q", URL)
		defer downloadSpinner.End(false)
		content, err := o.fsys.ReadFile(URL)
		if err != nil {
			return err
		}
		info, err := o.fsys.Stat(URL)
		if err != nil {
			return err
		}
		err = o.fsys.WriteFile(dest, content, info.Mode().Perm())
		if err != nil {
			return err
		}
		downloadSpinner.End(true)
	}

	return nil
}

// downloadFromRegistry downloads a devfile from the provided registry and saves it in dest
// If registryName is empty, will try to download the devfile from the list of registries in preferences
func (o *InitOptions) downloadFromRegistry(registryName string, devfile string, dest string) error {
	var downloadSpinner *log.Status
	var forceRegistry bool
	if registryName == "" {
		downloadSpinner = log.Spinnerf("Downloading devfile %q", devfile)
		forceRegistry = false
	} else {
		downloadSpinner = log.Spinnerf("Downloading devfile %q from registry %q", devfile, registryName)
		forceRegistry = true
	}
	defer downloadSpinner.End(false)

	registries := o.preferenceClient.RegistryList()
	var reg preference.Registry
	for _, reg = range *registries {
		if forceRegistry && reg.Name == registryName {
			err := o.registryClient.PullStackFromRegistry(reg.URL, devfile, dest, segment.GetRegistryOptions())
			if err != nil {
				return err
			}
			downloadSpinner.End(true)
			return nil
		} else if !forceRegistry {
			err := o.registryClient.PullStackFromRegistry(reg.URL, devfile, dest, segment.GetRegistryOptions())
			if err != nil {
				continue
			}
			downloadSpinner.End(true)
			return nil
		}
	}

	return fmt.Errorf("unable to find the registry with name %q", devfile)
}

// downloadStarterProject downloads the starter project referenced in devfile and stores it in dest directory
// WARNING: This will first remove all the content of dest.
func (o *InitOptions) downloadStarterProject(devfile parser.DevfileObj, project string, dest string) error {
	projects, err := devfile.Data.GetStarterProjects(common.DevfileOptions{})
	if err != nil {
		return err
	}
	var prj v1alpha2.StarterProject
	var found bool
	for _, prj = range projects {
		if prj.Name == project {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("starter project %q does not exist in devfile", project)
	}
	downloadSpinner := log.Spinnerf("Downloading starter project %q", prj.Name)
	err = o.registryClient.DownloadStarterProject(&prj, "", dest, false)
	if err != nil {
		downloadSpinner.End(false)
		return err
	}
	downloadSpinner.End(true)
	return nil
}

// NewCmdInit implements the odo command
func NewCmdInit(name, fullName string) *cobra.Command {
	backends := []params.ParamsBuilder{
		&params.FlagsBuilder{},
		params.NewInteractiveBuilder(asker.NewSurveyAsker(), catalog.NewCatalogClient(filesystem.DefaultFs{})),
	}
	prefClient, err := preference.NewClient()
	if err != nil {
		odoutil.LogErrorAndExit(err, "unable to set preference, something is wrong with odo, kindly raise an issue at https://github.com/redhat-developer/odo/issues/new?template=Bug.md")
	}

	o := NewInitOptions(backends, filesystem.DefaultFs{}, prefClient, registry.NewRegistryClient())
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

	initCmd.Flags().StringVar(&o.Name, params.FLAG_NAME, "", "name of the component to create")
	initCmd.Flags().StringVar(&o.Devfile, params.FLAG_DEVFILE, "", "name of the devfile in devfile registry")
	initCmd.Flags().StringVar(&o.DevfileRegistry, params.FLAG_DEVFILE_REGISTRY, "", "name of the devfile registry (as configured in \"odo registry list\"). It can be used in combination with --devfile, but not with --devfile-path")
	initCmd.Flags().StringVar(&o.Starter, params.FLAG_STARTER, "", "name of the starter project")
	initCmd.Flags().StringVar(&o.DevfilePath, params.FLAG_DEVFILE_PATH, "", "path to a devfile. This is an alternative to using devfile from Devfile registry. It can be local filesystem path or http(s) URL")

	// Add a defined annotation in order to appear in the help menu
	initCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	return initCmd
}
