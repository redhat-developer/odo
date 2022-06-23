package add

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
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/registry"
	"github.com/redhat-developer/odo/pkg/util"
)

const registryCommandName = "registry"

// "odo preference add registry" command description and examples
var (
	addLongDesc = ktemplates.LongDesc(`Add devfile registry`)

	addExample = ktemplates.Examples(`# Add devfile registry
	%[1]s CheRegistry https://che-devfile-registry.openshift.io
	`)
)

// RegistryOptions encapsulates the options for the "odo preference add registry" command
type RegistryOptions struct {
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

// NewRegistryOptions creates a new RegistryOptions instance
func NewRegistryOptions() *RegistryOptions {
	return &RegistryOptions{}
}

func (o *RegistryOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete completes RegistryOptions after they've been created
func (o *RegistryOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.operation = "add"
	o.registryName = args[0]
	o.registryURL = args[1]
	o.user = "default"
	return nil
}

// Validate validates the RegistryOptions based on completed values
func (o *RegistryOptions) Validate() (err error) {
	err = util.ValidateURL(o.registryURL)
	if err != nil {
		return err
	}
	isGithubRegistry, err := registry.IsGithubBasedRegistry(o.registryURL)
	if err != nil {
		return err
	}
	if isGithubRegistry {
		return &registry.ErrGithubRegistryNotSupported{}
	}
	return nil
}

// Run contains the logic for "odo preference add registry" command
func (o *RegistryOptions) Run(ctx context.Context) (err error) {
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

// NewCmdRegistry implements the "odo preference add registry" command
func NewCmdRegistry(name, fullName string) *cobra.Command {
	o := NewRegistryOptions()
	registryCmd := &cobra.Command{
		Use:     fmt.Sprintf("%s <registry name> <registry URL>", name),
		Short:   addLongDesc,
		Long:    addLongDesc,
		Example: fmt.Sprintf(fmt.Sprint(addExample), fullName),
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(registryCmd, clientset.PREFERENCE)

	registryCmd.Flags().StringVar(&o.tokenFlag, "token", "", "Token to be used to access secure registry")

	return registryCmd
}
