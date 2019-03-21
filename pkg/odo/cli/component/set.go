package component

import (
	"fmt"
	"github.com/openshift/odo/pkg/odo/genericclioptions"

	"github.com/openshift/odo/pkg/component"
	"github.com/openshift/odo/pkg/log"
	appCmd "github.com/openshift/odo/pkg/odo/cli/application"
	"github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/util/completion"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

// SetRecommendedCommandName is the recommended push command name
const SetRecommendedCommandName = "set"

var setExample = ktemplates.Examples(`  # Set component named 'frontend' as active
%[1]s frontend
`)

// SetOptions encapsulates component set options
type SetOptions struct {
	*ComponentOptions
}

// NewSetOptions returns new instance of SetOptions
func NewSetOptions() *SetOptions {
	return &SetOptions{&ComponentOptions{}}
}

// Complete completes get args
func (sto *SetOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	err = sto.ComponentOptions.Complete(name, cmd, args)
	return
}

// Validate validates the get parameters
func (sto *SetOptions) Validate() (err error) {
	return
}

// Run has the logic to perform the required actions as part of command
func (sto *SetOptions) Run() (err error) {
	err = component.SetCurrent(sto.componentName, sto.Context.Application, sto.Context.Project)
	if err != nil {
		return err
	}
	log.Infof("Switched to component: %v", sto.componentName)
	return
}

// NewCmdSet implements odo component set command
func NewCmdSet(name, fullName string) *cobra.Command {
	sto := NewSetOptions()

	var componentSetCmd = &cobra.Command{
		Use:     name,
		Short:   "Set active component.",
		Long:    "Set component as active.",
		Example: fmt.Sprintf(setExample, fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(sto, cmd, args)
		},
	}

	//Adding `--project` flag
	project.AddProjectFlag(componentSetCmd)
	//Adding `--application` flag
	appCmd.AddApplicationFlag(componentSetCmd)

	completion.RegisterCommandHandler(componentSetCmd, completion.ComponentNameCompletionHandler)

	return componentSetCmd
}
