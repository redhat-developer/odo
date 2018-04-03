package cmd

import (
	"fmt"
	"net/url"
	"os"

	"github.com/redhat-developer/ocdev/pkg/application"
	"github.com/redhat-developer/ocdev/pkg/component"
	"github.com/redhat-developer/ocdev/pkg/project"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [component name]",
	Short: "Push source code to component",
	Long:  `Push source code to component.`,
	Example: `  # Push source code to the to the current component
  ocdev push

  # Push data to the current component from original source.
  ocdev push

  # Push source code in ~/home/mycode to component called my-component
  ocdev push my-component --local ~/home/mycode
	`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		projectName := project.GetCurrent(client)

		applicationName, err := application.GetCurrent(client)
		if err != nil {
			fmt.Println(errors.Wrap(err, "unable to get current application"))
			os.Exit(1)
		}

		var componentName string
		if len(args) == 0 {
			var err error
			log.Debug("No component name passed, assuming current component")
			componentName, err = component.GetCurrent(client)
			if err != nil {
				fmt.Println(errors.Wrap(err, "unable to get current component"))
				os.Exit(1)
			}
			if componentName == "" {
				fmt.Println("No component is set as active.")
				fmt.Println("Use 'ocdev component set <component name> to set and existing component as active or call this command with component name as and argument.")
				os.Exit(1)
			}
		} else {
			componentName = args[0]
		}
		fmt.Printf("Pushing changes to component: %v\n", componentName)

		sourceType, sourcePath, err := component.GetComponentSource(client, componentName, applicationName, projectName)
		if err != nil {
			fmt.Println(errors.Wrap(err, "unable to get current component"))
			os.Exit(1)
		}

		switch sourceType {
		case "local":
			// use value of '--dir' as source if it was used
			if len(componentLocal) != 0 {
				sourcePath = componentLocal
			}
			u, err := url.Parse(sourcePath)
			if err != nil {
				fmt.Printf("Unable to parse source %s from component %s", sourcePath, componentName)
				os.Exit(1)
			}
			if u.Scheme != "" && u.Scheme != "file" {
				fmt.Printf("Component %s has invalid source path %s", componentName, u.Scheme)
				os.Exit(1)
			}

			if err := component.PushLocal(client, componentName, u.Path); err != nil {
				fmt.Printf("failed to push component: %v", componentName)
				os.Exit(1)
			}
		case "git":
			// currently we don't support changing build type
			// it doesn't make sense to use --dir with git build
			if len(componentLocal) != 0 {
				fmt.Println("unable to push local directory to component that uses git repository as source")
				os.Exit(1)
			}
			if err := component.RebuildGit(client, componentName); err != nil {
				fmt.Printf("failed to push component: %v", componentName)
				os.Exit(1)
			}

		}

		fmt.Printf("changes successfully pushed to component: %v\n", componentName)
	},
}

func init() {
	pushCmd.Flags().StringVar(&componentLocal, "local", "", "Use given local directory as a source for component. (It must be a local component)")

	rootCmd.AddCommand(pushCmd)
}
