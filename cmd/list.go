package cmd

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"os"
	"text/tabwriter"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/spf13/cobra"
)

var componentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all components in the current application",
	Long:  "List all components in the current application.",
	Example: `  # List all components in the application
  odo list
	`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := util.GetOcClient()

		projectName := util.GetAndSetNamespace(client)
		applicationName := util.GetAppName(client)

		currentComponent, err := component.GetCurrent(applicationName, projectName)
		util.CheckError(err, "")
		components, err := component.List(client, applicationName)
		util.CheckError(err, "")

		if len(components) == 0 {
			fmt.Println("There are no components deployed.")
			return
		}

		activeMark := " "
		w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
		fmt.Fprintln(w, "ACTIVE", "\t", "NAME", "\t", "TYPE")
		for _, comp := range components {
			if comp.Name == currentComponent {
				activeMark = "*"
			}
			fmt.Fprintln(w, activeMark, "\t", comp.Name, "\t", comp.Type)
			activeMark = " "
		}
		w.Flush()

	},
}

func init() {
	// Add a defined annotation in order to appear in the help menu
	componentListCmd.Annotations = map[string]string{"command": "component"}

	//Adding `--project` flag
	addProjectFlag(componentListCmd)
	//Adding `--application` flag
	addApplicationFlag(componentListCmd)

	rootCmd.AddCommand(componentListCmd)
}
