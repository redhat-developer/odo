// cmdline package provides an abstration of a cmdline utility
package cmdline

import "github.com/spf13/cobra"

type Cmdline interface {
	GetCmd() *cobra.Command // TODO temporary, to be removed
}
