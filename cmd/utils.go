package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// utilsCmd represents the utils command
var utilsCmd = &cobra.Command{
	Use:   "utils",
	Short: "Utilities for completion and terminal commands",
	Long:  `Utilities for completion and terminal commands`,
	Example: fmt.Sprintf("%s\n%s",
		completionCmd.Example,
		terminalCmd.Example),
}

func init() {
	utilsCmd.Annotations = map[string]string{"command": "utility"}
	utilsCmd.SetUsageTemplate(cmdUsageTemplate)

	utilsCmd.AddCommand(completionCmd)
	utilsCmd.AddCommand(terminalCmd)
	rootCmd.AddCommand(utilsCmd)
}
