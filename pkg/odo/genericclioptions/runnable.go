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
	events.DispatchEvent(cmd, events.PreRun)
	util.LogErrorAndExit(o.Complete(cmd.Name(), cmd, args), "")
	events.DispatchEvent(cmd, events.PostComplete)
	util.LogErrorAndExit(o.Validate(), "")
	events.DispatchEvent(cmd, events.PostValidate)
	util.LogErrorAndExit(o.Run(), "")
	events.DispatchEvent(cmd, events.PostRun)
}
