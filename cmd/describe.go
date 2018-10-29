package cmd

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"os"

	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe [component_name]",
	Short: "Describe the given component",
	Long:  `Describe the given component.`,
	Example: `  # Describe nodejs component,
  odo describe nodejs
	`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		client := util.GetOcClient()
		// Application
		currentApplication, err := application.GetCurrent(client)
		util.CheckError(err, "")
		// Project
		currentProject := project.GetCurrent(client)
		var currentComponent string
		if len(args) == 0 {
			var err error
			currentComponent, err = component.GetCurrent(client, currentApplication, currentProject)
			util.CheckError(err, "")
		} else {
			currentComponent = args[0]
			//Check whether component exist or not
			exists, err := component.Exists(client, currentComponent, currentApplication, currentProject)
			util.CheckError(err, "")
			if !exists {
				fmt.Printf("component with the name %s does not exist\n", currentComponent)
				os.Exit(1)
			}
		}

		componentType, path, componentURL, appStore, err := component.GetComponentDesc(client, currentComponent, currentApplication, currentProject)
		util.CheckError(err, "")
		printComponentInfo(currentComponent, componentType, path, componentURL, appStore)
	},
}

func init() {
	// Add a defined annotation in order to appear in the help menu
	describeCmd.Annotations = map[string]string{"command": "component"}
	describeCmd.SetUsageTemplate(cmdUsageTemplate)

	rootCmd.AddCommand(describeCmd)
}
