package catalog

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/odo/cli/catalog/describe"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/redhat-developer/odo/pkg/occlient"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"

	"github.com/spf13/cobra"
)

// RecommendedCatalogCommandName is the recommended catalog command name
const RecommendedCatalogCommandName = "catalog"

func displayServices(services []occlient.Service) {
	w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "NAME", "\t", "PLANS")
	for _, service := range services {
		fmt.Fprintln(w, service.Name, "\t", strings.Join(service.PlanList, ","))
	}
	w.Flush()
}

func filterHiddenServices(services []occlient.Service) []occlient.Service {
	var filteredServices []occlient.Service
	for _, service := range services {
		if !service.Hidden {
			filteredServices = append(filteredServices, service)
		}
	}
	return filteredServices
}

// NewCmdCatalog implements the odo catalog command
func NewCmdCatalog(name, fullName string) *cobra.Command {
	catalogDescribeCmd := describe.NewCmdCatalogDescribe(describe.RecommendedCommandName, odoutil.GetFullName(fullName, describe.RecommendedCommandName))
	catalogCmd := &cobra.Command{
		Use:   fmt.Sprintf("%s [options]", name),
		Short: "Catalog related operations",
		Long:  "Catalog related operations",
		Example: fmt.Sprintf("%s\n%s\n%s",
			catalogListCmd.Example,
			catalogSearchCmd.Example,
			catalogDescribeCmd.Example),
	}

	catalogCmd.AddCommand(catalogSearchCmd)
	catalogCmd.AddCommand(catalogListCmd)
	catalogCmd.AddCommand(catalogDescribeCmd)
	catalogListCmd.AddCommand(catalogListComponentCmd)
	catalogListCmd.AddCommand(catalogListServiceCmd)
	catalogSearchCmd.AddCommand(catalogSearchComponentCmd)
	catalogSearchCmd.AddCommand(catalogSearchServiceCmd)

	// Add a defined annotation in order to appear in the help menu
	catalogCmd.Annotations = map[string]string{"command": "other"}
	catalogCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	return catalogCmd
}
