package cmd

import (
	"fmt"
	"os"
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
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Please specify component name to delete.")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("component delete called")
		log.Debugf("args: %#v", strings.Join(args, " "))
		client := getOcClient()
		componentName := args[0]
		var confirmDeletion string

		currentApp, err := application.GetCurrent(client)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if componentForceDeleteFlag {
			confirmDeletion = "y"
		} else {
			fmt.Printf("Are you sure you want to delete %v from %v? [y/N] ", componentName, currentApp)
			fmt.Scanln(&confirmDeletion)
		}

		if strings.ToLower(confirmDeletion) == "y" {
			output, err := component.Delete(client, componentName)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
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
