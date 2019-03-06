package genericclioptions

import "github.com/spf13/cobra"

// AddOutputFlag adds a `output` flag to the given cobra command
func AddOutputFlag(cmd *cobra.Command) {
	cmd.Flags().StringP(OutputFlagName, "o", "", "Specify output format, supported format: json")
}
