package component

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/redhat-developer/odo/pkg/segment"

	"github.com/redhat-developer/odo/pkg/odo/cli/component/ui"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"

	registryLibrary "github.com/devfile/registry-support/registry-library/library"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/log"
	registryUtil "github.com/redhat-developer/odo/pkg/odo/cli/registry/util"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/zalando/go-keyring"
	"k8s.io/klog"
)

type CreateMethod interface {
	// CheckConflicts checks for conflicts specific to a create method
	CheckConflicts(co *CreateOptions, args []string) error
	// FetchDevfileAndCreateComponent fetches devfile from registry, or a remote location, or a local file system, and creates a component
	// This method also updates the CreateOptions structure with co.devfileMetadata
	FetchDevfileAndCreateComponent(co *CreateOptions, cmdline cmdline.Cmdline, args []string) error
	// Rollback cleans the component context of any files that were created by odo (devfile.yaml, .odo e.g.)
	Rollback(devfile, componentContext string)
}

// InteractiveCreateMethod is used while creating a component interactively
type InteractiveCreateMethod struct{}

func (icm InteractiveCreateMethod) CheckConflicts(co *CreateOptions, args []string) error {
	return nil
}

func (icm InteractiveCreateMethod) FetchDevfileAndCreateComponent(co *CreateOptions, cmdline cmdline.Cmdline, args []string) error {
	catalogDevfileList, err := validateAndFetchRegistry(co.devfileMetadata.devfileRegistry.Name)
	if err != nil {
		return err
	}

	//SET METADATA
	// Component type: We provide devfile component list to let user choose
	componentType := ui.SelectDevfileComponentType(catalogDevfileList.Items)

	// Component name: User needs to specify the component name, by default it is component type that user chooses
	componentName := ui.EnterDevfileComponentName(componentType)

	// Component namespace: User needs to specify component namespace, by default it is the current active namespace
	var componentNamespace string
	if cmdline.IsFlagSet("project") {
		componentNamespace, err = cmdline.FlagValue("project")
		if err != nil {
			return err
		}
	} else {
		// if the user is logged in or if we have cluster information, display the default project
		componentNamespace = ui.EnterDevfileComponentProject(co.KClient.GetCurrentProjectName())
	}

	co.devfileMetadata.componentType = componentType
	co.devfileName = componentType
	co.devfileMetadata.componentName = componentName
	co.devfileMetadata.componentNamespace = componentNamespace

	co.devfileMetadata.devfileLink, co.devfileMetadata.devfileRegistry, err = findDevfileFromRegistry(catalogDevfileList, co.devfileMetadata.devfileRegistry.Name, co.devfileMetadata.componentType)
	if err != nil {
		return err
	}
	return fetchDevfileFromRegistry(co.devfileMetadata.devfileRegistry, co.devfileMetadata.devfileLink, co.DevfilePath, co.devfileMetadata.componentType, co.contextFlag)
}

func (icm InteractiveCreateMethod) Rollback(devfile, componentContext string) {
	deleteDevfile(devfile)
	deleteOdoDir(componentContext)
}

// DirectCreateMethod is used with the basic odo create; `odo create nodejs mynode`
type DirectCreateMethod struct{}

func (dcm DirectCreateMethod) CheckConflicts(co *CreateOptions, args []string) error {
	return nil
}

func (dcm DirectCreateMethod) FetchDevfileAndCreateComponent(co *CreateOptions, cmdline cmdline.Cmdline, args []string) error {
	catalogDevfileList, err := validateAndFetchRegistry(co.devfileMetadata.devfileRegistry.Name)
	if err != nil {
		return err
	}
	// SET METADATA
	// The first argument passed will always be considered as component type
	co.devfileMetadata.componentType = args[0]
	co.devfileName = args[0]

	var componentName string
	if len(args) == 2 {
		// If more than one argument is passed, then the second one will be considered as component name
		componentName = args[1]
	} else {
		// If only component type is passed, then the component name will be created by odo
		componentName, err = createDefaultComponentName(
			co.devfileMetadata.componentType,
			co.contextFlag,
		)
		if err != nil {
			return err
		}
	}
	co.devfileMetadata.componentName = componentName
	co.devfileMetadata.devfileLink, co.devfileMetadata.devfileRegistry, err = findDevfileFromRegistry(catalogDevfileList, co.devfileMetadata.devfileRegistry.Name, co.devfileMetadata.componentType)
	if err != nil {
		return err
	}
	return fetchDevfileFromRegistry(co.devfileMetadata.devfileRegistry, co.devfileMetadata.devfileLink, co.DevfilePath, co.devfileMetadata.componentType, co.contextFlag)
}

