package application

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/application"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
	"os"
	"text/tabwriter"
)

const listRecommendedCommandName = "list"

var (
	listExample = ktemplates.Examples(`  # List all applications in the current project
  %[1]s

  # List all applications in the specified project
  %[1]s --project myproject`)
)

// ListOptions encapsulates the options for the odo command
type ListOptions struct {
	*genericclioptions.Context
}

// NewListOptions creates a new ListOptions instance
func NewListOptions() *ListOptions {
	return &ListOptions{}
}

// Complete completes ListOptions after they've been created
func (o *ListOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	return
}

// Validate validates the ListOptions based on completed values
func (o *ListOptions) Validate() (err error) {
	return
}

// Run contains the logic for the odo command
func (o *ListOptions) Run() (err error) {
	apps, err := application.ListInProject(o.Client)
	if err != nil {
		return fmt.Errorf("unable to get list of applications: %v", err)
	}

	if len(apps) > 0 {
		log.Infof("The project '%v' has the following applications:", o.Project)
		tabWriter := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
		fmt.Fprintln(tabWriter, "ACTIVE", "\t", "NAME")
		for _, app := range apps {
			activeMark := " "
			if app.Active {
				activeMark = "*"
			}
			fmt.Fprintln(tabWriter, activeMark, "\t", app.Name)
		}
		tabWriter.Flush()
	} else {
		log.Infof("There are no applications deployed in the project '%v'.", o.Project)
	}
	return
}

// NewCmdList implements the odo command.
func NewCmdList(name, fullName string) *cobra.Command {
	o := NewListOptions()
	command := &cobra.Command{
		Use:     name,
		Short:   "List all applications in the current project",
		Long:    "List all applications in the current project",
		Example: fmt.Sprintf(listExample, fullName),
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			util.LogErrorAndExit(o.Complete(name, cmd, args), "")
			util.LogErrorAndExit(o.Validate(), "")
			util.LogErrorAndExit(o.Run(), "")
		},
	}
	project.AddProjectFlag(command)
	return command
}
