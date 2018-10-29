package cmd

import (
	"github.com/redhat-developer/odo/pkg/odo/util"
	"os"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/spf13/cobra"
)

var (
	logFollow bool
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
		client := util.GetOcClient()

		getAndSetNamespace(client)
		applicationName := getAppName(client)

		var argComponent string

		if len(args) == 1 {
			argComponent = args[0]
		}

		// Retrieve and set the currentComponent
		currentComponent := getComponent(client, argComponent, applicationName)

		// Retrieve the log
		err := component.GetLogs(client, currentComponent, applicationName, logFollow, stdout)
		util.CheckError(err, "Unable to retrieve logs, does your component exist?")
	},
}

func init() {
	logCmd.Flags().BoolVarP(&logFollow, "follow", "f", false, "Follow logs")

	// Add a defined annotation in order to appear in the help menu
	logCmd.Annotations = map[string]string{"command": "component"}
	logCmd.SetUsageTemplate(cmdUsageTemplate)

	//Adding `--project` flag
	addProjectFlag(logCmd)
	//Adding `--application` flag
	addApplicationFlag(logCmd)

	rootCmd.AddCommand(logCmd)
}
