package set

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/odo/cli/set/namespace"
	"github.com/redhat-developer/odo/pkg/odo/util"

	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended namespace command name
const RecommendedCommandName = "set"

// NewCmdSet implements the namespace odo command
func NewCmdSet(name, fullName string) *cobra.Command {

	namespaceSetCmd := namespace.NewCmdNamespaceSet(namespace.RecommendedCommandName,
		util.GetFullName(fullName, namespace.RecommendedCommandName))
	setCmd := &cobra.Command{
		Use:   name + " [options]",
		Short: "Perform set operation",
		Long:  "Perform set operation",
		Example: fmt.Sprintf("%s\n",
			namespaceSetCmd.Example,
		),
	}

	setCmd.AddCommand(namespaceSetCmd)

	util.SetCommandGroup(setCmd, util.ManagementGroup)
	setCmd.SetUsageTemplate(util.CmdUsageTemplate)

	return setCmd
}