func (dcm DirectCreateMethod) Rollback(devfile, componentContext string) {
	deleteDevfile(devfile)
	deleteOdoDir(componentContext)
}

// UserCreatedDevfileMethod is used when a devfile is present in the context directory
type UserCreatedDevfileMethod struct{}

func (ucdm UserCreatedDevfileMethod) CheckConflicts(co *CreateOptions, args []string) error {
	// More than one arguments should not be allowed when a devfile exists
	if len(args) > 1 {
		return &DevfileExistsExtraArgsError{len(args)}
	}
	//Check if the directory already contains a devfile when --devfile flag is passed
	if co.devfileMetadata.devfilePath.value != "" && !util.PathEqual(co.DevfilePath, co.devfileMetadata.devfilePath.value) {
		return &DevfileExistsDevfileFlagError{}
	}
	return nil
}

func (ucdm UserCreatedDevfileMethod) FetchDevfileAndCreateComponent(co *CreateOptions, cmdline cmdline.Cmdline, args []string) error {

	//	Existing devfile Mode; co.devfileName = ""
	devfileAbsolutePath, err := filepath.Abs(co.DevfilePath)
	if err != nil {
		return err
	}
	devfileSpinner := log.Spinnerf("odo will create a devfile component from the existing devfile: %s", devfileAbsolutePath)
	defer devfileSpinner.End(true)
	co.devfileMetadata.componentName, co.devfileMetadata.componentType, err = getMetadataForExistingDevfile(co, args)
	return err
}

func (ucdm UserCreatedDevfileMethod) Rollback(devfile, componentContext string) {
	deleteOdoDir(componentContext)
}

// HTTPCreateMethod is used when --devfile flag is used with a remote file; `odo create --devfile https://example.com/devfile.yaml`
type HTTPCreateMethod struct{}

func (hcm HTTPCreateMethod) CheckConflicts(co *CreateOptions, args []string) error {
	return conflictCheckForDevfileFlag(args, co.devfileMetadata.devfileRegistry.Name)
}

func (hcm HTTPCreateMethod) FetchDevfileAndCreateComponent(co *CreateOptions, cmdline cmdline.Cmdline, args []string) error {
	devfileSpinner := log.Spinnerf("Creating a devfile component from devfile path: %s", co.devfileMetadata.devfilePath.value)
	defer devfileSpinner.End(false)

	params := util.HTTPRequestParams{
		URL:   co.devfileMetadata.devfilePath.value,
		Token: co.devfileMetadata.token,
	}
	devfileData, err := util.DownloadFileInMemory(params)
	if err != nil {
		return errors.Wrapf(err, "failed to download devfile for devfile component from %s", co.devfileMetadata.devfilePath.value)
	}
	err = ioutil.WriteFile(co.DevfilePath, devfileData, 0644) // #nosec G306
	if err != nil {
		return errors.Wrapf(err, "unable to save devfile to %s", co.DevfilePath)
	}
	devfileSpinner.End(true)

	// SET METADATA
	co.devfileMetadata.devfilePath.protocol = "http(s)"
	co.devfileMetadata.componentName, co.devfileMetadata.componentType, err = getMetadataForExistingDevfile(co, args)
	return err
}

func (hcm HTTPCreateMethod) Rollback(devfile, componentContext string) {
	deleteDevfile(devfile)
	deleteOdoDir(componentContext)
}

