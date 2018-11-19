package component

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	"github.com/fatih/color"
	"github.com/golang/glog"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
)

var (
	componentBinary  string
	componentGit     string
	componentLocal   string
	componentPorts   []string
	componentEnvVars []string
	memoryMax        string
	memoryMin        string
	memory           string
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

  # Create new Node.js component named 'frontend' with the source in './frontend' directory
  odo create nodejs frontend --local ./frontend

  # Create new Node.js component with source from remote git repository.
  odo create nodejs --git https://github.com/openshift/nodejs-ex.git

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
		var componentPathType component.CreateType

		if len(componentBinary) != 0 {
			componentPath = componentBinary
			componentPathType = component.BINARY
			checkFlag++
		}
		if len(componentGit) != 0 {
			componentPath = componentGit
			componentPathType = component.GIT
			checkFlag++
		}
		if len(componentLocal) != 0 {
			componentPath = componentLocal
			componentPathType = component.SOURCE
			checkFlag++
		}

		if checkFlag > 1 {
			fmt.Println("The source can be either --binary or --local or --git")
			os.Exit(1)
		}

		componentImageName, componentType, _, componentVersion := util.ParseCreateCmdArgs(args)

		// Fetch list of existing components in-order to attempt generation of unique component name
		componentList, err := component.List(client, applicationName)
		odoutil.CheckError(err, "")

		// Generate unique name for component
		componentName, err := component.GetDefaultComponentName(
			componentPath,
			componentPathType,
			componentType,
			componentList,
		)
		odoutil.CheckError(err, "")

		// Check to see if the catalog type actually exists
		exists, err := catalog.Exists(client, componentType)
		odoutil.CheckError(err, "")
		if !exists {
			fmt.Printf("Invalid component type: %v\nRun 'odo catalog list components' to see a list of supported component types\n", componentType)
			os.Exit(1)
		}

		// Check to see if that particular version exists
		versionExists, err := catalog.VersionExists(client, componentType, componentVersion)
		odoutil.CheckError(err, "")
		if !versionExists {
			fmt.Printf("Invalid component version: %v\nRun 'odo catalog list components' to see a list of supported component type versions\n", componentVersion)
			os.Exit(1)
		}

		// Retrieve the componentName, if the componentName isn't specified, we will use the default image name
		if len(args) == 2 {
			componentName = args[1]
		}

		// Validate component name
		err = odoutil.ValidateName(componentName)
		odoutil.CheckError(err, "")
		exists, err = component.Exists(client, componentName, applicationName)
		odoutil.CheckError(err, "")
		if exists {
			fmt.Printf("component with the name %s already exists in the current application\n", componentName)
			os.Exit(1)
		}

		// If min-memory, max-memory and memory are passed, memory will be ignored as the other 2 have greater precedence.
		// Emit a message indicating the same
		if memoryMin != "" && memoryMax != "" {
			if memory != "" {
				fmt.Printf("%s will be ignored as minimum memory %s and maximum memory %s passed carry greater precedence\n", memory, memoryMin, memoryMax)
			}
		}

		memoryQuantity := util.FetchResourceQunatity(corev1.ResourceMemory, memoryMin, memoryMax, memory)
		resourceQuantity := []util.ResourceRequirementInfo{}
		if memoryQuantity != nil {
			resourceQuantity = append(resourceQuantity, *memoryQuantity)
		}

		// Deploy the component with Git
		if len(componentGit) != 0 {

			// Use Git
			err := component.CreateFromGit(client, componentName, componentImageName, componentGit, applicationName, componentPorts, componentEnvVars, resourceQuantity)
			odoutil.CheckError(err, "")
			fmt.Printf("Triggering build from %s.\n\n", componentGit)

			// Git is the only one using BuildConfig since we need to retrieve the git
			err = component.Build(client, componentName, applicationName, true, true, stdout)
			odoutil.CheckError(err, "")

		} else if len(componentLocal) != 0 {

			// Use the absolute path for the component
			dir, err := filepath.Abs(componentLocal)
			odoutil.CheckError(err, "")
			fileInfo, err := os.Stat(dir)
			odoutil.CheckError(err, "")
			if !fileInfo.IsDir() {
				fmt.Println("Please provide a path to the directory")
				os.Exit(1)
			}

			// Create
			err = component.CreateFromPath(client, componentName, componentImageName, dir, applicationName, "local", componentPorts, componentEnvVars, resourceQuantity)
			odoutil.CheckError(err, "")

		} else if len(componentBinary) != 0 {
			// Deploy the component with a binary

			// Retrieve the path of the binary
			path, err := filepath.Abs(componentBinary)
			odoutil.CheckError(err, "")

			// Create
			err = component.CreateFromPath(client, componentName, componentImageName, path, applicationName, "binary", componentPorts, componentEnvVars, resourceQuantity)
			odoutil.CheckError(err, "")

		} else {
			// If the user does not provide anything (local, git or binary), use the current absolute path and deploy it
			dir, err := filepath.Abs("./")
			odoutil.CheckError(err, "")

			// Create
			err = component.CreateFromPath(client, componentName, componentImageName, dir, applicationName, "local", componentPorts, componentEnvVars, resourceQuantity)
			odoutil.CheckError(err, "")
		}

		ports, err := component.GetComponentPorts(client, componentName, applicationName)
		odoutil.CheckError(err, "")
		fmt.Printf("Component '%s' was created", componentName)

		if len(ports) > 1 {
			fmt.Printf(" and ports %s were opened\n", strings.Join(ports, ","))
		} else if len(ports) == 1 {
			fmt.Printf(" and port %s was opened\n", ports[0])
		}

		if len(componentGit) == 0 {
			fmt.Printf("To push source code to the component run 'odo push'\n")
		}
		// after component is successfully created, set is as active
		err = application.SetCurrent(client, applicationName)
		odoutil.CheckError(err, "")
		err = component.SetCurrent(componentName, applicationName, projectName)
		odoutil.CheckError(err, "")
		fmt.Printf("\nComponent '%s' is now set as active component.\n", componentName)
	},
}

