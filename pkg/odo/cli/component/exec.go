package component

import (
	"fmt"

	projectCmd "github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// ExecRecommendedCommandName is the recommended exec command name
const ExecRecommendedCommandName = "exec"

var execExample = ktemplates.Examples(`  # Executes a command inside the component
%[1]s -- ls -a
`)

// ExecOptions contains exec options
type ExecOptions struct {
	// Component context
	componentOptions *ComponentOptions

	// Parameters
	command []string

	// Flags
	contextFlag string
}

// NewExecOptions returns new instance of ExecOptions
func NewExecOptions() *ExecOptions {
	return &ExecOptions{
		componentOptions: &ComponentOptions{},
	}
}

// Complete completes exec args
func (eo *ExecOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	// gets the command args passed after the dash i.e `--`
	eo.command, err = cmdline.GetArgsAfterDashes(args)
	if err != nil || len(eo.command) <= 0 {
		return fmt.Errorf(`no command was given for the exec command
Please provide a command to execute, odo exec -- <command to be execute>`)
	}

	// checks if something is passed between `odo exec` and the dash `--`
	if len(eo.command) != len(args) {
		return fmt.Errorf("no parameter is expected for the command")
	}

	eo.componentOptions.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(eo.contextFlag))
	return err
}

// Validate validates the exec parameters
func (eo *ExecOptions) Validate() (err error) {
	return
}

// Run has the logic to perform the required actions as part of command
func (eo *ExecOptions) Run() (err error) {
	return eo.DevfileComponentExec(eo.command)
}

// NewCmdExec implements the exec odo command
func NewCmdExec(name, fullName string) *cobra.Command {
	o := NewExecOptions()

	var execCmd = &cobra.Command{
		Use:         name,
		Short:       "Executes a command inside the component",
		Long:        `Executes a command inside the component`,
		Example:     fmt.Sprintf(execExample, fullName),
		Annotations: map[string]string{"command": "component"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	execCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(execCmd, completion.ComponentNameCompletionHandler)
	odoutil.AddContextFlag(execCmd, &o.contextFlag)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(execCmd)

	return execCmd
}
