package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/redhat-developer/ocdev/pkg/component"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// componentCmd represents the component command
var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Components of application.",
	// 'ocdev component' is the same as 'ocdev component get'
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 && args[0] != "get" && args[0] != "set" {
			componentSetCmd.Run(cmd, args)
		} else {
			componentGetCmd.Run(cmd, args)
		}
	},
}

var componentGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get currently active component",
	Long:  "Get currently active component.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("component get called")
		client := getOcClient()
		component, err := component.GetCurrent(client)
		if err != nil {
			fmt.Println(errors.Wrap(err, "unable to get current component"))
			os.Exit(1)
		}
		if componentShortFlag {
			fmt.Print(component)
		} else {
			if component == "" {
				fmt.Printf("No component is set as current\n")
				return
			}
			fmt.Printf("The current component is: %v\n", component)
		}
	},
}

var componentSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set active component.",
	Long:  "Set component as active.",
	Example: `  # Set component named 'frontend' as active
  ocdev set component frontend
  `,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("Please provide component name")
		}
		if len(args) > 1 {
			return fmt.Errorf("Only one argument (component name) is allowed")
		}
		return nil
	}, Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		err := component.SetCurrent(client, args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Switched to component: %v\n", args[0])
	},
}

func init() {

	componentGetCmd.Flags().BoolVarP(&componentShortFlag, "short", "q", false, "If true, display only the component name")

	// add flags from 'get' to component command
	componentCmd.Flags().AddFlagSet(applicationGetCmd.Flags())

	componentCmd.AddCommand(componentGetCmd)
	componentCmd.AddCommand(componentSetCmd)

	rootCmd.AddCommand(componentCmd)
}
