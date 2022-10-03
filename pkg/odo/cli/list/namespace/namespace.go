package namespace

import (
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	"github.com/redhat-developer/odo/pkg/project"
)

const RecommendedCommandName = "namespace"

var (
	listExample = ktemplates.Examples(`
	# List all the namespaces
    %[1]s`)
	listLongDesc = ktemplates.LongDesc(`
	List all the namespaces
`)
)

// NamespaceListOptions encapsulates the options for the odo list project command
type NamespaceListOptions struct {
	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset

	commandName string
}

var _ genericclioptions.Runnable = (*NamespaceListOptions)(nil)

// NewNamespaceListOptions creates a new NamespaceListOptions instance
func NewNamespaceListOptions() *NamespaceListOptions {
	return &NamespaceListOptions{}
}

func (o *NamespaceListOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete completes NamespaceListOptions after they've been created
func (plo *NamespaceListOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	plo.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline))
	return err
}

// Validate validates the NamespaceListOptions based on completed values
func (plo *NamespaceListOptions) Validate(ctx context.Context) (err error) {
	return nil
}

// Run contains the logic for the odo list project command
func (plo *NamespaceListOptions) Run(ctx context.Context) error {
	namespaces, err := plo.clientset.ProjectClient.List()
	if err != nil {
		return err
	}

	return HumanReadableOutput(os.Stdout, namespaces, plo.commandName)
}

// NewCmdNamespaceList implements the odo list project command.
func NewCmdNamespaceList(name, fullName string) *cobra.Command {
	o := NewNamespaceListOptions()
	projectListCmd := &cobra.Command{
		Use:     name,
		Short:   listLongDesc,
		Long:    listLongDesc,
		Example: fmt.Sprintf(listExample, fullName),
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
		Aliases: []string{"namespaces", "project", "projects"},
	}
	clientset.Add(projectListCmd, clientset.PROJECT)
	return projectListCmd
}

// HumanReadableOutput outputs the list of namespaces in a human readable format
func HumanReadableOutput(w io.Writer, o project.ProjectList, commandName string) error {
	if len(o.Items) == 0 {
		return fmt.Errorf("you are not a member of any %[1]ss. You can request a %[1]s to be created using the `odo create %[1]s <%[1]s_name>` command", commandName)
	}
	wr := tabwriter.NewWriter(w, 5, 2, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(wr, "ACTIVE", "\t", "NAME")
	for _, project := range o.Items {
		activeMark := " "
		if project.Status.Active {
			activeMark = "*"
		}
		fmt.Fprintln(wr, activeMark, "\t", project.Name)
	}
	wr.Flush()
	return nil
}
