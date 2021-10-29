package component

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	registryLibrary "github.com/devfile/registry-support/registry-library/library"
	registryConsts "github.com/openshift/odo/pkg/odo/cli/registry/consts"
	"github.com/openshift/odo/pkg/preference"

	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/odo/cli/component/ui"

	registryUtil "github.com/openshift/odo/pkg/odo/cli/registry/util"
	"github.com/zalando/go-keyring"

	odoDevfile "github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/envinfo"

	"github.com/devfile/library/pkg/devfile"

	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/openshift/odo/pkg/odo/genericclioptions"

	"github.com/openshift/odo/pkg/devfile/location"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type CreateMethod interface {
	FetchDevfile() (bool, error)
	Rollback() error
}

func getContext(now bool, cmd *cobra.Command) (*genericclioptions.Context, error) {
	if now {
		return genericclioptions.NewContextCreatingAppIfNeeded(cmd)
	}
	return genericclioptions.NewOfflineContext(cmd)
}

// DevfileParseFromFile reads, parses and validates a devfile from a file without flattening it
func devfileParseFromFile(devfilePath string, resolved bool) (parser.DevfileObj, error) {
	devObj, _, err := devfile.ParseDevfileAndValidate(parser.ParserArgs{Path: devfilePath, FlattenedDevfile: &resolved})
	if err != nil {
		return parser.DevfileObj{}, err
	}

	return devObj, nil
}
func (co *CreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	// GETTERS
	// Get context
	co.Context, err = getContext(co.now, cmd)
	if err != nil {
		return err
	}
	// Get the app name
	co.appName = genericclioptions.ResolveAppFlag(cmd)
	// Get the project name
	co.devfileMetadata.componentNamespace = co.Context.Project
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

	// Initialize envinfo
	err = co.InitEnvInfoFromContext()
	if err != nil {
		return err
	}

	// Set the starter project token if required
	if co.devfileMetadata.starter != "" {
		secure, err := registryUtil.IsSecure(co.devfileMetadata.devfileRegistry.Name)
		if err != nil {
			return err
		}
		if co.devfileMetadata.starterToken == "" && secure {
			var token string
			token, err = keyring.Get(fmt.Sprintf("%s%s", util.CredentialPrefix, co.devfileMetadata.devfileRegistry.Name), registryUtil.RegistryUser)
			if err != nil {
				return errors.Wrap(err, "unable to get secure registry credential from keyring")
			}
			co.devfileMetadata.starterToken = token
		}
	}

	log.Info("Devfile Object Creation")
	//--------------------------------------------------------------------------------------------------------------------------------------------
	//	Existing devfile Mode; co.devfileName = ""
	if co.devfileMetadata.userCreatedDevfile {
		devfileAbsolutePath, err := filepath.Abs(co.DevfilePath)
		if err != nil {
			return err
		}
		devfileSpinner := log.Spinnerf("Creating a devfile component from devfile path %s", devfileAbsolutePath)
		defer devfileSpinner.End(true)

		// Set the component name
		co.devfileMetadata.componentName, err = getComponentNameForExistingDevfile(co.DevfilePath, args)
		if err != nil {
			return err
		}

		// Set the component type
		co.devfileMetadata.componentType, err = getComponentType(co.DevfilePath)
		if err != nil {
			return err
		}
		return nil
	}
	//--------------------------------------------------------------------------------------------------------------------------------------------
	//--devfile Mode; co.devfileName = ""
	if co.devfileMetadata.devfilePath.value != "" {
		fileErr := util.ValidateFile(co.devfileMetadata.devfilePath.value)
		urlErr := util.ValidateURL(co.devfileMetadata.devfilePath.value)
		if fileErr != nil && urlErr != nil {
			return errors.Errorf("the devfile path you specify is invalid with either file error %q or url error %q", fileErr, urlErr)
		} else if fileErr == nil {
			co.devfileMetadata.devfilePath.protocol = "file"
		} else if urlErr == nil {
			co.devfileMetadata.devfilePath.protocol = "http(s)"
		}

		var devfileAbsolutePath string
		var devfileData []byte
		if co.devfileMetadata.devfilePath.protocol == "file" {
			devfileAbsolutePath, err = filepath.Abs(co.devfileMetadata.devfilePath.value)
			if err != nil {
				return err
			}

			devfileData, err = ioutil.ReadFile(co.devfileMetadata.devfilePath.value)
			if err != nil {
				return errors.Wrapf(err, "failed to read devfile from %s", co.devfileMetadata.devfilePath)
			}

		} else if co.devfileMetadata.devfilePath.protocol == "http(s)" {
			devfileAbsolutePath = co.devfileMetadata.devfilePath.value
			params := util.HTTPRequestParams{
				URL:   co.devfileMetadata.devfilePath.value,
				Token: co.devfileMetadata.token,
			}
			devfileData, err = util.DownloadFileInMemory(params)
			if err != nil {
				return errors.Wrapf(err, "failed to download devfile for devfile component from %s", co.devfileMetadata.devfilePath.value)
			}
		}
		err = ioutil.WriteFile(co.DevfilePath, devfileData, 0644) // #nosec G306
		if err != nil {
			return errors.Wrapf(err, "unable to save devfile to %s", co.DevfilePath)
		}
		devfileSpinner := log.Spinnerf("Creating a devfile component from devfile path: %s", devfileAbsolutePath)
		defer devfileSpinner.End(true)

		// get the custom component name
		co.devfileMetadata.componentName, err = getComponentNameForExistingDevfile(co.DevfilePath, args)
		if err != nil {
			return err
		}
		co.devfileMetadata.componentType, err = getComponentType(co.DevfilePath)
		if err != nil {
			return err
		}
		return nil
	}

	//Interactive Mode
	if co.interactive {
		// Get available devfile components for checking devfile compatibility
		catalogDevfileList, err := catalog.ListDevfileComponents(co.devfileMetadata.devfileRegistry.Name)
		if err != nil {
			return err
		}

		if co.devfileMetadata.devfileRegistry.Name != "" && catalogDevfileList.Items == nil {
			return errors.Errorf("can't create devfile component from registry %s", co.devfileMetadata.devfileRegistry.Name)
		}

		if len(catalogDevfileList.DevfileRegistries) == 0 {
			return errors.New("Registry is empty, please run `odo registry add <registry name> <registry URL>` to add a registry\n")
		}

		// Component type: We provide devfile component list to let user choose
		componentType := ui.SelectDevfileComponentType(catalogDevfileList.Items)

		// Component name: User needs to specify the component name, by default it is component type that user chooses
		componentName := ui.EnterDevfileComponentName(componentType)

		// Component namespace: User needs to specify component namespace, by default it is the current active namespace
		var componentNamespace string
		if cmd.Flags().Changed("project") {
			componentNamespace, err = cmd.Flags().GetString("project")
			if err != nil {
				return err
			}
		} else {
			client, e := genericclioptions.Client()
			// if the user is logged in or if we have cluster information, display the default project
			if e == nil {
				componentNamespace = ui.EnterDevfileComponentProject(client.GetCurrentProjectName())
			}
		}

		co.devfileMetadata.componentType = componentType
		co.devfileName = componentType
		co.devfileMetadata.componentName = componentName
		co.devfileMetadata.componentNamespace = componentNamespace

		// Validate if the component type is available
		hasComponent := false
		var devfileExistSpinner *log.Status
		log.Info("Devfile Object Validation")
		devfileExistSpinner = log.Spinner("Checking if the devfile type exists")
		defer devfileExistSpinner.End(false)

		for _, devfileComponent := range catalogDevfileList.Items {
			if co.devfileMetadata.componentType == devfileComponent.Name {
				hasComponent = true
				co.devfileMetadata.devfileLink = devfileComponent.Link
				co.devfileMetadata.devfileRegistry = devfileComponent.Registry
				break
			}
		}
		if hasComponent {
			devfileExistSpinner.End(true)
		} else {
			devfileExistSpinner.End(false)
			return fmt.Errorf("devfile component type %q is not supported, please run `odo catalog list components` for a list of supported devfile component types", co.devfileMetadata.componentType)

		}

		registrySpinner := log.Spinnerf("Creating a devfile component from registry: %s", co.devfileMetadata.devfileRegistry.Name)
		if registryUtil.IsGitBasedRegistry(co.devfileMetadata.devfileRegistry.URL) {
			registryUtil.PrintGitRegistryDeprecationWarning()
		}

		// Download devfile from registry
		var params util.HTTPRequestParams
		// For GitHub based registries
		if strings.Contains(co.devfileMetadata.devfileRegistry.URL, "github") {
			params = util.HTTPRequestParams{
				URL: co.devfileMetadata.devfileRegistry.URL + co.devfileMetadata.devfileLink,
			}

			secure, e := registryUtil.IsSecure(co.devfileMetadata.devfileRegistry.Name)
			if e != nil {
				return e
			}

			if secure {
				var token string
				token, err = keyring.Get(fmt.Sprintf("%s%s", util.CredentialPrefix, co.devfileMetadata.devfileRegistry.Name), registryUtil.RegistryUser)
				if err != nil {
					return errors.Wrap(err, "unable to get secure registry credential from keyring")
				}
				params.Token = token
			}

			cfg, err := preference.New()
			if err != nil {
				return err
			}
			devfileData, err := util.DownloadFileInMemoryWithCache(params, cfg.GetRegistryCacheTime())
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(co.DevfilePath, devfileData, 0644)
			if err != nil {
				return err
			}
		} else {
			err = registryLibrary.PullStackFromRegistry(co.devfileMetadata.devfileRegistry.URL, co.devfileMetadata.componentType, co.componentContext, false, registryConsts.TelemetryClient)
			if err != nil {
				return err
			}
			// if the function fails, remove this newly created devfile
			defer func() {
				if err != nil {
					os.Remove(co.DevfilePath)
				}
			}()
		}
		registrySpinner.End(true)
		return nil
	}
	//Normal Mode
	// Component type: Get from full command's first argument (mandatory in direct mode)
	componentType := args[0]

	co.devfileMetadata.componentType = componentType
	co.devfileName = componentType

	// Get available devfile components for checking devfile compatibility
	catalogDevfileList, err := catalog.ListDevfileComponents(co.devfileMetadata.devfileRegistry.Name)
	if err != nil {
		return err
	}

	if co.devfileMetadata.devfileRegistry.Name != "" && catalogDevfileList.Items == nil {
		return errors.Errorf("can't create devfile component from registry %s", co.devfileMetadata.devfileRegistry.Name)
	}

	if len(catalogDevfileList.DevfileRegistries) == 0 {
		return errors.New("Registry is empty, please run `odo registry add <registry name> <registry URL>` to add a registry\n")
	}
	// Validate if the component type is available
	hasComponent := false
	var devfileExistSpinner *log.Status
	log.Info("Devfile Object Validation")
	devfileExistSpinner = log.Spinner("Checking if the devfile type exists")
	defer devfileExistSpinner.End(false)

	for _, devfileComponent := range catalogDevfileList.Items {
		if co.devfileMetadata.componentType == devfileComponent.Name {
			hasComponent = true
			co.devfileMetadata.devfileLink = devfileComponent.Link
			co.devfileMetadata.devfileRegistry = devfileComponent.Registry
			break
		}
	}
	if hasComponent {
		devfileExistSpinner.End(true)
	} else {
		devfileExistSpinner.End(false)
		return fmt.Errorf("devfile component type %q is not supported, please run `odo catalog list components` for a list of supported devfile component types", co.devfileMetadata.componentType)

	}

	registrySpinner := log.Spinnerf("Creating a devfile component from registry: %s", co.devfileMetadata.devfileRegistry.Name)
	if registryUtil.IsGitBasedRegistry(co.devfileMetadata.devfileRegistry.URL) {
		registryUtil.PrintGitRegistryDeprecationWarning()
	}

	// Download devfile from registry
	var params util.HTTPRequestParams
	// For GitHub based registries
	if strings.Contains(co.devfileMetadata.devfileRegistry.URL, "github") {
		params = util.HTTPRequestParams{
			URL: co.devfileMetadata.devfileRegistry.URL + co.devfileMetadata.devfileLink,
		}

		secure, e := registryUtil.IsSecure(co.devfileMetadata.devfileRegistry.Name)
		if e != nil {
			return e
		}

		if secure {
			var token string
			token, err = keyring.Get(fmt.Sprintf("%s%s", util.CredentialPrefix, co.devfileMetadata.devfileRegistry.Name), registryUtil.RegistryUser)
			if err != nil {
				return errors.Wrap(err, "unable to get secure registry credential from keyring")
			}
			params.Token = token
		}

		cfg, err := preference.New()
		if err != nil {
			return err
		}
		devfileData, err := util.DownloadFileInMemoryWithCache(params, cfg.GetRegistryCacheTime())
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(co.DevfilePath, devfileData, 0644)
		if err != nil {
			return err
		}
	} else {
		err = registryLibrary.PullStackFromRegistry(co.devfileMetadata.devfileRegistry.URL, co.devfileMetadata.componentType, co.componentContext, false, registryConsts.TelemetryClient)
		if err != nil {
			return err
		}
		// if the function fails, remove this newly created devfile
		defer func() {
			if err != nil {
				os.Remove(co.DevfilePath)
			}
		}()
	}
	registrySpinner.End(true)
	var componentName string
	if len(args) == 2 {
		componentName = args[1]
	} else {
		componentName, err = createDefaultComponentName(
			componentType,
			co.componentContext,
		)
		if err != nil {
			return err
		}
	}
	co.devfileMetadata.componentName = componentName

	return nil
}

