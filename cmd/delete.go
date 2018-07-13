package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/project"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	componentShortFlag       bool
	componentForceDeleteFlag bool
)

var componentDeleteCmd = &cobra.Command{
	Use:   "delete <component_name>",
	Short: "Delete an existing component",
	Long:  "Delete an existing component.",
	Example: `  # Delete component named 'frontend'. 
  odo delete frontend
	`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("component delete called")
		log.Debugf("args: %#v", strings.Join(args, " "))
		client := getOcClient()

		// Get all necessary names (current application + project)
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)

		// Get the current component if no arguments have been passed
		var componentName string

		// If no arguments have been passed, get the current component
		// else, use the first argument and check to see if it exists
		if len(args) == 0 {
			componentName = getComponent(client, "", applicationName, projectName)
		} else {

			componentName = args[0]

			// Checks to see if the component actually exists
			exists, err := component.Exists(client, componentName, applicationName, projectName)
			checkError(err, "")
			if !exists {
				fmt.Printf("Component with the name %s does not exist in the current application\n", componentName)
				os.Exit(1)
			}
		}

		var confirmDeletion string
		if componentForceDeleteFlag {
			confirmDeletion = "y"
		} else {
			fmt.Printf("Are you sure you want to delete %v from %v? [y/N] ", componentName, applicationName)
			fmt.Scanln(&confirmDeletion)
		}

		if strings.ToLower(confirmDeletion) == "y" {
			err := component.Delete(client, componentName, applicationName, projectName)
			checkError(err, "")
			fmt.Printf("Component %s from application %s has been deleted\n", componentName, applicationName)

			currentComponent, err := component.GetCurrent(client, applicationName, projectName)
			checkError(err, "Unable to get current component")

			if currentComponent == "" {
				fmt.Println("No default component has been set")
			} else {
				fmt.Printf("Default component set to: %s\n", currentComponent)
			}

		} else {
			fmt.Printf("Aborting deletion of component: %v\n", componentName)
		}
	},
}

func init() {
	componentDeleteCmd.Flags().BoolVarP(&componentForceDeleteFlag, "force", "f", false, "Delete component without prompting")

	// Add a defined annotation in order to appear in the help menu
	componentDeleteCmd.Annotations = map[string]string{"command": "component"}
	componentDeleteCmd.SetUsageTemplate(cmdUsageTemplate)

	rootCmd.AddCommand(componentDeleteCmd)
}
