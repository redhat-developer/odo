package list

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

const servicesRecommendedCommandName = "services"

var servicesExample = `  # Get the supported services from service catalog
  %[1]s`

func NewCmdCatalogListServices(name, fullName string) *cobra.Command {
	return &cobra.Command{
		Use:     name,
		Short:   "Lists all available services",
		Long:    "Lists all available services",
		Example: fmt.Sprintf(servicesExample, fullName),
		Args:    cobra.ExactArgs(0),
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

}
