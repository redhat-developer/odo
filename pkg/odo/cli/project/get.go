package project

import (
	"fmt"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/project"
	"github.com/spf13/cobra"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const getRecommendedCommandName = "get"

var (
	getExample = ktemplates.Examples(`
	# Get the active project
	%[1]s 
	`)

	getLongDesc = ktemplates.LongDesc(`Get the active project`)

	getShortDesc = `Get the active project`
)

// ProjectGetOptions encapsulates the options for the odo project get command
type ProjectGetOptions struct {

	// if supplied then only print the project name
	projectShortFlag bool

	// generic context options common to all commands
	*genericclioptions.Context
}

// NewProjectGetOptions creates a ProjectGetOptions instance
func NewProjectGetOptions() *ProjectGetOptions {
	return &ProjectGetOptions{}
}

// Complete completes ProjectGetOptions after they've been created
func (pgo *ProjectGetOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	pgo.Context, err = genericclioptions.NewContext(cmd)
	return
}

// Validate validates the parameters of the ProjectGetOptions
func (pgo *ProjectGetOptions) Validate() (err error) {
	return
}

// Run the project get command
func (pgo *ProjectGetOptions) Run(cmd *cobra.Command) (err error) {
	currentProject := pgo.Context.GetProject()

	if pgo.projectShortFlag {
		fmt.Print(currentProject)
		return
	}

	log.Infof("The current project is: %v", currentProject)

	if log.IsJSON() {
		prj := project.NewProject(currentProject, true)
		machineoutput.OutputSuccess(prj)
	}

	return
}

// NewCmdProjectGet creates the project get command
func NewCmdProjectGet(name, fullName string) *cobra.Command {
	o := NewProjectGetOptions()

	projectGetCmd := &cobra.Command{
		Use:         name,
		Short:       getShortDesc,
		Long:        getLongDesc,
		Example:     fmt.Sprintf(getExample, fullName),
		Args:        cobra.ExactArgs(0),
		Annotations: map[string]string{"machineoutput": "json"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	projectGetCmd.Flags().BoolVarP(&o.projectShortFlag, "short", "q", false, "If true, display only the project name")

	return projectGetCmd
}
