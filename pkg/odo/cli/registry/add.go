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

const addCommandName = "add"

// "odo registry add" command description and examples
var (
	addLongDesc = ktemplates.LongDesc(`Add devfile registry`)

	addExample = ktemplates.Examples(`# Add devfile registry
	%[1]s cheregistry https://che-devfile-registry.openshift.io/
	`)
)

// AddOptions encapsulates the options for the "odo registry add" command
type AddOptions struct {
	operation    string
	registryName string
	registryURL  string
}

// NewAddOptions creates a new AddOptions instance
func NewAddOptions() *AddOptions {
	return &AddOptions{}
}

// Complete completes AddOptions after they've been created
func (o *AddOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.operation = "add"
	o.registryName = args[0]
	o.registryURL = args[1]
	return
}

// Validate validates the AddOptions based on completed values
func (o *AddOptions) Validate() (err error) {
	return
}

// Run contains the logic for "odo registry add" command
func (o *AddOptions) Run() (err error) {

	cfg, err := preference.New()
	if err != nil {
		return errors.Wrapf(err, "Unable to add registry")
	}

	err = cfg.RegistryHandler(o.operation, o.registryName, o.registryURL)
	if err != nil {
		return err
	}

	log.Info("New registry was successfully added")
	return nil
}

// NewCmdAdd implements the "odo registry add" command
func NewCmdAdd(name, fullName string) *cobra.Command {
	o := NewAddOptions()
	registryAddCmd := &cobra.Command{
		Use:     fmt.Sprintf("%s <registry name> <registry URL>", name),
		Short:   addLongDesc,
		Long:    addLongDesc,
		Example: fmt.Sprintf(fmt.Sprint(addExample), fullName),
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	return registryAddCmd
}
