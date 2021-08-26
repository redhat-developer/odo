package search

import (
	"fmt"

	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/odo/cli/catalog/util"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/spf13/cobra"
)

const serviceRecommendedCommandName = "service"

var serviceExample = `  # Search for a service
  %[1]s mysql`

// SearchServiceOptions encapsulates the options for the odo catalog describe service command
type SearchServiceOptions struct {
	searchTerm string
	services   catalog.ServiceTypeList
	// generic context options common to all commands
	csvs *olm.ClusterServiceVersionList
	*genericclioptions.Context
}

// NewSearchServiceOptions creates a new SearchServiceOptions instance
func NewSearchServiceOptions() *SearchServiceOptions {
	return &SearchServiceOptions{}
}

// Complete completes SearchServiceOptions after they've been created
func (o *SearchServiceOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	var noCsvs, noServices bool
	o.Context, err = genericclioptions.NewContext(cmd)
	if err != nil {
		return err
	}
	o.searchTerm = args[0]
	o.csvs, err = o.KClient.SearchClusterServiceVersionList(o.searchTerm)
	if err != nil {
		// Error only occurs when OperatorHub is not installed/enabled on the
		// Kubernetes or OpenShift 4.x cluster. It doesn't occur when there are
		// no operators installed.
		noCsvs = true
	}

	// Checks service catalog, but if its not available, we do not error.
	o.services, err = catalog.SearchService(o.Client, o.searchTerm)
	if err != nil {
		// Error occurs if Service Catalog is not enabled on the OpenShift
		// 3.x/4.x cluster
		noServices = true
		// But we don't care about the Service Catalog not being enabled if
		// it's 4.x or k8s cluster
		if !noCsvs {
			err = nil
		}
	}

	if noCsvs && noServices {
		// Neither OperatorHub nor Service Catalog is enabled on the cluster
		return fmt.Errorf("unable to list services because neither Service Catalog nor Operator Hub is enabled in your cluster: %v", err)
	}
	o.services = util.FilterHiddenServices(o.services)

	return err
}

// Validate validates the SearchServiceOptions based on completed values
func (o *SearchServiceOptions) Validate() (err error) {
	if len(o.services.Items) == 0 && len(o.csvs.Items) == 0 {
		return fmt.Errorf("no service matched the query: %s", o.searchTerm)
	}
	return
}

// Run contains the logic for the command associated with SearchServiceOptions
func (o *SearchServiceOptions) Run(cmd *cobra.Command) (err error) {
	if len(o.csvs.Items) > 0 {
		util.DisplayClusterServiceVersions(o.csvs)
	}
	if len(o.services.Items) > 0 {
		util.DisplayServices(o.services)
	}

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
services from operator hub services.
`,
		Example: fmt.Sprintf(serviceExample, fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}

}
