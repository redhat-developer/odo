package project

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/project"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const listRecommendedCommandName = "list"

var (
	listExample = ktemplates.Examples(`
	# List all the projects
    %[1]s`)
	listLongDesc = ktemplates.LongDesc(`
	List all the projects
`)
)

// ProjectListOptions encapsulates the options for the odo project list command
type ProjectListOptions struct {
	*genericclioptions.Context
}

// NewProjectListOptions creates a new ProjectListOptions instance
func NewProjectListOptions() *ProjectListOptions {
	return &ProjectListOptions{}
}

// Complete completes ProjectListOptions after they've been created
func (plo *ProjectListOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	plo.Context = genericclioptions.NewContext(cmd)
	return
}

// Validate validates the ProjectListOptions based on completed values
func (plo *ProjectListOptions) Validate() (err error) {
	return
}

// Run contains the logic for the odo project list command
func (plo *ProjectListOptions) Run() (err error) {

	projects, err := project.List(plo.Client)
	if err != nil {
		return err
	}
	if log.IsJSON() {
		machineoutput.OutputSuccess(projects)
	} else {

		if len(projects.Items) == 0 {
			return fmt.Errorf("You are not a member of any projects. You can request a project to be created using the `odo project create <project_name>` command")
		}
		w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
		fmt.Fprintln(w, "ACTIVE", "\t", "NAME")
		for _, project := range projects.Items {
			activeMark := " "
			if project.Status.Active {
				activeMark = "*"
			}
			fmt.Fprintln(w, activeMark, "\t", project.Name)
		}
		w.Flush()
	}
	return
}

// NewCmdProjectList implements the odo project list command.
func NewCmdProjectList(name, fullName string) *cobra.Command {
	o := NewProjectListOptions()
	projectListCmd := &cobra.Command{
		Use:         name,
		Short:       listLongDesc,
		Long:        listLongDesc,
		Example:     fmt.Sprintf(listExample, fullName),
		Args:        cobra.ExactArgs(0),
		Annotations: map[string]string{"machineoutput": "json"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	return projectListCmd
}
