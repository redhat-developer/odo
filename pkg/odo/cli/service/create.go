package service

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	svc "github.com/redhat-developer/odo/pkg/service"
	"github.com/spf13/cobra"
	"strings"
)

const createRecommendedCommandName = "create"

type ServiceCreateOptions struct {
	parameters  []string
	plan        string
	serviceType string
	serviceName string
	*genericclioptions.Context
}

func NewServiceCreateOptions() *ServiceCreateOptions {
	return &ServiceCreateOptions{}
}
func (o *ServiceCreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	o.serviceType = args[0]
	// if only one arg is given, then it is considered as service name and service type both
	o.serviceName = o.serviceType
	// if two args are given, first is service type and second one is service name
	if len(args) == 2 {
		o.serviceName = args[1]
	}
	o.Context = genericclioptions.NewContextCreatingAppIfNeeded(cmd)
	return err
}
func (o *ServiceCreateOptions) Validate() (err error) {
	// make sure the service type exists
	matchingService, err := svc.GetSvcByType(o.Client, o.serviceType)
	if err != nil {
		return fmt.Errorf("unable to create service because Service Catalog is not enabled in your cluster:\n%v", err)
	}
	if matchingService == nil {
		return fmt.Errorf("service %v doesn't exist\nRun 'odo service catalog' to see a list of supported services.\n", o.serviceType)
	}
	if len(o.plan) == 0 {
		// when the plan has not been supplied, if there is only one available plan, we select it
		if len(matchingService.PlanList) == 1 {
			o.plan = matchingService.PlanList[0]
			glog.V(4).Infof("Plan %s was automatically selected since it's the only one available for service %s", o.plan, o.serviceType)
		} else {
			return fmt.Errorf("no plan was supplied for service %v.\nPlease select one of: %v\n", o.serviceType, strings.Join(matchingService.PlanList, ","))
		}
	} else {
		// when the plan has been supplied, we need to make sure it exists
		planFound := false
		for _, candidatePlan := range matchingService.PlanList {
			if o.plan == candidatePlan {
				planFound = true
				break
			}
		}
		if !planFound {
			return fmt.Errorf("plan %s is invalid for service %v.\nPlease select one of: %v\n", o.plan, o.serviceType, strings.Join(matchingService.PlanList, ","))
		}
	}
	//validate service name
	err = util.ValidateName(o.serviceName)
	if err != nil {
		return err
	}
	exists, err := svc.SvcExists(o.Client, o.serviceName, o.Application)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%s service already exists in the current application", o.serviceName)
	}
	return err
}
func (o *ServiceCreateOptions) Run() (err error) {
	err = svc.CreateService(o.Client, o.serviceName, o.serviceType, o.plan, o.parameters, o.Application)
	fmt.Printf("Service '%s' was created.\n", o.serviceName)
	return
}

// NewCmdServiceCreate implements the odo service create command.
func NewCmdServiceCreate(name string) *cobra.Command {
	o := NewServiceCreateOptions()
	serviceCreateCmd := &cobra.Command{
		Use:   name + " <service_type> --plan <plan_name> [service_name]",
		Short: "Create a new service",
		Long: `Create a new service from service catalog using the plan defined and deploy it on OpenShift.
If service name is not provided, service type value will be used. The plan to be used must be passed along the service type
using this convention <service_type>/<plan>. The parameters to configure the service are passed as a list of key=value pairs.
The list of the parameters and their type is defined according to the plan selected.
A full list of service types that can be deployed are available using: 'odo catalog list services'`,
		Example: `  # Create new postgresql service from service catalog using dev plan and name my-postgresql-db.
  odo service create dh-postgresql-apb my-postgresql-db --plan dev -p postgresql_user=luke -p postgresql_password=secret
	`,
		Args: cobra.RangeArgs(1, 2),
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckError(o.Complete(name, cmd, args), "")
			util.CheckError(o.Validate(), "")
			util.CheckError(o.Run(), "")
		},
	}
	serviceCreateCmd.Flags().StringVar(&o.plan, "plan", "", "The name of the plan of the service to be created")
	serviceCreateCmd.Flags().StringSliceVarP(&o.parameters, "parameters", "p", []string{}, "Parameters of the plan where a parameter is expressed as <key>=<value")
	addProjectFlag(serviceCreateCmd)
	genericclioptions.AddApplicationFlag(serviceCreateCmd)
	completion.RegisterCommandHandler(serviceCreateCmd, completion.ServiceClassCompletionHandler)
	return serviceCreateCmd
}
