package utils

import (
	"fmt"

	odoutil "github.com/openshift/odo/v2/pkg/odo/util"

	"github.com/spf13/cobra"
)

const RecommendedCommandName = "utils"

// NewCmdUtils implements the utils odo command
func NewCmdUtils(name, fullName string) *cobra.Command {
	terminalCmd := NewCmdTerminal(terminalCommandName, odoutil.GetFullName(fullName, terminalCommandName))
	utilsCmd := &cobra.Command{
		Use:   name,
		Short: "Utilities for terminal commands and modifying odo configurations",
		Long:  "Utilities for terminal commands and modifying odo configurations",
		Example: fmt.Sprintf("%s\n",
			terminalCmd.Example),
	}

	utilsCmd.Annotations = map[string]string{"command": "utility"}
	utilsCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	utilsCmd.AddCommand(terminalCmd)
	return utilsCmd
}
