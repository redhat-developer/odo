package service

import (
	"fmt"
	"strings"

	"github.com/redhat-developer/odo/pkg/service"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/ui"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
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
	// Context
	*genericclioptions.Context

	// Parameters
	serviceName string

	// Flags
	forceFlag   bool
	contextFlag string

	// Backend is the service provider backend that was used to create the service
	Backend ServiceProviderBackend
}

// NewDeleteOptions creates a new DeleteOptions instance
func NewDeleteOptions() *DeleteOptions {
	return &DeleteOptions{}
}

// Complete completes DeleteOptions after they've been created
func (o *DeleteOptions) Complete(name string, cmdline cmdline.Cmdline, args []string) (err error) {
	o.Context, err = genericclioptions.New(genericclioptions.NewCreateParameters(cmdline).NeedDevfile(o.contextFlag))
	if err != nil {
		return err
	}

	err = validDevfileDirectory(o.contextFlag)
	if err != nil {
		return err
	}

	o.serviceName = args[0]
	_, _, err = service.SplitServiceKindName(o.serviceName)
	if err != nil {
		return fmt.Errorf("invalid service name")
	}
	o.Backend = NewOperatorBackend()

	return nil
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
	return nil
}

// Run contains the logic for the odo service delete command
func (o *DeleteOptions) Run() (err error) {
	if o.forceFlag || ui.Proceed(fmt.Sprintf("Are you sure you want to delete %v", o.serviceName)) {
		err = o.Backend.DeleteService(o, o.serviceName, o.GetApplication())
		if err != nil {
			return err
		}
		log.Infof("Service %q has been successfully deleted; do 'odo push' to delete service from the cluster", o.serviceName)
	} else {
		log.Errorf("Aborting deletion of service: %v", o.serviceName)
	}
	return nil
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
	serviceDeleteCmd.Flags().BoolVarP(&o.forceFlag, "force", "f", false, "Delete service without prompting")
	genericclioptions.AddContextFlag(serviceDeleteCmd, &o.contextFlag)
	return serviceDeleteCmd
}
