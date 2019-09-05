package list

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/odo/cli/catalog/util"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"
)

const componentsRecommendedCommandName = "components"

var componentsExample = `  # Get the supported components
  %[1]s`

// ListComponentsOptions encapsulates the options for the odo catalog list components command
type ListComponentsOptions struct {
	// list of known images
	catalogList []catalog.CatalogImage
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewListComponentsOptions creates a new ListComponentsOptions instance
func NewListComponentsOptions() *ListComponentsOptions {
	return &ListComponentsOptions{}
}

// Complete completes ListComponentsOptions after they've been created
func (o *ListComponentsOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	o.catalogList, err = catalog.List(o.Client)
	if err != nil {
		return err
	}
	o.catalogList = util.FilterHiddenComponents(o.catalogList)

	return
}

// Validate validates the ListComponentsOptions based on completed values
func (o *ListComponentsOptions) Validate() (err error) {
	if len(o.catalogList) == 0 {
		return fmt.Errorf("no deployable components found")
	}

	return err
}

// Run contains the logic for the command associated with ListComponentsOptions
func (o *ListComponentsOptions) Run() (err error) {
	w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
	var supCatalogList, unsupCatalogList []catalog.CatalogImage

	for _, image := range o.catalogList {
		supported, unsupported := catalog.SpliceSupportedTags(image)

		if len(supported) != 0 {
			image.NonHiddenTags = supported
			supCatalogList = append(supCatalogList, image)
		}
		if len(unsupported) != 0 {
			image.NonHiddenTags = unsupported
			unsupCatalogList = append(unsupCatalogList, image)
		}
	}

	if len(supCatalogList) != 0 {
		fmt.Fprintln(w, "Component types fully supported in odo (including debugging capabilities):")
		o.printCatalogList(w, supCatalogList)
		fmt.Fprintln(w)

	}

	if len(unsupCatalogList) != 0 {
		fmt.Fprintln(w, "Component types without full odo support:")
		o.printCatalogList(w, unsupCatalogList)
	}

	w.Flush()
	return
}

// NewCmdCatalogListComponents implements the odo catalog list components command
func NewCmdCatalogListComponents(name, fullName string) *cobra.Command {
	o := NewListComponentsOptions()

	return &cobra.Command{
		Use:     name,
		Short:   "List all components",
		Long:    "List all available component types from OpenShift's Ima",
		Example: fmt.Sprintf(componentsExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

}

func (o *ListComponentsOptions) printCatalogList(w io.Writer, catalogList []catalog.CatalogImage) {
	fmt.Fprintln(w, "NAME", "\t", "PROJECT", "\t", "TAGS")

	for _, component := range catalogList {
		componentName := component.Name
		if component.Namespace == o.Project {
			/*
				If current namespace is same as the current component namespace,
				Loop through every other component,
				If there exists a component with same name but in different namespaces,
				mark the one from current namespace with (*)
			*/
			for _, comp := range catalogList {
				if comp.Name == component.Name && component.Namespace != comp.Namespace {
					componentName = fmt.Sprintf("%s (*)", component.Name)
				}
			}
		}
		fmt.Fprintln(w, componentName, "\t", component.Namespace, "\t", strings.Join(component.NonHiddenTags, ","))
	}
}
