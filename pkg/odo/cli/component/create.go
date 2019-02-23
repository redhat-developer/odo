package component

import (
	"fmt"
	"os"
	"strings"

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
	localConfigInfo   *config.LocalConfigInfo
	*genericclioptions.Context
	componentBinary  string
	componentGit     string
	componentGitRef  string
	componentContext string
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

var createLongDesc = ktemplates.LongDesc(`Create a configuration describing component to be deployed by  on OpenShift.

If a component name is not provided, it'll be auto-generated.

By default, builder images will be used from the current namespace. You can explicitly supply a namespace by using: odo create namespace/name:version
If version is not specified by default, latest wil be chosen as the version.

A full list of component types that can be deployed is available using: 'odo catalog list'`)

var createExample = ktemplates.Examples(`  # Create new Node.js component with the source in current directory.
%[1]s nodejs

# A specific image version may also be specified
%[1]s nodejs:latest

# Create new Node.js component named 'frontend' with the source in './frontend' directory
%[1]s nodejs frontend --context ./frontend

# Create a new Node.js component of version 6 from the 'openshift' namespace
%[1]s openshift/nodejs:6 --context /nodejs-ex

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
	cmpSrcType := config.LOCAL
	co.componentSettings.SourceType = &cmpSrcType
	co.componentSettings.Application = &(co.Context.Application)

	if len(co.componentBinary) != 0 {
		cPath, err := util.GetAbsPath(co.componentBinary)
		if err != nil {
			return err
		}
		co.componentSettings.SourceLocation = &cPath
		cmpSrcType = config.BINARY
		co.componentSettings.SourceType = &cmpSrcType
		componentCnt++
	} else if len(co.componentGit) != 0 {
		co.componentSettings.SourceLocation = &(co.componentGit)
		cmpSrcType = config.GIT
		co.componentSettings.SourceType = &cmpSrcType
		componentCnt++
	} else {
		componentCnt++
		if len(co.componentContext) > 0 {
			co.componentContext, err = util.GetAbsPath(co.componentContext)
			if err != nil {
				return errors.Wrapf(err, "please provide the context relative to your current directory")
			}
			co.componentSettings.SourceLocation = &co.componentContext
		} else {
			currDir, err := os.Getwd()
			if err != nil {
				return errors.Wrap(err, "failed to set component source location. Please pass a valid path")
			}
			co.componentSettings.SourceLocation = &currDir
		}
	}

	if componentCnt > 1 {
		return fmt.Errorf("The source can be either --binary or --local or --git")
	}

	if len(co.componentGitRef) != 0 {
		co.componentSettings.Ref = &(co.componentGitRef)
	}

	if len(co.componentGit) == 0 && len(co.componentGitRef) != 0 {
		return fmt.Errorf("The --ref flag is only valid for --git flag")
	}

	return
}

func (co *CreateOptions) setCmpName(args []string) (err error) {
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

func createDefaultComponentName(context *genericclioptions.Context, componentType string, sourceType config.SrcType, sourcePath string) (string, error) {
	// Retrieve the componentName, if the componentName isn't specified, we will use the default image name
	componentName, err := component.GetDefaultComponentName(
		sourcePath,
		sourceType,
		componentType,
		component.ComponentList{},
	)

	if err != nil {
		return "", nil
	}

	return componentName, nil
}

func (co *CreateOptions) setResourceLimits() {
	ensureAndLogProperResourceUsage(co.memory, co.memoryMin, co.memoryMax, "memory")

	ensureAndLogProperResourceUsage(co.cpu, co.cpuMin, co.cpuMax, "cpu")

	memoryQuantity := util.FetchResourceQuantity(corev1.ResourceMemory, co.memoryMin, co.memoryMax, co.memory)
	if memoryQuantity != nil {
		minMemory := memoryQuantity.MinQty.String()
		maxMemory := memoryQuantity.MaxQty.String()
		co.componentSettings.MinMemory = &minMemory
		co.componentSettings.MaxMemory = &maxMemory
	}

	cpuQuantity := util.FetchResourceQuantity(corev1.ResourceCPU, co.cpuMin, co.cpuMax, co.cpu)
	if cpuQuantity != nil {
		minCPU := cpuQuantity.MinQty.String()
		maxCPU := cpuQuantity.MaxQty.String()
		co.componentSettings.MinCPU = &minCPU
		co.componentSettings.MaxCPU = &maxCPU
	}
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
	co.componentSettings = co.localConfigInfo.GetComponentSettings()

	co.Context = genericclioptions.NewContextCreatingAppIfNeeded(cmd)

	if co.interactive {
		client := co.Client

		componentTypeCandidates, err := catalog.List(client)
		if err != nil {
			return err
		}
		componentTypeCandidates = catalogutil.FilterHiddenComponents(componentTypeCandidates)
		selectedComponentType := ui.SelectComponentType(componentTypeCandidates)
		selectedImageTag := ui.SelectImageTag(componentTypeCandidates, selectedComponentType)
		componentType := selectedComponentType + ":" + selectedImageTag
		co.componentSettings.Type = &componentType

		currentDirectory, err := os.Getwd()
		if err != nil {
			return err
		}

		co.componentContext, err = util.GetAbsPath(ui.EnterInputTypePath("context", currentDirectory, currentDirectory))
		if err != nil {
			return err
		}

		selectedSourceType := ui.SelectSourceType([]config.SrcType{config.LOCAL, config.GIT, config.BINARY})
		co.componentSettings.SourceType = &selectedSourceType
		selectedSourcePath := ""

		if selectedSourceType == config.BINARY {
			selectedSourcePath = ui.EnterInputTypePath("binary", currentDirectory)
			selectedSourcePath, err = util.GetAbsPath(selectedSourcePath)
			if err != nil {
				return err
			}
			co.componentSettings.SourceLocation = &selectedSourcePath
		} else if selectedSourceType == config.GIT {
			cmpSrcLOC, selectedGitRef := ui.EnterGitInfo()
			co.componentSettings.SourceLocation = &cmpSrcLOC
			co.componentSettings.Ref = &selectedGitRef
		} else if selectedSourceType == config.LOCAL {
			if len(co.componentContext) > 0 {
				co.componentContext, err = util.GetAbsPath(co.componentContext)
				if err != nil {
					return errors.Wrap(err, "failed to create component config")
				}
				co.componentSettings.SourceLocation = &(co.componentContext)
			} else {
				co.componentContext = currentDirectory
				co.componentSettings.SourceLocation = &currentDirectory
			}
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

				memoryQuantity := util.FetchResourceQuantity(corev1.ResourceMemory, memMin, memMax, "")
				if memoryQuantity != nil {
					co.componentSettings.MinMemory = &memMin
					co.componentSettings.MaxMemory = &memMax
				}
				cpuQuantity := util.FetchResourceQuantity(corev1.ResourceCPU, cpuMin, cpuMax, "")
				if cpuQuantity != nil {
					co.componentSettings.MinCPU = &cpuMin
					co.componentSettings.MaxCPU = &cpuMax
				}
			}
		}
	} else {
		err = co.setCmpSourceAttrs()
		if err != nil {
			return err
		}
		err = co.setCmpName(args)
		if err != nil {
			return err
		}
		co.setResourceLimits()
		if len(co.componentPorts) > 0 {
			co.componentSettings.Ports = &(co.componentPorts)
		}
	}

	co.componentSettings.Project = &(co.Context.Project)

	return
}

// Validate validates the create parameters
func (co *CreateOptions) Validate() (err error) {
	return component.ValidateComponentCreateRequest(co.Context.Client, co.componentSettings, false)
}

// Run has the logic to perform the required actions as part of command
func (co *CreateOptions) Run() (err error) {
	err = co.localConfigInfo.SetComponentSettings(co.componentSettings)
	if err != nil {
		return errors.Wrapf(err, "failed to persist the component settings to config file")
	}
	log.Infof("Please use `odo push` command to create the component with source deployed")
	return
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
	componentCreateCmd.Flags().StringVarP(&co.componentBinary, "binary", "b", "", "Use a binary as the source file for the component")
	componentCreateCmd.Flags().StringVarP(&co.componentGit, "git", "g", "", "Use a git repository as the source file for the component")
	componentCreateCmd.Flags().StringVarP(&co.componentGitRef, "ref", "r", "", "Use a specific ref e.g. commit, branch or tag of the git repository")
	componentCreateCmd.Flags().StringVar(&co.componentContext, "context", "", "Use context to indicate the path where the component settings need to be saved and this directory should contain component source for local and binary components")
	componentCreateCmd.Flags().StringVar(&co.memory, "memory", "", "Amount of memory to be allocated to the component. ex. 100Mi")
	componentCreateCmd.Flags().StringVar(&co.memoryMin, "min-memory", "", "Limit minimum amount of memory to be allocated to the component. ex. 100Mi")
	componentCreateCmd.Flags().StringVar(&co.memoryMax, "max-memory", "", "Limit maximum amount of memory to be allocated to the component. ex. 100Mi")
	componentCreateCmd.Flags().StringVar(&co.cpu, "cpu", "", "Amount of cpu to be allocated to the component. ex. 100m or 0.1")
	componentCreateCmd.Flags().StringVar(&co.cpuMin, "min-cpu", "", "Limit minimum amount of cpu to be allocated to the component. ex. 100m")
	componentCreateCmd.Flags().StringVar(&co.cpuMax, "max-cpu", "", "Limit maximum amount of cpu to be allocated to the component. ex. 1")
	componentCreateCmd.Flags().StringSliceVarP(&co.componentPorts, "port", "p", []string{}, "Ports to be used when the component is created (ex. 8080,8100/tcp,9100/udp)")
	componentCreateCmd.Flags().StringSliceVar(&co.componentEnvVars, "env", []string{}, "Environmental variables for the component. For example --env VariableName=Value")

	// Add a defined annotation in order to appear in the help menu
	componentCreateCmd.Annotations = map[string]string{"command": "component"}
	componentCreateCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(componentCreateCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(componentCreateCmd)

	completion.RegisterCommandHandler(componentCreateCmd, completion.CreateCompletionHandler)
	completion.RegisterCommandFlagHandler(componentCreateCmd, "context", completion.FileCompletionHandler)
	completion.RegisterCommandFlagHandler(componentCreateCmd, "binary", completion.FileCompletionHandler)

	return componentCreateCmd
}
