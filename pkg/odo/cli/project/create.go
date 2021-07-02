package project

import (
	"fmt"

	"github.com/openshift/odo/pkg/segment"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/project"
	scontext "github.com/openshift/odo/pkg/segment/context"

	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const createRecommendedCommandName = "create"

var (
	createExample = ktemplates.Examples(`
	# Create a new project
	%[1]s myproject
	`)

	createLongDesc = ktemplates.LongDesc(`Create a new project.
	This command directly performs actions on the cluster and doesn't require a push.
	`)

	createShortDesc = `Create a new project`
)

// ProjectCreateOptions encapsulates the options for the odo project create command
type ProjectCreateOptions struct {
	// name of the project
	projectName string
	wait        bool

	// generic context options common to all commands
	*genericclioptions.Context
}

// NewProjectCreateOptions creates a ProjectCreateOptions instance
func NewProjectCreateOptions() *ProjectCreateOptions {
	return &ProjectCreateOptions{}
}

// Complete completes ProjectCreateOptions after they've been created
func (pco *ProjectCreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	pco.projectName = args[0]
	pco.Context, err = genericclioptions.NewContext(cmd)
	return
}

// Validate validates the parameters of the ProjectCreateOptions
func (pco *ProjectCreateOptions) Validate() (err error) {
	return
}

// Run runs the project create command
func (pco *ProjectCreateOptions) Run(cmd *cobra.Command) (err error) {
	if segment.IsTelemetryEnabled(nil) {
		scontext.SetClusterType(cmd.Context(), pco.Client)
	}
	successMessage := fmt.Sprintf(`Project '%s' is ready for use`, pco.projectName)

	// Create the "spinner"
	s := &log.Status{}

	// If the --wait parameter has been passed, we add a spinner..
	if pco.wait {
		s = log.Spinner("Waiting for project to come up")
		defer s.End(false)
	}

	// Create the project & end the spinner (if there is any..)
	err = project.Create(pco.Context, pco.projectName, pco.wait)
	if err != nil {
		return err
	}
	s.End(true)
	log.Successf(successMessage)

	// Set the current project when created. If it's json output, we output a json output error
	err = project.SetCurrent(pco.Context, pco.projectName)
	if err != nil {
		return err
	}

	// If -o json has been passed, let's output the appropriate json output.
	if log.IsJSON() {
		project.MachineReadableSuccessOutput(pco.projectName, successMessage)
	} else {
		log.Successf("New project created and now using project: %v", pco.projectName)
	}
	return
}

// NewCmdProjectCreate creates the project create command
func NewCmdProjectCreate(name, fullName string) *cobra.Command {
	o := NewProjectCreateOptions()

	projectCreateCmd := &cobra.Command{
		Use:         name,
		Short:       createShortDesc,
		Long:        createLongDesc,
		Example:     fmt.Sprintf(createExample, fullName),
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{"machineoutput": "json"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	projectCreateCmd.Flags().BoolVarP(&o.wait, "wait", "w", false, "Wait until the project is ready")
	return projectCreateCmd
}
