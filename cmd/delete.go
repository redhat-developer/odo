package cmd

import (
	"fmt"
	"strings"

	"github.com/redhat-developer/ocdev/pkg/application"
	"github.com/redhat-developer/ocdev/pkg/component"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	componentShortFlag       bool
	componentForceDeleteFlag bool
)

var componentDeleteCmd = &cobra.Command{
	Use:   "delete <component_name>",
	Short: "Delete existing component",
	Long:  "Delete existing component.",
	Example: `  # Delete component named 'frontend'. 
  ocdev delete frontend
	`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("component delete called")
		log.Debugf("args: %#v", strings.Join(args, " "))
		client := getOcClient()
		componentName := args[0]
		var confirmDeletion string

		currentApp, err := application.GetCurrent(client)
		checkError(err, "")

		if componentForceDeleteFlag {
			confirmDeletion = "y"
		} else {
			fmt.Printf("Are you sure you want to delete %v from %v? [y/N] ", componentName, currentApp)
			fmt.Scanln(&confirmDeletion)
		}

		if strings.ToLower(confirmDeletion) == "y" {
			output, err := component.Delete(client, componentName)
			checkError(err, "")
			fmt.Println(output)
		} else {
			fmt.Printf("Aborting deletion of component: %v\n", componentName)
		}
	},
}

func init() {
	componentDeleteCmd.Flags().BoolVarP(&componentForceDeleteFlag, "force", "f", false, "Delete component without prompting")
	rootCmd.AddCommand(componentDeleteCmd)
}
