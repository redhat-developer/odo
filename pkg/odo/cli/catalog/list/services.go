package list

import (
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/odo/pkg/catalog"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/machineoutput"
	"github.com/openshift/odo/pkg/odo/cli/catalog/util"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	olm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/spf13/cobra"
)

const servicesRecommendedCommandName = "services"

var servicesExample = `  # Get the supported services from service catalog
  %[1]s`

// ListServicesOptions encapsulates the options for the odo catalog list services command
type ListServicesOptions struct {
	// list of known services
	services catalog.ServiceTypeList
	// list of clusterserviceversions (installed by Operators)
	csvs *olm.ClusterServiceVersionList
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewListServicesOptions creates a new ListServicesOptions instance
func NewListServicesOptions() *ListServicesOptions {
	return &ListServicesOptions{}
}

// Complete completes ListServicesOptions after they've been created
func (o *ListServicesOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	var noCsvs, noServices bool
	o.Context = genericclioptions.NewContext(cmd)
	o.csvs, err = o.KClient.GetClusterServiceVersionList()
	if err != nil {
		// Error only occurs when OperatorHub is not installed/enabled on the
		// Kubernetes or OpenShift 4.x cluster. It doesn't occur when there are
		// no operators installed.
		noCsvs = true
	}

	o.services, err = catalog.ListServices(o.Client)
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

	return
}

// Validate validates the ListServicesOptions based on completed values
func (o *ListServicesOptions) Validate() (err error) {
	if len(o.services.Items) == 0 && len(o.csvs.Items) == 0 {
		return fmt.Errorf("no deployable services/operators found")
	}
	return
}

// Run contains the logic for the command associated with ListServicesOptions
func (o *ListServicesOptions) Run() (err error) {
	if log.IsJSON() {
		machineoutput.OutputSuccess(newCatalogListOutput(&o.services, o.csvs))
	} else {
		if len(o.csvs.Items) > 0 {
			util.DisplayClusterServiceVersions(o.csvs)
		}
		if len(o.services.Items) > 0 {
			util.DisplayServices(o.services)
		}
	}
	return
}

// NewCmdCatalogListServices implements the odo catalog list services command
func NewCmdCatalogListServices(name, fullName string) *cobra.Command {
	o := NewListServicesOptions()
	return &cobra.Command{
		Use:         name,
		Short:       "Lists all available services",
		Long:        "Lists all available services",
		Example:     fmt.Sprintf(servicesExample, fullName),
		Args:        cobra.ExactArgs(0),
		Annotations: map[string]string{"machineoutput": "json"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
}

type catalogListOutput struct {
	v1.TypeMeta   `json:",inline"`
	v1.ObjectMeta `json:"metadata,omitempty"`
	Services      *catalog.ServiceTypeList `json:"services,omitempty"`
	// list of clusterserviceversions (installed by Operators)
	Operators *olm.ClusterServiceVersionList `json:"operators,omitempty"`
}

func newCatalogListOutput(services *catalog.ServiceTypeList, operators *olm.ClusterServiceVersionList) catalogListOutput {
	return catalogListOutput{
		TypeMeta: v1.TypeMeta{
			Kind:       "List",
			APIVersion: machineoutput.APIVersion,
		},
		Services:  services,
		Operators: operators,
	}
}
