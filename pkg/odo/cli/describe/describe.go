package describe

import (
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/odo/util"

	"context"

	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended delete command name
const RecommendedCommandName = "describe"

// NewCmdDescribe implements the describe odo command
func NewCmdDescribe(ctx context.Context, name, fullName string, testClientset clientset.Clientset) *cobra.Command {
	var describeCmd = &cobra.Command{
		Use:   name,
		Short: "Describe resource",
	}

	componentCmd := NewCmdComponent(ctx, ComponentRecommendedCommandName, util.GetFullName(fullName, ComponentRecommendedCommandName), testClientset)
	bindingCmd := NewCmdBinding(BindingRecommendedCommandName, util.GetFullName(fullName, BindingRecommendedCommandName), testClientset)
	describeCmd.AddCommand(componentCmd, bindingCmd)
	util.SetCommandGroup(describeCmd, util.ManagementGroup)
	describeCmd.SetUsageTemplate(util.CmdUsageTemplate)

	return describeCmd
}
