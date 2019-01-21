package search

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

const RecommendedCommandName = "search"

func NewCmdCatalogSearch(name, fullName string) *cobra.Command {
	component := NewCmdCatalogSearchComponent(componentRecommendedCommandName, util.GetFullName(fullName, componentRecommendedCommandName))
	service := NewCmdCatalogSearchService(serviceRecommendedCommandName, util.GetFullName(fullName, serviceRecommendedCommandName))
	catalogSearchCmd := &cobra.Command{
		Use:   name,
		Short: "Search available component & service types.",
		Long: `Search available component & service types..

This searches for a partial match for the given search term in all the available
components & services.
`,
		Example: fmt.Sprintf("%s\n%s\n", component.Example, service.Example),
	}
	catalogSearchCmd.AddCommand(component, service)

	return catalogSearchCmd
}
