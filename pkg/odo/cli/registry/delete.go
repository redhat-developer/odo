package registry

import (
	// Built-in packages
	"fmt"

	// Third-party packages
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	// odo packages
	registryUtil "github.com/redhat-developer/odo/pkg/odo/cli/registry/util"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/util"
)

const deleteCommandName = "delete"

// "odo registry delete" command description and examples
var (
	deleteLongDesc = ktemplates.LongDesc(`Delete devfile registry`)

	deleteExample = ktemplates.Examples(`# Delete devfile registry
	%[1]s CheRegistry
	`)
)

// DeleteOptions encapsulates the options for the "odo registry delete" command
type DeleteOptions struct {
	// Parameters
	registryName string

	// Flags
	forceFlag bool

	operation   string
	registryURL string
	user        string
}

// NewDeleteOptions creates a new DeleteOptions instance
func NewDeleteOptions() *DeleteOptions {
	return &DeleteOptions{}
}

// Complete completes DeleteOptions after they've been created
func (o *DeleteOptions) Complete(name string, cmdline cmdline.Cmdline, args []string) (err error) {
	o.operation = "delete"
	o.registryName = args[0]
	o.registryURL = ""
	o.user = "default"
	return nil
}

// Validate validates the DeleteOptions based on completed values
func (o *DeleteOptions) Validate() (err error) {
	return nil
}

// Run contains the logic for "odo registry delete" command
func (o *DeleteOptions) Run() (err error) {
	isSecure, err := registryUtil.IsSecure(o.registryName)
	if err != nil {
		return err
	}
	cfg, err := preference.New()
	if err != nil {
		return errors.Wrap(err, "unable to delete registry")
	}
	err = cfg.RegistryHandler(o.operation, o.registryName, o.registryURL, o.forceFlag, false)
	if err != nil {
		return err
	}

	if isSecure {
		err = keyring.Delete(util.CredentialPrefix+o.registryName, o.user)
		if err != nil {
			return errors.Wrap(err, "unable to delete registry credential from keyring")
		}
	}

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
