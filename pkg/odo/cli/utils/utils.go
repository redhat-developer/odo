package utils

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/odo/cli/utils/config"

	"github.com/redhat-developer/odo/pkg/odo/util"

	"github.com/spf13/cobra"
)

const RecommendedCommandName = "utils"

// NewCmdUtils implements the utils odo command
func NewCmdUtils(name, fullName string) *cobra.Command {
	terminalCmd := NewCmdTerminal(terminalCommandName, util.GetFullName(fullName, terminalCommandName))
	configurationCmd := config.NewCmdConfiguration(config.RecommendedCommandName, util.GetFullName(fullName, config.RecommendedCommandName))
	utilsCmd := &cobra.Command{
		Use:   name,
		Short: "Utilities for terminal commands and modifying Odo configurations",
		Long:  `Utilities for terminal commands and modifying Odo configurations`,
		Example: fmt.Sprintf("%s\n%s",
			terminalCmd.Example,
			configurationCmd.Example),
	}

	utilsCmd.Annotations = map[string]string{"command": "utility"}
	utilsCmd.SetUsageTemplate(util.CmdUsageTemplate)

	utilsCmd.AddCommand(configurationCmd, terminalCmd)
	return utilsCmd
}
