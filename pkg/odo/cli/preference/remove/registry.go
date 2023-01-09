package remove

import (
	"context"
	// Built-in packages
	"fmt"

	// Third-party packages
	dfutil "github.com/devfile/library/v2/pkg/util"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	// odo packages
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	registryUtil "github.com/redhat-developer/odo/pkg/registry"
)

const registryCommandName = "registry"

// "odo preference delete registry" command description and examples
var (
	removeLongDesc = ktemplates.LongDesc(`Remove devfile registry`)

	removeExample = ktemplates.Examples(`# Remove devfile registry
	%[1]s CheRegistry
	`)
)

// RegistryOptions encapsulates the options for the "odo preference remove registry" command
type RegistryOptions struct {
	// Clients
	clientset *clientset.Clientset

	// Parameters
	registryName string

	// Flags
	forceFlag bool

	operation   string
	registryURL string
	user        string
}

var _ genericclioptions.Runnable = (*RegistryOptions)(nil)

// NewRegistryOptions creates a new RegistryOptions instance
func NewRegistryOptions() *RegistryOptions {
	return &RegistryOptions{}
}

func (o *RegistryOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete completes RegistryOptions after they've been created
func (o *RegistryOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	o.operation = "remove"
	o.registryName = args[0]
	o.registryURL = ""
	o.user = "default"
	return nil
}

// Validate validates the RegistryOptions based on completed values
func (o *RegistryOptions) Validate(ctx context.Context) (err error) {
	return nil
}

// Run contains the logic for "odo preference remove registry" command
func (o *RegistryOptions) Run(ctx context.Context) (err error) {
	isSecure := registryUtil.IsSecure(o.clientset.PreferenceClient, o.registryName)
	err = o.clientset.PreferenceClient.RegistryHandler(o.operation, o.registryName, o.registryURL, o.forceFlag, false)
	if err != nil {
		return err
	}

	if isSecure {
		err = keyring.Delete(dfutil.CredentialPrefix+o.registryName, o.user)
		if err != nil {
			return fmt.Errorf("unable to remove registry credential from keyring: %w", err)
		}
	}

	return nil
}

// NewCmdRegistry implements the "odo preference remove registry" command
func NewCmdRegistry(name, fullName string) *cobra.Command {
	o := NewRegistryOptions()
	registryDeleteCmd := &cobra.Command{
		Use:     fmt.Sprintf("%s <registry name>", name),
		Short:   removeLongDesc,
		Long:    removeLongDesc,
		Example: fmt.Sprintf(fmt.Sprint(removeExample), fullName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(registryDeleteCmd, clientset.PREFERENCE)

	registryDeleteCmd.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Don't ask for confirmation, remove the registry directly")

	return registryDeleteCmd
}
