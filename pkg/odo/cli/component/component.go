package component

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/odo/cli/application"
	"github.com/redhat-developer/odo/pkg/odo/cli/project"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	"github.com/golang/glog"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

var componentShortFlag bool

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
		context := genericclioptions.NewContext(cmd)
		component := context.ComponentAllowingEmpty(true)

		if componentShortFlag {
			fmt.Print(component)
		} else {
			if component == "" {
				log.Error("No component is set as current")
				return
			}
			log.Infof("The current component is: %v", component)
		}
	},
}

var componentSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set active component.",
	Long:  "Set component as active.",
	Example: `  # Set component named 'frontend' as active
  odo component set frontend
  `,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		projectName := context.Project
		applicationName := context.Application
		componentName := context.Component(args[0])

		err := component.SetCurrent(componentName, applicationName, projectName)
		odoutil.LogErrorAndExit(err, "")
		log.Infof("Switched to component: %v", componentName)
	},
}

// NewCmdComponent implements the component odo command
func NewCmdComponent() *cobra.Command {

	componentGetCmd.Flags().BoolVarP(&componentShortFlag, "short", "q", false, "If true, display only the component name")

	// add flags from 'get' to component command
	componentCmd.Flags().AddFlagSet(componentGetCmd.Flags())

	componentCmd.AddCommand(componentGetCmd)
	componentCmd.AddCommand(componentSetCmd)

	//Adding `--project` flag
	project.AddProjectFlag(componentGetCmd)
	project.AddProjectFlag(componentSetCmd)
	//Adding `--application` flag
	application.AddApplicationFlag(componentGetCmd)
	application.AddApplicationFlag(componentSetCmd)

	// Add a defined annotation in order to appear in the help menu
	componentCmd.Annotations = map[string]string{"command": "component"}
	componentCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	return componentCmd
}

// AddComponentFlag adds a `component` flag to the given cobra command
// Also adds a completion handler to the flag
func AddComponentFlag(cmd *cobra.Command) {
	cmd.Flags().String(genericclioptions.ComponentFlagName, "", "Component, defaults to active component.")
	completion.RegisterCommandFlagHandler(cmd, "component", completion.ComponentNameCompletionHandler)
}
