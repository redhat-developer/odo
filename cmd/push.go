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

var pushCmd = &cobra.Command{
	Use:   "push [component name]",
	Short: "Push source code to component",
	Long:  `Push source code to component.`,
	Example: `  # Push source code to the to the current component
  odo push

  # Push data to the current component from original source.
  odo push

  # Push source code in ~/home/mycode to component called my-component
  odo push my-component --local ~/home/mycode
	`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)

		var componentName string
		if len(args) == 0 {
			var err error
			log.Debug("No component name passed, assuming current component")
			componentName, err = component.GetCurrent(client, applicationName, projectName)
			checkError(err, "unable to get current component")
			if componentName == "" {
				fmt.Println("No component is set as active.")
				fmt.Println("Use 'odo component set <component name> to set and existing component as active or call this command with component name as and argument.")
				os.Exit(1)
			}
		} else {
			componentName = args[0]
		}
		fmt.Printf("Pushing changes to component: %v\n", componentName)

		sourceType, sourcePath, err := component.GetComponentSource(client, componentName, applicationName, projectName)
		checkError(err, "unable to get current component")

		if sourceType == "git" {
			// currently we don't support changing build type
			// it doesn't make sense to use --dir with git build
			if len(componentLocal) != 0 {
				fmt.Println("unable to push local directory to component that uses git repository as source")
				os.Exit(1)
			}
			err := component.RebuildGit(client, componentName)
			checkError(err, fmt.Sprintf("failed to push component: %v", componentName))
		} else if sourceType == "binary" || sourceType == "local" {
			// use value of '--dir' as source if it was used
			if len(componentLocal) != 0 {
				sourcePath = componentLocal
			}
			u, err := url.Parse(sourcePath)
			checkError(err, fmt.Sprintf("unable to parse source %s from component %s", sourcePath, componentName))

			if u.Scheme != "" && u.Scheme != "file" {
				fmt.Printf("Component %s has invalid source path %s", componentName, u.Scheme)
				os.Exit(1)
			}
			_, err = os.Stat(u.Path)
			if err != nil {
				checkError(err, "")
			}

			if sourceType == "local" {
				err = component.PushLocal(client, componentName, u.Path, false)
				checkError(err, fmt.Sprintf("failed to push component: %v", componentName))
			} else if sourceType == "binary" {
				err = component.PushLocal(client, componentName, u.Path, true)
				checkError(err, fmt.Sprintf("failed to push component: %v", componentName))
			}
		}

		fmt.Printf("changes successfully pushed to component: %v\n", componentName)
	},
}

func init() {
	pushCmd.Flags().StringVar(&componentLocal, "local", "", "Use given local directory as a source for component. (It must be a local component)")

	rootCmd.AddCommand(pushCmd)
}
