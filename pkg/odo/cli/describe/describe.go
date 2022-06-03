package describe

import (
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended delete command name
const RecommendedCommandName = "describe"

// NewCmdDescribe implements the describe odo command
func NewCmdDescribe(name, fullName string) *cobra.Command {
	var describeCmd = &cobra.Command{
		Use:   name,
		Short: "Describe resource",
	}

	componentCmd := NewCmdComponent(ComponentRecommendedCommandName, util.GetFullName(fullName, ComponentRecommendedCommandName))
	bindingCmd := NewCmdBinding(BindingRecommendedCommandName, util.GetFullName(fullName, BindingRecommendedCommandName))
	describeCmd.AddCommand(componentCmd, bindingCmd)
	describeCmd.Annotations = map[string]string{"command": "main"}
	describeCmd.SetUsageTemplate(util.CmdUsageTemplate)

	return describeCmd
}
