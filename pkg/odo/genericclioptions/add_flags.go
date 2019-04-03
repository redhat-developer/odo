package genericclioptions

import "github.com/spf13/cobra"

// AddOutputFlag adds a `output` flag to the given cobra command
func AddOutputFlag(cmd *cobra.Command) {
	cmd.Flags().StringP(OutputFlagName, "o", "", "Specify output format, supported format: json")
}

// AddContextFlag adds `context` flag to given cobra command
func AddContextFlag(cmd *cobra.Command, setValueTo *string) {
	cmd.Flags().StringVar(setValueTo, "context", "", "Use context to indicate the path where the component settings need to be saved and this directory should contain component source for local and binary components")
}
