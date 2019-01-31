package application

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"
)

var applicationGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the active application",
	Long:  "Get the active application",
	Example: `  # Get the currently active application
  odo app get
	`,
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		projectName := context.Project
		app := context.Application
		if applicationShortFlag {
			fmt.Print(app)
			return
		}
		if app == "" {
			log.Infof("There's no active application.\nYou can create one by running 'odo application create <name>'.")
			return
		}
		log.Infof("The current application is: %v in project: %v", app, projectName)
	},
}
