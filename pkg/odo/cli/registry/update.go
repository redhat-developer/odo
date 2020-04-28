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

const updateCommandName = "update"

// "odo registry update" command description and examples
var (
	updateLongDesc = ktemplates.LongDesc(`Update devfile registry`)

	updateExample = ktemplates.Examples(`# Update devfile registry
	%[1]s cheregistry https://che-devfile-registry-update.openshift.io/
	`)
)

// UpdateOptions encapsulates the options for the "odo registry update" command
type UpdateOptions struct {
	operation    string
	registryName string
	registryURL  string
	forceFlag    bool
}

// NewUpdateOptions creates a new UpdateOptions instance
func NewUpdateOptions() *UpdateOptions {
	return &UpdateOptions{}
}

// Complete completes UpdateOptions after they've been created
func (o *UpdateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.operation = "update"
	o.registryName = args[0]
	o.registryURL = args[1]
	return
}

// Validate validates the UpdateOptions based on completed values
func (o *UpdateOptions) Validate() (err error) {
	return
}

// Run contains the logic for "odo registry update" command
func (o *UpdateOptions) Run() (err error) {

	cfg, err := preference.New()
	if err != nil {
		return errors.Wrapf(err, "Unable to update registry")
	}

	if !o.forceFlag {
		if !ui.Proceed(fmt.Sprintf("Are you sure you want to update registry %s", o.registryName)) {
			log.Info("Aborted by the user")
			return nil
		}
	}

	err = cfg.RegistryHandler(o.operation, o.registryName, o.registryURL)
	if err != nil {
		return err
	}

	log.Info("Successfully updated registry")
	return nil
}

// NewCmdUpdate implements the "odo registry update" command
func NewCmdUpdate(name, fullName string) *cobra.Command {
	o := NewUpdateOptions()
	registryUpdateCmd := &cobra.Command{
		Use:     name,
		Short:   updateLongDesc,
		Long:    updateLongDesc,
		Example: fmt.Sprintf(fmt.Sprint(updateExample), fullName),
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	registryUpdateCmd.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Don't ask for confirmation, update the registry directly")

	return registryUpdateCmd
}
