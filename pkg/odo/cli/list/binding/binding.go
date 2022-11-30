package binding

import (
	"context"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/runtime/schema"
	ktemplates "k8s.io/kubectl/pkg/util/templates"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
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
	// Clients
	clientset *clientset.Clientset

	// Flags
	namespaceFlag string
}

var _ genericclioptions.Runnable = (*BindingListOptions)(nil)
var _ genericclioptions.JsonOutputter = (*BindingListOptions)(nil)

// NewBindingListOptions creates a new BindingListOptions instance
func NewBindingListOptions() *BindingListOptions {
	return &BindingListOptions{}
}

func (o *BindingListOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete completes BindingListOptions after they've been created
func (o *BindingListOptions) Complete(ctx context.Context, cmdline cmdline.Cmdline, args []string) (err error) {
	if o.namespaceFlag != "" {
		o.clientset.KubernetesClient.SetNamespace(o.namespaceFlag)
	}
	return nil
}

// Validate validates the BindingListOptions based on completed values
func (o *BindingListOptions) Validate(ctx context.Context) (err error) {
	return nil
}

// Run contains the logic for the odo list binding command
func (o *BindingListOptions) Run(ctx context.Context) error {
	listSpinner := log.Spinnerf("Listing ServiceBindings from the namespace %q", o.clientset.KubernetesClient.GetCurrentNamespace())
	defer listSpinner.End(false)

	list, err := o.run(ctx)
	if err != nil {
		return err
	}

	listSpinner.End(true)

	HumanReadableOutput(o.clientset.KubernetesClient.GetCurrentNamespace(), list)
	return nil
}

func (o *BindingListOptions) RunForJsonOutput(ctx context.Context) (out interface{}, err error) {
	return o.run(ctx)
}

func (o *BindingListOptions) run(ctx context.Context) (api.ResourcesList, error) {
	var (
		workingDir = odocontext.GetWorkingDirectory(ctx)
		devfileObj = odocontext.GetDevfileObj(ctx)
	)
	bindings, inDevfile, err := o.clientset.BindingClient.ListAllBindings(devfileObj, workingDir)
	if err != nil {
		return api.ResourcesList{}, err
	}
	return api.ResourcesList{
		BindingsInDevfile: inDevfile,
		Bindings:          bindings,
	}, nil

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
		RunE: func(cmd *cobra.Command, args []string) error {
			return genericclioptions.GenericRun(o, cmd, args)
		},
		Aliases: []string{"bindings"},
	}
	clientset.Add(bindingListCmd, clientset.KUBERNETES, clientset.BINDING, clientset.FILESYSTEM)
	bindingListCmd.Flags().StringVar(&o.namespaceFlag, "namespace", "", "Namespace for odo to scan for bindings")
	commonflags.UseOutputFlag(bindingListCmd)
	return bindingListCmd
}

// HumanReadableOutput outputs the list of bindings in a human readable format
func HumanReadableOutput(namespace string, list api.ResourcesList) {
	bindings := list.Bindings
	if len(bindings) == 0 {
		log.Errorf("There are no service bindings in the %q namespace.", namespace)
		return

	}

	t := ui.NewTable()

	// Create the header and then sort accordingly
	t.AppendHeader(table.Row{"NAME", "APPLICATION", "SERVICES", "RUNNING IN"})
	t.SortBy([]table.SortBy{
		{Name: "NAME", Mode: table.Asc},
	})

	for _, binding := range bindings {

		// Mark the name as yellow in the index to it's easier to see.
		name := text.Colors{text.FgHiYellow}.Sprint(binding.Name)

		for _, inDevfile := range list.BindingsInDevfile {
			if binding.Name == inDevfile {
				name = fmt.Sprintf("* %s", name)
				break
			}
		}

		appSpec := binding.Spec.Application
		application := fmt.Sprintf("%s (%s)", appSpec.Name, appSpec.Kind)

		servicesSpecs := binding.Spec.Services
		services := ""
		for i, serviceSpec := range servicesSpecs {
			if i > 0 {
				services += "\n"
			}
			group := schema.FromAPIVersionAndKind(serviceSpec.APIVersion, "").Group
			if group != "" {
				group = "." + group
			}
			svcDesc := fmt.Sprintf("%s (%s%s)",
				serviceSpec.Name,
				serviceSpec.Kind,
				group,
			)
			if serviceSpec.Namespace != "" {
				svcDesc += fmt.Sprintf(" (namespace: %s)", serviceSpec.Namespace)
			}
			services += svcDesc
		}

		runningIn := "None"
		if binding.Status != nil && len(binding.Status.RunningIn) > 0 {
			runningIn = binding.Status.RunningIn.String()
		}
		t.AppendRow(table.Row{name, application, services, runningIn})
	}
	t.Render()
}
