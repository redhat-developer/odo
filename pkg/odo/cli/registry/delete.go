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
	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/preference"
)

const deleteCommandName = "delete"

// "odo registry delete" command description and examples
var (
	deleteLongDesc = ktemplates.LongDesc(`Delete devfile registry`)

	deleteExample = ktemplates.Examples(`# Delete devfile registry
	%[1]s cheregistry
	`)
)

// DeleteOptions encapsulates the options for the "odo registry delete" command
type DeleteOptions struct {
	operation    string
	registryName string
	registryURL  string
	forceFlag    bool
}

// NewDeleteOptions creates a new DeleteOptions instance
func NewDeleteOptions() *DeleteOptions {
	return &DeleteOptions{}
}

// Complete completes DeleteOptions after they've been created
func (o *DeleteOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.operation = "delete"
	o.registryName = args[0]
	o.registryURL = ""
	return
}

// Validate validates the DeleteOptions based on completed values
func (o *DeleteOptions) Validate() (err error) {
	return
}

// Run contains the logic for "odo registry delete" command
func (o *DeleteOptions) Run() (err error) {

	cfg, err := preference.New()
	if err != nil {
		return errors.Wrapf(err, "Unable to delete registry")
	}

	if !o.forceFlag {
		if !ui.Proceed(fmt.Sprintf("Are you sure you want to delete registry %s", o.registryName)) {
			log.Info("Aborted by the user")
			return nil
		}
	}

	err = cfg.RegistryHandler(o.operation, o.registryName, o.registryURL)
	if err != nil {
		return err
	}

	log.Info("Successfully deleted registry")
	return nil
}

// NewCmdDelete implements the "odo registry delete" command
func NewCmdDelete(name, fullName string) *cobra.Command {
	o := NewDeleteOptions()
	registryDeleteCmd := &cobra.Command{
		Use:     fmt.Sprintf("%s <registry name>", name),
		Short:   deleteLongDesc,
		Long:    deleteLongDesc,
		Example: fmt.Sprintf(fmt.Sprint(deleteExample), fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	registryDeleteCmd.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Don't ask for confirmation, delete the registry directly")

	return registryDeleteCmd
}
