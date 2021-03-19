/*
	This file contains code for various service backends supported by odo. Different backends have different logics for
	Complete, Validate and Run functions. These are covered in this file.
*/
package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	scv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/openshift/odo/pkg/log"
	"github.com/openshift/odo/pkg/odo/cli/service/ui"
	commonui "github.com/openshift/odo/pkg/odo/cli/ui"
	svc "github.com/openshift/odo/pkg/service"
	olm "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

// This CompleteServiceCreate contains logic to complete the "odo service create" call for the case of Operator backend
func (b *OperatorBackend) CompleteServiceCreate(o *CreateOptions, cmd *cobra.Command, args []string) (err error) {
	// since interactive mode is not supported for Operators yet, set it to false
	o.interactive = false

	// if user has just used "odo service create", simply return
	if o.fromFile == "" && len(args) == 0 {
		return
	}

	// if user wants to create service from file and use a name given on CLI
	if o.fromFile != "" {
		if len(args) == 1 {
			o.ServiceName = args[0]
		}
		return
	}

	// split the name provided on CLI and populate servicetype & customresource
	o.ServiceType, b.CustomResource, err = svc.SplitServiceKindName(args[0])
	if err != nil {
		return fmt.Errorf("invalid service name, use the format <operator-type>/<crd-name>")
	}

	// if two args are given, first is service type and second one is service name
	if len(args) == 2 {
		o.ServiceName = args[1]
	}

	return nil
}

func (b *OperatorBackend) ValidateServiceCreate(o *CreateOptions) (err error) {
	d := NewDynamicCRD()
	// if the user wants to create service from a file, we check for
	// existence of file and validate if the requested operator and CR
	// exist on the cluster
	if o.fromFile != "" {
		if _, err := os.Stat(o.fromFile); err != nil {
			return errors.Wrap(err, "unable to find specified file")
		}

		// Parse the file to find Operator and CR info
		fileContents, err := ioutil.ReadFile(o.fromFile)
		if err != nil {
			return err
		}

		err = yaml.Unmarshal(fileContents, &d.OriginalCRD)
		if err != nil {
			return err
		}

		// Check if the operator and the CR exist on cluster
		var csv olm.ClusterServiceVersion
		b.CustomResource, csv, err = svc.GetCSV(o.KClient, d.OriginalCRD)
		if err != nil {
			return err
		}

		// all is well, let's populate the fields required for creating operator backed service
		b.group, b.version, b.resource, err = svc.GetGVRFromOperator(csv, b.CustomResource)
		if err != nil {
			return err
		}

		err = d.validateMetadataInCRD()
		if err != nil {
			return err
		}

		if o.ServiceName != "" && !o.DryRun {
			// First check if service with provided name already exists
			svcFullName := strings.Join([]string{b.CustomResource, o.ServiceName}, "/")
			exists, err := svc.OperatorSvcExists(o.KClient, svcFullName)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("service %q already exists; please provide a different name or delete the existing service first", svcFullName)
			}

			d.setServiceName(o.ServiceName)
		} else {
			o.ServiceName, err = d.getServiceNameFromCRD()
			if err != nil {
				return err
			}
		}

		// CRD is valid. We can use it further to create a service from it.
		b.CustomResourceDefinition = d.OriginalCRD

		return nil
	} else if b.CustomResource != "" {
		// make sure that CSV of the specified ServiceType exists
		csv, err := o.KClient.GetClusterServiceVersion(o.ServiceType)
		if err != nil {
			// error only occurs when OperatorHub is not installed.
			// k8s does't have it installed by default but OCP does
			return err
		}

		almExample, err := svc.GetAlmExample(csv, b.CustomResource, o.ServiceType)
		if err != nil {
			return err
		}

		d.OriginalCRD = almExample

		b.group, b.version, b.resource, err = svc.GetGVRFromOperator(csv, b.CustomResource)
		if err != nil {
			return err
		}

		if o.ServiceName != "" && !o.DryRun {
			// First check if service with provided name already exists
			svcFullName := strings.Join([]string{b.CustomResource, o.ServiceName}, "/")
			exists, err := svc.OperatorSvcExists(o.KClient, svcFullName)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("service %q already exists; please provide a different name or delete the existing service first", svcFullName)
			}

			d.setServiceName(o.ServiceName)
		}

		err = d.validateMetadataInCRD()
		if err != nil {
			return err
		}

		// CRD is valid. We can use it further to create a service from it.
		b.CustomResourceDefinition = d.OriginalCRD

		if o.ServiceName == "" {
			o.ServiceName, err = d.getServiceNameFromCRD()
			if err != nil {
				return err
			}
		}

		return nil
	} else {
		// This block is executed only when user has neither provided a
		// file nor a valid `odo service create <operator-name>` to start
		// the service from an Operator. So we raise an error because the
		// correct way is to execute:
		// `odo service create <operator-name>/<crd-name>`

		return fmt.Errorf("please use a valid command to start an Operator backed service; desired format: %q", "odo service create <operator-name>/<crd-name>")
	}
}