// FileCreateMethod is used when --devfile flag is used with a local file; `odo create --devfile /tmp/comp/devfile.yaml`
type FileCreateMethod struct{}

func (fcm FileCreateMethod) CheckConflicts(co *CreateOptions, args []string) error {
	return conflictCheckForDevfileFlag(args, co.devfileMetadata.devfileRegistry.Name)
}

func (fcm FileCreateMethod) FetchDevfileAndCreateComponent(co *CreateOptions, cmdline cmdline.Cmdline, args []string) error {
	devfileAbsolutePath, err := filepath.Abs(co.devfileMetadata.devfilePath.value)
	if err != nil {
		return err
	}
	devfileSpinner := log.Spinnerf("Creating a devfile component from devfile path: %s", devfileAbsolutePath)
	defer devfileSpinner.End(false)
	devfileData, err := ioutil.ReadFile(co.devfileMetadata.devfilePath.value)
	if err != nil {
		return errors.Wrapf(err, "failed to read devfile from %s", co.devfileMetadata.devfilePath)
	}
	err = ioutil.WriteFile(co.DevfilePath, devfileData, 0644) // #nosec G306
	if err != nil {
		return errors.Wrapf(err, "unable to save devfile to %s", co.DevfilePath)
	}
	devfileSpinner.End(true)

	//	SET METADATA
	co.devfileMetadata.devfilePath.protocol = "file"
	co.devfileMetadata.componentName, co.devfileMetadata.componentType, err = getMetadataForExistingDevfile(co, args)
	return err
}

func (fcm FileCreateMethod) Rollback(devfile, componentContext string) {
	deleteDevfile(devfile)
	deleteOdoDir(componentContext)
}

// conflictCheckForDevfileFlag checks for the common conflicts while using --devfile flag
func conflictCheckForDevfileFlag(args []string, registryName string) error {
	// More than one arguments should not be allowed when --devfile is used
	if len(args) > 1 {
		return &DevfileExistsExtraArgsError{len(args)}
	}
	// Check if both --devfile and --registry flag are used, in which case raise an error
	if registryName != "" {
		return &DevfileFlagWithRegistryFlagError{}
	}
	return nil
}

// validateAndFetchRegistry validates if the provided registryName exists and returns the devfile listed in the registy;
// if the registryName is "", then it returns devfiles of all the available registries
func validateAndFetchRegistry(registryName string) (catalog.DevfileComponentTypeList, error) {
	// Validate if the component type is available
	if registryName != "" {
		registryExistSpinner := log.Spinnerf("Checking if the registry %q exists", registryName)
		defer registryExistSpinner.End(false)
		registryList, e := catalog.GetDevfileRegistries(registryName)
		if e != nil {
			return catalog.DevfileComponentTypeList{}, errors.Wrap(e, "failed to get registry")
		}
		if len(registryList) == 0 {
			return catalog.DevfileComponentTypeList{}, errors.Errorf("registry %s doesn't exist, please specify a valid registry via --registry", registryName)
		}
		registryExistSpinner.End(true)
	}

	klog.V(4).Infof("Fetching the available devfile components")
	// Get available devfile components for checking devfile compatibility
	catalogDevfileList, err := catalog.ListDevfileComponents(registryName)
	if err != nil {
		return catalog.DevfileComponentTypeList{}, err
	}

	if registryName != "" && catalogDevfileList.Items == nil {
		return catalog.DevfileComponentTypeList{}, errors.Errorf("can't create devfile component from registry %s", registryName)
	}

	if len(catalogDevfileList.DevfileRegistries) == 0 {
		return catalog.DevfileComponentTypeList{}, errors.New("Registry is empty, please run `odo registry add <registry name> <registry URL>` to add a registry\n")
	}
	return catalogDevfileList, nil
}

