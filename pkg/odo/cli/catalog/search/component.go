package search

import (
	"fmt"

	"github.com/openshift/odo/pkg/catalog"

	"github.com/openshift/odo/pkg/odo/cli/catalog/util"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"
)

const componentRecommendedCommandName = "component"

var componentExample = `  # Search for a component
  %[1]s python`

// SearchComponentOptions encapsulates the options for the odo catalog describe service command
type SearchComponentOptions struct {
	searchTerm string
	components []string
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewSearchComponentOptions creates a new SearchComponentOptions instance
func NewSearchComponentOptions() *SearchComponentOptions {
	return &SearchComponentOptions{}
}

// Complete completes SearchComponentOptions after they've been created
func (o *SearchComponentOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmd))
	if err != nil {
		return err
	}
	o.searchTerm = args[0]

	o.components, err = catalog.SearchComponent(o.Client, o.searchTerm)
	return err
}

// Validate validates the SearchComponentOptions based on completed values
func (o *SearchComponentOptions) Validate() (err error) {
	if len(o.components) == 0 {
		return fmt.Errorf("no component matched the query: %s", o.searchTerm)
	}

	return
}

// Run contains the logic for the command associated with SearchComponentOptions
func (o *SearchComponentOptions) Run(cmd *cobra.Command) (err error) {
	util.DisplayComponents(o.components)
	return
}

// NewCmdCatalogSearchComponent implements the odo catalog search component command
func NewCmdCatalogSearchComponent(name, fullName string) *cobra.Command {
	o := NewSearchComponentOptions()
	return &cobra.Command{
		Use:   name,
		Short: "Search component type in catalog",
		Long: `Search component type in catalog.

This searches for a partial match for the given search term in all the available
components.
`,
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf(componentExample, fullName),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
}
