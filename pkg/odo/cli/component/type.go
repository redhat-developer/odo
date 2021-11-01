package component

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	registryLibrary "github.com/devfile/registry-support/registry-library/library"
	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/cli/component/ui"
	registryConsts "github.com/openshift/odo/pkg/odo/cli/registry/consts"
	registryUtil "github.com/openshift/odo/pkg/odo/cli/registry/util"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/preference"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	"k8s.io/klog"
)

// Get Metadata
// If the devfile does not exist, then check and Validate Registry
// Fetch Devfile
// If devfile was manually created, rollback

type CreateMethod interface {
	// FetchDevfileAndCreateComponent fetches devfile from registry, or a remote location, or a local file system, and create a component
	FetchDevfileAndCreateComponent(co *CreateOptions, catalogDevfileList catalog.DevfileComponentTypeList) error
	// SetMetadata sets the necessary metadata, mainly component name and component type
	SetMetadata(co *CreateOptions, cmd *cobra.Command, args []string, catalogDevfileList catalog.DevfileComponentTypeList) error
}

// UserCreateDevfileMethod is used when a devfile is present in the context directory
type UserCreatedDevfileMethod struct{}

// InteractiveCreateMethod is used while creating a component interactively
type InteractiveCreateMethod struct{}

// DirectCreateMethod is used with the basic odo create; `odo create nodejs mynode`
type DirectCreateMethod struct{}

// HTTPCreateMethod is used when --devfile flag is used with a remote file; `odo create --devfile https://example.com/devfile.yaml`
type HTTPCreateMethod struct{}

// FileCreateMethod is used when --devfile flag is used with a local file; `odo create --devfile /tmp/comp/devfile.yaml`
type FileCreateMethod struct{}

func (ucdm UserCreatedDevfileMethod) FetchDevfileAndCreateComponent(co *CreateOptions, catalogDevfileList catalog.DevfileComponentTypeList) error {
	//	Existing devfile Mode; co.devfileName = ""
	devfileAbsolutePath, err := filepath.Abs(co.DevfilePath)
	if err != nil {
		return err
	}
	devfileSpinner := log.Spinnerf("Creating a devfile component from devfile path %s", devfileAbsolutePath)
	defer devfileSpinner.End(true)

	return nil
}
func (ucdm UserCreatedDevfileMethod) SetMetadata(co *CreateOptions, cmd *cobra.Command, args []string, catalogDevfileList catalog.DevfileComponentTypeList) error {
	return setMetadataForExistingDevfile(co, args)
}

func (icm InteractiveCreateMethod) FetchDevfileAndCreateComponent(co *CreateOptions, catalogDevfileList catalog.DevfileComponentTypeList) error {
	return fetchDevfile(co, catalogDevfileList)
}
func (icm InteractiveCreateMethod) SetMetadata(co *CreateOptions, cmd *cobra.Command, args []string, catalogDevfileList catalog.DevfileComponentTypeList) error {
	var err error
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

	return err
}

func (dcm DirectCreateMethod) FetchDevfileAndCreateComponent(co *CreateOptions, catalogDevfileList catalog.DevfileComponentTypeList) error {
	return fetchDevfile(co, catalogDevfileList)
}
func (dcm DirectCreateMethod) SetMetadata(co *CreateOptions, cmd *cobra.Command, args []string, catalogDevfileList catalog.DevfileComponentTypeList) error {
	var err error
	componentType := args[0]

	co.devfileMetadata.componentType = componentType
	co.devfileName = componentType
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
	return err
}

func (hcm HTTPCreateMethod) FetchDevfileAndCreateComponent(co *CreateOptions, catalogDevfileList catalog.DevfileComponentTypeList) error {
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
	return nil
}
func (hcm HTTPCreateMethod) SetMetadata(co *CreateOptions, cmd *cobra.Command, args []string, catalogDevfileList catalog.DevfileComponentTypeList) error {
	return setMetadataForExistingDevfile(co, args)
}

func (fcm FileCreateMethod) FetchDevfileAndCreateComponent(co *CreateOptions, catalogDevfileList catalog.DevfileComponentTypeList) error {
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
	return nil
}
func (fcm FileCreateMethod) SetMetadata(co *CreateOptions, cmd *cobra.Command, args []string, catalogDevfileList catalog.DevfileComponentTypeList) error {
	return setMetadataForExistingDevfile(co, args)
}

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

func fetchDevfile(co *CreateOptions, catalogDevfileList catalog.DevfileComponentTypeList) error {
	hasComponent := false
	var devfileExistSpinner *log.Status
	if co.devfileMetadata.devfileRegistry.Name != "" {
		devfileExistSpinner = log.Spinnerf("Checking if the devfile for %q exists on registry %q", co.devfileMetadata.componentType, co.devfileMetadata.devfileRegistry.Name)
	} else {
		devfileExistSpinner = log.Spinnerf("Checking if the devfile for %q exists on available registries", co.devfileMetadata.componentType)
	}
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

	registrySpinner := log.Spinnerf("Creating a devfile component from registry %q", co.devfileMetadata.devfileRegistry.Name)
	defer registrySpinner.End(false)
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
			token, err := keyring.Get(fmt.Sprintf("%s%s", util.CredentialPrefix, co.devfileMetadata.devfileRegistry.Name), registryUtil.RegistryUser)
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

		err = ioutil.WriteFile(co.DevfilePath, devfileData, 0644) // #nosec G306
		if err != nil {
			return err
		}
	} else {
		err := registryLibrary.PullStackFromRegistry(co.devfileMetadata.devfileRegistry.URL, co.devfileMetadata.componentType, co.componentContext, false, registryConsts.TelemetryClient)
		if err != nil {
			return err
		}
	}
	registrySpinner.End(true)
	return nil
}

func setMetadataForExistingDevfile(co *CreateOptions, args []string) error {
	var err error
	var componentName string
	devObj, err := devfileParseFromFile(co.DevfilePath, false)
	if err != nil {
		return err
	}
	co.devfileMetadata.componentType = component.GetComponentTypeFromDevfileMetadata(devObj.Data.GetMetadata())

	// Set component name
	if len(args) > 0 {
		// user provided name: `odo create mynode`
		componentName = args[1]
	} else {
		if devObj.GetMetadataName() != "" {
			// devfile provided name: `.metadata.name`
			componentName = devObj.GetMetadataName()
		} else {
			// default name
			currentDirPath, err := os.Getwd()
			if err != nil {
				return err
			}
			componentName = filepath.Base(currentDirPath)
		}
	}
	co.devfileMetadata.componentName = componentName
	return nil
}
