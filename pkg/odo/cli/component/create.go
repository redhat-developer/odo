package component

import (
	"os"
	"strings"

	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	"github.com/fatih/color"
	"github.com/golang/glog"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

var (
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
)

var componentCreateCmd = &cobra.Command{
	Use:   "create <component_type> [component_name] [flags]",
	Short: "Create a new component",
	Long: `Create a new component to deploy on OpenShift.

If a component name is not provided, it'll be auto-generated.

By default, builder images will be used from the current namespace. You can explicitly supply a namespace by using: odo create namespace/name:version
If version is not specified by default, latest wil be chosen as the version.

A full list of component types that can be deployed is available using: 'odo catalog list'`,
	Example: `  # Create new Node.js component with the source in current directory.
  odo create nodejs

  # A specific image version may also be specified
  odo create nodejs:latest

  # Passing memory limits
  odo create nodejs:latest --memory 150Mi
  odo create nodejs:latest --min-memory 150Mi --max-memory 300 Mi

  # Passing cpu limits
  odo create nodejs:latest --cpu 2
  odo create nodejs:latest --min-cpu 0.25 --max-cpu 2
  odo create nodejs:latest --min-cpu 200m --max-cpu 2

  # Create new Node.js component named 'frontend' with the source in './frontend' directory
  odo create nodejs frontend --local ./frontend

  # Create new Node.js component with source from remote git repository
  odo create nodejs --git https://github.com/openshift/nodejs-ex.git

  # Create new Node.js git component while specifying a ref
  odo create nodejs --git https://github.com/openshift/nodejs-ex.git --ref develop

  # Create a new Node.js component of version 6 from the 'openshift' namespace
  odo create openshift/nodejs:6 --local /nodejs-ex

  # Create new Wildfly component with binary named sample.war in './downloads' directory
  odo create wildfly wildly --binary ./downloads/sample.war

  # Create new Node.js component with the source in current directory and ports 8080-tcp,8100-tcp and 9100-udp exposed
  odo create nodejs --port 8080,8100/tcp,9100/udp

  # Create new Node.js component with the source in current directory and env variables key=value and key1=value1 exposed
  odo create nodejs --env key=value,key1=value1

  # For more examples, visit: https://github.com/redhat-developer/odo/blob/master/docs/examples.md
  odo create python --git https://github.com/openshift/django-ex.git
	`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {

		stdout := color.Output
		glog.V(4).Infof("Component create called with args: %#v, flags: binary=%s, git=%s, local=%s", strings.Join(args, " "), componentBinary, componentGit, componentLocal)

		context := genericclioptions.NewContextCreatingAppIfNeeded(cmd)
		client := context.Client
		projectName := context.Project
		applicationName := context.Application

		checkFlag := 0
		componentPath := ""
		var componentPathType occlient.CreateType

		if len(componentBinary) != 0 {
			componentPath = componentBinary
			path, err := util.GetAbsPath(componentPath)
			odoutil.LogErrorAndExit(err, "Failed to resolve %s to absolute path", componentPath)
			componentPath = path
			componentPathType = occlient.BINARY
			checkFlag++
		}
		if len(componentGit) != 0 {
			componentPath = componentGit
			componentPathType = occlient.GIT
			checkFlag++
		}
		if len(componentLocal) != 0 {
			componentPath = componentLocal
			path, err := util.GetAbsPath(componentPath)
			odoutil.LogErrorAndExit(err, "Failed to resolve %s to absolute path", componentPath)
			componentPath = path
			componentPathType = occlient.LOCAL
			checkFlag++
		}

		if checkFlag > 1 {
			log.Error("The source can be either --binary or --local or --git")
			os.Exit(1)
		}

		// if --git is not specified but --ref is still given then error has to be thrown
		if len(componentGit) == 0 && len(componentGitRef) != 0 {
			log.Errorf("The --ref flag is only valid for --git flag")
			os.Exit(1)
		}

		componentImageName, componentType, _, componentVersion := util.ParseCreateCmdArgs(args)

		// Fetch list of existing components in-order to attempt generation of unique component name
		componentList, err := component.List(client, applicationName)
		odoutil.LogErrorAndExit(err, "")

		// Generate unique name for component
		componentName, err := component.GetDefaultComponentName(
			componentPath,
			componentPathType,
			componentType,
			componentList,
		)
		odoutil.LogErrorAndExit(err, "")

		// Check to see if the catalog type actually exists
		exists, err := catalog.Exists(client, componentType)
		odoutil.LogErrorAndExit(err, "")
		if !exists {
			log.Errorf("Invalid component type: %v", componentType)
			log.Info("Run 'odo catalog list components' for a list of supported component types")
			os.Exit(1)
		}

		// Check to see if that particular version exists
		versionExists, err := catalog.VersionExists(client, componentType, componentVersion)
		odoutil.LogErrorAndExit(err, "")
		if !versionExists {
			log.Errorf("Invalid component version: %v", componentVersion)
			log.Info("Run 'odo catalog list components' to see a list of supported component type versions")
			os.Exit(1)
		}

		// Retrieve the componentName, if the componentName isn't specified, we will use the default image name
		if len(args) == 2 {
			componentName = args[1]
		}

		// Validate component name
		err = odoutil.ValidateName(componentName)
		odoutil.LogErrorAndExit(err, "")
		exists, err = component.Exists(client, componentName, applicationName)
		odoutil.LogErrorAndExit(err, "")
		if exists {
			log.Errorf("component with the name %s already exists in the current application", componentName)
			os.Exit(1)
		}

		log.Successf("Initializing '%s' component", componentName)
		ensureAndLogProperResourceUsage(memory, memoryMin, memoryMax, "memory")

		ensureAndLogProperResourceUsage(cpu, cpuMin, cpuMax, "cpu")

		resourceQuantity := []util.ResourceRequirementInfo{}
		memoryQuantity := util.FetchResourceQuantity(corev1.ResourceMemory, memoryMin, memoryMax, memory)
		if memoryQuantity != nil {
			resourceQuantity = append(resourceQuantity, *memoryQuantity)
		}
		cpuQuantity := util.FetchResourceQuantity(corev1.ResourceCPU, cpuMin, cpuMax, cpu)
		if cpuQuantity != nil {
			resourceQuantity = append(resourceQuantity, *cpuQuantity)
		}

		// Deploy the component with Git
		if len(componentGit) != 0 {

			// Use Git
			err := component.CreateFromGit(
				client,
				occlient.CreateArgs{
					Name:       componentName,
					SourcePath: componentGit,
					SourceRef:  componentGitRef,
					SourceType: occlient.GIT,

					ImageName:       componentImageName,
					EnvVars:         componentEnvVars,
					Ports:           componentPorts,
					Resources:       resourceQuantity,
					ApplicationName: applicationName,
				},
			)
			odoutil.LogErrorAndExit(err, "")

			// Git is the only one using BuildConfig since we need to retrieve the git
			err = component.Build(client, componentName, applicationName, wait, stdout)
			odoutil.CheckError(err, "")

		} else if len(componentLocal) != 0 {
			fileInfo, err := os.Stat(componentPath)
			odoutil.LogErrorAndExit(err, "")
			if !fileInfo.IsDir() {
				log.Errorf("Please provide a path to the directory")
				os.Exit(1)
			}

			// Create
			err = component.CreateFromPath(
				client,
				occlient.CreateArgs{
					Name:            componentName,
					SourcePath:      componentPath,
					SourceType:      occlient.LOCAL,
					ImageName:       componentImageName,
					EnvVars:         componentEnvVars,
					Ports:           componentPorts,
					Resources:       resourceQuantity,
					ApplicationName: applicationName,
					Wait:            wait,
				},
			)
			odoutil.LogErrorAndExit(err, "")

		} else if len(componentBinary) != 0 {
			// Deploy the component with a binary

			// Create
			err = component.CreateFromPath(
				client,
				occlient.CreateArgs{
					Name:            componentName,
					SourcePath:      componentPath,
					SourceType:      occlient.BINARY,
					ImageName:       componentImageName,
					EnvVars:         componentEnvVars,
					Ports:           componentPorts,
					Resources:       resourceQuantity,
					ApplicationName: applicationName,
					Wait:            wait,
				},
			)
			odoutil.LogErrorAndExit(err, "")

		} else {
			// If the user does not provide anything (local, git or binary), use the current absolute path and deploy it
			dir, err := util.GetAbsPath("./")
			odoutil.LogErrorAndExit(err, "")

			// Create
			err = component.CreateFromPath(
				client,
				occlient.CreateArgs{
					Name:            componentName,
					SourcePath:      dir,
					SourceType:      occlient.LOCAL,
					ImageName:       componentImageName,
					EnvVars:         componentEnvVars,
					Ports:           componentPorts,
					Resources:       resourceQuantity,
					ApplicationName: applicationName,
					Wait:            wait,
				},
			)
			odoutil.LogErrorAndExit(err, "")
		}

		ports, err := component.GetComponentPorts(client, componentName, applicationName)
		odoutil.LogErrorAndExit(err, "")

		if len(ports) > 1 {
			log.Successf("Component '%s' was created and ports %s were opened", componentName, strings.Join(ports, ","))
		} else if len(ports) == 1 {
			log.Successf("Component '%s' was created and port %s was opened", componentName, ports[0])
		}

		// after component is successfully created, set is as active
		err = application.SetCurrent(client, applicationName)
		odoutil.LogErrorAndExit(err, "")
		err = component.SetCurrent(componentName, applicationName, projectName)
		odoutil.LogErrorAndExit(err, "")
		log.Successf("Component '%s' is now set as active component", componentName)

		if len(componentGit) == 0 {
			log.Info("To push source code to the component run 'odo push'")
		}

		if !wait {
			log.Info("This may take few moments to be ready\n")
		}
	},
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
func NewCmdCreate() *cobra.Command {
	componentCreateCmd.Flags().StringVarP(&componentBinary, "binary", "b", "", "Use a binary as the source file for the component")
	componentCreateCmd.Flags().StringVarP(&componentGit, "git", "g", "", "Use a git repository as the source file for the component")
	componentCreateCmd.Flags().StringVarP(&componentGitRef, "ref", "r", "", "Use a specific ref e.g. commit, branch or tag of the git repository")
	componentCreateCmd.Flags().StringVarP(&componentLocal, "local", "l", "", "Use local directory as a source file for the component")
	componentCreateCmd.Flags().StringVar(&memory, "memory", "", "Amount of memory to be allocated to the component. ex. 100Mi")
	componentCreateCmd.Flags().StringVar(&memoryMin, "min-memory", "", "Limit minimum amount of memory to be allocated to the component. ex. 100Mi")
	componentCreateCmd.Flags().StringVar(&memoryMax, "max-memory", "", "Limit maximum amount of memory to be allocated to the component. ex. 100Mi")
	componentCreateCmd.Flags().StringVar(&cpu, "cpu", "", "Amount of cpu to be allocated to the component. ex. 100m or 0.1")
	componentCreateCmd.Flags().StringVar(&cpuMin, "min-cpu", "", "Limit minimum amount of cpu to be allocated to the component. ex. 100m")
	componentCreateCmd.Flags().StringVar(&cpuMax, "max-cpu", "", "Limit maximum amount of cpu to be allocated to the component. ex. 1")
	componentCreateCmd.Flags().StringSliceVarP(&componentPorts, "port", "p", []string{}, "Ports to be used when the component is created (ex. 8080,8100/tcp,9100/udp)")
	componentCreateCmd.Flags().StringSliceVar(&componentEnvVars, "env", []string{}, "Environmental variables for the component. For example --env VariableName=Value")
	componentCreateCmd.Flags().BoolVarP(&wait, "wait", "w", false, "Wait until the component is ready")

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

	completion.RegisterCommandHandler(updateCmd, completion.ComponentNameCompletionHandler)
	completion.RegisterCommandFlagHandler(updateCmd, "local", completion.FileCompletionHandler)
	completion.RegisterCommandFlagHandler(updateCmd, "binary", completion.FileCompletionHandler)

	completion.RegisterCommandHandler(componentSetCmd, completion.ComponentNameCompletionHandler)
	completion.RegisterCommandHandler(componentDeleteCmd, completion.ComponentNameCompletionHandler)
	completion.RegisterCommandHandler(describeCmd, completion.ComponentNameCompletionHandler)
	completion.RegisterCommandHandler(watchCmd, completion.ComponentNameCompletionHandler)
	completion.RegisterCommandHandler(logCmd, completion.ComponentNameCompletionHandler)
	completion.RegisterCommandHandler(pushCmd, completion.ComponentNameCompletionHandler)

	return componentCreateCmd
}
