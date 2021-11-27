package search

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"

	"github.com/redhat-developer/odo/pkg/odo/cli/catalog/util"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"
)

const componentRecommendedCommandName = "component"

var componentExample = `  # Search for a component
  %[1]s python`

// SearchComponentOptions encapsulates the options for the odo catalog describe service command
type SearchComponentOptions struct {
	// Context
	*genericclioptions.Context

	// Parameters
	searchTerm string

	// components matching the search query
	components []string
}

// NewSearchComponentOptions creates a new SearchComponentOptions instance
func NewSearchComponentOptions() *SearchComponentOptions {
	return &SearchComponentOptions{}
}

// Complete completes SearchComponentOptions after they've been created
func (o *SearchComponentOptions) Complete(name string, cmdline cmdline.Cmdline, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline))
	if err != nil {
		return err
	}
	o.searchTerm = args[0]

	o.components, err = catalog.SearchComponent(o.KClient, o.searchTerm)
	return err
}

// Validate validates the SearchComponentOptions based on completed values
func (o *SearchComponentOptions) Validate() error {
	if len(o.components) == 0 {
		return fmt.Errorf("no component matched the query: %s", o.searchTerm)
	}

	return nil
}

// Run contains the logic for the command associated with SearchComponentOptions
func (o *SearchComponentOptions) Run() error {
	util.DisplayComponents(o.components)
	return nil
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
