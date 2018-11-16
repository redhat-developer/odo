package utils

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/cli"

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

func init() {
	utilsCmd.Annotations = map[string]string{"command": "utility"}
	utilsCmd.SetUsageTemplate(cli.CmdUsageTemplate)

	configurationCmd.AddCommand(configurationViewCmd)
	configurationCmd.AddCommand(configurationSetCmd)

	configurationCmd.SetUsageTemplate(cli.CmdUsageTemplate)
	utilsCmd.AddCommand(configurationCmd)
	utilsCmd.AddCommand(terminalCmd)
	cli.RootCmd().AddCommand(utilsCmd)
}
