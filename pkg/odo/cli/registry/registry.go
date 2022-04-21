package registry

import (
	"context"
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions/clientset"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/registry"
	"github.com/redhat-developer/odo/pkg/util"
	"github.com/spf13/cobra"
)

const RecommendedCommandName = "registry"

var Example = `  # Get all devfile components
  %[1]s

# Filter by name
%[1]s --filter nodejs

# Show more details
%[1]s --details`

// ListOptions encapsulates the options for the odo registry command
type ListOptions struct {
	clientset *clientset.Clientset

	// List of known devfiles
	DevfileList registry.DevfileStackList

	// Flags
	filterFlag          string
	devfileFlag         string
	devfileRegistryFlag string
	detailsFlag         bool
}

// NewListOptions creates a new ListOptions instance
func NewListOptions() *ListOptions {
	return &ListOptions{}
}

func (o *ListOptions) SetClientset(clientset *clientset.Clientset) {
	o.clientset = clientset
}

// Complete completes ListOptions after they've been created
func (o *ListOptions) Complete(cmdline cmdline.Cmdline, args []string) (err error) {

	o.DevfileList, err = o.clientset.RegistryClient.ListDevfileStacks("")
	if err != nil {
		return err
	}

	if o.DevfileList.DevfileRegistries == nil {
		log.Warning("Please run 'odo preference registry add <registry name> <registry URL>' to add registry for listing devfile components\n")
	}

	return nil
}

// Validate validates the ListOptions based on completed values
func (o *ListOptions) Validate() error {
	if len(o.DevfileList.Items) == 0 {
		return fmt.Errorf("no deployable components found")
	}
	return nil
}

// Run contains the logic for the command associated with ListOptions
func (o *ListOptions) Run(ctx context.Context) (err error) {
	o.printDevfileList(o.DevfileList.Items)
	return nil
}

func NewCmdRegistry(name, fullName string) *cobra.Command {
	o := NewListOptions()

	var listCmd = &cobra.Command{
		Use:     name,
		Short:   "List all components from the Devfile registry",
		Long:    "List all components from the Devfile registry",
		Example: fmt.Sprintf(Example, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

	clientset.Add(listCmd, clientset.REGISTRY)

	// Flags
	listCmd.Flags().StringVar(&o.filterFlag, "filter", "", "Filter based on the name of the component")
	listCmd.Flags().StringVar(&o.devfileFlag, "devfile", "", "Only the specific Devfile component")
	listCmd.Flags().StringVar(&o.devfileRegistryFlag, "devfile-registry", "", "Only show components from the specific Devfile registry")
	listCmd.Flags().BoolVar(&o.detailsFlag, "details", false, "Show details of each component")

	// Add a defined annotation in order to appear in the help menu
	listCmd.Annotations["command"] = "main"
	listCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	return listCmd
}

func (o *ListOptions) printDevfileList(DevfileList []registry.DevfileStack) {

	// Create the table and use our own style
	t := table.NewWriter()

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

	t.AppendHeader(table.Row{"NAME", "REGISTRY", "DESCRIPTION"})
	t.SortBy([]table.SortBy{
		{Name: "NAME", Mode: table.Asc},
		{Name: "REGISTRY", Mode: table.Asc},
	})

	for _, devfileComponent := range DevfileList {

		// Mark the name as yellow in the index to it's easier to see.
		name := text.Colors{text.FgHiYellow}.Sprint(devfileComponent.Name)

		// If the user has specified a filter with variable o.filterFlag, then only show the components
		// containing that specific string.
		if o.filterFlag != "" {
			if !strings.Contains(devfileComponent.Name, o.filterFlag) {
				continue
			}
		}

		// If the user passed in --devfile-registry <REGISTRY-NAME>, then only show the components from that Devfile stack
		if o.devfileRegistryFlag != "" {
			if !strings.Contains(devfileComponent.Registry.Name, o.devfileRegistryFlag) {
				continue
			}
		}

		// If the user passed in --devfile <NAME> only show that specific component matching that name
		if o.devfileFlag != "" {
			if devfileComponent.Name != o.devfileFlag {
				continue
			}
		}

		if o.detailsFlag {

			// Output the details of the component
			fmt.Printf(`%s: %s
%s: %s
%s: %s
%s: %s
%s: %s
%s: %s 
%s: %s
%s: %s
%s: %s
%s:
  - %s
%s`,
				log.Sbold("Name"), name,
				log.Sbold("Display Name"), devfileComponent.DisplayName,
				log.Sbold("Registry"), devfileComponent.Registry.Name,
				log.Sbold("Registry URL"), devfileComponent.Registry.URL,
				log.Sbold("Version"), devfileComponent.Version,
				log.Sbold("Description"), devfileComponent.Description,
				log.Sbold("Tags"), strings.Join(devfileComponent.Tags[:], ", "),
				log.Sbold("Project Type"), devfileComponent.ProjectType,
				log.Sbold("Language"), devfileComponent.Language,
				log.Sbold("Starter Projects"), strings.Join(devfileComponent.StarterProjects, "\n  - "),
				// TODO, showing dev / deploy / debug NOT yet implemented
				// log.Sbold("Supported odo Features"), "Y", "Y", "Y",
				"\n")
		} else {
			// Create a simplified row only showing the name, registry and description.
			t.AppendRow(table.Row{name, devfileComponent.Registry.Name, util.TruncateString(devfileComponent.Description, 40, "...")})
		}

	}

	if !o.detailsFlag {
		t.Render()
	}

}
