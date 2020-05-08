package component

import (
	"fmt"
	"os"

	"github.com/openshift/odo/pkg/odo/genericclioptions"

	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	projectCmd "github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/util/completion"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	odoutil "github.com/openshift/odo/pkg/odo/util"

	"github.com/openshift/odo/pkg/component"
	"github.com/spf13/cobra"
)

// LogRecommendedCommandName is the recommended watch command name
const LogRecommendedCommandName = "log"

var logExample = ktemplates.Examples(`  # Get the logs for the nodejs component
%[1]s nodejs
`)

// LogOptions contains log options
type LogOptions struct {
	logFollow        bool
	componentContext string
	*ComponentOptions
}

// NewLogOptions returns new instance of LogOptions
func NewLogOptions() *LogOptions {
	return &LogOptions{false, "", &ComponentOptions{}}
}

// Complete completes log args
func (lo *LogOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	err = lo.ComponentOptions.Complete(name, cmd, args)
	return
}

// Validate validates the log parameters
func (lo *LogOptions) Validate() (err error) {
	return
}

// Run has the logic to perform the required actions as part of command
func (lo *LogOptions) Run() (err error) {
	stdout := os.Stdout

	// Retrieve the log
	err = component.GetLogs(lo.Context.Client, lo.componentName, lo.Context.Application, lo.logFollow, stdout)
	return
}

// NewCmdLog implements the log odo command
func NewCmdLog(name, fullName string) *cobra.Command {
	o := NewLogOptions()

	var logCmd = &cobra.Command{
		Use:         fmt.Sprintf("%s [component_name]", name),
		Short:       "Retrieve the log for the given component",
		Long:        `Retrieve the log for the given component`,
		Example:     fmt.Sprintf(logExample, fullName),
		Args:        cobra.RangeArgs(0, 1),
		Annotations: map[string]string{"command": "component"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	logCmd.Flags().BoolVarP(&o.logFollow, "follow", "f", false, "Follow logs")

	logCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	completion.RegisterCommandHandler(logCmd, completion.ComponentNameCompletionHandler)
	// Adding `--context` flag
	genericclioptions.AddContextFlag(logCmd, &o.componentContext)

	//Adding `--project` flag
	projectCmd.AddProjectFlag(logCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(logCmd)

	return logCmd
}
