package project

import (
	"context"
	"fmt"

	odoerrors "github.com/redhat-developer/odo/pkg/errors"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"
	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const setRecommendedCommandName = "set"

var (
	setExample = ktemplates.Examples(`
	# Set the active project
	%[1]s myproject
	`)

	setLongDesc = ktemplates.LongDesc(`Set the active project.
	This command directly performs actions on the cluster and doesn't require a push.
	`)

	setShortDesc = `Set the current active project`
)

// ProjectSetOptions encapsulates the options for the odo project set command
type ProjectSetOptions struct {
	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset

	// Parameters
	projectName string

	// Flags
	shortFlag bool
}

var _ genericclioptions.Runnable = (*ProjectSetOptions)(nil)

// NewProjectSetOptions creates a ProjectSetOptions instance
func NewProjectSetOptions() *ProjectSetOptions {
	return &ProjectSetOptions{}
}

func (o *ProjectSetOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete completes ProjectSetOptions after they've been created
func (pso *ProjectSetOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	pso.projectName = args[0]
	pso.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline))
	if err != nil {
		return err
	}
	if scontext.GetTelemetryStatus(cmdline.Context()) {
		scontext.SetClusterType(cmdline.Context(), pso.KClient)
	}
	return nil
}

// Validate validates the parameters of the ProjectSetOptions
func (pso *ProjectSetOptions) Validate() (err error) {

	exists, err := pso.clientset.ProjectClient.Exists(pso.projectName)
	if kerrors.IsForbidden(err) {
		return &odoerrors.Unauthorized{}
	}
	if !exists {
		//revive:disable:error-strings This is a top-level error message displayed as is to the end user
		return fmt.Errorf("The project %s does not exist", pso.projectName)
		//revive:enable:error-strings
	}

	return nil
}

// Run runs the project set command
func (pso *ProjectSetOptions) Run(ctx context.Context) (err error) {
	current := pso.GetProject()
	err = pso.clientset.ProjectClient.SetCurrent(pso.projectName)
	if err != nil {
		return err
	}
	if pso.shortFlag {
		fmt.Print(pso.projectName)
	} else {
		if current == pso.projectName {
			log.Infof("Already on project : %v", pso.projectName)
		} else {
			log.Infof("Switched to project : %v", pso.projectName)
		}
	}
	return nil
}

// NewCmdProjectSet creates the project set command
func NewCmdProjectSet(name, fullName string) *cobra.Command {
	o := NewProjectSetOptions()

	projectSetCmd := &cobra.Command{
		Use:     name,
		Short:   setShortDesc,
		Long:    setLongDesc,
		Example: fmt.Sprintf(setExample, fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(projectSetCmd, clientset.PROJECT)

	projectSetCmd.Flags().BoolVarP(&o.shortFlag, "short", "q", false, "If true, display only the project name")

	return projectSetCmd
}
