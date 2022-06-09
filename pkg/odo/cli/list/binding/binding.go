package binding

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
)

const RecommendedCommandName = "binding"

var (
	listExample = ktemplates.Examples(`
	# List all the bindings
    %[1]s`)
	listLongDesc = ktemplates.LongDesc(`
	List all the bindings
`)
)

// BindingListOptions encapsulates the options for the odo list binding command
type BindingListOptions struct {
	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset
}

// NewBindingListOptions creates a new BindingListOptions instance
func NewBindingListOptions() *BindingListOptions {
	return &BindingListOptions{}
}

func (o *BindingListOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete completes BindingListOptions after they've been created
func (o *BindingListOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline))
	return err
}

// Validate validates the BindingListOptions based on completed values
func (o *BindingListOptions) Validate() (err error) {
	return nil
}

// Run contains the logic for the odo list binding command
func (o *BindingListOptions) Run(ctx context.Context) error {

	return HumanReadableOutput(os.Stdout)
}
func (lo *BindingListOptions) RunForJsonOutput(ctx context.Context) (out interface{}, err error) {
	//list, err := lo.run(ctx)
	//if err != nil {
	//	return nil, err
	//}
	return nil, nil
}

// NewCmdBindingList implements the odo list binding command.
func NewCmdBindingList(name, fullName string) *cobra.Command {
	o := NewBindingListOptions()
	bindingListCmd := &cobra.Command{
		Use:     name,
		Short:   listLongDesc,
		Long:    listLongDesc,
		Example: fmt.Sprintf(listExample, fullName),
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	//	clientset.Add(bindingListCmd, clientset.PROJECT)
	machineoutput.UsedByCommand(bindingListCmd)
	return bindingListCmd
}

// HumanReadableOutput outputs the list of bindings in a human readable format
func HumanReadableOutput(w io.Writer) error {
	return nil
}
