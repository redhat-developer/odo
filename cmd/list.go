package cmd

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/spf13/cobra"
)

var componentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all components in the current application",
	Long:  "List all components in the current application.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)

		components, err := component.List(client, applicationName, projectName)
		checkError(err, "")

		if len(components) == 0 {
			fmt.Println("There are no components deployed.")
			return
		}

		fmt.Println("You have deployed:")
		for _, comp := range components {
			fmt.Printf("%s using the %s component\n", comp.Name, comp.Type)
		}

	},
}

func init() {
	rootCmd.AddCommand(componentListCmd)
}
