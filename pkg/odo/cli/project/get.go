package project

import (
	"context"
	"fmt"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
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
	// Context
	*genericclioptions.Context

	// Flags
	shortFlag bool
}

var _ genericclioptions.Runnable = (*ProjectGetOptions)(nil)

// NewProjectGetOptions creates a ProjectGetOptions instance
func NewProjectGetOptions() *ProjectGetOptions {
	return &ProjectGetOptions{}
}

func (o *ProjectGetOptions) SetClientset(clientset *clientset.Clientset) {
}

// Complete completes ProjectGetOptions after they've been created
func (pgo *ProjectGetOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	pgo.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline))
	return err
}

// Validate validates the parameters of the ProjectGetOptions
func (pgo *ProjectGetOptions) Validate() (err error) {
	return nil
}

// Run the project get command
func (pgo *ProjectGetOptions) Run(ctx context.Context) (err error) {
	currentProject := pgo.Context.GetProject()

	if pgo.shortFlag {
		fmt.Print(currentProject)
		return nil
	}

	log.Infof("The current project is: %v", currentProject)
	return nil
}

// NewCmdProjectGet creates the project get command
func NewCmdProjectGet(name, fullName string) *cobra.Command {
	o := NewProjectGetOptions()

	projectGetCmd := &cobra.Command{
		Use:     name,
		Short:   getShortDesc,
		Long:    getLongDesc,
		Example: fmt.Sprintf(getExample, fullName),
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	projectGetCmd.Flags().BoolVarP(&o.shortFlag, "short", "q", false, "If true, display only the project name")

	return projectGetCmd
}