func (b *OperatorBackend) RunServiceCreate(o *CreateOptions) (err error) {
	s := &log.Status{}

	// in case of an Operator backed service, name of the service is
	// provided by the yaml specification in alm-examples. It might also
	// happen that a user wants to spin up Service Catalog based service in
	// spite of having 4.x cluster mode but we're not supporting
	// interacting with both Operator Hub and Service Catalog on 4.x. So
	// the user won't get to see service name in the log message
	if !o.DryRun {
		log.Infof("Deploying service %q of type: %q", o.ServiceName, b.CustomResource)
		s = log.Spinner("Deploying service")
		defer s.End(false)
	}

	// if cluster has resources of type CSV and o.CustomResource is not
	// empty, we're expected to create an Operator backed service
	if o.DryRun {
		// if it's dry run, only print the alm-example (o.CustomResourceDefinition) and exit
		jsonCR, err := json.MarshalIndent(b.CustomResourceDefinition, "", "  ")
		if err != nil {
			return err
		}

		// convert json to yaml
		yamlCR, err := yaml.JSONToYAML(jsonCR)
		if err != nil {
			return err
		}

		log.Info(string(yamlCR))

		return nil
	} else {
		err = svc.CreateOperatorService(o.KClient, o.EnvSpecificInfo, o.ServiceName, b.group, b.version, b.resource, b.CustomResourceDefinition)
		if err != nil {
			// TODO: logic to remove CRD info from devfile because service creation failed.
			return err
		} else {
			s.End(true)
			log.Successf(`Service %q was created`, o.ServiceName)
		}

		crdYaml, err := yaml.Marshal(b.CustomResourceDefinition)
		if err != nil {
			return err
		}

		err = svc.AddKubernetesComponentToDevfile(string(crdYaml), o.ServiceName, o.EnvSpecificInfo.GetDevfileObj())
		if err != nil {
			return err
		}
	}
	s.End(true)

	return
}

func (b *OperatorBackend) CompleteServiceDelete(o *DeleteOptions, cmd *cobra.Command, args []string) (err error) {
	return
}

func (b *OperatorBackend) ServiceExists(o *DeleteOptions) (bool, error) {
	return svc.OperatorSvcExists(o.KClient, o.serviceName)
}

func (b *OperatorBackend) DeleteService(o *DeleteOptions, name string, application string) error {
	err := svc.DeleteOperatorService(o.KClient, o.serviceName)
	if err != nil {
		return err
	}

	// "name" is of the form CR-Name/Instance-Name so we split it
	// we ignore the error because the function used below is called in the call to "DeleteOperatorService" above.
	_, instanceName, _ := svc.SplitServiceKindName(name)

	err = svc.DeleteKubernetesComponentFromDevfile(instanceName, o.EnvSpecificInfo.GetDevfileObj())
	if err != nil {
		return errors.Wrap(err, "failed to delete service from the devfile")
	}

	return nil
}

func (b *ServiceCatalogBackend) CompleteServiceDelete(o *DeleteOptions, cmd *cobra.Command, args []string) (err error) {
	return nil
}

// This CompleteServiceCreate contains logic to complete the "odo service create" call for the case of Service Catalog backend
func (b *ServiceCatalogBackend) CompleteServiceCreate(o *CreateOptions, cmd *cobra.Command, args []string) (err error) {
	if len(args) == 0 && !cmd.HasFlags() {
		o.interactive = true

	}

	var class scv1beta1.ClusterServiceClass

	if o.interactive {
		classesByCategory, err := o.Client.GetKubeClient().ListServiceClassesByCategory()
		if err != nil {
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

		// we convert the param list provided in the format of key=value list
		// to a map
		o.ParametersMap = make(map[string]string)
		for _, kv := range o.parameters {
			kvSlice := strings.Split(kv, "=")
			// key value not provided in format of key=value
			if len(kvSlice) != 2 {
				return errors.New("parameters not provided in key=value format")
			}
			o.ParametersMap[kvSlice[0]] = kvSlice[1]
		}
	}
	return nil
}

func (b *ServiceCatalogBackend) ValidateServiceCreate(o *CreateOptions) (err error) {
	// make sure the service type exists
	classPtr, err := o.Client.GetKubeClient().GetClusterServiceClass(o.ServiceType)
	if err != nil {
		return errors.Wrap(err, "unable to create service because Service Catalog is not enabled in your cluster")
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

func (b *ServiceCatalogBackend) RunServiceCreate(o *CreateOptions) (err error) {
	s := &log.Status{}

	log.Infof("Deploying service %q of type: %q", o.ServiceName, o.ServiceType)
	// create a ServiceInstance
	serviceInstance, err := svc.CreateService(o.Client, o.EnvSpecificInfo, o.ServiceName, o.ServiceType, o.Plan, o.ParametersMap, o.Application)
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
