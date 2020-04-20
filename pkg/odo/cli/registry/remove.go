package registry

import (
	// Build-in packages
	"fmt"

	// Third-party packages
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/util/templates"

	// odo packages
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/preference"
)

const removeCommandName = "remove"

// "odo registry remove" command description and examples
var (
	removeLongDesc = ktemplates.LongDesc(`Remove devfile registry`)

	removeExample = ktemplates.Examples(`# Remove devfile registry
	%[1]s cheregistry
	`)
)

// RemoveOptions encapsulates the options for the "odo registry remove" command
type RemoveOptions struct {
	operation    string
	registryName string
	registryURL  string
}

// NewRemoveOptions creates a new RemoveOptions instance
func NewRemoveOptions() *RemoveOptions {
	return &RemoveOptions{}
}

// Complete completes RemoveOptions after they've been created
func (o *RemoveOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.operation = "remove"
	o.registryName = args[0]
	o.registryURL = ""
	return
}

// Validate validates the RemoveOptions based on completed values
func (o *RemoveOptions) Validate() (err error) {
	return
}

// Run contains the logic for "odo registry remove" command
func (o *RemoveOptions) Run() (err error) {

	cfg, err := preference.New()
	if err != nil {
		return errors.Wrapf(err, "Unable to remove registry")
	}

	err = cfg.RegistryHandler(o.operation, o.registryName, o.registryURL)
	if err != nil {
		return err
	}

	log.Info("Successfully removed registry")
	return nil
}

// NewCmdRemove implements the "odo registry remove" command
func NewCmdRemove(name, fullName string) *cobra.Command {
	o := NewRemoveOptions()
	registryRemoveCmd := &cobra.Command{
		Use:     name,
		Short:   removeLongDesc,
		Long:    removeLongDesc,
		Example: fmt.Sprintf(fmt.Sprint(removeExample), fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	return registryRemoveCmd
}
