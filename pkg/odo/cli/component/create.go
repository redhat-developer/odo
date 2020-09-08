package component

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	"k8s.io/klog"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/envinfo"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	catalogutil "github.com/openshift/odo/pkg/odo/cli/catalog/util"
	"github.com/openshift/odo/pkg/odo/cli/component/ui"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	registryUtil "github.com/openshift/odo/pkg/odo/cli/registry/util"
	commonui "github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/odo/util/pushtarget"
	"github.com/openshift/odo/pkg/preference"
	"github.com/openshift/odo/pkg/util"

	corev1 "k8s.io/api/core/v1"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// CreateOptions encapsulates create options
type CreateOptions struct {
	componentSettings config.ComponentSettings
	componentBinary   string
	componentGit      string
	componentGitRef   string
	componentContext  string
	componentPorts    []string
	componentEnvVars  []string
	appName           string
	interactive       bool
	now               bool
	forceS2i          bool
	*PushOptions
	devfileMetadata DevfileMetadata
}

// Path of user's own devfile, user specifies the path via --devfile flag
type devfilePath struct {
	protocol string
	value    string
}

// DevfileMetadata includes devfile component metadata
type DevfileMetadata struct {
	componentType      string
	componentName      string
	componentNamespace string
	devfileSupport     bool
	devfileLink        string
	devfileRegistry    catalog.Registry
	devfilePath        devfilePath
	starter            string
	token              string
}

// CreateRecommendedCommandName is the recommended watch command name
const CreateRecommendedCommandName = "create"

// LocalDirectoryDefaultLocation is the default location of where --local files should always be..
// since the application will always be in the same directory as `.odo`, we will always set this as: ./
const LocalDirectoryDefaultLocation = "./"

var (
	envFile    = filepath.Join(".odo", "env", "env.yaml")
	configFile = filepath.Join(".odo", "config.yaml")
	envDir     = filepath.Join(".odo", "env")
)

// EnvFilePath is the path of env file for devfile component
var EnvFilePath = filepath.Join(LocalDirectoryDefaultLocation, envFile)

// ConfigFilePath is the path of config.yaml for s2i component
var ConfigFilePath = filepath.Join(LocalDirectoryDefaultLocation, configFile)

var createLongDesc = ktemplates.LongDesc(`Create a configuration describing a component.`)

var createExample = ktemplates.Examples(`# Create a new Node.JS component with existing sourcecode as well as specifying a name
%[1]s nodejs mynodejs

# Name is not required and will be automatically generated if not passed
%[1]s nodejs

# List all available components before deploying
odo catalog list components
%[1]s java-quarkus

# Download an example devfile and application before deploying
%[1]s nodejs --starter

# Using a specific devfile
%[1]s mynodejs --devfile ./devfile.yaml
%[1]s mynodejs --devfile https://raw.githubusercontent.com/odo-devfiles/registry/master/devfiles/nodejs/devfile.yaml

# Create new Node.js component named 'frontend' with the source in './frontend' directory
%[1]s nodejs frontend --context ./frontend

# Create new Java component with binary named sample.jar in './target' directory
%[1]s java:8  --binary target/sample.jar

# Create new Node.js component with source from remote git repository
%[1]s nodejs --git https://github.com/openshift/nodejs-ex.git

# Create new Node.js component with custom ports and environment variables
%[1]s nodejs --port 8080,8100/tcp,9100/udp --env key=value,key1=value1`)

const defaultStarterProjectName = "devfile-starter-project-name"

// NewCreateOptions returns new instance of CreateOptions
func NewCreateOptions() *CreateOptions {
	return &CreateOptions{
		PushOptions: NewPushOptions(),
	}
}

func (co *CreateOptions) setComponentSourceAttributes() (err error) {
	// Set the correct application context
	co.componentSettings.Application = &(co.Context.Application)

	// By default we set the source as LOCAL (if --local, --binary or --git isn't passed)
	componentSourceType := config.LOCAL

	// If --local, --binary or --git is passed, let's set the correct source type.
	if len(co.componentBinary) != 0 {
		componentSourceType = config.BINARY
	} else if len(co.componentGit) != 0 {
		componentSourceType = config.GIT
	}
	co.componentSettings.SourceType = &componentSourceType

	// Here we set the correct source path for each type
	switch componentSourceType {

	// --binary
	case config.BINARY:
		// Convert componentContext to absolute path, so it can be safely used in filepath.Rel
		// even when it is not set (empty). In this case filepath.Abs will return current directory.
		absContext, err := filepath.Abs(co.componentContext)
		if err != nil {
			return err
		}
		absPath, err := filepath.Abs(co.componentBinary)
		if err != nil {
			return err
		}
		// we need to store the SourceLocation relative to the componentContext
		relativePathToSource, err := filepath.Rel(absContext, absPath)
		if err != nil {
			return err
		}
		co.componentSettings.SourceLocation = &relativePathToSource

	// --git
	case config.GIT:
		co.componentSettings.SourceLocation = &(co.componentGit)
		componentSourceType = config.GIT
		co.componentSettings.SourceType = &componentSourceType

	// --local / default
	case config.LOCAL:

		directory := LocalDirectoryDefaultLocation
		co.componentSettings.SourceLocation = &directory

	// Error out by default if no type of sources were passed..
	default:
		return fmt.Errorf("the source can be either --binary or --local or --git")

	}

	// Set the Git reference if passed
	if len(co.componentGitRef) != 0 {
		co.componentSettings.Ref = &(co.componentGitRef)
	}

	// Error out if reference is passed but no --git flag passed
	if len(co.componentGit) == 0 && len(co.componentGitRef) != 0 {
		return fmt.Errorf("the --ref flag is only valid for --git flag")
	}

	return
}

