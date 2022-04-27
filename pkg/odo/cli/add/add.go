package add

import (
	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/odo/util"
)

// RecommendedCommandName is the recommended create command name
const RecommendedCommandName = "add"

// NewCmdDelete implements the delete odo command
func NewCmdAdd(name, fullName string) *cobra.Command {
	var createCmd = &cobra.Command{
		Use:   name,
		Short: "Add resources to devfile",
	}

	bindingCmd := NewCmdBinding(BindingRecommendedCommandName, util.GetFullName(fullName, BindingRecommendedCommandName))
	createCmd.AddCommand(bindingCmd)
	createCmd.Annotations = map[string]string{"command": "main"}
	createCmd.SetUsageTemplate(util.CmdUsageTemplate)

	return createCmd
}
