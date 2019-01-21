package search

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/catalog/util"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	svc "github.com/redhat-developer/odo/pkg/service"
	"github.com/spf13/cobra"
	"os"
)

const serviceRecommendedCommandName = "service"

var serviceExample = `  # Search for a service
  %[1]s mysql`

func NewCmdCatalogSearchService(name, fullName string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: "Search service type in catalog",
		Long: `Search service type in catalog.

This searches for a partial match for the given search term in all the available
services from service catalog.
`,
		Example: fmt.Sprintf(serviceExample, fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := genericclioptions.Client(cmd)
			searchTerm := args[0]
			services, err := svc.Search(client, searchTerm)
			odoutil.LogErrorAndExit(err, "unable to search for services")
			services = util.FilterHiddenServices(services)

			switch len(services) {
			case 0:
				log.Errorf("No service matched the query: %v", searchTerm)
				os.Exit(1)
			default:
				util.DisplayServices(services)
			}
		},
	}

}
