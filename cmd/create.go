package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/project"
	log "github.com/sirupsen/logrus"
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

A full list of component types that can be deployed is available using: 'odo component list'`,
	Example: `  # Create new Node.js component with the source in current directory. 
  odo create nodejs

  # A specific image version may also be specified
  odo create nodejs:latest

  # Create new Node.js component named 'frontend' with the source in './frontend' directory
  odo create nodejs frontend --local ./frontend

  # Create new Node.js component with source from remote git repository.
  odo create nodejs --git https://github.com/openshift/nodejs-ex.git

  # Create new Wildfly component with binary named sample.war in './downloads' directory
  odo create wildfly wildly --binary ./downloads/sample.war

  # Create new Node.js component with the source in current directory and ports 8080-tcp,8100-tcp and 9100-udp exposed
  odo create nodejs --ports 8080,8100/tcp,9100/udp

  # For more examples, visit: https://github.com/redhat-developer/odo/blob/master/docs/examples.md
  odo create python --git https://github.com/openshift/django-ex.git
	`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {

		stdout := color.Output
		log.Debugf("Component create called with args: %#v, flags: binary=%s, git=%s, local=%s", strings.Join(args, " "), componentBinary, componentGit, componentLocal)

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

		// We don't have to check it anymore, Args check made sure that args has at least one item
		// and no more than two

		// "Default" values
		componentImageName := args[0]
		componentType := args[0]
		componentName := args[0]
		componentVersion := "latest"

		// Check if componentType includes ":", if so, then we need to spit it into using versions
		if strings.ContainsAny(componentImageName, ":") {
			versionSplit := strings.Split(args[0], ":")
			componentType = versionSplit[0]
			componentName = versionSplit[0]
			componentVersion = versionSplit[1]
		}

		// Check to see if the catalog type actually exists
		exists, err := catalog.Exists(client, componentType)
		checkError(err, "")
		if !exists {
			fmt.Printf("Invalid component type: %v\nRun 'odo catalog list' to see a list of supported components\n", componentType)
			os.Exit(1)
		}

		// Check to see if that particular version exists
		versionExists, err := catalog.VersionExists(client, componentType, componentVersion)
		checkError(err, "")
		if !versionExists {
			fmt.Printf("Invalid component version: %v\nRun 'odo catalog list' to see a list of supported component versions\n", componentVersion)
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

		if len(componentGit) != 0 {
			err := component.CreateFromGit(client, componentName, componentImageName, componentGit, applicationName, componentPorts)
			checkError(err, "")
			fmt.Printf("Component '%s' was created", componentName)
			fmt.Printf("Triggering build from %s.\n\n", componentGit)
			err = component.Build(client, componentName, applicationName, true, true, stdout)
			checkError(err, "")
		} else if len(componentLocal) != 0 {
			// we want to use and save absolute path for component
			dir, err := filepath.Abs(componentLocal)
			checkError(err, "")
			fileInfo, err := os.Stat(dir)
			checkError(err, "")
			if !fileInfo.IsDir() {
				fmt.Println("Please provide a path to the directory")
				os.Exit(1)
			}
			err = component.CreateFromPath(client, componentName, componentImageName, dir, applicationName, "local", componentPorts)
			checkError(err, "")
			fmt.Printf("Please wait, creating %s component ...\n", componentName)
			err = component.Build(client, componentName, applicationName, false, true, stdout)
			checkError(err, "")
			fmt.Printf("Component '%s' was created", componentName)
		} else if len(componentBinary) != 0 {
			path, err := filepath.Abs(componentBinary)
			checkError(err, "")

			err = component.CreateFromPath(client, componentName, componentImageName, path, applicationName, "binary", componentPorts)
			checkError(err, "")
			fmt.Printf("Please wait, creating %s component ...\n", componentName)
			err = component.Build(client, componentName, applicationName, false, true, stdout)
			checkError(err, "")
			fmt.Printf("Component '%s' was created", componentName)
		} else {
			// we want to use and save absolute path for component
			dir, err := filepath.Abs("./")
			checkError(err, "")
			err = component.CreateFromPath(client, componentName, componentImageName, dir, applicationName, "local", componentPorts)
			checkError(err, "")
			fmt.Printf("Please wait, creating %s component ...\n", componentName)
			err = component.Build(client, componentName, applicationName, false, true, stdout)
			checkError(err, "")
			fmt.Printf("Component '%s' was created", componentName)
		}
		ports, err := component.GetComponentPorts(client, componentName, applicationName, false)
		checkError(err, "")
		if len(ports) > 1 {
			fmt.Printf(" and ports %s were opened\n", strings.Join(ports, ","))
		} else if len(ports) == 1 {
			fmt.Printf(" and port %s was opened\n", ports[0])
		}
		fmt.Printf("To push source code to the component run 'odo push'\n")
		// after component is successfully created, set is as active
		err = component.SetCurrent(client, componentName, applicationName, projectName)
		checkError(err, "")
		fmt.Printf("\nComponent '%s' is now set as active component.\n", componentName)
	},
}

func init() {
	componentCreateCmd.Flags().StringVar(&componentBinary, "binary", "", "Binary artifact")
	componentCreateCmd.Flags().StringVar(&componentGit, "git", "", "Git source")
	componentCreateCmd.Flags().StringVar(&componentLocal, "local", "", "Use local directory as a source for component")
	componentCreateCmd.Flags().StringSliceVar(&componentPorts, "ports", []string{}, "Ports to be used when the component is created")

	// Add a defined annotation in order to appear in the help menu
	componentCreateCmd.Annotations = map[string]string{"command": "component"}
	componentCreateCmd.SetUsageTemplate(cmdUsageTemplate)

	rootCmd.AddCommand(componentCreateCmd)
}