func (co *CreateOptions) setComponentName(args []string) (err error) {
	componentImageName, componentType, _, _ := util.ParseComponentImageName(args[0])
	co.componentSettings.Type = &componentImageName

	if len(args) == 2 {
		co.componentSettings.Name = &args[1]
		return
	}

	if co.componentSettings.SourceType == nil {
		return errors.Wrap(err, "component type is mandatory parameter to generate a default component name")
	}

	componentName, err := createDefaultComponentName(
		co.Context,
		componentType,
		*(co.componentSettings.SourceType),
		co.componentContext,
	)
	if err != nil {
		return err
	}

	co.componentSettings.Name = &componentName
	return
}

func getSourceLocation(componentContext string, currentDirectory string) (string, error) {

	// After getting the path relative to the current directory, we set the SourceLocation
	sourceLocation, err := filepath.Rel(currentDirectory, componentContext)
	if err != nil {
		return "", errors.Wrapf(err, "unable to get a path relative to the current directory")
	}

	// If the paths are the same (currentDirectory vs co.componentSettings.SourceLocation)
	// then we use the default location
	if sourceLocation == "." {
		return LocalDirectoryDefaultLocation, nil
	}

	return sourceLocation, nil
}

func createDefaultComponentName(context *genericclioptions.Context, componentType string, sourceType config.SrcType, sourcePath string) (string, error) {
	// Retrieve the componentName, if the componentName isn't specified, we will use the default image name
	var err error
	finalSourcePath := sourcePath
	// we only get absolute path for local source type
	if sourceType == config.LOCAL {
		if sourcePath == "" {
			wd, err := os.Getwd()
			if err != nil {
				return "", err
			}
			finalSourcePath = wd
		} else {
			finalSourcePath, err = filepath.Abs(sourcePath)
			if err != nil {
				return "", err
			}
		}
	}

	componentName, err := component.GetDefaultComponentName(
		finalSourcePath,
		sourceType,
		componentType,
		component.ComponentList{},
	)

	if err != nil {
		return "", nil
	}

	return componentName, nil
}

func (co *CreateOptions) checkConflictingFlags() (err error) {
	if err = co.checkConflictingS2IFlags(); err != nil {
		return
	}

	if err = co.checkConflictingDevfileFlags(); err != nil {
		return
	}

	return nil
}

func (co *CreateOptions) checkConflictingS2IFlags() error {
	if !co.forceS2i {
		errorString := "flag --%s, requires --s2i flag to be set, when deploying S2I (Source-to-Image) components."

		var flagName string
		if len(co.componentBinary) != 0 {
			flagName = "binary"
		} else if len(co.componentGit) != 0 {
			flagName = "git"
		} else if len(co.componentEnvVars) != 0 {
			flagName = "env"
		} else if len(co.componentPorts) != 0 {
			flagName = "port"
		}

		if len(flagName) != 0 {
			return errors.New(fmt.Sprintf(errorString, flagName))
		}
	}
	return nil
}

func (co *CreateOptions) checkConflictingDevfileFlags() error {
	if co.forceS2i {
		errorString := "you can't set --s2i flag as true if you want to use the %s via --%s flag"

		var flagName string
		if len(co.devfileMetadata.devfilePath.value) != 0 {
			flagName = "devfile"
		} else if len(co.devfileMetadata.devfileRegistry.Name) != 0 {
			flagName = "registry"
		} else if len(co.devfileMetadata.token) != 0 {
			flagName = "token"
		} else if len(co.devfileMetadata.starter) != 0 {
			flagName = "starter"
		}

		if len(flagName) != 0 {
			return errors.New(fmt.Sprintf(errorString, flagName, flagName))
		}
	}
	return nil
}

