package catalog

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	svc "github.com/redhat-developer/odo/pkg/service"
	"github.com/spf13/cobra"
	"os"
)

var catalogSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search available component & service types.",
	Long: `Search available component & service types..

This searches for a partial match for the given search term in all the available
components & services.
`,
	Example: `  # Search for a component
  odo catalog search component python

  # Search for a service
  odo catalog search service mysql
	`,
}

var catalogSearchComponentCmd = &cobra.Command{
	Use:   "component",
	Short: "Search component type in catalog",
	Long: `Search component type in catalog.

This searches for a partial match for the given search term in all the available
components.
`,
	Args: cobra.ExactArgs(1),
	Example: `  # Search for a component
  odo catalog search component python
	`,
	Run: func(cmd *cobra.Command, args []string) {
		client := genericclioptions.Client(cmd)
		searchTerm := args[0]
		components, err := catalog.Search(client, searchTerm)
		odoutil.LogErrorAndExit(err, "unable to search for components")

		switch len(components) {
		case 0:
			log.Errorf("No component matched the query: %v", searchTerm)
			os.Exit(1)
		default:
			log.Infof("The following components were found:")
			for _, component := range components {
				fmt.Printf("- %v\n", component)
			}
		}
	},
}

var catalogSearchServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Search service type in catalog",
	Long: `Search service type in catalog.

This searches for a partial match for the given search term in all the available
services from service catalog.
`,
	Example: `  # Search for a service
  odo catalog search service mysql
	`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := genericclioptions.Client(cmd)
		searchTerm := args[0]
		services, err := svc.Search(client, searchTerm)
		odoutil.LogErrorAndExit(err, "unable to search for services")
		services = filterHiddenServices(services)

		switch len(services) {
		case 0:
			log.Errorf("No service matched the query: %v", searchTerm)
			os.Exit(1)
		default:
			displayServices(services)
		}
	},
}