func getComponentNameForExistingDevfile(devfilePath string, args []string) (string, error) {
	// get the custom component name
	if len(args) > 0 {
		return args[0], nil
	}
	devObj, err := devfileParseFromFile(devfilePath, false)
	if err != nil {
		currentDirPath, err := os.Getwd()
		if err != nil {
			return "", err
		}
		return filepath.Base(currentDirPath), nil
	}
	return devObj.GetMetadataName(), nil
}

func getComponentType(devfilePath string) (string, error) {
	devObj, err := devfileParseFromFile(devfilePath, false)
	if err != nil {
		return "", err
	}
	return component.GetComponentTypeFromDevfileMetadata(devObj.Data.GetMetadata()), nil
}

func (co *CreateOptions) Validate() (err error) {
	log.Info("Validation")
	// Validate if the devfile component name that user wants to create adheres to the k8s naming convention
	spinner := log.Spinner("Validating if devfile name is correct")
	defer spinner.End(false)

	err = util.ValidateK8sResourceName("component name", co.devfileMetadata.componentName)
	if err != nil {
		return err
	}
	spinner.End(true)

	// Validate if the devfile is compatible with odo; this checks the resolved/flattened version of devfile
	spinner = log.Spinner("Validating the devfile for odo")
	defer spinner.End(false)

	_, err = odoDevfile.ParseAndValidateFromFile(co.DevfilePath)
	if err != nil {
		return err
	}
	spinner.End(true)

	return nil
}

