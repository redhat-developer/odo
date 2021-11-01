package service

import (
	"fmt"

	"github.com/openshift/odo/pkg/odo/genericclioptions"
	svc "github.com/openshift/odo/pkg/service"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const listRecommendedCommandName = "list"

var (
	listExample = ktemplates.Examples(`
    # List all services in the application
    %[1]s`)
	listLongDesc = ktemplates.LongDesc(`
List all services in the current application
`)
)

// ServiceListOptions encapsulates the options for the odo service list command
type ServiceListOptions struct {
	*genericclioptions.Context
	// Context to use when listing service. This will use app and project values from the context
	componentContext string
	// If true, Operator Hub is installed on the cluster
	csvSupport bool
}

// NewServiceListOptions creates a new ServiceListOptions instance
func NewServiceListOptions() *ServiceListOptions {
	return &ServiceListOptions{}
}

// Complete completes ServiceListOptions after they've been created
func (o *ServiceListOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.csvSupport, err = svc.IsCSVSupported()
	if err != nil {
		return err
	}
	if !o.csvSupport {
		return fmt.Errorf("failed to list Operator backed services, make sure you have installed the Operators on the cluster")
	}
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmd).NeedDevfile().SetComponentContext(o.componentContext))
	return err
}

// Validate validates the ServiceListOptions based on completed values
func (o *ServiceListOptions) Validate() (err error) {
	return
}

// Run contains the logic for the odo service list command
func (o *ServiceListOptions) Run(cmd *cobra.Command) (err error) {
	return o.listOperatorServices()
}

// NewCmdServiceList implements the odo service list command.
func NewCmdServiceList(name, fullName string) *cobra.Command {
	o := NewServiceListOptions()
	serviceListCmd := &cobra.Command{
		Use:         name,
		Short:       "List all services in the current application",
		Long:        listLongDesc,
		Example:     fmt.Sprintf(listExample, fullName),
		Args:        cobra.NoArgs,
		Annotations: map[string]string{"machineoutput": "json"},
		Run: func(cmd *cobra.Command, args []string) {
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	genericclioptions.AddContextFlag(serviceListCmd, &o.componentContext)
	return serviceListCmd
}
