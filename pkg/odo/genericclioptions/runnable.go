package genericclioptions

import (
	api "github.com/metacosm/odo-event-api/odo/api/events"
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
	exitIfAbort(events.DispatchEvent(cmd, api.PreRun, args), cmd)
	util.LogErrorAndExit(o.Complete(cmd.Name(), cmd, args), "")
	exitIfAbort(events.DispatchEvent(cmd, api.PostComplete, o), cmd)
	util.LogErrorAndExit(o.Validate(), "")
	exitIfAbort(events.DispatchEvent(cmd, api.PostValidate, o), cmd)
	util.LogErrorAndExit(o.Run(), "")
	exitIfAbort(events.DispatchEvent(cmd, api.PostRun, o), cmd)
}

func exitIfAbort(err error, cmd *cobra.Command) {
	if api.IsEventCausedAbort(err) {
		log.Errorf("Processing of %s command was aborted: %v", cmd.Name(), err)
		os.Exit(1)
	}
}
