package component

import (
	"fmt"
	"github.com/openshift/odo/pkg/odo/cli/component/ui"
	commonui "github.com/openshift/odo/pkg/odo/cli/ui"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/openshift/odo/pkg/odo/util/validation"

	"github.com/fatih/color"
	"github.com/golang/glog"
	"github.com/openshift/odo/pkg/application"
	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/occlient"
	catalogutil "github.com/openshift/odo/pkg/odo/cli/catalog/util"
	"github.com/openshift/odo/pkg/util"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

// CreateOptions encapsulates create options
type CreateOptions struct {
	occlient.CreateArgs
	*genericclioptions.Context
	componentBinary  string
	componentGit     string
	componentGitRef  string
	componentLocal   string
	componentPorts   []string
	componentEnvVars []string
	memoryMax        string
	memoryMin        string
	memory           string
	cpuMax           string
	cpuMin           string
	cpu              string
	wait             bool
	interactive      bool
}

// CreateRecommendedCommandName is the recommended watch command name
const CreateRecommendedCommandName = "create"

var createLongDesc = ktemplates.LongDesc(`Create a new component to deploy on OpenShift.

If a component name is not provided, it'll be auto-generated.

By default, builder images will be used from the current namespace. You can explicitly supply a namespace by using: odo create namespace/name:version
If version is not specified by default, latest wil be chosen as the version.

A full list of component types that can be deployed is available using: 'odo catalog list'`)

var createExample = ktemplates.Examples(`  # Create new Node.js component with the source in current directory.
%[1]s nodejs

# A specific image version may also be specified
%[1]s nodejs:latest

# Create new Node.js component named 'frontend' with the source in './frontend' directory
%[1]s nodejs frontend --local ./frontend

# Create a new Node.js component of version 6 from the 'openshift' namespace
%[1]s openshift/nodejs:6 --local /nodejs-ex

# Create new Wildfly component with binary named sample.war in './downloads' directory
%[1]s wildfly wildly --binary ./downloads/sample.war

# Create new Node.js component with source from remote git repository
%[1]s nodejs --git https://github.com/openshift/nodejs-ex.git

# Create new Node.js git component while specifying a branch, tag or commit ref
%[1]s nodejs --git https://github.com/openshift/nodejs-ex.git --ref master

# Create new Node.js git component while specifying a tag
%[1]s nodejs --git https://github.com/openshift/nodejs-ex.git --ref v1.0.1

# Create new Node.js component with the source in current directory and ports 8080-tcp,8100-tcp and 9100-udp exposed
%[1]s nodejs --port 8080,8100/tcp,9100/udp

# Create new Node.js component with the source in current directory and env variables key=value and key1=value1 exposed
%[1]s nodejs --env key=value,key1=value1

# For more examples, visit: https://github.com/openshift/odo/blob/master/docs/examples.md
%[1]s python --git https://github.com/openshift/django-ex.git

# Passing memory limits
%[1]s nodejs --memory 150Mi
%[1]s nodejs --min-memory 150Mi --max-memory 300 Mi

# Passing cpu limits
%[1]s nodejs --cpu 2
%[1]s nodejs --min-cpu 200m --max-cpu 2

  `)

// NewCreateOptions returns new instance of CreateOptions
func NewCreateOptions() *CreateOptions {
	return &CreateOptions{}
}

func (co *CreateOptions) setCmpSourceAttrs() (err error) {
	componentCnt := 0

	if len(co.componentBinary) != 0 {
		if co.CreateArgs.SourcePath, err = util.GetAbsPath(co.componentBinary); err != nil {
			return err
		}
		co.CreateArgs.SourceType = occlient.BINARY
		componentCnt++
	}
	if len(co.componentGit) != 0 {
		co.CreateArgs.SourcePath = co.componentGit
		co.CreateArgs.SourceType = occlient.GIT
		componentCnt++
	}
	if len(co.componentLocal) != 0 {
		if co.CreateArgs.SourcePath, err = util.GetAbsPath(co.componentLocal); err != nil {
			return err
		}
		co.CreateArgs.SourceType = occlient.LOCAL
		componentCnt++
	}

	if componentCnt > 1 {
		return fmt.Errorf("The source can be either --binary or --local or --git")
	}

	if len(co.componentGitRef) != 0 {
		co.CreateArgs.SourceRef = co.componentGitRef
	}

	if len(co.componentGit) == 0 && len(co.componentGitRef) != 0 {
		return fmt.Errorf("The --ref flag is only valid for --git flag")
	}

	return
}

func (co *CreateOptions) setCmpName(args []string) (err error) {
	componentImageName, componentType, _, _ := util.ParseComponentImageName(args[0])
	co.CreateArgs.ImageName = componentImageName

	if len(args) == 2 {
		co.CreateArgs.Name = args[1]
		return
	}

	componentName, err := createDefaultComponentName(co.Context, componentType, co.CreateArgs.SourceType, co.CreateArgs.SourcePath)
	if err != nil {
		return err
	}

	co.CreateArgs.Name = componentName
	return
}

func createDefaultComponentName(context *genericclioptions.Context, componentType string, sourceType occlient.CreateType, sourcePath string) (string, error) {
	// Fetch list of existing components in-order to attempt generation of unique component name
	componentList, err := component.List(context.Client, context.Application)
	if err != nil {
		return "", err
	}

	// Retrieve the componentName, if the componentName isn't specified, we will use the default image name
	componentName, err := component.GetDefaultComponentName(
		sourcePath,
		sourceType,
		componentType,
		componentList,
	)

	if err != nil {
		return "", nil
	}

	return componentName, nil
}

func (co *CreateOptions) setResourceLimits() {
	ensureAndLogProperResourceUsage(co.memory, co.memoryMin, co.memoryMax, "memory")

	ensureAndLogProperResourceUsage(co.cpu, co.cpuMin, co.cpuMax, "cpu")

	resourceQuantity := []util.ResourceRequirementInfo{}
	memoryQuantity := util.FetchResourceQuantity(corev1.ResourceMemory, co.memoryMin, co.memoryMax, co.memory)
	if memoryQuantity != nil {
		resourceQuantity = append(resourceQuantity, *memoryQuantity)
	}
	cpuQuantity := util.FetchResourceQuantity(corev1.ResourceCPU, co.cpuMin, co.cpuMax, co.cpu)
	if cpuQuantity != nil {
		resourceQuantity = append(resourceQuantity, *cpuQuantity)
	}
	co.CreateArgs.Resources = resourceQuantity
}

// Complete completes create args
func (co *CreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 || !cmd.HasFlags() {
		co.interactive = true
	}

	co.Context = genericclioptions.NewContextCreatingAppIfNeeded(cmd)
	co.CreateArgs.ApplicationName = co.Context.Application

	if co.interactive {
		client := co.Client

		componentTypeCandidates, err := catalog.List(client)
		if err != nil {
			return err
		}
		componentTypeCandidates = catalogutil.FilterHiddenComponents(componentTypeCandidates)
		selectedComponentType := ui.SelectComponentType(componentTypeCandidates)
		selectedImageTag := ui.SelectImageTag(componentTypeCandidates, selectedComponentType)
		co.CreateArgs.ImageName = selectedComponentType + ":" + selectedImageTag

		selectedSourceType := ui.SelectSourceType([]occlient.CreateType{occlient.LOCAL, occlient.GIT, occlient.BINARY})
		co.CreateArgs.SourceType = selectedSourceType
		selectedSourcePath := ""
		currentDirectory, err := os.Getwd()
		if err != nil {
			return err
		}
		if selectedSourceType == occlient.LOCAL {
			selectedSourcePath = ui.EnterInputTypePath("local", currentDirectory, ".")
			selectedSourcePath, err = util.GetAbsPath(selectedSourcePath)
			if err != nil {
				return err
			}
		} else if selectedSourceType == occlient.BINARY {
			selectedSourcePath = ui.EnterInputTypePath("binary", currentDirectory)
			selectedSourcePath, err = util.GetAbsPath(selectedSourcePath)
			if err != nil {
				return err
			}
		} else if selectedSourceType == occlient.GIT {
			var selectedGitRef string
			selectedSourcePath, selectedGitRef = ui.EnterGitInfo()
			co.CreateArgs.SourceRef = selectedGitRef
		}
		co.CreateArgs.SourcePath = selectedSourcePath

		defaultComponentName, err := createDefaultComponentName(co.Context, selectedComponentType, selectedSourceType, selectedSourcePath)
		if err != nil {
			return err
		}
		co.CreateArgs.Name = ui.EnterComponentName(defaultComponentName, co.Context)

		if commonui.Proceed("Do you wish to set advanced options") {
			co.CreateArgs.Ports = ui.EnterPorts()
			co.CreateArgs.EnvVars = ui.EnterEnvVars()

			if commonui.Proceed("Do you wish to set resource limits") {
				memMax := ui.EnterMemory("maximum", "512Mi")
				memMin := ui.EnterMemory("minimum", memMax)
				cpuMax := ui.EnterCPU("maximum", "1")
				cpuMin := ui.EnterCPU("minimum", cpuMax)

				resourceQuantity := []util.ResourceRequirementInfo{}
				memoryQuantity := util.FetchResourceQuantity(corev1.ResourceMemory, memMin, memMax, "")
				if memoryQuantity != nil {
					resourceQuantity = append(resourceQuantity, *memoryQuantity)
				}
				cpuQuantity := util.FetchResourceQuantity(corev1.ResourceCPU, cpuMin, cpuMax, "")
				if cpuQuantity != nil {
					resourceQuantity = append(resourceQuantity, *cpuQuantity)
				}
				co.CreateArgs.Resources = resourceQuantity
			}
		}

		co.CreateArgs.Wait = commonui.Proceed("Would you wish to wait until the component is fully ready after after creation")
		// needed in order to avoid showing a misleading message at the end of process
		co.wait = co.CreateArgs.Wait

	} else {
		co.CreateArgs.Wait = co.wait
		err = co.setCmpSourceAttrs()
		if err != nil {
			return err
		}
		err = co.setCmpName(args)
		if err != nil {
			return err
		}
		co.setResourceLimits()
		co.CreateArgs.Ports = co.componentPorts
		co.CreateArgs.EnvVars = co.componentEnvVars
	}

	return
}

// Validate validates the create parameters
func (co *CreateOptions) Validate() (err error) {
	_, componentType, _, componentVersion := util.ParseComponentImageName(co.CreateArgs.ImageName)
	// Check to see if the catalog type actually exists
	exists, err := catalog.Exists(co.Context.Client, componentType)
	if err != nil {
		return errors.Wrapf(err, "Failed to create component of type %s", componentType)
	}
	if !exists {
		log.Info("Run 'odo catalog list components' for a list of supported component types")
		return fmt.Errorf("Failed to find component of type %s", componentType)
	}

	// Check to see if that particular version exists
	versionExists, err := catalog.VersionExists(co.Context.Client, componentType, componentVersion)
	if err != nil {
		return errors.Wrapf(err, "Failed to create component of type %s of version %s", componentType, componentVersion)
	}
	if !versionExists {
		log.Info("Run 'odo catalog list components' to see a list of supported component type versions")
		return fmt.Errorf("Invalid component version %s:%s", componentType, componentVersion)
	}

	// Validate component name
	err = validation.ValidateName(co.CreateArgs.Name)
	if err != nil {
		return errors.Wrapf(err, "failed to create component of name %s", co.CreateArgs.Name)
	}

	exists, err = component.Exists(co.Context.Client, co.CreateArgs.Name, co.Context.Application)
	if err != nil {
		return errors.Wrapf(err, "failed to check if component of name %s exists in application %s", co.CreateArgs.Name, co.Context.Application)
	}
	if exists {
		return fmt.Errorf("component with name %s already exists in application %s", co.CreateArgs.Name, co.Context.Application)
	}

	return
}

// createComponent creates the component
func (co *CreateOptions) createComponent(stdout io.Writer) (err error) {
	log.Successf("Initializing '%s' component", co.CreateArgs.Name)
	switch co.CreateArgs.SourceType {
	case occlient.GIT:
		// Use Git
		if err = component.CreateFromGit(
			co.Context.Client,
			co.CreateArgs,
		); err != nil {
			return errors.Wrapf(err, "failed to create component with args %+v", co.CreateArgs)
		}
		// Git is the only one using BuildConfig since we need to retrieve the git
		if err = component.Build(co.Context.Client, co.CreateArgs.Name, co.CreateArgs.ApplicationName, co.wait, stdout); err != nil {
			return errors.Wrapf(err, "failed to build component with args %+v", co)
		}
	case occlient.LOCAL:
		fileInfo, err := os.Stat(co.CreateArgs.SourcePath)
		if err != nil {
			return errors.Wrapf(err, "failed to get info of path %+v of component %+v", co.CreateArgs.SourcePath, co.CreateArgs)
		}
		if !fileInfo.IsDir() {
			return fmt.Errorf("component creation with args %+v as path needs to be a directory", co.CreateArgs)
		}
		// Create
		if err = component.CreateFromPath(co.Context.Client, co.CreateArgs); err != nil {
			return errors.Wrapf(err, "failed to create component with args %+v", co.CreateArgs)
		}
	case occlient.BINARY:
		if err = component.CreateFromPath(co.Context.Client, co.CreateArgs); err != nil {
			return errors.Wrapf(err, "failed to create component with args %+v", co.CreateArgs)
		}
	default:
		// If the user does not provide anything (local, git or binary), use the current absolute path and deploy it
		co.CreateArgs.SourceType = occlient.LOCAL
		dir, err := os.Getwd()
		if err != nil {
			return errors.Wrapf(err, "cannot create component with current path as local source path since no component source details are passed")
		}
		co.CreateArgs.SourcePath = dir
		if err = component.CreateFromPath(co.Context.Client, co.CreateArgs); err != nil {
			return errors.Wrapf(err, "")
		}
	}
	return
}

// Run has the logic to perform the required actions as part of command
func (co *CreateOptions) Run() (err error) {
	stdout := color.Output
	glog.V(4).Infof("Component create called with args: %#v", co.CreateArgs)

	err = co.createComponent(stdout)
	if err != nil {
		return err
	}

	ports, err := component.GetComponentPorts(co.Context.Client, co.CreateArgs.Name, co.Context.Application)
	if err != nil {
		return errors.Wrapf(err, "error getting ports for component with details %+v", co.CreateArgs)
	}

	if len(ports) > 1 {
		log.Successf("Component '%s' was created and ports %s were opened", co.CreateArgs.Name, strings.Join(ports, ","))
	} else if len(ports) == 1 {
		log.Successf("Component '%s' was created and port %s was opened", co.CreateArgs.Name, ports[0])
	}

	// after component is successfully created, set is as active
	if err = application.SetCurrent(co.Context.Client, co.Context.Application); err != nil {
		return errors.Wrapf(err, "failed to set %s application as current", co.Context.Application)
	}
	if err = component.SetCurrent(co.CreateArgs.Name, co.Context.Application, co.Context.Project); err != nil {
		return errors.Wrapf(err, "failed to set %s as current component", co.CreateArgs.Name)
	}
	log.Successf("Component '%s' is now set as active component", co.CreateArgs.Name)

	if len(co.componentGit) == 0 {
		log.Info("To push source code to the component run 'odo push'")
	}

	if !co.wait {
		log.Info("This may take a few moments to be ready")
	}

	return
}

// The general cpu/memory is used as a fallback when it's set and both min-cpu/memory max-cpu/memory are not set
// when the only thing specified is the min or max value, we exit the application
func ensureAndLogProperResourceUsage(resource, resourceMin, resourceMax, resourceName string) {
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
	componentCreateCmd.Flags().StringVarP(&co.componentBinary, "binary", "b", "", "Use a binary as the source file for the component")
	componentCreateCmd.Flags().StringVarP(&co.componentGit, "git", "g", "", "Use a git repository as the source file for the component")
	componentCreateCmd.Flags().StringVarP(&co.componentGitRef, "ref", "r", "", "Use a specific ref e.g. commit, branch or tag of the git repository")
	componentCreateCmd.Flags().StringVarP(&co.componentLocal, "local", "l", "", "Use local directory as a source file for the component")
	componentCreateCmd.Flags().StringVar(&co.memory, "memory", "", "Amount of memory to be allocated to the component. ex. 100Mi")
	componentCreateCmd.Flags().StringVar(&co.memoryMin, "min-memory", "", "Limit minimum amount of memory to be allocated to the component. ex. 100Mi")
	componentCreateCmd.Flags().StringVar(&co.memoryMax, "max-memory", "", "Limit maximum amount of memory to be allocated to the component. ex. 100Mi")
	componentCreateCmd.Flags().StringVar(&co.cpu, "cpu", "", "Amount of cpu to be allocated to the component. ex. 100m or 0.1")
	componentCreateCmd.Flags().StringVar(&co.cpuMin, "min-cpu", "", "Limit minimum amount of cpu to be allocated to the component. ex. 100m")
	componentCreateCmd.Flags().StringVar(&co.cpuMax, "max-cpu", "", "Limit maximum amount of cpu to be allocated to the component. ex. 1")
	componentCreateCmd.Flags().StringSliceVarP(&co.componentPorts, "port", "p", []string{}, "Ports to be used when the component is created (ex. 8080,8100/tcp,9100/udp)")
	componentCreateCmd.Flags().StringSliceVar(&co.componentEnvVars, "env", []string{}, "Environmental variables for the component. For example --env VariableName=Value")
	componentCreateCmd.Flags().BoolVarP(&co.wait, "wait", "w", false, "Wait until the component is ready")

	// Add a defined annotation in order to appear in the help menu
	componentCreateCmd.Annotations = map[string]string{"command": "component"}
	componentCreateCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(componentCreateCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(componentCreateCmd)

	completion.RegisterCommandHandler(componentCreateCmd, completion.CreateCompletionHandler)
	completion.RegisterCommandFlagHandler(componentCreateCmd, "local", completion.FileCompletionHandler)
	completion.RegisterCommandFlagHandler(componentCreateCmd, "binary", completion.FileCompletionHandler)

	return componentCreateCmd
}
