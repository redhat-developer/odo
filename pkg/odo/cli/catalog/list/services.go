package list

import (
	"fmt"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	olm "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/odo/cli/catalog/util"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/spf13/cobra"
)

const servicesRecommendedCommandName = "services"

var servicesExample = `  # Get the supported services
  %[1]s`

// ServiceOptions encapsulates the options for the odo catalog list services command
type ServiceOptions struct {
	// Context
	*genericclioptions.Context

	// list of clusterserviceversions (installed by Operators)
	csvs *olm.ClusterServiceVersionList
}

// NewServiceOptions creates a new ListServicesOptions instance
func NewServiceOptions() *ServiceOptions {
	return &ServiceOptions{}
}

// Complete completes ListServicesOptions after they've been created
func (o *ServiceOptions) Complete(name string, cmdline cmdline.Cmdline, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline))
	if err != nil {
		return err
	}

	o.csvs, err = catalog.ListOperatorServices(o.KClient)
	if err != nil && !strings.Contains(err.Error(), "could not find specified operator") {
		return err
	}
	return nil
}

// Validate validates the ListServicesOptions based on completed values
func (o *ServiceOptions) Validate() error {
	return nil
}

// Run contains the logic for the command associated with ListServicesOptions
func (o *ServiceOptions) Run(cmd *cobra.Command) error {
	if log.IsJSON() {
		machineoutput.OutputSuccess(newCatalogListOutput(o.csvs))
	} else {
		if len(o.csvs.Items) == 0 {
			log.Info("no deployable operators found")
			return nil
		}

		if len(o.csvs.Items) > 0 {
			util.DisplayClusterServiceVersions(o.csvs)
		}
	}
	return nil
}

// NewCmdCatalogListServices implements the odo catalog list services command
func NewCmdCatalogListServices(name, fullName string) *cobra.Command {
	o := NewServiceOptions()
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
	// list of clusterserviceversions (installed by Operators)
	Operators *olm.ClusterServiceVersionList `json:"operators,omitempty"`
}

func newCatalogListOutput(operators *olm.ClusterServiceVersionList) catalogListOutput {
	return catalogListOutput{
		TypeMeta: v1.TypeMeta{
			Kind:       "List",
			APIVersion: machineoutput.APIVersion,
		},
		Operators: operators,
	}
}
