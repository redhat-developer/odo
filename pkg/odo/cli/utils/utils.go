package utils

import (
	"fmt"

	"github.com/openshift/odo/pkg/odo/util"

	"github.com/spf13/cobra"
)

const RecommendedCommandName = "utils"

// NewCmdUtils implements the utils odo command
func NewCmdUtils(name, fullName string) *cobra.Command {
	terminalCmd := NewCmdTerminal(terminalCommandName, util.GetFullName(fullName, terminalCommandName))
	utilsCmd := &cobra.Command{
		Use:   name,
		Short: "Utilities for terminal commands and modifying Odo configurations",
		Long:  `Utilities for terminal commands and modifying Odo configurations`,
		Example: fmt.Sprintf("%s\n",
			terminalCmd.Example),
	}

	utilsCmd.Annotations = map[string]string{"command": "utility"}
	utilsCmd.SetUsageTemplate(util.CmdUsageTemplate)

	utilsCmd.AddCommand(terminalCmd)
	return utilsCmd
}

// VisitCommands visits each command within Cobra.
// Adapted from: https://github.com/cppforlife/knctl/blob/612840d3c9729b1c57b20ca0450acab0d6eceeeb/pkg/knctl/cobrautil/misc.go#L23
func VisitCommands(cmd *cobra.Command, f func(*cobra.Command)) {
	f(cmd)
	for _, child := range cmd.Commands() {
		VisitCommands(child, f)
	}
}
