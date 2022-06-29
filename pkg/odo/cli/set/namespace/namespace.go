package namespace

import (
	"context"
	"fmt"
	"os"

	dfutil "github.com/devfile/library/pkg/util"

	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"

	"k8s.io/klog"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/spf13/cobra"
)

const RecommendedCommandName = "namespace"

var (
	setExample = ktemplates.Examples(`
	# Set the specified namespace as the current active namespace in the config
	%[1]s my-namespace
	`)

	setLongDesc = ktemplates.LongDesc(`Set the current active namespace.
	
	If executed inside a component directory, this command will not update the namespace of the existing component.
	`)

	setShortDesc = `Set the current active namespace`
)

// SetOptions encapsulates the options for the odo namespace create command
type SetOptions struct {
	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset

	// Destination directory
	contextDir string

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
func (so *SetOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	so.namespaceName = args[0]
	so.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline))
	if err != nil {
		return err
	}
	so.contextDir, err = so.clientset.FS.Getwd()
	if err != nil {
		return err
	}
	if scontext.GetTelemetryStatus(cmdline.Context()) {
		scontext.SetClusterType(cmdline.Context(), so.KClient)
	}
	return nil
}

// Validate validates the parameters of the SetOptions
func (so *SetOptions) Validate() error {
	return dfutil.ValidateK8sResourceName("namespace name", so.namespaceName)
}

// Run runs the 'set namespace' command
func (so *SetOptions) Run(ctx context.Context) error {
	devfilePresent, err := location.DirectoryContainsDevfile(so.clientset.FS, so.contextDir)
	if err != nil {
		// Purposely ignoring the error, as it is not mandatory for this command
		klog.V(2).Infof("Unexpected error while checking if running inside a component directory: %v", err)
	}
	if devfilePresent {
		log.Warningf("This is being executed inside a component directory. This will not update the %s of the existing component",
			so.commandName)
	}

	err = so.clientset.ProjectClient.SetCurrent(so.namespaceName)
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

	clientset.Add(namespaceSetCmd, clientset.FILESYSTEM, clientset.PROJECT)

	return namespaceSetCmd
}