// findDevfileFromRegistry finds the devfile and returns necessary information related to it
func findDevfileFromRegistry(catalogDevfileList catalog.DevfileComponentTypeList, registryName, componentType string) (devfileLink string, devfileRegistry catalog.Registry, err error) {
	devfileExistSpinner := log.Spinnerf("Checking if the devfile for %q exists on available registries", componentType)
	defer devfileExistSpinner.End(false)
	if registryName != "" {
		devfileExistSpinner = log.Spinnerf("Checking if the devfile for %q exists on registry %q", componentType, registryName)
	}

	// Find the request devfile from the registry
	for _, devfileComponent := range catalogDevfileList.Items {
		if componentType == devfileComponent.Name {
			devfileExistSpinner.End(true)
			return devfileComponent.Link, devfileComponent.Registry, nil
		}
	}
	return "", catalog.Registry{}, fmt.Errorf("devfile component type %q is not supported, please run `odo catalog list components` for a list of supported devfile component types", componentType)
}

// fetchDevfileFromRegistry fetches the required devfile from the list catalogDevfileList
func fetchDevfileFromRegistry(registry catalog.Registry, devfileLink, devfilePath, componentType, componentContext string) (err error) {
	// Download devfile from registry
	registrySpinner := log.Spinnerf("Creating a devfile component from registry %q", registry.Name)
	defer registrySpinner.End(false)

	// For GitHub based registries
	if registryUtil.IsGitBasedRegistry(registry.URL) {
		registryUtil.PrintGitRegistryDeprecationWarning()

		params := util.HTTPRequestParams{
			URL: registry.URL + devfileLink,
		}

		secure, err := registryUtil.IsSecure(registry.Name)
		if err != nil {
			return err
		}

		if secure {
			var token string
			token, err = keyring.Get(fmt.Sprintf("%s%s", util.CredentialPrefix, registry.Name), registryUtil.RegistryUser)
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

		err = ioutil.WriteFile(devfilePath, devfileData, 0644) // #nosec G306
		if err != nil {
			return err
		}
	} else {
		err := registryLibrary.PullStackFromRegistry(registry.URL, componentType, componentContext, segment.GetRegistryOptions())
		if err != nil {
			return err
		}
	}
	registrySpinner.End(true)
	return nil
}

// getMetadataForExistingDevfile sets metadata for a user provided devfile; UserCreatedDevfileCreateMethod, HTTPCreateMethod, and FileCreateMethod
func getMetadataForExistingDevfile(co *CreateOptions, args []string) (componentName, componentType string, err error) {
	devObj, err := devfileParseFromFile(co.DevfilePath, false)
	if err != nil {
		return "", "", err
	}
	componentType = component.GetComponentTypeFromDevfileMetadata(devObj.Data.GetMetadata())

	// Set component name
	if len(args) > 0 {
		// user provided name: `odo create mynode`
		componentName = args[0]
	} else {
		if devObj.GetMetadataName() != "" {
			// devfile provided name: `.metadata.name`
			componentName = devObj.GetMetadataName()
		} else {
			// default name
			componentName, err = createDefaultComponentName(co.devfileMetadata.componentType, co.contextFlag)
			if err != nil {
				return "", "", err
			}
		}
	}

	return
}

// createDefaultComponentName creates a default unique component name with the help of component context
func createDefaultComponentName(componentType string, sourcePath string) (string, error) {
	var finalSourcePath string
	var err error
	if sourcePath != "" {
		finalSourcePath, err = filepath.Abs(sourcePath)
	} else {
		finalSourcePath, err = os.Getwd()
	}
	if err != nil {
		return "", err
	}

	return component.GetDefaultComponentName(
		finalSourcePath,
		componentType,
		component.ComponentList{},
	)
}

// deleteDevfile deletes the devfile if it exists in case of rollback
func deleteDevfile(devfile string) {
	if util.CheckPathExists(devfile) {
		_ = os.Remove(devfile)
	}
}

//deleteOdoDir deletes the .odo directory in case of rollback
func deleteOdoDir(componentContext string) {
	odoDir := filepath.Join(componentContext, ".odo")
	if util.CheckPathExists(odoDir) {
		_ = util.DeletePath(odoDir)
	}
}
