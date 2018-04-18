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
	Short: "Delete existing component",
	Long:  "Delete existing component.",
	Example: `  # Delete component named 'frontend'. 
  odo delete frontend
	`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("component delete called")
		log.Debugf("args: %#v", strings.Join(args, " "))
		client := getOcClient()
		componentName := args[0]
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)

		exists, err := component.Exists(client, componentName, applicationName, projectName)
		checkError(err, "")
		if !exists {
			fmt.Printf("Component with the name %s does not exist in the current application\n", componentName)
			os.Exit(1)
		}

		var confirmDeletion string
		if componentForceDeleteFlag {
			confirmDeletion = "y"
		} else {
			fmt.Printf("Are you sure you want to delete %v from %v? [y/N] ", componentName, applicationName)
			fmt.Scanln(&confirmDeletion)
		}

		if strings.ToLower(confirmDeletion) == "y" {
			output, err := component.Delete(client, componentName, applicationName, projectName)
			checkError(err, "")
			log.Debug(output)
			fmt.Printf("Component %s from application %s has been deleted\n", componentName, applicationName)

			currentComponent, err := component.GetCurrent(client, applicationName, projectName)
			if err != nil {
				fmt.Printf("Unable to get current component\n")
				os.Exit(1)
			}

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
	rootCmd.AddCommand(componentDeleteCmd)
}
