package component

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/config"
	"github.com/openshift/odo/pkg/log"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	catalogutil "github.com/openshift/odo/pkg/odo/cli/catalog/util"
	"github.com/openshift/odo/pkg/odo/cli/component/ui"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	commonui "github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/util"

	corev1 "k8s.io/api/core/v1"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
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
	memoryMax         string
	memoryMin         string
	memory            string
	cpuMax            string
	cpuMin            string
	cpu               string
	wait              bool
	interactive       bool
	now               bool
	*CommonPushOptions
}

// CreateRecommendedCommandName is the recommended watch command name
const CreateRecommendedCommandName = "create"

// LocalDirectoryDefaultLocation is the default location of where --local files should always be..
// since the application will always be in the same directory as `.odo`, we will always set this as: ./
const LocalDirectoryDefaultLocation = "./"

var createLongDesc = ktemplates.LongDesc(`Create a configuration describing a component.

If a component name is not provided, it'll be auto-generated.

A full list of component types that can be deployed is available using: 'odo catalog list'

By default, builder images (component type) will be used from the current namespace. You can explicitly supply a namespace by using: odo create namespace/name:version
If version is not specified by default, latest will be chosen as the version.`)

var createExample = ktemplates.Examples(`  # Create new Node.js component with the source in current directory.
%[1]s nodejs

# Create new Node.js component named 'frontend' with the source in './frontend' directory
%[1]s nodejs frontend --context ./frontend

# Create new Java component with binary named sample.jar in './target' directory
%[1]s java:8  --binary target/sample.jar

# Create new Node.js component with source from remote git repository
%[1]s nodejs --git https://github.com/openshift/nodejs-ex.git

# Create new Node.js component with custom ports, additional environment variables and memory and cpu limits
%[1]s nodejs --port 8080,8100/tcp,9100/udp --env key=value,key1=value1 --memory 4Gi --cpu 2`)

