package application

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

var applicationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all applications in the current project",
	Long:  "List all applications in the current project.",
	Example: `  # List all applications in the current project
  odo app list

  # List all applications in the specified project
  odo app list --project myproject
	`,
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		projectName := context.Project

		apps, err := application.ListInProject(client)
		util.LogErrorAndExit(err, "unable to get list of applications")
		if len(apps) > 0 {
			log.Infof("The project '%v' has the following applications:", projectName)
			tabWriter := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
			fmt.Fprintln(tabWriter, "ACTIVE", "\t", "NAME")
			for _, app := range apps {
				activeMark := " "
				if app.Active {
					activeMark = "*"
				}
				fmt.Fprintln(tabWriter, activeMark, "\t", app.Name)
			}
			tabWriter.Flush()
		} else {
			log.Infof("There are no applications deployed in the project '%v'.", projectName)
		}
	},
}
