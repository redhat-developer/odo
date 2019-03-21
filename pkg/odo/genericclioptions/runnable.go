package genericclioptions

import (
	"github.com/openshift/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

type Runnable interface {
	Complete(name string, cmd *cobra.Command, args []string) error
	Validate() error
	Run() error
}

func GenericRun(o Runnable, cmd *cobra.Command, args []string) {
	util.LogErrorAndExit(o.Complete(cmd.Name(), cmd, args), "")
	util.LogErrorAndExit(o.Validate(), "")
	util.LogErrorAndExit(o.Run(), "")
}
