package search

import (
	"fmt"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/odo/cli/catalog/util"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	svc "github.com/openshift/odo/pkg/service"
	"github.com/spf13/cobra"
)

const serviceRecommendedCommandName = "service"

var serviceExample = `  # Search for a service
  %[1]s mysql`

// SearchServiceOptions encapsulates the options for the odo catalog describe service command
type SearchServiceOptions struct {
	searchTerm string
	services   []occlient.Service
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewSearchServiceOptions creates a new SearchServiceOptions instance
func NewSearchServiceOptions() *SearchServiceOptions {
	return &SearchServiceOptions{}
}

// Complete completes SearchServiceOptions after they've been created
func (o *SearchServiceOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	o.searchTerm = args[0]

	o.services, err = svc.Search(o.Client, o.searchTerm)
	if err != nil {
		return fmt.Errorf("unable to list services because Service Catalog is not enabled in your cluster: %v", err)
	}
	o.services = util.FilterHiddenServices(o.services)

	return err
}

// Validate validates the SearchServiceOptions based on completed values
func (o *SearchServiceOptions) Validate() (err error) {
	if len(o.services) == 0 {
		return fmt.Errorf("no service matched the query: %s", o.searchTerm)
	}

	return
}

// Run contains the logic for the command associated with SearchServiceOptions
func (o *SearchServiceOptions) Run() (err error) {
	util.DisplayServices(o.services)
	return
}

// NewCmdCatalogSearchService implements the odo catalog search service command
func NewCmdCatalogSearchService(name, fullName string) *cobra.Command {
	o := NewSearchServiceOptions()
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
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

}
