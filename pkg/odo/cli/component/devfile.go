package component

import (
	"fmt"

	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/odo/cli/project"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

// examples
var pushDevFileExample = ktemplates.Examples(`
Devfile support is an experimental feature which extends the support for the use of Che devfiles in odo
for performing various odo operations.

The devfile support progress can be tracked by:
https://github.com/openshift/odo/issues/2467

Please note that this feature is currently under development and the "push-devfile" command has been
temporarily exposed only for experimental purposes, and may/will be removed in future releases.
  `)

const PushDevfileRecommendedCommandName = "push-devfile"

// PushDevfileOptions encapsulates odo component push-devfile  options
type PushDevfileOptions struct {
	devfilePath string
	*genericclioptions.Context
}

// NewPushDevfileOptions returns new instance of PushDevfileOptions
func NewPushDevfileOptions() *PushDevfileOptions {
	return &PushDevfileOptions{}
}

// Complete completes  args
func (pdo *PushDevfileOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	return nil
}

// Validate validates the  parameters
func (pdo *PushDevfileOptions) Validate() (err error) {
	return nil
}

// Run has the logic to perform the required actions as part of command
func (pdo *PushDevfileOptions) Run() (err error) {

	// Parse devfile
	devObj, err := devfile.Parse(pdo.devfilePath)
	if err != nil {
		return err
	}

	// Write back devfile yaml
	err = devObj.WriteYamlDevfile()
	if err != nil {
		return err
	}

	return nil
}

// NewCmdPushDevfile implements odo push-devfile  command
func NewCmdPushDevfile(name, fullName string) *cobra.Command {
	o := NewPushDevfileOptions()

	var pushDevfileCmd = &cobra.Command{
		Use:     name,
		Short:   "Push component using devfile.",
		Long:    "Push component using devfile.",
		Example: fmt.Sprintf(pushDevFileExample, fullName),
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	pushDevfileCmd.Flags().StringVar(&o.devfilePath, "devfile", "./devfile.yaml", "Path to a devfile.yaml")
	project.AddProjectFlag(pushDevfileCmd)

	return pushDevfileCmd
}
