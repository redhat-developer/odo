package registry

import (
	// Built-in packages
	"fmt"

	util2 "github.com/redhat-developer/odo/pkg/odo/cli/registry/util"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"

	// Third-party packages
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	// odo packages
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/util"
)

const addCommandName = "add"

// "odo registry add" command description and examples
var (
	addLongDesc = ktemplates.LongDesc(`Add devfile registry`)

	addExample = ktemplates.Examples(`# Add devfile registry
	%[1]s CheRegistry https://che-devfile-registry.openshift.io

	%[1]s RegistryFromGitHub https://github.com/elsony/devfile-registry
	`)
)

// AddOptions encapsulates the options for the "odo registry add" command
type AddOptions struct {
	// Parameters
	registryName string
	registryURL  string

	// Flags
	tokenFlag string

	operation string
	user      string
}

// NewAddOptions creates a new AddOptions instance
func NewAddOptions() *AddOptions {
	return &AddOptions{}
}

// Complete completes AddOptions after they've been created
func (o *AddOptions) Complete(name string, cmdline cmdline.Cmdline, args []string) (err error) {
	o.operation = "add"
	o.registryName = args[0]
	o.registryURL = args[1]
	o.user = "default"
	return nil
}

// Validate validates the AddOptions based on completed values
func (o *AddOptions) Validate() (err error) {
	err = util.ValidateURL(o.registryURL)
	if err != nil {
		return err
	}
	if util2.IsGitBasedRegistry(o.registryURL) {
		util2.PrintGitRegistryDeprecationWarning()
	}
	return nil
}

// Run contains the logic for "odo registry add" command
func (o *AddOptions) Run() (err error) {
	isSecure := false
	if o.tokenFlag != "" {
		isSecure = true
	}

	cfg, err := preference.New()
	if err != nil {
		return errors.Wrap(err, "unable to add registry")
	}
	err = cfg.RegistryHandler(o.operation, o.registryName, o.registryURL, false, isSecure)
	if err != nil {
		return err
	}

	if o.tokenFlag != "" {
		err = keyring.Set(util.CredentialPrefix+o.registryName, o.user, o.tokenFlag)
		if err != nil {
			return errors.Wrap(err, "unable to store registry credential to keyring")
		}
	}

	log.Info("New registry successfully added")
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

	registryAddCmd.Flags().StringVar(&o.tokenFlag, "token", "", "Token to be used to access secure registry")

	return registryAddCmd
}
