package component

import (
	"fmt"
	"strings"

	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"

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
			log.Askf("Are you sure you want to delete %v from %v? [y/N]: ", componentName, applicationName)
			fmt.Scanln(&confirmDeletion)
		}

		if strings.ToLower(confirmDeletion) == "y" {
			err := component.Delete(client, componentName, applicationName)
			odoutil.CheckError(err, "")
			log.Successf("Component %s from application %s has been deleted", componentName, applicationName)

			currentComponent, err := component.GetCurrent(applicationName, projectName)
			odoutil.CheckError(err, "Unable to get current component")

			if currentComponent == "" {
				log.Info("No default component has been set")
			} else {
				log.Infof("Default component set to: %s", currentComponent)
			}

		} else {
			log.Infof("Aborting deletion of component: %v", componentName)
		}
	},
}

// NewCmdDelete implements the delete odo command
func NewCmdDelete() *cobra.Command {
	componentDeleteCmd.Flags().BoolVarP(&componentForceDeleteFlag, "force", "f", false, "Delete component without prompting")

	// Add a defined annotation in order to appear in the help menu
	componentDeleteCmd.Annotations = map[string]string{"command": "component"}
	componentDeleteCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(componentDeleteCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(componentDeleteCmd)

	return componentDeleteCmd
}
