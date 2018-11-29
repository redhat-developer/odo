package utils

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/util"

	"github.com/spf13/cobra"
)

// utilsCmd represents the utils command
var utilsCmd = &cobra.Command{
	Use:   "utils",
	Short: "Utilities for terminal commands and modifying Odo configurations",
	Long:  `Utilities for terminal commands and modifying Odo configurations`,
	Example: fmt.Sprintf("%s\n%s\n%s",
		terminalCmd.Example,
		configurationSetCmd.Example,
		configurationViewCmd.Example),
}

// NewCmdUtils implements the utils odo command
func NewCmdUtils() *cobra.Command {
	utilsCmd.Annotations = map[string]string{"command": "utility"}
	utilsCmd.SetUsageTemplate(util.CmdUsageTemplate)

	configurationCmd.AddCommand(configurationViewCmd)
	configurationCmd.AddCommand(configurationSetCmd)

	configurationCmd.SetUsageTemplate(util.CmdUsageTemplate)
	utilsCmd.AddCommand(configurationCmd)
	utilsCmd.AddCommand(terminalCmd)
	return utilsCmd
}
