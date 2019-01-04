package component

import (
	"os"

	appCmd "github.com/redhat-developer/odo/pkg/odo/cli/application"
	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"

	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"

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

		context := genericclioptions.NewContext(cmd)
		client := context.Client
		applicationName := context.Application

		var argComponent string
		if len(args) == 1 {
			argComponent = args[0]
		}
		componentName := context.Component(argComponent)

		// Retrieve the log
		err := component.GetLogs(client, componentName, applicationName, logFollow, stdout)
		odoutil.LogErrorAndExit(err, "Unable to retrieve logs, does your component exist?")
	},
}

// NewCmdLog implements the log odo command
func NewCmdLog() *cobra.Command {
	logCmd.Flags().BoolVarP(&logFollow, "follow", "f", false, "Follow logs")

	// Add a defined annotation in order to appear in the help menu
	logCmd.Annotations = map[string]string{"command": "component"}
	logCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(logCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(logCmd)

	return logCmd
}
