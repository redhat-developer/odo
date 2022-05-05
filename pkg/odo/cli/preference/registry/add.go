package registry

import (
	"context"
	// Built-in packages
	"fmt"

	// Third-party packages
	dfutil "github.com/devfile/library/pkg/util"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	// odo packages
	"github.com/redhat-developer/odo/pkg/log"
	util2 "github.com/redhat-developer/odo/pkg/odo/cli/preference/registry/util"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/util"
)

const addCommandName = "add"

// "odo preference registry add" command description and examples
var (
	addLongDesc = ktemplates.LongDesc(`Add devfile registry`)

	addExample = ktemplates.Examples(`# Add devfile registry
	%[1]s CheRegistry https://che-devfile-registry.openshift.io

	%[1]s RegistryFromGitHub https://github.com/elsony/devfile-registry
	`)
)

// AddOptions encapsulates the options for the "odo preference registry add" command
type AddOptions struct {
	// Clients
	clientset *clientset.Clientset

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

func (o *AddOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete completes AddOptions after they've been created
func (o *AddOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
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
	isGithubRegistry, err := util2.IsGithubBasedRegistry(o.registryURL)
	if err != nil {
		return err
	}
	if isGithubRegistry {
		return util2.ErrGithubRegistryNotSupported
	}
	return nil
}

// Run contains the logic for "odo preference registry add" command
func (o *AddOptions) Run(ctx context.Context) (err error) {
	isSecure := false
	if o.tokenFlag != "" {
		isSecure = true
	}

	err = o.clientset.PreferenceClient.RegistryHandler(o.operation, o.registryName, o.registryURL, false, isSecure)
	if err != nil {
		return err
	}

	if o.tokenFlag != "" {
		err = keyring.Set(dfutil.CredentialPrefix+o.registryName, o.user, o.tokenFlag)
		if err != nil {
			return fmt.Errorf("unable to store registry credential to keyring: %w", err)
		}
	}

	log.Info("New registry successfully added")
	return nil
}

// NewCmdAdd implements the "odo preference registry add" command
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
	clientset.Add(registryAddCmd, clientset.PREFERENCE)

	registryAddCmd.Flags().StringVar(&o.tokenFlag, "token", "", "Token to be used to access secure registry")

	return registryAddCmd
}
