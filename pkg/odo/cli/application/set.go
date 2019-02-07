package application

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

const setRecommendedCommandName = "set"

var (
	setExample = ktemplates.Examples(`  # Set an application as active
  %[1]s myapp`)
)

// SetOptions encapsulates the options for the odo command
type SetOptions struct {
	appName string
	*genericclioptions.Context
}

// NewSetOptions creates a new SetOptions instance
func NewSetOptions() *SetOptions {
	return &SetOptions{}
}

// Complete completes SetOptions after they've been created
func (o *SetOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	o.appName = args[0]
	return
}

// Validate validates the SetOptions based on completed values
func (o *SetOptions) Validate() (err error) {
	return ensureAppExists(o.Client, o.appName, o.Project)
}

// Run contains the logic for the odo command
func (o *SetOptions) Run() (err error) {
	err = application.SetCurrent(o.Client, o.appName)
	if err != nil {
		return err
	}
	log.Infof("Switched to application: %v in project: %v", o.appName, o.Project)

	// TODO: updating the app name should be done via SetCurrent and passing the Context
	// not strictly needed here but Context should stay in sync
	o.Context.Application = o.appName
	return
}

// NewCmdSet implements the odo command.
func NewCmdSet(name, fullName string) *cobra.Command {
	o := NewSetOptions()
	command := &cobra.Command{
		Use:     name,
		Short:   "Set application as active",
		Long:    "Set application as active",
		Example: fmt.Sprintf(setExample, fullName),
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("Only one argument (application name) is allowed")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	project.AddProjectFlag(command)
	completion.RegisterCommandHandler(command, completion.AppCompletionHandler)
	return command
}
