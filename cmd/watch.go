package cmd

import (
	"fmt"
	"net/url"
	"os"

	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/project"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var watchLocal string

var watchCmd = &cobra.Command{
	Use:   "watch [component name]",
	Short: "Watch for changes, update component on change",
	Long:  `Watch for changes, update component on change.`,
	Example: `  # Watch for changes in directory for current component
  odo watch

  # Watch for changes in directory for component called frontend 
  odo watch frontend
	`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		stdout := os.Stdout
		client := getOcClient()
		projectName := project.GetCurrent(client)
		applicationName, err := application.GetCurrent(client)
		checkError(err, "Unable to get current application.")

		var componentName string
		if len(args) == 0 {
			var err error
			log.Debug("No component name passed, assuming current component")
			componentName, err = component.GetCurrent(client, applicationName, projectName)
			checkError(err, "")
			if componentName == "" {
				fmt.Println("No component is set as active.")
				fmt.Println("Use 'odo component set <component name> to set and existing component as active or call this command with component name as and argument.")
				os.Exit(1)
			}
		} else {
			componentName = args[0]
		}

		sourceType, sourcePath, err := component.GetComponentSource(client, componentName, applicationName, projectName)
		checkError(err, "Unable to get source for %s component.", componentName)
		if sourceType != "local" {
			fmt.Println("You use watch only with components with local source.")
			os.Exit(1)
		}

		var watchPath string
		if watchLocal != "" {
			watchPath = watchLocal
		} else {
			u, err := url.Parse(sourcePath)
			checkError(err, "Unable to parse source %s from component %s.", sourcePath, componentName)

			if u.Scheme != "" && u.Scheme != "file" {
				fmt.Printf("Component %s has invalid source path %s.", componentName, u.Scheme)
				os.Exit(1)
			}
			watchPath = u.Path
		}

		err = component.WatchAndPush(client, componentName, applicationName, watchPath, stdout)
		checkError(err, "Error while trying to watch %s", watchPath)
	},
}

func init() {
	watchCmd.Flags().StringVar(&watchLocal, "local", "", "Use given local directory as a source for component. (It must be a local component)")

	rootCmd.AddCommand(watchCmd)
}
