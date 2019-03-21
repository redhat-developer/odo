package application

import (
	"fmt"
	"github.com/openshift/odo/pkg/application"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/validation"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

const createRecommendedCommandName = "create"

var (
	createExample = ktemplates.Examples(`  # Create an application
  %[1]s myapp
  %[1]s`)
	createLongDesc = ktemplates.LongDesc(`Create an application.
If no app name is passed, a default app name will be auto-generated.`)
)

// CreateOptions encapsulates the options for the odo command
type CreateOptions struct {
	appName string
	*genericclioptions.Context
}

// NewCreateOptions creates a new CreateOptions instance
func NewCreateOptions() *CreateOptions {
	return &CreateOptions{}
}

// Complete completes CreateOptions after they've been created
func (o *CreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	if len(args) == 1 {
		// The only arg passed is the app name
		o.appName = args[0]
	} else {
		// Desired app name is not passed so, generate a new app name
		// Fetch existing list of apps
		apps, err := application.List(o.Client)
		if err != nil {
			return err
		}

		// Generate a random name that's not already in use for the existing apps
		o.appName, err = application.GetDefaultAppName(apps)
		if err != nil {
			return err
		}
	}
	return
}

// Validate validates the CreateOptions based on completed values
func (o *CreateOptions) Validate() (err error) {
	return validation.ValidateName(o.appName)
}

// Run contains the logic for the odo command
func (o *CreateOptions) Run() (err error) {
	log.Progressf("Creating application: %v in project: %v", o.appName, o.Project)
	err = application.Create(o.Client, o.appName)
	if err != nil {
		return err
	}

	err = application.SetCurrent(o.Client, o.appName)
	if err != nil {
		return err
	}

	// TODO: updating the app name should be done via SetCurrent and passing the Context
	// not strictly needed here but Context should stay in sync
	o.Context.Application = o.appName

	log.Infof("Switched to application: %v in project: %v", o.appName, o.Project)
	return
}

// NewCmdCreate implements the odo command.
func NewCmdCreate(name, fullName string) *cobra.Command {
	o := NewCreateOptions()
	command := &cobra.Command{
		Use:     name,
		Short:   "Create an application",
		Long:    createLongDesc,
		Example: fmt.Sprintf(createExample, fullName),
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	project.AddProjectFlag(command)
	return command
}
