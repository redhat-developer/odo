package genericclioptions

import (
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/events"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	"os"
)

type Runnable interface {
	Complete(name string, cmd *cobra.Command, args []string) error
	Validate() error
	Run() error
}

func GenericRun(o Runnable, cmd *cobra.Command, args []string) {
	exitIfAbort(events.DispatchEvent(cmd, events.PreRun, args), cmd)
	util.CheckError(o.Complete(cmd.Name(), cmd, args), "")
	exitIfAbort(events.DispatchEvent(cmd, events.PostComplete, o), cmd)
	util.CheckError(o.Validate(), "")
	exitIfAbort(events.DispatchEvent(cmd, events.PostValidate, o), cmd)
	util.CheckError(o.Run(), "")
	exitIfAbort(events.DispatchEvent(cmd, events.PostRun, o), cmd)
}

func exitIfAbort(err error, cmd *cobra.Command) {
	if events.IsEventCausedAbort(err) {
		log.Errorf("Processing of %s command was aborted: %v", cmd.Name(), err)
		os.Exit(1)
	}
}
