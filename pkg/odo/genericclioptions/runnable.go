package genericclioptions

import (
	"github.com/redhat-developer/odo/pkg/odo/events"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

type Runnable interface {
	Complete(name string, cmd *cobra.Command, args []string) error
	Validate() error
	Run() error
}

func GenericRun(o Runnable, cmd *cobra.Command, args []string) {
	events.DispatchEvent(cmd, events.PreRun, args)
	util.CheckError(o.Complete(cmd.Name(), cmd, args), "")
	events.DispatchEvent(cmd, events.PostComplete, o)
	util.CheckError(o.Validate(), "")
	events.DispatchEvent(cmd, events.PostValidate, o)
	util.CheckError(o.Run(), "")
	events.DispatchEvent(cmd, events.PostRun, o)
}
