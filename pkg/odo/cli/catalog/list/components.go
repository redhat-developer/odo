package list

import (
	"fmt"
	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	odoutil "github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"text/tabwriter"
)

const componentsRecommendedCommandName = "components"

var componentsExample = `  # Get the supported components
  %[1]s`

func NewCmdCatalogListComponents(name, fullName string) *cobra.Command {
	return &cobra.Command{
		Use:     name,
		Short:   "List all components available.",
		Long:    "List all available component types from OpenShift's Image Builder.",
		Example: fmt.Sprintf(componentsExample, fullName),
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

}
