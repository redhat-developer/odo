package cmd

import (
	"os"

	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log [component_name]",
	Short: "Retrieve the log for the given component.",
	Long:  `Retrieve the log for the given component.`,
	Example: `  # Get the logs for the nodejs component
  odo log nodejs
	`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {

		// Retrieve stdout / io.Writer
		stdout := os.Stdout

		// Retrieve the client
		client := getOcClient()

		// Application
		currentApplication, err := application.GetCurrent(client)
		checkError(err, "")

		// Project
		currentProject := project.GetCurrent(client)

		var argComponent string

		if len(args) == 1 {
			argComponent = args[0]
		}

		// Retrieve and set the currentComponent
		currentComponent := getComponent(client, argComponent, currentApplication, currentProject)

		err = component.GetLogs(client, currentComponent, stdout)
		checkError(err, "Unable to retrieve logs, does your component exist?")
	},
}

func init() {
	// Add a defined annotation in order to appear in the help menu
	logCmd.Annotations = map[string]string{"command": "component"}
	logCmd.SetUsageTemplate(cmdUsageTemplate)

	rootCmd.AddCommand(logCmd)
}
