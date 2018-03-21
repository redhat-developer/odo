package cmd

import (
	"fmt"
	"os"

	"github.com/redhat-developer/ocdev/pkg/component"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push [component name]",
	Short: "Push source code to component",
	Long:  `Push source code to component.`,
	Example: `  # Push source code in current directory to current component
  ocdev push

  # Push source code in ~/home/mycode to component called my-component
  ocdev push my-component --dir ~/home/mycode
	`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("component push called")
		client := getOcClient()
		var componentName string
		if len(args) == 0 {
			var err error
			log.Debug("No component name passed, assuming current component")
			componentName, err = component.GetCurrent(client)
			if err != nil {
				fmt.Println(errors.Wrap(err, "unable to get current component"))
				os.Exit(1)
			}
		} else {
			componentName = args[0]
		}
		fmt.Printf("pushing changes to component: %v\n", componentName)

		if len(componentDir) == 0 {
			componentDir = "."
		}

		if _, err := component.Push(client, componentName, componentDir); err != nil {
			fmt.Printf("failed to push component: %v", componentName)
			os.Exit(1)
		}
		fmt.Printf("changes successfully pushed to component: %v\n", componentName)
	},
}

func init() {
	pushCmd.Flags().StringVar(&componentDir, "dir", "", "specify directory to push changes from")

	rootCmd.AddCommand(pushCmd)
}
