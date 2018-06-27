package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/project"
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
		client := getOcClient()
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)
		currentComponent, err := component.GetCurrent(client, applicationName, projectName)
		checkError(err, "")
		components, err := component.List(client, applicationName, projectName)
		checkError(err, "")

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

	rootCmd.AddCommand(componentListCmd)
}
