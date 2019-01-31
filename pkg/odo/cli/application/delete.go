package application

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var applicationDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the given application",
	Long:  "Delete the given application",
	Example: `  # Delete the application
  odo app delete myapp
	`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		projectName := context.Project
		appName := context.Application
		if len(args) == 1 {
			// If app name passed, consider it for deletion
			appName = args[0]
		}

		var confirmDeletion string

		// Print App Information which will be deleted
		err := printDeleteAppInfo(client, appName, projectName)
		util.LogErrorAndExit(err, "")
		exists, err := application.Exists(client, appName)
		util.LogErrorAndExit(err, "")
		if !exists {
			log.Errorf("Application %v in project %v does not exist", appName, projectName)
			os.Exit(1)
		}

		if applicationForceDeleteFlag {
			confirmDeletion = "y"
		} else {
			log.Askf("Are you sure you want to delete the application: %v from project: %v? [y/N]: ", appName, projectName)
			fmt.Scanln(&confirmDeletion)
		}

		if strings.ToLower(confirmDeletion) == "y" {
			err := application.Delete(client, appName)
			util.LogErrorAndExit(err, "")
			log.Infof("Deleted application: %s from project: %v", appName, projectName)
		} else {
			log.Infof("Aborting deletion of application: %v", appName)
		}
	},
}
