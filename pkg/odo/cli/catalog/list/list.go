package list

import (
	"fmt"

	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended command name
const RecommendedCommandName = "list"

// NewCmdCatalogList implements the odo catalog list command
func NewCmdCatalogList(name, fullName string) *cobra.Command {
	components := NewCmdCatalogListComponents(componentsRecommendedCommandName, util.GetFullName(fullName, componentsRecommendedCommandName))

	catalogListCmd := &cobra.Command{
		Use:     name,
		Short:   "List all available component types.",
		Long:    "List all available component types from OpenShift",
		Example: fmt.Sprintf("%s\n", components.Example),
	}

	catalogListCmd.AddCommand(
		components,
	)

	return catalogListCmd
}
