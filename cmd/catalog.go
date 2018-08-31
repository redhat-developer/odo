package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/redhat-developer/odo/pkg/catalog"
	svc "github.com/redhat-developer/odo/pkg/service"
	"github.com/spf13/cobra"
)

var catalogCmd = &cobra.Command{
	Use:   "catalog [options]",
	Short: "Catalog related operations",
	Long:  "Catalog related operations",
	Example: fmt.Sprintf("%s\n%s",
		catalogListCmd.Example,
		catalogSearchCmd.Example),
}

var catalogListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available component & service types.",
	Long:  "List all available component and service types from OpenShift",
	Example: `  # Get the supported components
  odo catalog list components

  # Get the supported services from service catalog
  odo catalog list services
`,
}

var catalogListComponentCmd = &cobra.Command{
	Use:   "components",
	Short: "List all available component types.",
	Long:  "List all available component types from OpenShift's Image Builder.",
	Example: `  # Get the supported components
  odo catalog list components

  # Search for a supported component
  odo catalog search components nodejs
`,
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		catalogList, err := catalog.List(client)
		checkError(err, "unable to list components")
		switch len(catalogList) {
		case 0:
			fmt.Printf("No deployable components found\n")
		default:
			currentProject := client.GetCurrentProjectName()
			w := tabwriter.NewWriter(os.Stdout, 5, 2, 3, ' ', tabwriter.TabIndent)
			fmt.Fprintln(w, "NAME", "\t", "PROJECT", "\t", "TAGS")
			for _, component := range catalogList {
				componentName := component.Name
				if component.Namespace == currentProject {
					/*
						If current namespace is same as the current component namespace,
						Loop through every other component,
						If there exists a component with same name but in different namespaces,
						mark the one from current namespace with (*)
					*/
					for _, comp := range catalogList {
						if comp.Name == component.Name && component.Namespace != comp.Namespace {
							componentName = fmt.Sprintf("%s (*)", component.Name)
						}
					}
				}
				fmt.Fprintln(w, componentName, "\t", component.Namespace, "\t", strings.Join(component.Tags, ","))
			}
			w.Flush()
		}
	},
}

var catalogListServiceCmd = &cobra.Command{
	Use:   "services",
	Short: "Lists all the services from service catalog",
	Long:  "Lists all the services from service catalog",
	Example: `  # List all services
  odo catalog list services

 # Search for a supported services
  odo catalog search services mysql
	`,
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		catalogList, err := svc.ListCatalog(client)
		checkError(err, "unable to list services because Service Catalog is not enabled in your cluster")
		switch len(catalogList) {
		case 0:
			fmt.Printf("No deployable services found\n")
		default:
			fmt.Println("The following services can be deployed:")
			for _, service := range catalogList {
				fmt.Printf("- %v\n", service)
			}
		}
	},
}

var catalogSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search available component & service types.",
	Long: `Search available component & service types..

This searches for a partial match for the given search term in all the available
components & services.
`,
	Example: `  # Search for a component
  odo catalog search components python

  # Search for a service
  odo catalog search service mysql
	`,
}

var catalogSearchComponentCmd = &cobra.Command{
	Use:   "components",
	Short: "Search component type in catalog",
	Long: `Search component type in catalog.

This searches for a partial match for the given search term in all the available
components.
`,
	Args: cobra.ExactArgs(1),
	Example: `  # Search for a component
  odo catalog search components python
	`,
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		searchTerm := args[0]
		components, err := catalog.Search(client, searchTerm)
		checkError(err, "unable to search for components")

		switch len(components) {
		case 0:
			fmt.Printf("No component matched the query: %v\n", searchTerm)
		default:
			fmt.Println("The following components were found:")
			for _, component := range components {
				fmt.Printf("- %v\n", component)
			}
		}
	},
}

var catalogSearchServiceCmd = &cobra.Command{
	Use:   "services",
	Short: "Search service type in catalog",
	Long: `Search service type in catalog.

This searches for a partial match for the given search term in all the available
services from service catalog.
`,
	Example: `  # Search for a service
  odo catalog search services mysql
	`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := getOcClient()
		searchTerm := args[0]
		components, err := svc.Search(client, searchTerm)
		checkError(err, "unable to search for services")

		switch len(components) {
		case 0:
			fmt.Printf("No service matched the query: %v\n", searchTerm)
		default:
			fmt.Println("The following services were found:")
			for _, component := range components {
				fmt.Printf("- %v\n", component)
			}
		}
	},
}

func init() {
	catalogCmd.AddCommand(catalogSearchCmd)
	catalogCmd.AddCommand(catalogListCmd)
	catalogListCmd.AddCommand(catalogListComponentCmd)
	catalogListCmd.AddCommand(catalogListServiceCmd)
	catalogSearchCmd.AddCommand(catalogSearchComponentCmd)
	catalogSearchCmd.AddCommand(catalogSearchServiceCmd)
	// Add a defined annotation in order to appear in the help menu
	catalogCmd.Annotations = map[string]string{"command": "other"}
	catalogCmd.SetUsageTemplate(cmdUsageTemplate)

	rootCmd.AddCommand(catalogCmd)
}
