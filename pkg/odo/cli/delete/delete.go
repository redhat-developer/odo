package delete

import (
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended delete command name
const RecommendedCommandName = "delete"

// NewCmdDelete implements the delete odo command
func NewCmdDelete(name, fullName string) *cobra.Command {
	var deleteCmd = &cobra.Command{
		Use:   name,
		Short: "Delete resources",
	}

	componentCmd := NewCmdComponent(ComponentRecommendedCommandName, util.GetFullName(fullName, ComponentRecommendedCommandName))
	deleteCmd.AddCommand(componentCmd)
	deleteCmd.SetUsageTemplate(util.CmdUsageTemplate)

	return deleteCmd
}
