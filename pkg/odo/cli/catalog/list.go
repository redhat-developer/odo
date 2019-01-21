package catalog

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/catalog/util"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	svc "github.com/redhat-developer/odo/pkg/service"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"text/tabwriter"
)

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
	Short: "List all components available.",
	Long:  "List all available component types from OpenShift's Image Builder.",
	Example: `  # Get the supported components
  odo catalog list components

  # Search for a supported component
  odo catalog search component nodejs
`,
	Run: func(cmd *cobra.Command, args []string) {
		client := genericclioptions.Client(cmd)
		catalogList, err := catalog.List(client)
		odoutil.LogErrorAndExit(err, "unable to list components")
		switch len(catalogList) {
		case 0:
			log.Errorf("No deployable components found")
			os.Exit(1)
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
	Short: "Lists all available services",
	Long:  "Lists all available services",
	Example: `  # List all services
  odo catalog list services

 # Search for a supported service
  odo catalog search service mysql
	`,
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		client := genericclioptions.Client(cmd)
		catalogList, err := svc.ListCatalog(client)
		odoutil.LogErrorAndExit(err, "unable to list services because Service Catalog is not enabled in your cluster")
		catalogList = util.FilterHiddenServices(catalogList)
		switch len(catalogList) {
		case 0:
			log.Errorf("No deployable services found")
			os.Exit(1)
		default:
			util.DisplayServices(catalogList)

		}
	},
}
