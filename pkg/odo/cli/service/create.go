package service

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/odo/cli/service/ui"
	"github.com/redhat-developer/odo/pkg/odo/genericclioptions"
	"github.com/redhat-developer/odo/pkg/odo/util"
	"github.com/redhat-developer/odo/pkg/odo/util/completion"
	svc "github.com/redhat-developer/odo/pkg/service"
	"github.com/spf13/cobra"
	ktemplates "k8s.io/kubernetes/pkg/kubectl/cmd/templates"
)

const createRecommendedCommandName = "create"

var (
	createExample = ktemplates.Examples(`
    # Create new postgresql service from service catalog using dev plan and name my-postgresql-db.
    %[1]s dh-postgresql-apb my-postgresql-db --plan dev -p postgresql_user=luke -p postgresql_password=secret`)

	createShortDesc = `Create a new service from service catalog using the plan defined and deploy it on OpenShift.`

	createLongDesc = ktemplates.LongDesc(`
Create a new service from service catalog using the plan defined and deploy it on OpenShift.

A --plan must be passed along with the service type. Parameters to configure the service are passed as key=value pairs.

For a full list of service types, use: 'odo catalog list services'`)
)

// ServiceCreateOptions encapsulates the options for the odo service create command
type ServiceCreateOptions struct {
	// parameters hold the user-provided values for service class parameters via flags (populated by cobra)
	parameters []string
	// plan is the selected service plan
	plan string
	// serviceType corresponds to the service class name
	serviceType string
	// serviceName is how the service will be named and known by odo
	serviceName string
	// parametersMap is populated from the flag-provided values (parameters) and/or the interactive mode and is the expected format by the business logic
	parametersMap map[string]string
	// interactive specifies whether the command operates in interactive mode or not
	interactive bool
	// generic context options common to all commands
	*genericclioptions.Context
}

// NewServiceCreateOptions creates a new ServiceCreateOptions instance
func NewServiceCreateOptions() *ServiceCreateOptions {
	return &ServiceCreateOptions{}
}

// Complete completes ServiceCreateOptions after they've been created
func (o *ServiceCreateOptions) Complete(name string, cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 || !cmd.HasFlags() {
		o.interactive = true
	}

	o.Context = genericclioptions.NewContextCreatingAppIfNeeded(cmd)

	client := o.Client

	var class scv1beta1.ClusterServiceClass
	if o.interactive {
		classesByCategory, _ := client.GetServiceClassesByCategory()
		class, o.serviceType = ui.SelectClassInteractively(classesByCategory)

		plans, _ := client.GetMatchingPlans(class)

		var svcPlan scv1beta1.ClusterServicePlan
		// if there is only one available plan, we select it
		if len(plans) == 1 {
			for k, v := range plans {
				o.plan = k
				svcPlan = v
			}
			glog.V(4).Infof("Plan %s was automatically selected since it's the only one available for service %s", o.plan, o.serviceType)
		} else {
			// otherwise select the plan interactively
			o.plan = ui.SelectPlanNameInteractively(plans, "Which service plan should we use ")
			svcPlan = plans[o.plan]
		}

		o.parametersMap = ui.EnterServicePropertiesInteractively(svcPlan)
		o.serviceName = ui.EnterServiceNameInteractively(o.serviceType, "How should we name your service ", validateNameFn)
	} else {
		o.serviceType = args[0]
		// if only one arg is given, then it is considered as service name and service type both
		o.serviceName = o.serviceType
		// if two args are given, first is service type and second one is service name
		if len(args) == 2 {
			o.serviceName = args[1]
		}
	}

	return
}

// validateServiceName adopts the Validator interface and checks that the name of the service being created is valid
func (o *ServiceCreateOptions) validateServiceName(i interface{}) (err error) {
	s := i.(string)
	err = util.ValidateName(s)
	if err != nil {
		return err
	}
	exists, err := svc.SvcExists(o.Client, s, o.Application)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%s service already exists in the current application", o.serviceName)
	}
	return
}

var validateNameFn = func(i interface{}) error {
	s := i.(string)
	return util.ValidateName(s)
}

// Validate validates the ServiceCreateOptions based on completed values
func (o *ServiceCreateOptions) Validate() (err error) {
	// if we are in interactive mode, all values are already valid
	if o.interactive {
		return nil
	}

	// make sure the service type exists
	classPtr, err := o.Client.GetClusterServiceClass(o.serviceType)
	if err != nil {
		return fmt.Errorf("unable to create service because Service Catalog is not enabled in your cluster:\n%v", err)
	}
	if classPtr == nil {
		return fmt.Errorf("service %v doesn't exist\nRun 'odo catalog list services' to see a list of supported services.\n", o.serviceType)
	}
	if len(o.plan) == 0 {
		plans, err := o.Client.GetMatchingPlans(*classPtr)
		if err != nil {
			return err
		}
		// when the plan has not been supplied, if there is only one available plan, we select it
		if len(plans) == 1 {
			for k := range plans {
				o.plan = k
			}
			glog.V(4).Infof("Plan %s was automatically selected since it's the only one available for service %s", o.plan, o.serviceType)
		} else {
			return fmt.Errorf("no plan was supplied for service %v.\nPlease select one of: %v\n", o.serviceType, strings.Join(ui.GetServicePlanNames(plans), ","))
		}
	} else {
		// when the plan has been supplied, we need to make sure it exists
		if ok, err := o.Client.DoesPlanExist(o.plan); !ok {
			if err != nil {
				return err
			}

			plans, err := o.Client.GetMatchingPlans(*classPtr)
			if err != nil {
				return err
			}
			return fmt.Errorf("plan %s is invalid for service %v.\nPlease select one of: %v\n", o.plan, o.serviceType, strings.Join(ui.GetServicePlanNames(plans), ","))
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

// Run contains the logic for the odo service create command
func (o *ServiceCreateOptions) Run() (err error) {
	err = svc.CreateService(o.Client, o.serviceName, o.serviceType, o.plan, o.parametersMap, o.Application)
	log.Successf(`Service '%s' was created`, o.serviceName)
	log.Info(`Progress of the provisioning will not be reported and might take a long time.
You can see the current status by executing 'odo service list'`)
	return
}

// NewCmdServiceCreate implements the odo service create command.
func NewCmdServiceCreate(name, fullName string) *cobra.Command {
	o := NewServiceCreateOptions()
	serviceCreateCmd := &cobra.Command{
		Use:     name + " <service_type> --plan <plan_name> [service_name]",
		Short:   createShortDesc,
		Long:    createLongDesc,
		Example: fmt.Sprintf(createExample, fullName),
		Args:    cobra.RangeArgs(0, 2),
		Run: func(cmd *cobra.Command, args []string) {
			util.LogErrorAndExit(o.Complete(name, cmd, args), "")
			util.LogErrorAndExit(o.Validate(), "")
			util.LogErrorAndExit(o.Run(), "")
		},
	}
	serviceCreateCmd.Flags().StringVar(&o.plan, "plan", "", "The name of the plan of the service to be created")
	serviceCreateCmd.Flags().StringSliceVarP(&o.parameters, "parameters", "p", []string{}, "Parameters of the plan where a parameter is expressed as <key>=<value")
	completion.RegisterCommandHandler(serviceCreateCmd, completion.ServiceClassCompletionHandler)
	completion.RegisterCommandFlagHandler(serviceCreateCmd, "plan", completion.ServicePlanCompletionHandler)
	completion.RegisterCommandFlagHandler(serviceCreateCmd, "parameters", completion.ServiceParameterCompletionHandler)
	return serviceCreateCmd
}
