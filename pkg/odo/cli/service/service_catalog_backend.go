package service

import (
	"fmt"
	"strings"

	"github.com/openshift/odo/pkg/odo/util/validation"

	svc "github.com/openshift/odo/pkg/service"

	scv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/cli/service/ui"
	commonui "github.com/openshift/odo/pkg/odo/cli/ui"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

// This CompleteServiceCreate contains logic to complete the "odo service create" call for the case of Service Catalog backend
func (b *ServiceCatalogBackend) CompleteServiceCreate(o *CreateOptions, cmd *cobra.Command, args []string) (err error) {
	var class scv1beta1.ClusterServiceClass

	if o.interactive {
		classesByCategory, err := o.Client.GetKubeClient().ListServiceClassesByCategory()
		if err != nil {
			// this error indicates that Service Catalog is not properly setup
			// we inform the user that if they're trying interactive mode for Operators, it's not yet supported.
			// TODO: remove the warning when interactive mode for Operators is supported
			log.Warning("odo doesn't support interactive mode for creating Operator backed service yet; refer \"odo service create -h\"")
			return fmt.Errorf("unable to retrieve service classes: %v", err)
		}

		if len(classesByCategory) == 0 {
			return fmt.Errorf("no available service classes")
		}

		class, o.ServiceType = ui.SelectClassInteractively(classesByCategory)

		plans, err := o.Client.GetKubeClient().ListMatchingPlans(class)
		if err != nil {
			return fmt.Errorf("couldn't retrieve plans for class %s: %v", class.GetExternalName(), err)
		}

		var svcPlan scv1beta1.ClusterServicePlan
		// if there is only one available plan, we select it
		if len(plans) == 1 {
			for k, v := range plans {
				o.Plan = k
				svcPlan = v
			}
			klog.V(4).Infof("Plan %s was automatically selected since it's the only one available for service %s", o.Plan, o.ServiceType)
		} else {
			// otherwise select the plan interactively
			o.Plan = ui.SelectPlanNameInteractively(plans, "Which service plan should we use ")
			svcPlan = plans[o.Plan]
		}

		o.ParametersMap = ui.EnterServicePropertiesInteractively(svcPlan)
		o.ServiceName = ui.EnterServiceNameInteractively(o.ServiceType, "How should we name your service ", o.validateServiceName)
		o.outputCLI = commonui.Proceed("Output the non-interactive version of the selected options")
		o.wait = commonui.Proceed("Wait for the service to be ready")
	} else {

		o.ServiceType = args[0]

		// if two args are given, first is service type and second one is service name
		if len(args) == 2 {
			o.ServiceName = args[1]
		} else {
			o.ServiceName = o.ServiceType
		}

	}
	return nil
}

func (b *ServiceCatalogBackend) ValidateServiceCreate(o *CreateOptions) (err error) {
	// make sure the service type exists
	classPtr, err := o.Client.GetKubeClient().GetClusterServiceClass(o.ServiceType)
	if err != nil {
		return fmt.Errorf("unable to create service because Service Catalog is not enabled in your cluster")
	}
	if classPtr == nil {
		return fmt.Errorf("service %v doesn't exist\nRun 'odo catalog list services' to see a list of supported services.\n", o.ServiceType)
	}

	// check plan
	plans, err := o.Client.GetKubeClient().ListMatchingPlans(*classPtr)
	if err != nil {
		return err
	}
	if len(o.Plan) == 0 {
		// when the plan has not been supplied, if there is only one available plan, we select it
		if len(plans) == 1 {
			for k := range plans {
				o.Plan = k
			}
			klog.V(4).Infof("Plan %s was automatically selected since it's the only one available for service %s", o.Plan, o.ServiceType)
		} else {
			return fmt.Errorf("no plan was supplied for service %v.\nPlease select one of: %v\n", o.ServiceType, strings.Join(ui.GetServicePlanNames(plans), ","))
		}
	} else {
		// when the plan has been supplied, we need to make sure it exists
		if _, ok := plans[o.Plan]; !ok {
			return fmt.Errorf("plan %s is invalid for service %v.\nPlease select one of: %v\n", o.Plan, o.ServiceType, strings.Join(ui.GetServicePlanNames(plans), ","))
		}
	}
	//validate service name
	return o.validateServiceName(o.ServiceName)
}

// validateServiceName adopts the Validator interface and checks that the name of the service being created is valid
func (o *CreateOptions) validateServiceName(i interface{}) (err error) {
	s := i.(string)
	err = validation.ValidateName(s)
	if err != nil {
		return err
	}
	exists, err := svc.SvcExists(o.Client, s, o.Application)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%s service already exists in the current application", o.ServiceName)
	}
	return
}

func (b *ServiceCatalogBackend) RunServiceCreate(o *CreateOptions) (err error) {
	s := &log.Status{}

	log.Infof("Deploying service %q of type: %q", o.ServiceName, o.ServiceType)
	// create a ServiceInstance
	serviceInstance, err := svc.CreateService(o.Client, o.ServiceName, o.ServiceType, o.Plan, o.ParametersMap, o.Application)
	if err != nil {
		return err
	}

	err = svc.AddKubernetesComponentToDevfile(serviceInstance, o.ServiceName, o.EnvSpecificInfo.GetDevfileObj())
	if err != nil {
		return err
	}

	s.End(true)

	if o.wait {
		s = log.Spinner("Waiting for service to come up")
		_, err = o.Client.GetKubeClient().WaitAndGetSecret(o.ServiceName, o.Project)
		if err == nil {
			s.End(true)
			log.Successf(`Service %q is ready for use`, o.ServiceName)
		}
	} else {
		log.Successf(`Service %q was created`, o.ServiceName)
		log.Italic("\nProgress of the provisioning will not be reported and might take a long time\nYou can see the current status by executing 'odo service list'")
	}
	return
}

// ServiceDefined returns true if the service is defined in the devfile
func (b *ServiceCatalogBackend) ServiceDefined(o *DeleteOptions) (bool, error) {
	return svc.IsDefined(o.serviceName, o.EnvSpecificInfo.GetDevfileObj())
}

func (b *ServiceCatalogBackend) ServiceExists(o *DeleteOptions) (bool, error) {
	return svc.SvcExists(o.Client, o.serviceName, o.Application)
}

func (b *ServiceCatalogBackend) DeleteService(o *DeleteOptions, name string, application string) error {
	err := svc.DeleteServiceAndUnlinkComponents(o.Client, o.serviceName, o.Application)
	if err != nil {
		return err
	}

	err = svc.DeleteKubernetesComponentFromDevfile(o.serviceName, o.EnvSpecificInfo.GetDevfileObj())
	if err != nil {
		return err
	}

	return nil
}
