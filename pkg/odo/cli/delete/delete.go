package delete

import (
	"context"

	"github.com/redhat-developer/odo/pkg/odo/cli/delete/component"
	"github.com/redhat-developer/odo/pkg/odo/cli/delete/namespace"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended delete command name
const RecommendedCommandName = "delete"

// NewCmdDelete implements the delete odo command
func NewCmdDelete(ctx context.Context, name, fullName string) *cobra.Command {
	var deleteCmd = &cobra.Command{
		Use:   name,
		Short: "Delete resources",
	}

	componentCmd := component.NewCmdComponent(ctx, component.ComponentRecommendedCommandName,
		util.GetFullName(fullName, component.ComponentRecommendedCommandName))
	deleteCmd.AddCommand(componentCmd)

	namespaceDeleteCmd := namespace.NewCmdNamespaceDelete(namespace.RecommendedCommandName,
		util.GetFullName(fullName, namespace.RecommendedCommandName))
	deleteCmd.AddCommand(namespaceDeleteCmd)

	util.SetCommandGroup(deleteCmd, util.ManagementGroup)
	deleteCmd.SetUsageTemplate(util.CmdUsageTemplate)

	return deleteCmd
}
