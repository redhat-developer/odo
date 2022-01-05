package application

import (
	"fmt"

	odoUtil "github.com/redhat-developer/odo/pkg/odo/util"

	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const deleteRecommendedCommandName = "delete"

var (
	deleteExample = ktemplates.Examples(`  # Delete the application
  %[1]s myapp`)
)

// DeleteOptions encapsulates the options for the odo command
type DeleteOptions struct {
	// Context
	*genericclioptions.Context

	// Parameters
	appName string

	// Flags
	forceFlag bool
}

// NewDeleteOptions creates a new DeleteOptions instance
func NewDeleteOptions() *DeleteOptions {
	return &DeleteOptions{}
}

// Complete completes DeleteOptions after they've been created
func (o *DeleteOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline))
	if err != nil {
		return err
	}
	o.appName = o.GetApplication()
	if len(args) == 1 {
		// If app name passed, consider it for deletion
		o.appName = args[0]
	}

	return
}

// Validate validates the DeleteOptions based on completed values
func (o *DeleteOptions) Validate() (err error) {
	if o.Context.GetProject() == "" || o.appName == "" {
		return odoUtil.ThrowContextError()
	}

	exist, err := application.Exists(o.appName, o.KClient)
	if !exist {
		return fmt.Errorf("%s app does not exists", o.appName)
	}
	return err
}

// Run contains the logic for the odo command
func (o *DeleteOptions) Run() (err error) {
	if log.IsJSON() {
		err = application.Delete(o.KClient, o.appName)
		if err != nil {
			return err
		}
		return nil
	}

	// Print App Information which will be deleted
	err = printAppInfo(o.KClient, o.KClient, o.appName, o.GetProject())
	if err != nil {
		return err
	}

	if o.forceFlag || ui.Proceed(fmt.Sprintf("Are you sure you want to delete the application: %v from project: %v", o.appName, o.GetProject())) {
		err = application.Delete(o.KClient, o.appName)
		if err != nil {
			return err
		}
		log.Infof("Deleted application: %s from project: %v", o.appName, o.GetProject())
	} else {
		log.Infof("Aborting deletion of application: %v", o.appName)
	}
	return
}

// NewCmdDelete implements the odo command.
func NewCmdDelete(name, fullName string) *cobra.Command {
	o := NewDeleteOptions()
	command := &cobra.Command{
		Use:     name,
		Short:   "Delete the given application",
		Long:    "Delete the given application",
		Example: fmt.Sprintf(deleteExample, fullName),
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	command.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Delete application without prompting")

	project.AddProjectFlag(command)
	completion.RegisterCommandHandler(command, completion.AppCompletionHandler)
	return command
}
