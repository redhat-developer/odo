package namespace

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/odo/util"
	scontext "github.com/redhat-developer/odo/pkg/segment/context"
)

const RecommendedCommandName = "namespace"

var (
	deleteExample = ktemplates.Examples(`
	# Delete a namespace
    %[1]s my-namespace`)

	deleteLongDesc = ktemplates.LongDesc(`
	Delete the specified namespace
	`)

	deleteShortDesc = `Delete a namespace`
)

// DeleteOptions encapsulates the options for the 'odo delete namespace' command
type DeleteOptions struct {
	// Clients
	clientset *clientset.Clientset

	// Parameters
	namespaceName string

	// Flags
	waitFlag bool

	// forceFlag forces deletion
	forceFlag bool

	// value can be either 'project' or 'namespace', depending on what command is called
	commandName string
}

var _ genericclioptions.Runnable = (*DeleteOptions)(nil)

// NewDeleteOptions creates a new DeleteOptions instance
func NewDeleteOptions() *DeleteOptions {
	return &DeleteOptions{}
}

func (do *DeleteOptions) SetClientset(clientset *clientset.Clientset) {
	do.clientset = clientset
}

// Complete completes DeleteOptions after they've been created
func (do *DeleteOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	do.namespaceName = args[0]
	if scontext.GetTelemetryStatus(cmdline.Context()) {
		scontext.SetClusterType(cmdline.Context(), do.clientset.KubernetesClient)
	}
	return nil
}

// Validate validates the DeleteOptions based on completed values
func (do *DeleteOptions) Validate(ctx context.Context) (err error) {
	return nil
}

// Run contains the logic for the 'odo delete namespace' command
func (do *DeleteOptions) Run(ctx context.Context) error {
	log.Infof("Searching %s to delete, please wait...", do.commandName)
	exists, err := do.clientset.ProjectClient.Exists(do.namespaceName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("No %s named %q found", do.commandName, do.namespaceName)
	}
	proceed := do.forceFlag
	if !proceed {
		proceed, err = ui.Proceed(fmt.Sprintf("Are you sure you want to delete %s %q?", do.commandName, do.namespaceName))
		if err != nil {
			return err
		}
	}
	if proceed {
		// Create the "spinner"
		s := &log.Status{}

		// If the --wait parameter has been passed, we add a spinner..
		if do.waitFlag {
			s = log.Spinnerf("Waiting for %s %q to be deleted", do.commandName, do.namespaceName)
			defer s.End(false)
		}

		err := do.clientset.ProjectClient.Delete(do.namespaceName, do.waitFlag)
		if err != nil {
			return err
		}
		s.End(true)

		cmdNameTitled := cases.Title(language.Und).String(do.commandName)
		msg := fmt.Sprintf("%s %q deleted", cmdNameTitled, do.namespaceName)
		if !do.waitFlag {
			msg = fmt.Sprintf("%s %q will be deleted asynchronously", cmdNameTitled, do.namespaceName)
		}
		log.Success(msg)

		return nil
	}
	log.Errorf("Aborting %s deletion", do.commandName)
	return nil
}

// NewCmdNamespaceDelete implements the 'odo delete namespace' command.
func NewCmdNamespaceDelete(name, fullName string, testClientset clientset.Clientset) *cobra.Command {
	do := NewDeleteOptions()
	// To help the UI messages deal better with namespace vs project
	do.commandName = name
	if len(os.Args) > 2 {
		do.commandName = os.Args[2]
	}
	namespaceDeleteCmd := &cobra.Command{
		Use:     name,
		Short:   deleteShortDesc,
		Long:    deleteLongDesc,
		Example: fmt.Sprintf(deleteExample, fullName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return genericclioptions.GenericRun(do, testClientset, cmd, args)
		},
		Aliases: []string{"project"},
	}

	namespaceDeleteCmd.Flags().BoolVarP(&do.forceFlag, "force", "f", false, "Delete namespace without prompting")
	namespaceDeleteCmd.Flags().BoolVarP(
		&do.waitFlag,
		"wait", "w", false,
		"Wait until the namespace no longer exists")

	clientset.Add(namespaceDeleteCmd, clientset.KUBERNETES, clientset.PROJECT)
	util.SetCommandGroup(namespaceDeleteCmd, util.MainGroup)
	return namespaceDeleteCmd
}
