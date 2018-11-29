package component

import (
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe [component_name]",
	Short: "Describe the given component",
	Long:  `Describe the given component.`,
	Example: `  # Describe nodejs component,
  odo describe nodejs
	`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		context := genericclioptions.NewContext(cmd)
		client := context.Client
		applicationName := context.Application

		// If no arguments have been passed, get the current component
		// else, use the first argument and check to see if it exists
		var componentName string
		if len(args) == 0 {
			componentName = context.Component()
		} else {
			componentName = context.Component(args[0])
		}
		componentType, path, componentURL, appStore, err := component.GetComponentDesc(client, componentName, applicationName)
		util.CheckError(err, "")
		util.PrintComponentInfo(componentName, componentType, path, componentURL, appStore)
	},
}

// NewCmdDescribe implements the describe odo command
func NewCmdDescribe() *cobra.Command {
	// Add a defined annotation in order to appear in the help menu
	describeCmd.Annotations = map[string]string{"command": "component"}
	describeCmd.SetUsageTemplate(util.CmdUsageTemplate)

	//Adding `--project` flag
	addProjectFlag(describeCmd)
	//Adding `--application` flag
	genericclioptions.AddApplicationFlag(describeCmd)

	return describeCmd
}
