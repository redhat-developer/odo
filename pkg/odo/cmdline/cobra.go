package cmdline

import "github.com/spf13/cobra"

type Cobra struct {
	cmd *cobra.Command
}

func NewCobra(cmd *cobra.Command) *Cobra {
	return &Cobra{
		cmd: cmd,
	}
}

func (o *Cobra) GetCmd() *cobra.Command {
	return o.cmd
}
