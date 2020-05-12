package registry

import (
	// Built-in packages
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	// Third-party packages
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	// odo packages
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util"
	"github.com/openshift/odo/pkg/preference"
)

const listCommandName = "list"

// "odo registry list" command description and examples
var (
	listDesc = ktemplates.LongDesc(`List devfile registry`)

	listExample = ktemplates.Examples(`# List devfile registry
	%[1]s
	`)
)

// ListOptions encapsulates the options for "odo registry list" command
type ListOptions struct {
}

// NewListOptions creates a new ListOptions instance
func NewListOptions() *ListOptions {
	return &ListOptions{}
}

// Complete completes ListOptions after they've been created
func (o *ListOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	return
}

// Validate validates the ListOptions based on completed values
func (o *ListOptions) Validate() (err error) {
	return
}

// Run contains the logic for "odo registry list" command
func (o *ListOptions) Run() (err error) {
	cfg, err := preference.New()
	if err != nil {
		util.LogErrorAndExit(err, "")
	}

	w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "NAME", "\t", "URL")
	o.printRegistryList(w, cfg.OdoSettings.RegistryList)
	w.Flush()
	return
}

func (o *ListOptions) printRegistryList(w io.Writer, registryList *[]preference.Registry) {
	if registryList == nil {
		return
	}

	for _, registry := range *registryList {
		fmt.Fprintln(w, registry.Name, "\t", registry.URL)
	}
}

// NewCmdList implements the "odo registry list" command
func NewCmdList(name, fullName string) *cobra.Command {
	o := NewListOptions()
	registryListCmd := &cobra.Command{
		Use:     name,
		Short:   listDesc,
		Long:    listDesc,
		Example: fmt.Sprintf(fmt.Sprint(listExample), fullName),
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	return registryListCmd
}
