package catalog

import (
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var addCommand func(streams genericclioptions.IOStreams, cmd *cobra.Command)

// AddCommand adds the catalog subcommand to the given command.
// The subcommand is only added when built on supporting OSes (linux).
func AddCommand(streams genericclioptions.IOStreams, cmd *cobra.Command) {
	if addCommand == nil {
		return
	}

	addCommand(streams, cmd)
}
