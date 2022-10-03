package namespace

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	dfutil "github.com/devfile/library/pkg/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"
)

const RecommendedCommandName = "namespace"

var (
	createExample = ktemplates.Examples(`
	# Create a new namespace and set it as the current active namespace
	%[1]s my-namespace
	`)

	createLongDesc = ktemplates.LongDesc(`Create a new namespace.
	This command directly performs actions on the cluster and doesn't require a push.

	Any new namespace created with this command will also be set as the current active namespace.
	If executed inside a component directory, this command will not update the namespace of the existing component.
	`)

	createShortDesc = `Create a new namespace`
)

// NamespaceCreateOptions encapsulates the options for the odo namespace create command
type NamespaceCreateOptions struct {
	// Clients
	clientset *clientset.Clientset

	// Parameters
	namespaceName string

	// Flags
	waitFlag bool

	// value can be either 'project' or 'namespace', depending on what command is called
	commandName string
}

var _ genericclioptions.Runnable = (*NamespaceCreateOptions)(nil)

// NewNamespaceCreateOptions creates a NamespaceCreateOptions instance
func NewNamespaceCreateOptions() *NamespaceCreateOptions {
	return &NamespaceCreateOptions{}
}

func (o *NamespaceCreateOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete completes NamespaceCreateOptions after they've been created
func (nco *NamespaceCreateOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	nco.namespaceName = args[0]
	if scontext.GetTelemetryStatus(cmdline.Context()) {
		scontext.SetClusterType(cmdline.Context(), nco.clientset.KubernetesClient)
	}
	return nil
}

// Validate validates the parameters of the NamespaceCreateOptions
func (nco *NamespaceCreateOptions) Validate(ctx context.Context) error {
	return dfutil.ValidateK8sResourceName("namespace name", nco.namespaceName)
}

// Run runs the namespace create command
func (nco *NamespaceCreateOptions) Run(ctx context.Context) (err error) {
	// Create the "spinner"
	s := &log.Status{}

	// If the --wait parameter has been passed, we add a spinner..
	if nco.waitFlag {
		s = log.Spinnerf("Waiting for %s to come up", nco.commandName)
		defer s.End(false)
	}

	// Create the namespace & end the spinner (if there is any..)
	err = nco.clientset.ProjectClient.Create(nco.namespaceName, nco.waitFlag)
	if err != nil {
		return err
	}
	s.End(true)

	caser := cases.Title(language.Und)
	successMessage := fmt.Sprintf(`%s %q is ready for use`, caser.String(nco.commandName), nco.namespaceName)
	log.Successf(successMessage)

	// Set the current namespace when created
	err = nco.clientset.ProjectClient.SetCurrent(nco.namespaceName)
	if err != nil {
		return err
	}

	log.Successf("New %[1]s created and now using %[1]s: %v", nco.commandName, nco.namespaceName)

	return nil
}

// NewCmdNamespaceCreate creates the namespace create command
func NewCmdNamespaceCreate(name, fullName string) *cobra.Command {
	o := NewNamespaceCreateOptions()
	// To help the UI messages deal better with namespace vs project
	o.commandName = name
	if len(os.Args) > 2 {
		o.commandName = os.Args[2]
	}

	namespaceCreateCmd := &cobra.Command{
		Use:     name,
		Short:   createShortDesc,
		Long:    createLongDesc,
		Example: fmt.Sprintf(createExample, fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
		Annotations: map[string]string{"command": "main"},
		Aliases:     []string{"project"},
	}

	namespaceCreateCmd.Flags().BoolVarP(&o.waitFlag, "wait", "w", false, "Wait until the namespace is ready")

	clientset.Add(namespaceCreateCmd, clientset.KUBERNETES, clientset.PROJECT)

	return namespaceCreateCmd
}
