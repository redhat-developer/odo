package cmd

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/project"
	"github.com/spf13/cobra"
)

// componentCmd represents the component command
var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Components of application.",
	Example: fmt.Sprintf("%s\n%s",
		componentGetCmd.Example,
		componentSetCmd.Example),
	// 'odo component' is the same as 'odo component get'
	// 'odo component <component_name>' is the same as 'odo component set <component_name>'
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
	Example: `  # Get the currently active component
  odo component get
	`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		glog.V(4).Infof("component get called")
		client := getOcClient()
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)

		component, err := component.GetCurrent(client, applicationName, projectName)
		checkError(err, "unable to get current component")
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
  odo set component frontend
  `,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		applicationName, err := application.GetCurrent(client)
		checkError(err, "")
		projectName := project.GetCurrent(client)

		exists, err := component.Exists(client, args[0], applicationName, projectName)
		checkError(err, "")
		if !exists {
			fmt.Printf("Component %s does not exist in the current application\n", args[0])
			os.Exit(1)
		}

		err = component.SetCurrent(client, args[0], applicationName, projectName)
		checkError(err, "")
		fmt.Printf("Switched to component: %v\n", args[0])
	},
}

func init() {

	componentGetCmd.Flags().BoolVarP(&componentShortFlag, "short", "q", false, "If true, display only the component name")

	// add flags from 'get' to component command
	componentCmd.Flags().AddFlagSet(applicationGetCmd.Flags())

	componentCmd.AddCommand(componentGetCmd)
	componentCmd.AddCommand(componentSetCmd)

	// Add a defined annotation in order to appear in the help menu
	componentCmd.Annotations = map[string]string{"command": "component"}
	componentCmd.SetUsageTemplate(cmdUsageTemplate)

	rootCmd.AddCommand(componentCmd)
}