func NewCmdCreate() *cobra.Command {
	componentCreateCmd.Flags().StringVarP(&componentBinary, "binary", "b", "", "Use a binary as the source file for the component")
	componentCreateCmd.Flags().StringVarP(&componentGit, "git", "g", "", "Use a git repository as the source file for the component")
	componentCreateCmd.Flags().StringVarP(&componentLocal, "local", "l", "", "Use local directory as a source file for the component")
	componentCreateCmd.Flags().StringVar(&memory, "memory", "", "Amount of memory to be allocated to the component container. Ex: 100Mi")
	componentCreateCmd.Flags().StringVar(&memoryMin, "min-memory", "", "Limit minimum amount of memory to be allocated to the component container. Ex: 100Mi")
	componentCreateCmd.Flags().StringVar(&memoryMax, "max-memory", "", "Limit maximum amount of memory to be allocated to the component container. Ex: 100Mi")
	componentCreateCmd.Flags().StringSliceVarP(&componentPorts, "port", "p", []string{}, "Ports to be used when the component is created (ex. 8080,8100/tcp,9100/udp")
	componentCreateCmd.Flags().StringSliceVar(&componentEnvVars, "env", []string{}, "Environmental variables for the component. For example --env VariableName=Value")

	// Add a defined annotation in order to appear in the help menu
	componentCreateCmd.Annotations = map[string]string{"command": "component"}
	componentCreateCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	//Adding `--project` flag
	addProjectFlag(componentCreateCmd)
	//Adding `--application` flag
	genericclioptions.AddApplicationFlag(componentCreateCmd)

	completion.RegisterCommandHandler(componentCreateCmd, completion.CreateCompletionHandler)
	completion.RegisterCommandFlagHandler(componentCreateCmd, "local", completion.FileCompletionHandler)
	completion.RegisterCommandFlagHandler(componentCreateCmd, "binary", completion.FileCompletionHandler)

	return componentCreateCmd
}
