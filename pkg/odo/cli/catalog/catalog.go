package catalog

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/redhat-developer/odo/pkg/occlient"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"

	"github.com/spf13/cobra"
)

var catalogCmd = &cobra.Command{
	Use:   "catalog [options]",
	Short: "Catalog related operations",
	Long:  "Catalog related operations",
	Example: fmt.Sprintf("%s\n%s\n%s",
		catalogListCmd.Example,
		catalogSearchCmd.Example,
		catalogDescribeCmd.Example),
}

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
func NewCmdCatalog() *cobra.Command {
	catalogCmd.AddCommand(catalogSearchCmd)
	catalogCmd.AddCommand(catalogListCmd)
	catalogCmd.AddCommand(catalogDescribeCmd)
	catalogListCmd.AddCommand(catalogListComponentCmd)
	catalogListCmd.AddCommand(catalogListServiceCmd)
	catalogSearchCmd.AddCommand(catalogSearchComponentCmd)
	catalogSearchCmd.AddCommand(catalogSearchServiceCmd)
	catalogDescribeCmd.AddCommand(catalogDescribeServiceCmd)
	// Add a defined annotation in order to appear in the help menu
	catalogCmd.Annotations = map[string]string{"command": "other"}
	catalogCmd.SetUsageTemplate(odoutil.CmdUsageTemplate)

	return catalogCmd
}
