package project

import (
	"fmt"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/project"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

const createRecommendedCommandName = "create"

var (
	createExample = ktemplates.Examples(`
	# Create a new project
	%[1]s myproject
	`)

	createLongDesc = ktemplates.LongDesc(`Create a new project`)

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
	pco.Context = genericclioptions.NewContext(cmd)
	return
}

// Validate validates the parameters of the ProjectCreateOptions
func (pco *ProjectCreateOptions) Validate() (err error) {
	return
}

// Run runs the project create command
func (pco *ProjectCreateOptions) Run() (err error) {
	if pco.wait {
		s := log.Spinner("Waiting for project to come up")
		err = project.Create(pco.Client, pco.projectName, true)
		if err != nil {
			return err
		} else {
			s.End(true)
			log.Successf(`Project '%s' is ready for use`, pco.projectName)
		}
	} else {
		err = project.Create(pco.Client, pco.projectName, false)
		if err != nil {
			return err
		}
	}

	err = project.SetCurrent(pco.Client, pco.projectName)
	if err != nil {
		return err
	}
	log.Successf("New project created and now using project : %v", pco.projectName)
	return
}

// NewCmdProjectCreate creates the project create command
func NewCmdProjectCreate(name, fullName string) *cobra.Command {
	o := NewProjectCreateOptions()

	projectCreateCmd := &cobra.Command{
		Use:     name,
		Short:   createShortDesc,
		Long:    createLongDesc,
		Example: fmt.Sprintf(createExample, fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	projectCreateCmd.Flags().BoolVarP(&o.wait, "wait", "w", false, "Wait until the project is ready")
	return projectCreateCmd
}
