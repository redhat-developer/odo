package application

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/project"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

const getRecommendedCommandName = "get"

var (
	getExample = ktemplates.Examples(`  # Get the currently active application
  %[1]s`)
)

// GetOptions encapsulates the options for the odo command
type GetOptions struct {
	short bool
	*genericclioptions.Context
}

// NewGetOptions creates a new GetOptions instance
func NewGetOptions() *GetOptions {
	return &GetOptions{}
}

// Complete completes GetOptions after they've been created
func (o *GetOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	return
}

// Validate validates the GetOptions based on completed values
func (o *GetOptions) Validate() (err error) {
	return
}

// Run contains the logic for the odo command
func (o *GetOptions) Run() (err error) {
	if o.short {
		fmt.Print(o.Application)
		return
	}
	if o.Application == "" {
		log.Infof("There's no active application.\nYou can create one by running 'odo application create <name>'.")
		return
	}
	log.Infof("The current application is: %v in project: %v", o.Application, o.Project)
	return
}

// NewCmdGet implements the odo command.
func NewCmdGet(name, fullName string) *cobra.Command {
	o := NewGetOptions()
	command := &cobra.Command{
		Use:     name,
		Short:   "Get the active application",
		Long:    "Get the active application",
		Example: fmt.Sprintf(getExample, fullName),
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	project.AddProjectFlag(command)
	command.Flags().BoolVarP(&o.short, "short", "q", false, "If true, display only the application name")
	return command
}