// NewCreateOptions returns new instance of CreateOptions
func NewCreateOptions() *CreateOptions {
	return &CreateOptions{
		CommonPushOptions: NewCommonPushOptions(),
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
		cPath, err := filepath.EvalSymlinks(co.componentBinary)
		if err != nil {
			return err
		}
		// Convert componentContext to absolute path, so it can be safely used in filepath.Rel
		// even when it is not set (empty). In this case filepath.Abs will return current directory.
		absContext, err := filepath.Abs(co.componentContext)
		if err != nil {
			return errors.Wrapf(err, "unable to convert \"%s\" to absolute path", co.componentContext)
		}
		// we need to store the SourceLocation relative to the componentContext
		relativePathToSource, err := filepath.Rel(absContext, cPath)
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
		return fmt.Errorf("The source can be either --binary or --local or --git")

	}

	// Set the Git reference if passed
	if len(co.componentGitRef) != 0 {
		co.componentSettings.Ref = &(co.componentGitRef)
	}

	// Error out if reference is passed but no --git flag passed
	if len(co.componentGit) == 0 && len(co.componentGitRef) != 0 {
		return fmt.Errorf("The --ref flag is only valid for --git flag")
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

func (co *CreateOptions) setResourceLimits() error {
	ensureAndLogProperResourceUsage(co.memory, co.memoryMin, co.memoryMax, "memory")

	ensureAndLogProperResourceUsage(co.cpu, co.cpuMin, co.cpuMax, "cpu")

	memoryQuantity, err := util.FetchResourceQuantity(corev1.ResourceMemory, co.memoryMin, co.memoryMax, co.memory)
	if err != nil {
		return err
	}
	if memoryQuantity != nil {
		minMemory := memoryQuantity.MinQty.String()
		maxMemory := memoryQuantity.MaxQty.String()
		co.componentSettings.MinMemory = &minMemory
		co.componentSettings.MaxMemory = &maxMemory
	}

	cpuQuantity, err := util.FetchResourceQuantity(corev1.ResourceCPU, co.cpuMin, co.cpuMax, co.cpu)
	if err != nil {
		return err
	}
	if cpuQuantity != nil {
		minCPU := cpuQuantity.MinQty.String()
		maxCPU := cpuQuantity.MaxQty.String()
		co.componentSettings.MinCPU = &minCPU
		co.componentSettings.MaxCPU = &maxCPU
	}

	return nil
}

// Complete completes create args
func (co *CreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 || !cmd.HasFlags() {
		co.interactive = true
	}

	co.localConfigInfo, err = config.NewLocalConfigInfo(co.componentContext)
	if err != nil {
		return errors.Wrap(err, "failed intiating local config")
	}

	// check to see if config file exists or not, if it does that
	// means we shouldn't allow the user to override the current component
	if co.localConfigInfo.ConfigFileExists() {
		return errors.New("this directory already contains a component")
	}

	co.componentSettings = co.localConfigInfo.GetComponentSettings()

	co.Context = genericclioptions.NewContextCreatingAppIfNeeded(cmd)

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
			glog.V(4).Infof("Context: %s", co.componentContext)

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

		appName := ui.EnterOpenshiftName(co.Context.Application, "Which application do you want the commponent to be associated with", co.Context)
		co.componentSettings.Application = &appName

		projectName := ui.EnterOpenshiftName(co.Context.Project, "Which project go you want the component to be created in", co.Context)
		co.componentSettings.Project = &projectName

		co.componentSettings.Name = &componentName

		if commonui.Proceed("Do you wish to set advanced options") {
			ports := ui.EnterPorts()
			if len(ports) > 0 {
				co.componentSettings.Ports = &ports
			}
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
		err = co.setResourceLimits()
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
	s := log.Spinner("Validating component")
	defer s.End(false)

	if err := component.ValidateComponentCreateRequest(co.Context.Client, co.componentSettings, false, false); err != nil {
		return err
	}

	s.End(true)
	return nil
}

// Run has the logic to perform the required actions as part of command
func (co *CreateOptions) Run() (err error) {
	err = co.localConfigInfo.SetComponentSettings(co.componentSettings)
	if err != nil {
		return errors.Wrapf(err, "failed to persist the component settings to config file")
	}
	if co.now {
		co.Context, co.localConfigInfo, err = genericclioptions.UpdatedContext(co.Context)

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
		log.Infof("Please use `odo push` command to create the component with source deployed")
	}
	return
}

func (co *CreateOptions) createCmpIfNotExistsAndApplyCmpConfig(stdout io.Writer) error {

	cmpName := co.localConfigInfo.GetName()
	appName := co.localConfigInfo.GetApplication()
	cmpType := co.localConfigInfo.GetType()

	isCmpExists, err := component.Exists(co.Context.Client, cmpName, appName)
	if err != nil {
		return errors.Wrapf(err, "failed to check if component %s exists or not", cmpName)
	}

	if !isCmpExists {
		log.Successf("Creating %s component with name %s", cmpType, cmpName)
		// Classic case of component creation
		if err = component.CreateComponent(co.Context.Client, *co.localConfigInfo, co.componentContext, stdout); err != nil {
			log.Errorf(
				"Failed to create component with name %s. Please use `odo config view` to view settings used to create component. Error: %+v",
				cmpName,
				err,
			)
			os.Exit(1)
		}
		log.Successf("Successfully created component %s", cmpName)
	}

	// Apply config
	err = component.ApplyConfig(co.Context.Client, *co.localConfigInfo, stdout, isCmpExists)
	if err != nil {
		odoutil.LogErrorAndExit(err, "Failed to update config to component deployed")
	}
	log.Successf("Successfully updated component with name: %v", cmpName)

	return nil
}

// The general cpu/memory is used as a fallback when it's set and both min-cpu/memory max-cpu/memory are not set
// when the only thing specified is the min or max value, we exit the application
func ensureAndLogProperResourceUsage(resource, resourceMin, resourceMax, resourceName string) {
	if strings.HasPrefix(resourceMin, "-") {
		log.Errorf("min-%s cannot be negative", resource)
		os.Exit(1)
	}
	if strings.HasPrefix(resourceMax, "-") {
		log.Errorf("max-%s cannot be negative", resource)
		os.Exit(1)
	}
	if strings.HasPrefix(resource, "-") {
		log.Errorf("%s cannot be negative", resource)
		os.Exit(1)
	}
	if resourceMin != "" && resourceMax != "" && resource != "" {
		log.Infof("`--%s` will be ignored as `--min-%s` and `--max-%s` has been passed\n", resourceName, resourceName, resourceName)
	}
	if (resourceMin == "") != (resourceMax == "") && resource != "" {
		log.Infof("Using `--%s` %s for min and max limits.\n", resourceName, resource)
	}
	if (resourceMin == "") != (resourceMax == "") && resource == "" {
		log.Errorf("`--min-%s` should accompany `--max-%s` or pass `--%s` to use same value for both min and max or try not passing any of them\n", resourceName, resourceName, resourceName)
		os.Exit(1)
	}
}

// NewCmdCreate implements the create odo command
func NewCmdCreate(name, fullName string) *cobra.Command {
	co := NewCreateOptions()
	var componentCreateCmd = &cobra.Command{
		Use:     fmt.Sprintf("%s <component_type> [component_name] [flags]", name),
		Short:   "Create a new component",
		Long:    createLongDesc,
		Example: fmt.Sprintf(createExample, fullName),
		Args:    cobra.RangeArgs(0, 2),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(co, cmd, args)
		},
	}
	componentCreateCmd.Flags().StringVarP(&co.componentBinary, "binary", "b", "", "Create a binary file component component using given artifact. Works only with Java components. File needs to be in the context directory.")
	componentCreateCmd.Flags().StringVarP(&co.componentGit, "git", "g", "", "Create a git component using this repository.")
	componentCreateCmd.Flags().StringVarP(&co.componentGitRef, "ref", "r", "", "Use a specific ref e.g. commit, branch or tag of the git repository")
	genericclioptions.AddContextFlag(componentCreateCmd, &co.componentContext)
	componentCreateCmd.Flags().StringVar(&co.memory, "memory", "", "Amount of memory to be allocated to the component. ex. 100Mi (sets min-memory and max-memory to this value)")
	componentCreateCmd.Flags().StringVar(&co.memoryMin, "min-memory", "", "Limit minimum amount of memory to be allocated to the component. ex. 100Mi")
	componentCreateCmd.Flags().StringVar(&co.memoryMax, "max-memory", "", "Limit maximum amount of memory to be allocated to the component. ex. 100Mi")
	componentCreateCmd.Flags().StringVar(&co.cpu, "cpu", "", "Amount of cpu to be allocated to the component. ex. 100m or 0.1 (sets min-cpu and max-cpu to this value)")
	componentCreateCmd.Flags().StringVar(&co.cpuMin, "min-cpu", "", "Limit minimum amount of cpu to be allocated to the component. ex. 100m")
	componentCreateCmd.Flags().StringVar(&co.cpuMax, "max-cpu", "", "Limit maximum amount of cpu to be allocated to the component. ex. 1")
	componentCreateCmd.Flags().StringSliceVarP(&co.componentPorts, "port", "p", []string{}, "Ports to be used when the component is created (ex. 8080,8100/tcp,9100/udp)")
	componentCreateCmd.Flags().StringSliceVar(&co.componentEnvVars, "env", []string{}, "Environmental variables for the component. For example --env VariableName=Value")
	// Add a defined annotation in order to appear in the help menu
	componentCreateCmd.Annotations = map[string]string{"command": "component"}
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
