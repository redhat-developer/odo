package service

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/odo/cli/ui"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
	"github.com/openshift/odo/pkg/odo/util/completion"
	svc "github.com/openshift/odo/pkg/service"
	"github.com/spf13/cobra"
	"k8s.io/klog"
	ktemplates "k8s.io/kubectl/pkg/util/templates"
)

const deleteRecommendedCommandName = "delete"

var (
	deleteExample = ktemplates.Examples(`
    # Delete the service named 'mysql-persistent'
    %[1]s mysql-persistent`)

	deleteLongDesc = ktemplates.LongDesc(`
	Delete an existing service`)
)

// ServiceDeleteOptions encapsulates the options for the odo service delete command
type ServiceDeleteOptions struct {
	serviceForceDeleteFlag bool
	serviceName            string
	*genericclioptions.Context
	// Context to use when listing service. This will use app and project values from the context
	componentContext string
}

// NewServiceDeleteOptions creates a new ServiceDeleteOptions instance
func NewServiceDeleteOptions() *ServiceDeleteOptions {
	return &ServiceDeleteOptions{}
}

// Complete completes ServiceDeleteOptions after they've been created
func (o *ServiceDeleteOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context = genericclioptions.NewContext(cmd)
	o.serviceName = args[0]

	return
}

// Validate validates the ServiceDeleteOptions based on completed values
func (o *ServiceDeleteOptions) Validate() (err error) {
	exists, err := svc.SvcExists(o.Client, o.serviceName, o.Application)
	if err != nil {
		return fmt.Errorf("unable to delete service because Service Catalog is not enabled in your cluster:\n%v", err)
	}
	if !exists {
		return fmt.Errorf("Service with the name %s does not exist in the current application\n", o.serviceName)
	}
	return
}

// Run contains the logic for the odo service delete command
func (o *ServiceDeleteOptions) Run() (err error) {
	if o.serviceForceDeleteFlag || ui.Proceed(fmt.Sprintf("Are you sure you want to delete %v from %v", o.serviceName, o.Application)) {
		err = svc.DeleteServiceAndUnlinkComponents(o.Client, o.serviceName, o.Application)
		if err != nil {
			return fmt.Errorf("unable to delete service %s:\n%v", o.serviceName, err)
		}
		log.Infof("Service %s from application %s has been deleted", o.serviceName, o.Application)
	} else {
		log.Errorf("Aborting deletion of service: %v", o.serviceName)
	}
	return
}

// NewCmdServiceDelete implements the odo service delete command.
func NewCmdServiceDelete(name, fullName string) *cobra.Command {
	o := NewServiceDeleteOptions()
	serviceDeleteCmd := &cobra.Command{
		Use:     name + " <service_name>",
		Short:   "Delete an existing service",
		Long:    deleteLongDesc,
		Example: fmt.Sprintf(deleteExample, fullName),
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			klog.V(4).Infof("service delete called\n args: %#v", strings.Join(args, " "))
			genericclioptions.GenericRun(o, cmd, args)
		},
	}
	serviceDeleteCmd.Flags().BoolVarP(&o.serviceForceDeleteFlag, "force", "f", false, "Delete service without prompting")
	genericclioptions.AddContextFlag(serviceDeleteCmd, &o.componentContext)
	completion.RegisterCommandHandler(serviceDeleteCmd, completion.ServiceCompletionHandler)
	return serviceDeleteCmd
}
