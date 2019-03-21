package list

import (
	"fmt"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/cli/catalog/util"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	svc "github.com/openshift/odo/pkg/service"
	"github.com/spf13/cobra"
)

const servicesRecommendedCommandName = "services"

var servicesExample = `  # Get the supported services from service catalog
  %[1]s`

// ListServicesOptions encapsulates the options for the odo catalog list services command
type ListServicesOptions struct {
	// list of known services
	services []occlient.Service
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewListServicesOptions creates a new ListServicesOptions instance
func NewListServicesOptions() *ListServicesOptions {
	return &ListServicesOptions{}
}

// Complete completes ListServicesOptions after they've been created
func (o *ListServicesOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	o.services, err = svc.ListCatalog(o.Client)
	if err != nil {
		return fmt.Errorf("unable to list services because Service Catalog is not enabled in your cluster: %v", err)
	}
	o.services = util.FilterHiddenServices(o.services)

	return
}

// Validate validates the ListServicesOptions based on completed values
func (o *ListServicesOptions) Validate() (err error) {
	if len(o.services) == 0 {
		return fmt.Errorf("no deployable services found")
	}
	return
}

// Run contains the logic for the command associated with ListServicesOptions
func (o *ListServicesOptions) Run() (err error) {
	util.DisplayServices(o.services)
	return
}

// NewCmdCatalogListServices implements the odo catalog list services command
func NewCmdCatalogListServices(name, fullName string) *cobra.Command {
	o := NewListServicesOptions()
	return &cobra.Command{
		Use:     name,
		Short:   "Lists all available services",
		Long:    "Lists all available services",
		Example: fmt.Sprintf(servicesExample, fullName),
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
}
