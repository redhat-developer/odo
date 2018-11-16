package cli

import (
	"fmt"

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
	utilsCmd.SetUsageTemplate(CmdUsageTemplate)

	configurationCmd.AddCommand(configurationViewCmd)
	configurationCmd.AddCommand(configurationSetCmd)

	configurationCmd.SetUsageTemplate(CmdUsageTemplate)
	utilsCmd.AddCommand(configurationCmd)
	utilsCmd.AddCommand(terminalCmd)
	rootCmd.AddCommand(utilsCmd)
}
