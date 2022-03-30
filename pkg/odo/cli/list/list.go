package list

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/spf13/cobra"

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
	localComponent  component.OdoComponent

	// Flags
	namespaceFlag string
}

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
		localComponent := component.OdoComponent{
			Name:      devObj.Data.GetMetadata().Name,
			ManagedBy: "",
			Modes:     map[string]bool{},
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

	listSpinner := log.Spinnerf("Listing components from namespace '%s'", lo.namespaceFilter)
	defer listSpinner.End(false)

	// Step 1.
	// Retrieve all related components from the Kubernetes cluster, from the given namespace
	devfileComponents, err := component.ListAllClusterComponents(lo.clientset.KubernetesClient, lo.namespaceFilter)
	if err != nil {
		return err
	}
	listSpinner.End(true)

	// Step 2.
	// If we have a local component, let's add it to the list of Devfiles
	// This checks lo.localComponent.Name. If it's empty, we didn't parse one in the Complete() function, so there is no local devfile.
	// We will only append the local component to the devfile if it doesn't exist in the list.
	if (lo.localComponent.Name != "") && !component.Contains(lo.localComponent, devfileComponents) {
		devfileComponents = append(devfileComponents, lo.localComponent)
	}

	if log.IsJSON() {
		machineoutput.OutputSuccess(devfileComponents)
	} else {
		lo.HumanReadableOutput(log.GetStdout(), devfileComponents)
	}

	return nil
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
		Annotations: map[string]string{"machineoutput": "json", "command": "component"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	clientset.Add(listCmd, clientset.KUBERNETES)

	listCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)
	listCmd.Flags().StringVar(&o.namespaceFlag, "namespace", "", "Namespace for odo to scan for components")

	completion.RegisterCommandFlagHandler(listCmd, "path", completion.FileCompletionHandler)

	return listCmd
}

func (lo *ListOptions) HumanReadableOutput(wr io.Writer, components []component.OdoComponent) {
	if len(components) == 0 {
		log.Info("There are no components deployed.")
		return
	}

	if len(components) != 0 {

		// Create the table and use our own style
		t := table.NewWriter()

		// Set the style of the table
		t.SetStyle(table.Style{
			Box: table.BoxStyle{
				PaddingLeft:  " ",
				PaddingRight: " ",
			},
			Color: table.ColorOptions{
				Header: text.Colors{text.FgHiGreen, text.Underline},
			},
			Format: table.FormatOptions{
				Footer: text.FormatUpper,
				Header: text.FormatUpper,
				Row:    text.FormatDefault,
			},
			Options: table.Options{
				DrawBorder:      false,
				SeparateColumns: false,
				SeparateFooter:  false,
				SeparateHeader:  false,
				SeparateRows:    false,
			},
		})
		t.SetOutputMirror(log.GetStdout())

		// Create the header and then sort accordingly
		t.AppendHeader(table.Row{"NAME", "PROJECT TYPE", "RUNNING IN", "MANAGED"})
		t.SortBy([]table.SortBy{
			{Name: "MANAGED", Mode: table.Asc},
			{Name: "NAME", Mode: table.Dsc},
		})

		// Go through each componment and add it to the table
		for _, comp := range components {

			// Mark the name as yellow in the index to it's easier to see.
			name := text.Colors{text.FgHiYellow}.Sprint(comp.Name)

			// Get the managed by label
			managedBy := comp.ManagedBy
			if managedBy == "" {
				managedBy = component.StateTypeUnknown
			}

			// Get the mode (dev or deploy)
			modes := comp.Modes
			var mode string
			if len(modes) == 0 {
				mode = component.StateTypeUnknown
			} else {
				keys := make([]string, 0, len(modes))
				for k := range modes {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				mode = strings.Join(keys, ", ")
			}

			// Get the type of the component
			componentType := comp.Type
			if componentType == "" {
				componentType = component.StateTypeUnknown
			}

			// If we find our local unpushed component, let's change the output appropriately.
			if (lo.localComponent.Name == comp.Name) && (lo.localComponent.Type == comp.Type) {
				name = fmt.Sprintf("* %s", name)

				// If we found the local component, but mode len is 0, we will set it to "None" since it's not pushed..
				if len(modes) == 0 {
					mode = component.StateTypeNone
				}

				if comp.ManagedBy == "" {
					managedBy = "odo"
				}
			}

			// If we are managing that component, output it as blue (our logo colour) to indicate it's used by odo
			if managedBy == "odo" {
				managedBy = text.Colors{text.FgBlue}.Sprint("odo")
			}

			t.AppendRow(table.Row{name, componentType, mode, managedBy})
		}
		t.Render()
	}
}
