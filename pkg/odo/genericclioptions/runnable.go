package genericclioptions

import (
	api "github.com/metacosm/odo-event-api/odo/api/events"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/events"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/plugin"
	"github.com/spf13/cobra"
	"os"
)

type Runnable interface {
	Complete(name string, cmd *cobra.Command, args []string) error
	Validate() error
	Run() error
}

func GenericRun(o Runnable, cmd *cobra.Command, args []string) {
	// gob.Register(o) // TODO: figure out how to properly serialize Runnable so that it can be passed as event payload
	exitIfAbort(events.DispatchEvent(cmd, api.PreRun, args), cmd)
	util.LogErrorAndExit(o.Complete(cmd.Name(), cmd, args), "")
	exitIfAbort(events.DispatchEvent(cmd, api.PostComplete, nil), cmd)
	util.LogErrorAndExit(o.Validate(), "")
	exitIfAbort(events.DispatchEvent(cmd, api.PostValidate, nil), cmd)
	util.LogErrorAndExit(o.Run(), "")
	exitIfAbort(events.DispatchEvent(cmd, api.PostRun, nil), cmd)
}

func exitIfAbort(err error, cmd *cobra.Command) {
	if err != nil {
		if api.IsEventCausedAbort(err) {
			log.Errorf("Processing of %s command was aborted: %v", cmd.Name(), err)
			plugin.CleanPlugins()
			os.Exit(1)
		}
		log.Errorf("Error(s) during event dispatch: %v", err)
	}
}
