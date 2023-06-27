package namespace

import (
	"context"
	"fmt"
	"github.com/redhat-developer/odo/pkg/project"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"k8s.io/klog"
	"os"
	"time"

	dfutil "github.com/devfile/library/v2/pkg/util"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/odo/util"
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
	createSpinner := log.Spinnerf("Creating the %s %q", nco.commandName, nco.namespaceName)
	defer createSpinner.End(false)

	// Create the namespace & end the spinner (if there is any..)
	err = nco.clientset.ProjectClient.Create(nco.namespaceName, nco.waitFlag)
	if err != nil {
		return err
	}
	createSpinner.End(true)

	// If the --wait parameter has been passed, we add a spinner..
	if nco.waitFlag {
		waitSpinner := log.Spinnerf("Waiting for the %s to come up", nco.commandName)
		defer waitSpinner.End(false)
		timeOut := time.After(nco.clientset.PreferenceClient.GetTimeout())
	L:
		for {
			select {
			case <-timeOut:
				return fmt.Errorf("timeout while waiting for %s %q to be ready; you can change the timeout preference by running `odo preference set timeout <duration>`", nco.commandName, nco.namespaceName)
			default:
				var nsList project.ProjectList
				nsList, err = nco.clientset.ProjectClient.List()
				if err != nil {
					klog.V(4).Infof("Failed to list %ss", nco.commandName)
				}
				for _, ns := range nsList.Items {
					if ns.Name == nco.namespaceName {
						break L
					}
				}
				time.Sleep(50 * time.Millisecond)
			}
		}
		waitSpinner.End(true)
	}

	// Set the current namespace when created
	err = nco.clientset.ProjectClient.SetCurrent(nco.namespaceName)
	if err != nil {
		return err
	}
	caser := cases.Title(language.Und)
	successMessage := fmt.Sprintf(`%s %q is ready for use`, caser.String(nco.commandName), nco.namespaceName)
	log.Successf(successMessage)
	log.Successf("New %[1]s created and now using %[1]s: %v", nco.commandName, nco.namespaceName)

	return nil
}

// NewCmdNamespaceCreate creates the namespace create command
func NewCmdNamespaceCreate(name, fullName string, testClientset clientset.Clientset) *cobra.Command {
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return genericclioptions.GenericRun(o, testClientset, cmd, args)
		},
		Aliases: []string{"project"},
	}

	namespaceCreateCmd.Flags().BoolVarP(&o.waitFlag, "wait", "w", false, "Wait until the namespace is ready")

	clientset.Add(namespaceCreateCmd, clientset.KUBERNETES, clientset.PROJECT, clientset.PREFERENCE)
	util.SetCommandGroup(namespaceCreateCmd, util.MainGroup)

	return namespaceCreateCmd
}
