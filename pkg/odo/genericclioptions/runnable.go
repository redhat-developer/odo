package genericclioptions

import (
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

type Runnable interface {
	Complete(name string, cmd *cobra.Command, args []string) error
	Validate() error
	Run() error
}

func GenericRun(o Runnable, cmd *cobra.Command, args []string) {
	util.CheckError(o.Complete(cmd.Name(), cmd, args), "")
	util.CheckError(o.Validate(), "")
	util.CheckError(o.Run(), "")
}
