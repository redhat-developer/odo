package run

import (
	"context"
	"fmt"

	"github.com/devfile/library/v2/pkg/devfile/parser/data/v2/common"
	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/odo/cli/errors"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const (
	RecommendedCommandName = "run"
)

type RunOptions struct {
	// Clients
	clientset *clientset.Clientset

	// Args
	commandName string
}

var _ genericclioptions.Runnable = (*RunOptions)(nil)

func NewRunOptions() *RunOptions {
	return &RunOptions{}
}

var runExample = ktemplates.Examples(`
	# Run the command "my-command" in the Dev mode
	%[1]s my-command

`)

func (o *RunOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

func (o *RunOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) error {
	o.commandName = args[0] // Value at 0 is expected to exist, thanks to ExactArgs(1)
	return nil
}

func (o *RunOptions) Validate(ctx context.Context) error {
	var (
		devfileObj = odocontext.GetEffectiveDevfileObj(ctx)
	)

	if devfileObj == nil {
		return genericclioptions.NewNoDevfileError(odocontext.GetWorkingDirectory(ctx))
	}

	commands, err := devfileObj.Data.GetCommands(common.DevfileOptions{
		FilterByName: o.commandName,
	})
	if err != nil || len(commands) != 1 {
		return errors.NewNoCommandNameInDevfileError(o.commandName)
	}
	return nil
}

func (o *RunOptions) Run(ctx context.Context) (err error) {
	return nil
}

func NewCmdRun(name, fullName string) *cobra.Command {
	o := NewRunOptions()
	runCmd := &cobra.Command{
		Use:     name,
		Short:   "Run a specific command in the Dev mode",
		Long:    `odo run executes a specific command of the Devfile during the Dev mode ("odo dev" needs to be running)`,
		Example: fmt.Sprintf(runExample, fullName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(runCmd,
		clientset.FILESYSTEM,
	)

	odoutil.SetCommandGroup(runCmd, odoutil.MainGroup)
	runCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	commonflags.UsePlatformFlag(runCmd)
	return runCmd
}
