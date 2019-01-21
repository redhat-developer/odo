package search

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	"os"
)

const componentRecommendedCommandName = "component"

var componentExample = `  # Search for a component
  %[1]s python`

func NewCmdCatalogSearchComponent(name, fullName string) *cobra.Command {
	return &cobra.Command{
		Use:   name,
		Short: "Search component type in catalog",
		Long: `Search component type in catalog.

This searches for a partial match for the given search term in all the available
components.
`,
		Args:    cobra.ExactArgs(1),
		Example: fmt.Sprintf(componentExample, fullName),
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
}
