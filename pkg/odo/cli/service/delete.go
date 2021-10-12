package service

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/service"

	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
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

// DeleteOptions encapsulates the options for the odo service delete command
type DeleteOptions struct {
	serviceForceDeleteFlag bool
	serviceName            string
	*genericclioptions.Context
	// Context to use when listing service. This will use app and project values from the context
	componentContext string
	// Backend is the service provider backend that was used to create the service
	Backend ServiceProviderBackend
}

// NewDeleteOptions creates a new DeleteOptions instance
func NewDeleteOptions() *DeleteOptions {
	return &DeleteOptions{}
}

// Complete completes DeleteOptions after they've been created
func (o *DeleteOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.CreateParameters{
		Cmd:              cmd,
		DevfilePath:      devfile.DevfileFilenamesProvider(o.componentContext),
		ComponentContext: o.componentContext,
	})
	if err != nil {
		return err
	}

	err = validDevfileDirectory(o.componentContext)
	if err != nil {
		return err
	}

	o.serviceName = args[0]
	_, _, err = service.SplitServiceKindName(o.serviceName)
	if err != nil {
		return fmt.Errorf("invalid service name")
	}
	o.Backend = NewOperatorBackend()

	return
}

// Validate validates the DeleteOptions based on completed values
func (o *DeleteOptions) Validate() (err error) {
	svcDefined, err := o.Backend.ServiceDefined(o.Context, o.serviceName)
	if err != nil {
		return err
	}

	if !svcDefined {
		return fmt.Errorf("couldn't find service named %q. Refer %q to see list of defined services", o.serviceName, "odo service list")
	}
	return
}

// Run contains the logic for the odo service delete command
func (o *DeleteOptions) Run(cmd *cobra.Command) (err error) {
	if o.serviceForceDeleteFlag || ui.Proceed(fmt.Sprintf("Are you sure you want to delete %v", o.serviceName)) {
		err = o.Backend.DeleteService(o, o.serviceName, o.Application)
		if err != nil {
			return err
		}
		log.Infof("Service %q has been successfully deleted; do 'odo push' to delete service from the cluster", o.serviceName)
	} else {
		log.Errorf("Aborting deletion of service: %v", o.serviceName)
	}
	return
}

// NewCmdServiceDelete implements the odo service delete command.
func NewCmdServiceDelete(name, fullName string) *cobra.Command {
	o := NewDeleteOptions()
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
	return serviceDeleteCmd
}