// Complete completes create args
func (co *CreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	// this populates the LocalConfigInfo as well
	co.Context = genericclioptions.NewContextCreatingAppIfNeeded(cmd)

	// If not using --s2i
	if !co.forceS2i {

		err = co.checkConflictingFlags()
		if err != nil {
			return
		}

		// Configure the context
		if co.componentContext != "" {
			DevfilePath = filepath.Join(co.componentContext, devFile)
			log.Infof("Devfile path: %s", co.DevfilePath)
			EnvFilePath = filepath.Join(co.componentContext, envFile)
			ConfigFilePath = filepath.Join(co.componentContext, configFile)
			co.PushOptions.componentContext = co.componentContext
		}
		co.DevfilePath = DevfilePath

		if util.CheckPathExists(ConfigFilePath) || util.CheckPathExists(EnvFilePath) {
			return errors.New("this directory already contains a component")
		}

		if util.CheckPathExists(co.DevfilePath) && co.devfileMetadata.devfilePath.value != "" && !util.PathEqual(co.DevfilePath, co.devfileMetadata.devfilePath.value) {
			return errors.New("this directory already contains a devfile, you can't specify devfile via --devfile")
		}

		co.appName = genericclioptions.ResolveAppFlag(cmd)

		var catalogList catalog.ComponentTypeList
		if co.forceS2i {
			client := co.Client
			catalogList, err = catalog.ListComponents(client)
			if err != nil {
				return err
			}
		}

		// Validate user specify devfile path
		if co.devfileMetadata.devfilePath.value != "" {
			fileErr := util.ValidateFile(co.devfileMetadata.devfilePath.value)
			urlErr := util.ValidateURL(co.devfileMetadata.devfilePath.value)
			if fileErr != nil && urlErr != nil {
				return errors.Errorf("the devfile path you specify is invalid with either file error \"%v\" or url error \"%v\"", fileErr, urlErr)
			} else if fileErr == nil {
				co.devfileMetadata.devfilePath.protocol = "file"
			} else if urlErr == nil {
				co.devfileMetadata.devfilePath.protocol = "http(s)"
			}
		}

		// Validate user specify registry
		if co.devfileMetadata.devfileRegistry.Name != "" {

			if co.devfileMetadata.devfilePath.value != "" {
				return errors.New("you can't specify registry via --registry if you want to use the devfile that is specified via --devfile")
			}

			registryList, err := catalog.GetDevfileRegistries(co.devfileMetadata.devfileRegistry.Name)
			if err != nil {
				return errors.Wrap(err, "failed to get registry")
			}
			if len(registryList) == 0 {
				return errors.Errorf("registry %s doesn't exist, please specify a valid registry via --registry", co.devfileMetadata.devfileRegistry.Name)
			}
		}

		// Can't use the existing devfile or download devfile from registry, go to interactive mode
		if len(args) == 0 && !util.CheckPathExists(co.DevfilePath) && co.devfileMetadata.devfilePath.value == "" {
			co.interactive = true
		}

		// Configure the default namespace
		var defaultComponentNamespace string
		// If the push target is set to Docker, we can't assume we have an active Kube context
		if !pushtarget.IsPushTargetDocker() {
			// Get current active namespace
			client, err := kclient.New()
			if err != nil {
				return err
			}
			defaultComponentNamespace = client.Namespace
		}

		var componentType string
		var componentName string
		var componentNamespace string
		var catalogDevfileList catalog.DevfileComponentTypeList
		isDevfileRegistryPresent := true // defaulted to true since odo ships with a default registry set

		if co.interactive && !co.forceS2i {
			// Interactive mode

			// Get available devfile components for checking devfile compatibility
			catalogDevfileList, err = catalog.ListDevfileComponents(co.devfileMetadata.devfileRegistry.Name)
			if err != nil {
				return err
			}

			if len(catalogDevfileList.DevfileRegistries) == 0 {
				isDevfileRegistryPresent = false
				log.Warning("Registry is empty, please run `odo registry add <registry name> <registry URL>` to add a registry\n")
			}

			if isDevfileRegistryPresent {
				// Component type: We provide devfile component list to let user choose
				componentType = ui.SelectDevfileComponentType(catalogDevfileList.Items)

				// Component name: User needs to specify the componet name, by default it is component type that user chooses
				componentName = ui.EnterDevfileComponentName(componentType)

				// Component namespace: User needs to specify component namespace, by default it is the current active namespace
				if cmd.Flags().Changed("project") && !pushtarget.IsPushTargetDocker() {
					componentNamespace, err = cmd.Flags().GetString("project")
					if err != nil {
						return err
					}
				} else if !pushtarget.IsPushTargetDocker() {
					componentNamespace = ui.EnterDevfileComponentNamespace(defaultComponentNamespace)
				}
			}

		} else {
			// Direct mode (User enters the full command)

			if util.CheckPathExists(co.DevfilePath) || co.devfileMetadata.devfilePath.value != "" {
				// Use existing devfile directly

				if co.forceS2i {
					return errors.Errorf("existing devfile component detected. Please remove the devfile component before creating an s2i component")
				}

				if len(args) > 1 {
					return errors.Errorf("accepts between 0 and 1 arg when using existing devfile, received %d", len(args))
				}

				// If user can use existing devfile directly, the first arg is component name instead of component type
				if len(args) == 1 {
					componentName = args[0]
				} else {
					currentDirPath, err := os.Getwd()
					if err != nil {
						return err
					}
					currentDirName := filepath.Base(currentDirPath)
					componentName = currentDirName
				}

				co.devfileMetadata.devfileSupport = true
			} else if len(args) > 0 {
				// Download devfile from registry

				// Component type: Get from full command's first argument (mandatory in direct mode)
				componentType = args[0]

				// Component name: Get from full command's second argument (optional in direct mode), by default it is component type from first argument
				if len(args) == 2 {
					componentName = args[1]
				} else {
					componentName = args[0]
				}

				// Get available devfile components for checking devfile compatibility
				catalogDevfileList, err = catalog.ListDevfileComponents(co.devfileMetadata.devfileRegistry.Name)
				if err != nil {
					return err
				}
				if co.devfileMetadata.devfileRegistry.Name != "" && catalogDevfileList.Items == nil {
					return errors.Errorf("can't create devfile component from registry %s", co.devfileMetadata.devfileRegistry.Name)
				}

				if len(catalogDevfileList.DevfileRegistries) == 0 {
					isDevfileRegistryPresent = false
					log.Warning("Registry is empty, please run `odo registry add <registry name> <registry URL>` to add a registry\n")
				}
			}

			// Component namespace: Get from --project flag or --namespace flag, by default it is the current active namespace
			if co.devfileMetadata.componentNamespace == "" && !pushtarget.IsPushTargetDocker() {

				// Check to see if we've passed in "project", if not, default to the standard Kubernetes namespace
				componentNamespace, err = retrieveCmdNamespace(cmd)
				if err != nil {
					return err
				}

			} else {
				componentNamespace = defaultComponentNamespace
			}
		}

		// Set devfileMetadata struct
		co.devfileMetadata.componentType = componentType
		co.devfileMetadata.componentName = strings.ToLower(componentName)
		co.devfileMetadata.componentNamespace = strings.ToLower(componentNamespace)

		if util.CheckPathExists(co.DevfilePath) || co.devfileMetadata.devfilePath.value != "" {
			// Categorize the sections
			log.Info("Validation")

			var devfileAbsolutePath string
			if util.CheckPathExists(co.DevfilePath) || co.devfileMetadata.devfilePath.protocol == "file" {
				var devfilePath string
				if util.CheckPathExists(co.DevfilePath) {
					devfilePath = co.DevfilePath
				} else {
					devfilePath = co.devfileMetadata.devfilePath.value
				}
				devfileAbsolutePath, err = filepath.Abs(devfilePath)
				if err != nil {
					return err
				}
			} else if co.devfileMetadata.devfilePath.protocol == "http(s)" {
				devfileAbsolutePath = co.devfileMetadata.devfilePath.value
			}
			devfileSpinner := log.Spinnerf("Creating a devfile component from devfile path: %s", devfileAbsolutePath)
			defer devfileSpinner.End(true)

			// Initialize envinfo
			err = co.InitEnvInfoFromContext()
			if err != nil {
				return err
			}

			return nil
		}

		if isDevfileRegistryPresent && !co.forceS2i {
			// Categorize the sections
			log.Info("Validation")

			// Since we need to support both devfile and s2i, so we have to check if the component type is
			// supported by devfile, if it is supported we return and will download the corresponding devfile later,
			// if it is not supported we still need to run all the codes related with s2i after devfile compatibility check

			hasComponent := false

			for _, devfileComponent := range catalogDevfileList.Items {
				if co.devfileMetadata.componentType == devfileComponent.Name {
					hasComponent = true
					co.devfileMetadata.devfileSupport = true
					co.devfileMetadata.devfileLink = devfileComponent.Link
					co.devfileMetadata.devfileRegistry = devfileComponent.Registry
					break
				}
			}

			if co.forceS2i && hasComponent {
				s2iOverride := false
				for _, item := range catalogList.Items {
					if item.Name == co.devfileMetadata.componentType {
						s2iOverride = true
						break
					}
				}
				if !s2iOverride {
					return errors.New("cannot select this devfile component type with --s2i flag")
				}
			}

			existSpinner := log.Spinner("Checking devfile existence")
			if hasComponent {
				existSpinner.End(true)
			} else {
				existSpinner.End(false)
			}

			supportSpinner := log.Spinner("Checking devfile compatibility")
			if co.devfileMetadata.devfileSupport {
				registrySpinner := log.Spinnerf("Creating a devfile component from registry: %s", co.devfileMetadata.devfileRegistry.Name)

				// Initialize envinfo
				err = co.InitEnvInfoFromContext()
				if err != nil {
					return err
				}

				supportSpinner.End(true)
				registrySpinner.End(true)
				return nil
			}
			supportSpinner.End(false)

			// Currently only devfile component supports --registry flag, so if user specifies --registry when creating devfile component,
			// we should error out instead of running s2i componet code and throw warning message
			if co.devfileMetadata.devfileRegistry.Name != "" {
				return errors.Errorf("Devfile component type %s is not supported, please run `odo catalog list components` for a list of supported devfile component types", co.devfileMetadata.componentType)
			}

			log.Warningf("Devfile component type %s is not supported, please run `odo catalog list components` for a list of supported devfile component types", co.devfileMetadata.componentType)
		}
	}

	if len(args) == 0 || !cmd.HasFlags() {
		co.interactive = true
	}

	// Do not execute S2I specific code on Kubernetes Cluster or Docker
	// return from here, if it is not an openshift cluster.
	var openshiftCluster bool
	if !pushtarget.IsPushTargetDocker() {
		openshiftCluster, _ = co.Client.IsImageStreamSupported()
	} else {
		openshiftCluster = false
	}
	if !openshiftCluster {
		return errors.New("component type not found")
	}

	// check to see if config file exists or not, if it does that
	// means we shouldn't allow the user to override the current component
	if co.LocalConfigInfo.ConfigFileExists() {
		return errors.New("this directory already contains a component")
	}

	co.componentSettings = co.LocalConfigInfo.GetComponentSettings()

	// Below code is for INTERACTIVE mode
	if co.interactive {
		client := co.Client

		catalogList, err := catalog.ListComponents(client)
		if err != nil {
			return err
		}

		componentTypeCandidates := catalogutil.FilterHiddenComponents(catalogList.Items)
		selectedComponentType := ui.SelectComponentType(componentTypeCandidates)
		selectedImageTag := ui.SelectImageTag(componentTypeCandidates, selectedComponentType)
		componentType := selectedComponentType + ":" + selectedImageTag
		co.componentSettings.Type = &componentType

		// Ask for the type of source if not provided
		selectedSourceType := ui.SelectSourceType([]config.SrcType{config.LOCAL, config.GIT, config.BINARY})
		co.componentSettings.SourceType = &selectedSourceType
		selectedSourcePath := LocalDirectoryDefaultLocation

		// Get the current directory
		currentDirectory, err := os.Getwd()
		if err != nil {
			return err
		}

		if selectedSourceType == config.BINARY {

			// We ask for the source of the component context
			co.componentContext = ui.EnterInputTypePath("context", currentDirectory, ".")
			klog.V(4).Infof("Context: %s", co.componentContext)

			// If it's a binary, we have to ask where the actual binary in relation
			// to the context
			selectedSourcePath = ui.EnterInputTypePath("binary", ".")

			// Get the correct source location
			sourceLocation, err := getSourceLocation(selectedSourcePath, co.componentContext)
			if err != nil {
				return errors.Wrapf(err, "unable to get source location")
			}
			co.componentSettings.SourceLocation = &sourceLocation

		} else if selectedSourceType == config.GIT {

			// For git, we ask for the Git URL and set that as the source location
			cmpSrcLOC, selectedGitRef := ui.EnterGitInfo()
			co.componentSettings.SourceLocation = &cmpSrcLOC
			co.componentSettings.Ref = &selectedGitRef

		} else if selectedSourceType == config.LOCAL {

			// We ask for the source of the component, in this case the "path"!
			co.componentContext = ui.EnterInputTypePath("path", currentDirectory, ".")

			// Get the correct source location
			if co.componentContext == "" {
				co.componentContext = LocalDirectoryDefaultLocation
			}
			co.componentSettings.SourceLocation = &co.componentContext

		}

		defaultComponentName, err := createDefaultComponentName(co.Context, selectedComponentType, selectedSourceType, selectedSourcePath)
		if err != nil {
			return err
		}
		componentName := ui.EnterComponentName(defaultComponentName, co.Context)

		appName := ui.EnterOpenshiftName(co.Context.Application, "Which application do you want the component to be associated with", co.Context)
		co.componentSettings.Application = &appName

		projectName := ui.EnterOpenshiftName(co.Context.Project, "Which project go you want the component to be created in", co.Context)
		co.componentSettings.Project = &projectName

		co.componentSettings.Name = &componentName

		var ports []string

		if commonui.Proceed("Do you wish to set advanced options") {
			// if the user doesn't opt for advanced options, ports field would remain unpopulated
			// we then set it at the end of this function
			ports = ui.EnterPorts()

			co.componentEnvVars = ui.EnterEnvVars()

			if commonui.Proceed("Do you wish to set resource limits") {
				memMax := ui.EnterMemory("maximum", "512Mi")
				memMin := ui.EnterMemory("minimum", memMax)
				cpuMax := ui.EnterCPU("maximum", "1")
				cpuMin := ui.EnterCPU("minimum", cpuMax)

				memoryQuantity, err := util.FetchResourceQuantity(corev1.ResourceMemory, memMin, memMax, "")
				if err != nil {
					return err
				}
				if memoryQuantity != nil {
					co.componentSettings.MinMemory = &memMin
					co.componentSettings.MaxMemory = &memMax
				}
				cpuQuantity, err := util.FetchResourceQuantity(corev1.ResourceCPU, cpuMin, cpuMax, "")
				if err != nil {
					return err
				}
				if cpuQuantity != nil {
					co.componentSettings.MinCPU = &cpuMin
					co.componentSettings.MaxCPU = &cpuMax
				}
			}
		}

		// if user didn't opt for advanced options, "ports" value remains empty which panics the "odo push"
		// so we set the ports here
		if len(ports) == 0 {
			ports, err = co.Client.GetPortsFromBuilderImage(*co.componentSettings.Type)
			if err != nil {
				return err
			}
		}

		co.componentSettings.Ports = &ports
		// Above code is for INTERACTIVE mode

	} else {
		// Else if NOT using interactive / UI
		err = co.setComponentSourceAttributes()
		if err != nil {
			return err
		}
		err = co.setComponentName(args)
		if err != nil {
			return err
		}

		var portList []string
		if len(co.componentPorts) > 0 {
			portList = co.componentPorts
		} else {
			portList, err = co.Client.GetPortsFromBuilderImage(*co.componentSettings.Type)
			if err != nil {
				return err
			}
		}

		co.componentSettings.Ports = &(portList)
	}

	co.componentSettings.Project = &(co.Context.Project)
	envs, err := config.NewEnvVarListFromSlice(co.componentEnvVars)
	if err != nil {
		return
	}
	co.componentSettings.Envs = envs
	co.ignores = []string{}
	if co.now {
		co.ResolveSrcAndConfigFlags()
		err = co.ResolveProject(co.Context.Project)
		if err != nil {
			return err
		}
	}
	return
}

