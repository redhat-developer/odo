package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// utilsCmd represents the utils command
var utilsCmd = &cobra.Command{
	Use:   "utils",
	Short: "Utilities for completion, terminal commands and modifying Odo configurations",
	Long:  `Utilities for completion, terminal commands and modifying Odo configurations`,
	Example: fmt.Sprintf("%s\n%s\n%s\n%s",
		completionCmd.Example,
		terminalCmd.Example,
		configurationSetCmd.Example,
		configurationViewCmd.Example),
}

func init() {
	utilsCmd.Annotations = map[string]string{"command": "utility"}
	utilsCmd.SetUsageTemplate(cmdUsageTemplate)

	configurationCmd.AddCommand(configurationViewCmd)
	configurationCmd.AddCommand(configurationSetCmd)

	configurationCmd.SetUsageTemplate(cmdUsageTemplate)
	utilsCmd.AddCommand(configurationCmd)
	utilsCmd.AddCommand(completionCmd)
	utilsCmd.AddCommand(terminalCmd)
	rootCmd.AddCommand(utilsCmd)
}
