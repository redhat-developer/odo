package catalog

import (
	"fmt"

	"github.com/openshift/odo/v2/pkg/odo/cli/catalog/describe"
	"github.com/openshift/odo/v2/pkg/odo/cli/catalog/list"
	"github.com/openshift/odo/v2/pkg/odo/cli/catalog/search"
	odoutil "github.com/openshift/odo/v2/pkg/odo/util"

	"github.com/spf13/cobra"
)

// RecommendedCommandName is the recommended catalog command name
const RecommendedCommandName = "catalog"

// NewCmdCatalog implements the odo catalog command
func NewCmdCatalog(name, fullName string) *cobra.Command {
	catalogDescribeCmd := describe.NewCmdCatalogDescribe(describe.RecommendedCommandName, odoutil.GetFullName(fullName, describe.RecommendedCommandName))
	catalogSearchCmd := search.NewCmdCatalogSearch(search.RecommendedCommandName, odoutil.GetFullName(fullName, search.RecommendedCommandName))
	catalogListCmd := list.NewCmdCatalogList(list.RecommendedCommandName, odoutil.GetFullName(fullName, list.RecommendedCommandName))

	catalogCmd := &cobra.Command{
		Use:   fmt.Sprintf("%s [options]", name),
		Short: "Catalog related operations",
		Long:  "Catalog related operations",
		Example: fmt.Sprintf("%s\n%s\n%s",
			catalogListCmd.Example,
			catalogSearchCmd.Example,
			catalogDescribeCmd.Example),
	}

	catalogCmd.AddCommand(catalogSearchCmd, catalogListCmd, catalogDescribeCmd)

	// Add a defined annotation in order to appear in the help menu
	catalogCmd.Annotations = map[string]string{"command": "main"}
	catalogCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	return catalogCmd
}