// Validate validates the create parameters
func (co *CreateOptions) Validate() (err error) {

	if !co.forceS2i && co.devfileMetadata.devfileSupport {
		// Validate if the devfile component that user wants to create already exists
		spinner := log.Spinner("Validating devfile component")
		defer spinner.End(false)

		err = util.ValidateK8sResourceName("component name", co.devfileMetadata.componentName)
		if err != nil {
			return err
		}

		// Only validate namespace if pushtarget isn't docker
		if !pushtarget.IsPushTargetDocker() {
			err := util.ValidateK8sResourceName("component namespace", co.devfileMetadata.componentNamespace)
			if err != nil {
				return err
			}
		}

		spinner.End(true)

		return nil
	}

	log.Info("Validation")

	supported, err := catalog.IsComponentTypeSupported(co.Context.Client, *co.componentSettings.Type)
	if err != nil {
		return err
	}

	if !supported {
		log.Infof("Warning: %s is not fully supported by odo, and it is not guaranteed to work", *co.componentSettings.Type)
	}

	s := log.Spinner("Validating component")
	defer s.End(false)
	if err := component.ValidateComponentCreateRequest(co.Context.Client, co.componentSettings, co.componentContext); err != nil {
		return err
	}

	s.End(true)
	return nil
}

// Downloads first starter project from list of starter projects in devfile
// Currenty type git with a non github url is not supported
func downloadStarterProject(devObj parser.DevfileObj, projectPassed string, interactive bool) error {
	if projectPassed == "" && !interactive {
		return nil
	}

	// Retrieve starter projects
	starterProjects := devObj.Data.GetStarterProjects()

	var starterProject *common.DevfileStarterProject
	var err error
	if interactive {
		starterProject = getStarterProjectInteractiveMode(starterProjects)
	} else {
		starterProject, err = getStarterProjectFromFlag(starterProjects, projectPassed)
		if err != nil {
			return err
		}
	}

	if starterProject == nil {
		return nil
	}

	// Retrieve the working directory in order to clone correctly
	path, err := os.Getwd()
	if err != nil {
		return errors.Wrapf(err, "Could not get the current working directory.")
	}

	if starterProject.ClonePath != "" {
		clonePath := filepath.FromSlash(starterProject.ClonePath)

		path = filepath.Join(path, clonePath)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err = os.MkdirAll(path, os.FileMode(0755))
			if err != nil {
				return errors.Wrapf(err, "failed creating folder with path: %s", path)
			}
		}
	}

	// We will check to see if the project has a valid directory
	err = util.IsValidProjectDir(path, DevfilePath)
	if err != nil {
		return err
	}

	log.Info("\nStarter Project")

	if starterProject.Git != nil || starterProject.Github != nil {

		projectSource := common.GitLikeProjectSource{}
		if starterProject.Git != nil {
			projectSource = starterProject.Git.GitLikeProjectSource
		} else {
			projectSource = starterProject.Github.GitLikeProjectSource
		}

		remoteName, remoteUrl, revision, err := projectSource.GetDefaultSource()
		if err != nil {
			return errors.Wrapf(err, "unable to get default project source for starter project %s", starterProject.Name)
		}

		if revision != "" {
			log.Warning("Specifying 'revision' in 'checkoutFrom' is not yet supported in odo.")
		}

		downloadSpinner := log.Spinnerf("Downloading starter project %s from %s", starterProject.Name, remoteUrl)
		defer downloadSpinner.End(false)

		_, err = git.PlainClone(path, false, &git.CloneOptions{
			URL:           remoteUrl,
			RemoteName:    remoteName,
			ReferenceName: plumbing.ReferenceName(revision),
			SingleBranch:  true,
			// we don't need history for starter projects
			Depth: 1,
		})
		if err != nil {
			return err
		}

		// we don't want to download project be a git repo
		err = os.RemoveAll(filepath.Join(path, ".git"))
		if err != nil {
			// we don't need to return (fail) if this happens
			log.Warning("Unable to delete .git from cloned starter project")
		}
		downloadSpinner.End(true)

	} else if starterProject.Zip != nil {
		url := starterProject.Zip.Location
		logUrl := starterProject.Zip.Location
		sparseDir := starterProject.Zip.SparseCheckoutDir
		downloadSpinner := log.Spinnerf("Downloading starter project %s from %s", starterProject.Name, logUrl)
		err := checkoutProject(sparseDir, url, path)
		if err != nil {
			downloadSpinner.End(false)
			return err
		}
		downloadSpinner.End(true)
	} else {
		return errors.Errorf("Project type not supported")
	}

	return nil
}

