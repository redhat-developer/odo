package component

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"strings"

	"github.com/golang/glog"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/spf13/cobra"
)

var componentForceDeleteFlag bool

var componentDeleteCmd = &cobra.Command{
	Use:   "delete <component_name>",
	Short: "Delete an existing component",
	Long:  "Delete an existing component.",
	Example: `  # Delete component named 'frontend'. 
  odo delete frontend
	`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		glog.V(4).Infof("component delete called")
		glog.V(4).Infof("args: %#v", strings.Join(args, " "))

		context := genericclioptions.NewContext(cmd)
		client := context.Client
		projectName := context.Project
		applicationName := context.Application

		// If no arguments have been passed, get the current component
		// else, use the first argument and check to see if it exists
		var componentName string
		if len(args) == 0 {
			componentName = context.Component()
		} else {
			componentName = context.Component(args[0])
		}

		var confirmDeletion string
		if componentForceDeleteFlag {
			confirmDeletion = "y"
		} else {
			fmt.Printf("Are you sure you want to delete %v from %v? [y/N] ", componentName, applicationName)
			fmt.Scanln(&confirmDeletion)
		}

		if strings.ToLower(confirmDeletion) == "y" {
			err := component.Delete(client, componentName, applicationName)
			util.CheckError(err, "")
			fmt.Printf("Component %s from application %s has been deleted\n", componentName, applicationName)

			currentComponent, err := component.GetCurrent(applicationName, projectName)
			util.CheckError(err, "Unable to get current component")

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

// NewCmdDelete implements the delete odo command
func NewCmdDelete() *cobra.Command {
	componentDeleteCmd.Flags().BoolVarP(&componentForceDeleteFlag, "force", "f", false, "Delete component without prompting")

	componentDeleteCmd.SetUsageTemplate(util.CmdUsageTemplate)

	//Adding `--project` flag
	addProjectFlag(componentDeleteCmd)
	//Adding `--application` flag
	genericclioptions.AddApplicationFlag(componentDeleteCmd)

	return componentDeleteCmd
}
