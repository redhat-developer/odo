package component

import (
	"fmt"

	"github.com/openshift/odo/pkg/devfile/location"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	odoutil "github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/odo/util/completion"

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
	componentContext string
	componentOptions *ComponentOptions
	devfilePath      string
	namespace        string

	command []string
}

// NewExecOptions returns new instance of ExecOptions
func NewExecOptions() *ExecOptions {
	return &ExecOptions{
		componentOptions: &ComponentOptions{},
	}
}

// Complete completes exec args
func (eo *ExecOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	if cmd.ArgsLenAtDash() <= -1 {
		return fmt.Errorf(`no command was given for the exec command
Please provide a command to execute, odo exec -- <command to be execute>`)
	}

	// gets the command args passed after the dash i.e `--`
	eo.command = args[cmd.ArgsLenAtDash():]

	if len(eo.command) <= 0 {
		return fmt.Errorf(`no command was given for the exec command.
Please provide a command to execute, odo exec -- <command to be execute>`)
	}

	// checks if something is passed between `odo exec` and the dash `--`
	if len(eo.command) != len(args) {
		return fmt.Errorf("no parameter is expected for the command")
	}
	eo.devfilePath = location.DevfileLocation(eo.componentContext)

	eo.componentOptions.Context, err = genericclioptions.NewContext(cmd)
	if err != nil {
		return err
	}

	// The namespace was retrieved from the --project flag (or from the kube client if not set) and stored in kclient when initializing the context
	eo.namespace = eo.componentOptions.KClient.GetCurrentNamespace()

	return nil
}

// Validate validates the exec parameters
func (eo *ExecOptions) Validate() (err error) {
	return
}

// Run has the logic to perform the required actions as part of command
func (eo *ExecOptions) Run(cmd *cobra.Command) (err error) {
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
	genericclioptions.AddContextFlag(execCmd, &o.componentContext)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(execCmd)

	// Adding `--app` flag
	appCmd.AddApplicationFlag(execCmd)

	return execCmd
}