func (co *CreateOptions) s2iRun() (err error) {
	err = co.LocalConfigInfo.SetComponentSettings(co.componentSettings)
	if err != nil {
		return errors.Wrapf(err, "failed to persist the component settings to config file")
	}
	if co.now {
		co.Context, co.LocalConfigInfo, err = genericclioptions.UpdatedContext(co.Context)

		if err != nil {
			return errors.Wrap(err, "unable to retrieve updated local config")
		}
		err = co.SetSourceInfo()
		if err != nil {
			return errors.Wrap(err, "unable to set source information")
		}
		err = co.Push()
		if err != nil {
			return errors.Wrapf(err, "failed to push the changes")
		}
	} else {
		log.Italic("\nPlease use `odo push` command to create the component with source deployed")
	}
	return nil
}

// Run has the logic to perform the required actions as part of command
func (co *CreateOptions) devfileRun() (err error) {
	var devfileData []byte
	// Use existing devfile directly from --devfile flag
	if co.devfileMetadata.devfilePath.value != "" {
		if co.devfileMetadata.devfilePath.protocol == "http(s)" {
			// User specify devfile path is http(s) URL
			params := util.HTTPRequestParams{
				URL:   co.devfileMetadata.devfilePath.value,
				Token: co.devfileMetadata.token,
			}
			devfileData, err = util.DownloadFileInMemory(params)
			if err != nil {
				return errors.Wrapf(err, "failed to download devfile for devfile component from %s", co.devfileMetadata.devfilePath.value)
			}
		} else if co.devfileMetadata.devfilePath.protocol == "file" {
			devfileData, err = ioutil.ReadFile(co.devfileMetadata.devfilePath.value)
			if err != nil {
				return errors.Wrapf(err, "failed to read devfile from %s", co.devfileMetadata.devfilePath)
			}
		}
	} else {
		if util.CheckPathExists(DevfilePath) {
			// if local devfile already exists read that
			// odo create command was exected in a directory already containing devfile
			devfileData, err = ioutil.ReadFile(DevfilePath)
			if err != nil {
				return errors.Wrapf(err, "failed to read devfile from %s", DevfilePath)
			}
		} else {
			// Download devfile from registry
			params := util.HTTPRequestParams{
				URL: co.devfileMetadata.devfileRegistry.URL + co.devfileMetadata.devfileLink,
			}

			if registryUtil.IsSecure(co.devfileMetadata.devfileRegistry.Name) {
				token, err := keyring.Get(util.CredentialPrefix+co.devfileMetadata.devfileRegistry.Name, "default")
				if err != nil {
					return errors.Wrap(err, "unable to get secure registry credential from keyring")
				}
				params.Token = token
			}

			cfg, err := preference.New()
			if err != nil {
				return err
			}
			devfileData, err = util.DownloadFileInMemoryWithCache(params, cfg.GetRegistryCacheTime())
			if err != nil {
				return errors.Wrapf(err, "failed to download devfile for devfile component from %s", co.devfileMetadata.devfileRegistry.URL+co.devfileMetadata.devfileLink)
			}
		}
	}
	devObj, err := devfile.ParseFromDataAndValidate(devfileData)
	if err != nil {
		return errors.Wrap(err, "unable to parse devfile")
	}

	err = downloadStarterProject(devObj, co.devfileMetadata.starter, co.interactive)
	if err != nil {
		return errors.Wrap(err, "failed to download project for devfile component")
	}

	// save devfile
	// use original devfileData to persist original formating of the devfile file
	err = ioutil.WriteFile(DevfilePath, devfileData, 0644) // #nosec G306
	if err != nil {
		return errors.Wrapf(err, "unable to save devfile to %s", DevfilePath)
	}

	// Generate env file
	err = co.EnvSpecificInfo.SetComponentSettings(envinfo.ComponentSettings{
		Name:      co.devfileMetadata.componentName,
		Namespace: co.devfileMetadata.componentNamespace,
		AppName:   co.appName,
	})
	if err != nil {
		return errors.Wrap(err, "failed to create env file for devfile component")
	}

	sourcePath, err := util.GetAbsPath(co.componentContext)
	if err != nil {
		return errors.Wrap(err, "unable to get source path")
	}

	ignoreFile, err := util.CheckGitIgnoreFile(sourcePath)
	if err != nil {
		return err
	}

	err = util.AddFileToIgnoreFile(ignoreFile, filepath.Join(co.componentContext, envDir))
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

// Run has the logic to perform the required actions as part of command
func (co *CreateOptions) Run() (err error) {

	// By default we run Devfile
	if !co.forceS2i && co.devfileMetadata.devfileSupport {
		return co.devfileRun()
	}

	// If not, we run s2i (if the --s2i parameter has been passed in).
	// It's implied that we have passed it in if Devfile did not run above
	err = co.s2iRun()
	if err != nil {
		return err
	}

	if log.IsJSON() {
		var componentDesc component.Component
		co.Context, co.LocalConfigInfo, err = genericclioptions.UpdatedContext(co.Context)
		if err != nil {
			return err
		}
		state := component.GetComponentState(co.Client, *co.componentSettings.Name, co.Context.Application)

		if state == component.StateTypeNotPushed || state == component.StateTypeUnknown {
			componentDesc, err = component.GetComponentFromConfig(co.LocalConfigInfo)
			componentDesc.Status.State = state
			if err != nil {
				return err
			}
		} else {
			componentDesc, err = component.GetComponent(co.Context.Client, *co.componentSettings.Name, co.Context.Application, co.Context.Project)
			if err != nil {
				return err
			}
		}

		componentDesc.Spec.Ports = co.LocalConfigInfo.GetPorts()
		machineoutput.OutputSuccess(componentDesc)
	}
	return
}

func checkoutProject(sparseCheckoutDir, zipURL, path string) error {

	if sparseCheckoutDir != "" {
		err := util.GetAndExtractZip(zipURL, path, sparseCheckoutDir)
		if err != nil {
			return errors.Wrap(err, "failed to download and extract project zip folder")
		}
	} else {
		// extract project to current working directory
		err := util.GetAndExtractZip(zipURL, path, "/")
		if err != nil {
			return errors.Wrap(err, "failed to download and extract project zip folder")
		}
	}
	return nil
}

// getStarterProjectInteractiveMode gets starter project value by asking user in interactive mode.
func getStarterProjectInteractiveMode(projects []common.DevfileStarterProject) *common.DevfileStarterProject {
	projectName := ui.SelectStarterProject(projects)

	// if user do not wish to download starter project or there are no projects in devfile, project name would be empty
	if projectName == "" {
		return nil
	}

	var project common.DevfileStarterProject

	for _, value := range projects {
		if value.Name == projectName {
			project = value
			break
		}
	}

	return &project
}

// getStarterProjectFromFlag gets starter project value from flag --starter.
func getStarterProjectFromFlag(projects []common.DevfileStarterProject, projectPassed string) (project *common.DevfileStarterProject, err error) {

	nOfProjects := len(projects)

	if nOfProjects == 0 {
		return nil, errors.Errorf("no starter project found in devfile.")
	}

	// Determine what project to be used
	if nOfProjects == 1 && projectPassed == defaultStarterProjectName {
		project = &projects[0]
	} else if nOfProjects > 1 && projectPassed == defaultStarterProjectName {
		project = &projects[0]
		log.Warning("There are multiple projects in this devfile but none have been specified in --starter. Downloading the first: " + project.Name)
	} else { //If the user has specified a project
		projectFound := false
		for indexOfProject, projectInfo := range projects {
			if projectInfo.Name == projectPassed { //Get the index
				project = &projects[indexOfProject]
				projectFound = true
			}
		}

		if !projectFound {
			return nil, errors.Errorf("the project: %s specified in --starter does not exist", projectPassed)
		}
	}

	return project, err

}

// NewCmdCreate implements the create odo command
func NewCmdCreate(name, fullName string) *cobra.Command {
	co := NewCreateOptions()
	var componentCreateCmd = &cobra.Command{
		Use:         fmt.Sprintf("%s <component_type> [component_name] [flags]", name),
		Short:       "Create a new component",
		Long:        createLongDesc,
		Example:     fmt.Sprintf(createExample, fullName),
		Args:        cobra.RangeArgs(0, 2),
		Annotations: map[string]string{"machineoutput": "json", "command": "component"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(co, cmd, args)
		},
	}
	componentCreateCmd.Flags().StringVarP(&co.componentBinary, "binary", "b", "", "Create a binary file component component using given artifact. Works only with Java components. File needs to be in the context directory.")
	componentCreateCmd.Flags().StringVarP(&co.componentGit, "git", "g", "", "Create a git component using this repository.")
	componentCreateCmd.Flags().StringVarP(&co.componentGitRef, "ref", "r", "", "Use a specific ref e.g. commit, branch or tag of the git repository (only valid for --git components)")
	genericclioptions.AddContextFlag(componentCreateCmd, &co.componentContext)
	componentCreateCmd.Flags().StringSliceVarP(&co.componentPorts, "port", "p", []string{}, "Ports to be used when the component is created (ex. 8080,8100/tcp,9100/udp)")
	componentCreateCmd.Flags().StringSliceVar(&co.componentEnvVars, "env", []string{}, "Environmental variables for the component. For example --env VariableName=Value")

	componentCreateCmd.Flags().StringVar(&co.devfileMetadata.starter, "starter", "", "Download a project specified in the devfile")
	componentCreateCmd.Flags().Lookup("starter").NoOptDefVal = defaultStarterProjectName //Default value to pass to the flag if one is not specified.
	componentCreateCmd.Flags().StringVar(&co.devfileMetadata.devfileRegistry.Name, "registry", "", "Create devfile component from specific registry")
	componentCreateCmd.Flags().StringVar(&co.devfileMetadata.devfilePath.value, "devfile", "", "Path to the user specified devfile")
	componentCreateCmd.Flags().StringVar(&co.devfileMetadata.token, "token", "", "Token to be used when downloading devfile from the devfile path that is specified via --devfile")
	componentCreateCmd.Flags().BoolVar(&co.forceS2i, "s2i", false, "Enforce S2I type components")

	componentCreateCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	// Adding `--now` flag
	genericclioptions.AddNowFlag(componentCreateCmd, &co.now)
	//Adding `--project` flag
	projectCmd.AddProjectFlag(componentCreateCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(componentCreateCmd)

	completion.RegisterCommandHandler(componentCreateCmd, completion.CreateCompletionHandler)
	completion.RegisterCommandFlagHandler(componentCreateCmd, "context", completion.FileCompletionHandler)
	completion.RegisterCommandFlagHandler(componentCreateCmd, "binary", completion.FileCompletionHandler)

	return componentCreateCmd
}
