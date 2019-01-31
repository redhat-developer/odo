package application

import (
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/validation"
	"github.com/spf13/cobra"
)

var applicationCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an application",
	Long: `Create an application.
If no app name is passed, a default app name will be auto-generated.
	`,
	Example: `  # Create an application
  odo app create myapp
  odo app create
	`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		projectName := context.Project

		var appName string
		if len(args) == 1 {
			// The only arg passed is the app name
			appName = args[0]
		} else {
			// Desired app name is not passed so, generate a new app name
			// Fetch existing list of apps
			apps, err := application.List(client)
			util.LogErrorAndExit(err, "")

			// Generate a random name that's not already in use for the existing apps
			appName, err = application.GetDefaultAppName(apps)
			util.LogErrorAndExit(err, "")
		}
		// validate application name
		err := validation.ValidateName(appName)
		util.LogErrorAndExit(err, "")
		log.Progressf("Creating application: %v in project: %v", appName, projectName)
		err = application.Create(client, appName)
		util.LogErrorAndExit(err, "")
		err = application.SetCurrent(client, appName)

		// TODO: updating the app name should be done via SetCurrent and passing the Context
		// not strictly needed here but Context should stay in sync
		context.Application = appName

		util.LogErrorAndExit(err, "")
		log.Infof("Switched to application: %v in project: %v", appName, projectName)
	},
}
