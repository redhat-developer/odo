package application

import (
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	"os"
)

var applicationSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set application as active",
	Long:  "Set application as active",
	Example: `  # Set an application as active
  odo app set myapp
	`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			log.Error("Please provide application name")
			os.Exit(1)
		}
		if len(args) > 1 {
			log.Error("Only one argument (application name) is allowed")
			os.Exit(1)
		}
		return nil
	}, Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		projectName := context.Project

		// error if application does not exist
		appName := args[0]
		exists, err := application.Exists(client, appName)
		util.LogErrorAndExit(err, "unable to check if application exists")
		if !exists {
			log.Errorf("Application %v does not exist", appName)
			os.Exit(1)
		}

		err = application.SetCurrent(client, appName)
		util.LogErrorAndExit(err, "")
		log.Infof("Switched to application: %v in project: %v", args[0], projectName)

		// TODO: updating the app name should be done via SetCurrent and passing the Context
		// not strictly needed here but Context should stay in sync
		context.Application = appName
	},
}
