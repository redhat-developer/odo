package list

import (
	"context"
	"errors"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/cli/list/binding"
	"github.com/redhat-developer/odo/pkg/odo/cli/list/namespace"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/util"

	dfutil "github.com/devfile/library/pkg/util"

	"github.com/redhat-developer/odo/pkg/component"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"

	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

// RecommendedCommandName is the recommended list name
const RecommendedCommandName = "list"

var listExample = ktemplates.Examples(`  # List all components in the application
%[1]s
  `)

// ListOptions ...
type ListOptions struct {
	// Context
	*genericclioptions.Context

	// Clients
	clientset *clientset.Clientset

	// Local variables
	project         string
	namespaceFilter string
	devfilePath     string
	localComponent  api.ComponentAbstract

	// Flags
	namespaceFlag string
}

var _ genericclioptions.Runnable = (*ListOptions)(nil)
var _ genericclioptions.JsonOutputter = (*ListOptions)(nil)

// NewListOptions ...
func NewListOptions() *ListOptions {
	return &ListOptions{}
}

func (o *ListOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete ...
func (lo *ListOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {

	// Check to see if KUBECONFIG exists, and if not, error the user that we would not be able to get cluster information
	// Do this before anything else, or else we will just error out with the:
	// invalid configuration: no configuration has been provided, try setting KUBERNETES_MASTER environment variable
	// instead
	if !dfutil.CheckKubeConfigExist() {
		return errors.New("KUBECONFIG not found. Unable to retrieve cluster information. Please set your Kubernetes configuration via KUBECONFIG env variable or ~/.kube/config")
	}

	// Create the local context and initial Kubernetes client configuration
	lo.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline))
	if err != nil {
		return err
	}

	// Check for the Devfile and then retrieve all information regarding the local Devfile
	lo.devfilePath = location.DevfileLocation("")
	if util.CheckPathExists(lo.devfilePath) {

		// Set the project / namespace based on the devfile context
		lo.project = lo.Context.GetProject()

		// Parse the devfile
		devObj, parseErr := devfile.ParseAndValidateFromFile(lo.devfilePath)
		if parseErr != nil {
			return parseErr
		}

		// Create a local component from the parse devfile
		localComponent := api.ComponentAbstract{
			Name:      devObj.Data.GetMetadata().Name,
			ManagedBy: "",
			RunningIn: []api.RunningMode{},
			Type:      component.GetComponentTypeFromDevfileMetadata(devObj.Data.GetMetadata()),
		}

		lo.localComponent = localComponent

	}

	// If the context is "" (devfile.yaml not found..), we get the active one from KUBECONFIG.
	if lo.project == "" {
		lo.project = lo.clientset.KubernetesClient.GetCurrentNamespace()
	}

	// If the namespace flag has been passed, we will search there.
	// if it hasn't, we will search from the default project / namespace.
	if lo.namespaceFlag != "" {
		lo.namespaceFilter = lo.namespaceFlag
	} else {
		lo.namespaceFilter = lo.project
	}

	return
}

// Validate ...
func (lo *ListOptions) Validate() (err error) {
	return nil
}

// Run has the logic to perform the required actions as part of command
func (lo *ListOptions) Run(ctx context.Context) error {
	list, err := lo.run(ctx)
	if err != nil {
		return err
	}
	humanReadableOutput(list)
	return nil
}

// Run contains the logic for the odo command
func (lo *ListOptions) RunForJsonOutput(ctx context.Context) (out interface{}, err error) {
	list, err := lo.run(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (lo *ListOptions) run(cts context.Context) (api.ResourcesList, error) {
	listSpinner := log.Spinnerf("Listing components from namespace '%s'", lo.namespaceFilter)
	defer listSpinner.End(false)

	// Step 1.
	// Retrieve all related components from the Kubernetes cluster, from the given namespace
	devfileComponents, err := component.ListAllClusterComponents(lo.clientset.KubernetesClient, lo.namespaceFilter)
	if err != nil {
		return api.ResourcesList{}, err
	}
	listSpinner.End(true)

	// Step 2.
	// If we have a local component, let's add it to the list of Devfiles
	// This checks lo.localComponent.Name. If it's empty, we didn't parse one in the Complete() function, so there is no local devfile.
	// We will only append the local component to the devfile if it doesn't exist in the list.
	componentInDevfile := ""
	if lo.localComponent.Name != "" {
		if !component.Contains(lo.localComponent, devfileComponents) {
			devfileComponents = append(devfileComponents, lo.localComponent)
		}
		componentInDevfile = lo.localComponent.Name
	}
	return api.ResourcesList{
		ComponentInDevfile: componentInDevfile,
		Components:         devfileComponents,
	}, nil
}

// NewCmdList implements the list odo command
func NewCmdList(name, fullName string) *cobra.Command {
	o := NewListOptions()

	var listCmd = &cobra.Command{
		Use:         name,
		Short:       "List all components in the current namespace",
		Long:        "List all components in the current namespace.",
		Example:     fmt.Sprintf(listExample, fullName),
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"command": "management"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(listCmd, clientset.KUBERNETES)

	namespaceCmd := namespace.NewCmdNamespaceList(namespace.RecommendedCommandName, odoutil.GetFullName(fullName, namespace.RecommendedCommandName))
	bindingCmd := binding.NewCmdBindingList(binding.RecommendedCommandName, odoutil.GetFullName(fullName, binding.RecommendedCommandName))
	listCmd.AddCommand(namespaceCmd, bindingCmd)

	listCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	listCmd.Flags().StringVar(&o.namespaceFlag, "namespace", "", "Namespace for odo to scan for components")

	completion.RegisterCommandFlagHandler(listCmd, "path", completion.FileCompletionHandler)
	machineoutput.UsedByCommand(listCmd)

	return listCmd
}

func humanReadableOutput(list api.ResourcesList) {
	components := list.Components
	if len(components) == 0 {
		log.Error("There are no components deployed.")
		return
	}

	t := ui.NewTable()

	// Create the header and then sort accordingly
	t.AppendHeader(table.Row{"NAME", "PROJECT TYPE", "RUNNING IN", "MANAGED"})
	t.SortBy([]table.SortBy{
		{Name: "MANAGED", Mode: table.Asc},
		{Name: "NAME", Mode: table.Dsc},
	})

	// Go through each component and add it to the table
	for _, comp := range components {

		// Mark the name as yellow in the index to it's easier to see.
		name := text.Colors{text.FgHiYellow}.Sprint(comp.Name)

		// Get the managed by label
		managedBy := comp.ManagedBy
		if managedBy == "" {
			managedBy = api.TypeUnknown
		}

		// Get the mode (dev or deploy)
		mode := comp.RunningIn.String()

		// Get the type of the component
		componentType := comp.Type
		if componentType == "" {
			componentType = api.TypeUnknown
		}

		// If we find our local unpushed component, let's change the output appropriately.
		if list.ComponentInDevfile == comp.Name {
			name = fmt.Sprintf("* %s", name)

			if comp.ManagedBy == "" {
				managedBy = "odo"
			}
		}

		// If we are managing that component, output it as blue (our logo colour) to indicate it's used by odo
		if managedBy == "odo" {
			managedBy = text.Colors{text.FgBlue}.Sprintf("odo (%s)", comp.ManagedByVersion)
		} else if managedBy != "" && comp.ManagedByVersion != "" {
			// this is done to maintain the color of the output
			managedBy += fmt.Sprintf("(%s)", comp.ManagedByVersion)
		}

		t.AppendRow(table.Row{name, componentType, mode, managedBy})
	}
	t.Render()

}
