package utils

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/odo/util"

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

// VisitCommands ...
func VisitCommands(cmd *cobra.Command, f func(*cobra.Command)) {
	f(cmd)
	for _, child := range cmd.Commands() {
		VisitCommands(child, f)
	}
}
