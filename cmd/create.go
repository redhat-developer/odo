package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/golang/glog"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/spf13/cobra"
)

var (
	componentBinary string
	componentGit    string
	componentLocal  string
	componentPorts  []string
)

var componentCreateCmd = &cobra.Command{
	Use:   "create <component_type> [component_name] [flags]",
	Short: "Create a new component",
	Long: `Create a new component to deploy on OpenShift.

If component name is not provided, component type value will be used for the name.

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

  # For more examples, visit: https://github.com/redhat-developer/odo/blob/master/docs/examples.md
  odo create python --git https://github.com/openshift/django-ex.git
	`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {

		stdout := color.Output
		glog.V(4).Infof("Component create called with args: %#v, flags: binary=%s, git=%s, local=%s", strings.Join(args, " "), componentBinary, componentGit, componentLocal)

		client := getOcClient()
		applicationName, err := application.GetCurrentOrGetCreateSetDefault(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)

		checkFlag := 0

		if len(componentBinary) != 0 {
			checkFlag++
		}
		if len(componentGit) != 0 {
			checkFlag++
		}
		if len(componentLocal) != 0 {
			checkFlag++
		}

		if checkFlag > 1 {
			fmt.Println("The source can be either --binary or --local or --git")
			os.Exit(1)
		}

		componentImageName, componentType, componentName, componentVersion := util.ParseCreateCmdArgs(args)

		// Check to see if the catalog type actually exists
		exists, err := catalog.Exists(client, componentType)
		checkError(err, "")
		if !exists {
			fmt.Printf("Invalid component type: %v\nRun 'odo catalog list components' to see a list of supported component types\n", componentType)
			os.Exit(1)
		}

		// Check to see if that particular version exists
		versionExists, err := catalog.VersionExists(client, componentType, componentVersion)
		checkError(err, "")
		if !versionExists {
			fmt.Printf("Invalid component version: %v\nRun 'odo catalog list components' to see a list of supported component type versions\n", componentVersion)
			os.Exit(1)
		}

		// Retrieve the componentName, if the componentName isn't specified, we will use the default image name
		if len(args) == 2 {
			componentName = args[1]
		}

		// Validate component name
		err = validateName(componentName)
		checkError(err, "")
		exists, err = component.Exists(client, componentName, applicationName, projectName)
		checkError(err, "")
		if exists {
			fmt.Printf("component with the name %s already exists in the current application\n", componentName)
			os.Exit(1)
		}

		// Deploy the component with Git
		if len(componentGit) != 0 {

			// Use Git
			err := component.CreateFromGit(client, componentName, componentImageName, componentGit, applicationName, componentPorts)
			checkError(err, "")
			fmt.Printf("Triggering build from %s.\n\n", componentGit)

			// Git is the only one using BuildConfig since we need to retrieve the git
			err = component.Build(client, componentName, applicationName, true, true, stdout)
			checkError(err, "")

		} else if len(componentLocal) != 0 {

			// Use the absolute path for the component
			dir, err := filepath.Abs(componentLocal)
			checkError(err, "")
			fileInfo, err := os.Stat(dir)
			checkError(err, "")
			if !fileInfo.IsDir() {
				fmt.Println("Please provide a path to the directory")
				os.Exit(1)
			}

			// Create
			err = component.CreateFromPath(client, componentName, componentImageName, dir, applicationName, "local", componentPorts)
			checkError(err, "")

		} else if len(componentBinary) != 0 {
			// Deploy the component with a binary

			// Retrieve the path of the binary
			path, err := filepath.Abs(componentBinary)
			checkError(err, "")

			// Create
			err = component.CreateFromPath(client, componentName, componentImageName, path, applicationName, "binary", componentPorts)
			checkError(err, "")

		} else {
			// If the user does not provide anything (local, git or binary), use the current absolute path and deploy it
			dir, err := filepath.Abs("./")
			checkError(err, "")

			// Create
			err = component.CreateFromPath(client, componentName, componentImageName, dir, applicationName, "local", componentPorts)
			checkError(err, "")
		}

		ports, err := component.GetComponentPorts(client, componentName, applicationName)
		checkError(err, "")
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
		err = component.SetCurrent(client, componentName, applicationName, projectName)
		checkError(err, "")
		fmt.Printf("\nComponent '%s' is now set as active component.\n", componentName)
	},
}

func init() {
	componentCreateCmd.Flags().StringVarP(&componentBinary, "binary", "b", "", "Use a binary as the source file for the component")
	componentCreateCmd.Flags().StringVarP(&componentGit, "git", "g", "", "Use a git repository as the source file for the component")
	componentCreateCmd.Flags().StringVarP(&componentLocal, "local", "l", "", "Use local directory as a source file for the component")
	componentCreateCmd.Flags().StringSliceVar(&componentPorts, "port", []string{}, "Ports to be used when the component is created")

	// Add a defined annotation in order to appear in the help menu
	componentCreateCmd.Annotations = map[string]string{"command": "component"}
	componentCreateCmd.SetUsageTemplate(cmdUsageTemplate)

	rootCmd.AddCommand(componentCreateCmd)
}
