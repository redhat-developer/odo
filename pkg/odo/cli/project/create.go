package project

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/project"
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

	// generic context options common to all commands
	*genericclioptions.Context
}

func NewProjectCreateOptions() *ProjectCreateOptions {
	return &ProjectCreateOptions{}
}

// Complete completes ProjectCreateOptions after they've been created
func (pco *ProjectCreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	pco.projectName = args[0]
	pco.Context = genericclioptions.NewContextCreatingAppIfNeeded(cmd)
	return
}

// Validate validates the parameters of the ProjectCreateOptions
func (pco *ProjectCreateOptions) Validate() (err error) {
	return
}

// Run runs the project create command
func (pco *ProjectCreateOptions) Run() (err error) {
	err = project.Create(pco.Client, pco.projectName)
	if err != nil {
		return err
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
	pco := NewProjectCreateOptions()

	projectCreateCmd := &cobra.Command{
		Use:     name,
		Short:   createShortDesc,
		Long:    createLongDesc,
		Example: fmt.Sprintf(createExample, fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			util.LogErrorAndExit(pco.Complete(name, cmd, args), "")
			util.LogErrorAndExit(pco.Validate(), "")
			util.LogErrorAndExit(pco.Run(), "")
		},
	}

	return projectCreateCmd
}
