package component

import (
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

// RecommendedComponentCommandName is the recommended component command name
const RecommendedComponentCommandName = "component"

// NewCmdComponent implements the component odo command
func NewCmdComponent(name, fullName string) *cobra.Command {

	componentGetCmd := NewCmdGet(RecommendedGetCommandName, odoutil.GetFullName(fullName, RecommendedGetCommandName))
	componentSetCmd := NewCmdSet(RecommendedSetCommandName, odoutil.GetFullName(fullName, RecommendedSetCommandName))

	// componentCmd represents the component command
	var componentCmd = &cobra.Command{
		Use:   name,
		Short: "Components of application.",
		/*Example: fmt.Sprintf("%s\n%s",
		componentGetCmd.Example,
		componentSetCmd.Example),*/
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

	componentCmd.AddCommand(componentGetCmd)

	// add flags from 'get' to component command
	componentCmd.Flags().AddFlagSet(componentGetCmd.Flags())
	componentCmd.AddCommand(componentSetCmd)

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