func (co *CreateOptions) Run(cmd *cobra.Command) (err error) {
	devObj, err := devfileParseFromFile(co.DevfilePath, false)
	if err != nil {
		return errors.New("Failed to parse the devfile")
	}

	devfileData, err := ioutil.ReadFile(co.DevfilePath)
	if err != nil {
		return err
	}
	// WARN: Starter Project uses go-git that overrides the directory content, there by deleting the existing devfile.
	err = decideAndDownloadStarterProject(devObj, co.devfileMetadata.starter, co.devfileMetadata.starterToken, co.interactive, co.componentContext)
	if err != nil {
		return errors.Wrap(err, "failed to download project for devfile component")
	}

	// TODO: We should not have to rewrite to the file. Fix the starter project.
	err = ioutil.WriteFile(co.DevfilePath, devfileData, 0644)
	if err != nil {
		return err
	}

	// If user provided a custom name, re-write the devfile
	// ENSURE: co.devfileMetadata.componentName != ""
	if co.devfileMetadata.componentName != devObj.GetMetadataName() {
		spinner := log.Spinnerf("Updating the devfile with component name: %v", co.devfileMetadata.componentName)
		defer spinner.End(false)

		err := devObj.SetMetadataName(co.devfileMetadata.componentName)
		if err != nil {
			return errors.New("Failed to update the devfile")
		}
		spinner.End(true)
	}

	// Generate env file
	err = co.EnvSpecificInfo.SetComponentSettings(envinfo.ComponentSettings{
		Name:               co.devfileMetadata.componentName,
		Project:            co.devfileMetadata.componentNamespace,
		AppName:            co.appName,
		UserCreatedDevfile: co.devfileMetadata.userCreatedDevfile,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create env file for devfile component")
	}

	sourcePath, err := util.GetAbsPath(co.componentContext)
	if err != nil {
		return errors.Wrap(err, "unable to get source path")
	}

	ignoreFile, err := util.TouchGitIgnoreFile(sourcePath)
	if err != nil {
		return err
	}

	err = util.AddFileToIgnoreFile(ignoreFile, filepath.Join(co.componentContext, EnvDirectory))
	if err != nil {
		return err
	}

	if co.now {
		err = co.DevfilePush()
		if err != nil {
			return fmt.Errorf("failed to push changes: %w", err)
		}
	} else {
		log.Italic("\nPlease use `odo push` command to create the component with source deployed")
	}
	return nil
}
