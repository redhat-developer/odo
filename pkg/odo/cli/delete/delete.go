package delete

import (
	"github.com/redhat-developer/odo/pkg/odo/cli/delete/component"
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

	componentCmd := component.NewCmdComponent(component.ComponentRecommendedCommandName,
		util.GetFullName(fullName, component.ComponentRecommendedCommandName))
	deleteCmd.AddCommand(componentCmd)
	deleteCmd.Annotations = map[string]string{"command": "main"}
	deleteCmd.SetUsageTemplate(util.CmdUsageTemplate)

	return deleteCmd
}
