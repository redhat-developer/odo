package genericclioptions

import "github.com/spf13/cobra"

// AddOutputFlag adds a `output` flag to the given cobra command
func AddOutputFlag(cmd *cobra.Command) {
	cmd.Flags().StringP(OutputFlagName, "o", "", "Specify output format, supported format: json")
}

// AddContextFlag adds `context` flag to given cobra command
func AddContextFlag(cmd *cobra.Command, setValueTo *string) {
	if setValueTo != nil {
		cmd.Flags().StringVar(setValueTo, "context", "", "Use given context directory as a source for component settings")
	} else {
		cmd.Flags().String("context", "", "Use given context directory as a source for component settings")
	}
}
