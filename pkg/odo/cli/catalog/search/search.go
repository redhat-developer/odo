package search

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "search"

// NewCmdCatalogSearch implements the odo catalog search command
func NewCmdCatalogSearch(name, fullName string) *cobra.Command {
	component := NewCmdCatalogSearchComponent(componentRecommendedCommandName, util.GetFullName(fullName, componentRecommendedCommandName))
	catalogSearchCmd := &cobra.Command{
		Use:   name,
		Short: "Search available component & service types.",
		Long: `Search available component & service types..

This searches for a partial match for the given search term in all the available
components & services.
`,
		Example: fmt.Sprintf("%s\n", component.Example),
	}
	catalogSearchCmd.AddCommand(component)

	return catalogSearchCmd
}
