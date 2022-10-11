package namespace

import (
	"context"
	"fmt"
	"os"

	dfutil "github.com/devfile/library/pkg/util"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"

	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/spf13/cobra"
)

const RecommendedCommandName = "namespace"

var (
	setExample = ktemplates.Examples(`
	# Set the specified namespace as the current active namespace in your local kubeconfig configuration
	%[1]s my-namespace
	`)

	setLongDesc = ktemplates.LongDesc(`Set the current active namespace.
	`)

	setShortDesc = `Set the current active namespace`
)

// SetOptions encapsulates the options for the odo namespace create command
type SetOptions struct {
	// Clients
	clientset *clientset.Clientset

	// Parameters
	namespaceName string

	// value can be either 'project' or 'namespace', depending on what command is called
	commandName string
}

var _ genericclioptions.Runnable = (*SetOptions)(nil)

// NewSetOptions creates a SetOptions instance
func NewSetOptions() *SetOptions {
	return &SetOptions{}
}

func (so *SetOptions) SetClientset(clientset *clientset.Clientset) {
	so.clientset = clientset
}

// Complete completes SetOptions after they've been created
func (so *SetOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	so.namespaceName = args[0]
	if scontext.GetTelemetryStatus(cmdline.Context()) {
		scontext.SetClusterType(cmdline.Context(), so.clientset.KubernetesClient)
	}
	return nil
}

// Validate validates the parameters of the SetOptions
func (so *SetOptions) Validate(ctx context.Context) error {
	return dfutil.ValidateK8sResourceName("namespace name", so.namespaceName)
}

// Run runs the 'set namespace' command
func (so *SetOptions) Run(ctx context.Context) error {
	err := so.clientset.ProjectClient.SetCurrent(so.namespaceName)
	if err != nil {
		return err
	}

	log.Successf("Current active %[1]s set to %q", so.commandName, so.namespaceName)

	return nil
}

// NewCmdNamespaceSet creates the 'set namespace' command
func NewCmdNamespaceSet(name, fullName string) *cobra.Command {
	o := NewSetOptions()
	// To help the UI messages deal better with namespace vs project
	o.commandName = name
	if len(os.Args) > 2 {
		o.commandName = os.Args[2]
	}

	namespaceSetCmd := &cobra.Command{
		Use:     name,
		Short:   setShortDesc,
		Long:    setLongDesc,
		Example: fmt.Sprintf(setExample, fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
		Annotations: map[string]string{"command": "main"},
		Aliases:     []string{"project"},
	}

	clientset.Add(namespaceSetCmd, clientset.KUBERNETES, clientset.FILESYSTEM, clientset.PROJECT)

	return namespaceSetCmd
}
