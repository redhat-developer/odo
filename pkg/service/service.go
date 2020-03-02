package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/odo/util/validation"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	appsv1 "github.com/openshift/api/apps/v1"

	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/occlient"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
)

const provisionedAndBoundStatus = "ProvisionedAndBound"
const provisionedAndLinkedStatus = "ProvisionedAndLinked"

// NewServicePlanParameter creates a new ServicePlanParameter instance with the specified state
func NewServicePlanParameter(name, typeName, defaultValue string, required bool) ServicePlanParameter {
	return ServicePlanParameter{
		Name:    name,
		Default: defaultValue,
		Validatable: validation.Validatable{
			Type:     typeName,
			Required: required,
		},
	}
}

type servicePlanParameters []ServicePlanParameter

func (params servicePlanParameters) Len() int {
	return len(params)
}

func (params servicePlanParameters) Less(i, j int) bool {
	return params[i].Name < params[j].Name
}

func (params servicePlanParameters) Swap(i, j int) {
	params[i], params[j] = params[j], params[i]
}

// CreateService creates new service from serviceCatalog
func CreateService(client *occlient.Client, serviceName string, serviceType string, servicePlan string, parameters map[string]string, applicationName string) error {
	labels := componentlabels.GetLabels(serviceName, applicationName, true)
	// save service type as label
	labels[componentlabels.ComponentTypeLabel] = serviceType
	err := client.CreateServiceInstance(serviceName, serviceType, servicePlan, parameters, labels)
	if err != nil {
		return errors.Wrap(err, "unable to create service instance")

	}
	return nil
}

// CreateOperatorService creates new service (actually a Deployment) from OperatorHub
func CreateOperatorService(client *kclient.Client, serviceName string, serviceType string, crd string, parameters map[string]string, applicationName, group, version, resource string, exampleCR map[string]interface{}) error {
	err := client.CreateDynamicDeployment(exampleCR, group, version, resource)
	if err != nil {
		return errors.Wrap(err, "Unable to create operator backed service")
	}
	return nil
}

// DeleteServiceAndUnlinkComponents will delete the service with the provided `name`
// it also removes links to that service in components of the application
func DeleteServiceAndUnlinkComponents(client *occlient.Client, serviceName string, applicationName string) error {
	// first we attempt to delete the service instance itself
	labels := componentlabels.GetLabels(serviceName, applicationName, false)
	err := client.DeleteServiceInstance(labels)
	if err != nil {
		return err
	}

	// lookup all the components of the application
	applicationSelector := fmt.Sprintf("%s=%s", applabels.ApplicationLabel, applicationName)
	componentsDCs, err := client.GetDeploymentConfigsFromSelector(applicationSelector)
	if err != nil {
		return errors.Wrapf(err, "unable to list the components in order to check if they need to be unlinked")
	}

	// go through the components and check if they have the service name as part of the envFrom configuration
	for _, dc := range componentsDCs {
		for _, envFromSourceName := range dc.Spec.Template.Spec.Containers[0].EnvFrom {
			if envFromSourceName.SecretRef.Name == serviceName {
				if componentName, ok := dc.Labels[componentlabels.ComponentLabel]; ok {
					err := client.UnlinkSecret(serviceName, componentName, applicationName)
					if err != nil {
						glog.Warningf("Unable to unlink component %s from service", componentName)
					} else {
						glog.V(2).Infof("Component %s was successfully unlinked from service", componentName)
					}
				}
			}
		}
	}

	return nil
}

// List lists all the deployed services
func List(client *occlient.Client, applicationName string) (ServiceList, error) {
	labels := map[string]string{
		applabels.ApplicationLabel: applicationName,
	}

	//since, service is associated with application, it consist of application label as well
	// which we can give as a selector
	applicationSelector := util.ConvertLabelsToSelector(labels)

	// get service instance list based on given selector
	serviceInstanceList, err := client.GetServiceInstanceList(applicationSelector)
	if err != nil {
		return ServiceList{}, errors.Wrapf(err, "unable to list services")
	}

	var services []Service
	// Iterate through serviceInstanceList and add to service
	for _, elem := range serviceInstanceList {
		conditions := elem.Status.Conditions
		var status string
		if len(conditions) == 0 {
			glog.Warningf("no condition in status for %+v, marking it as Unknown", elem)
			status = "Unknown"
		} else {
			status = conditions[0].Reason
		}

		// Check and make sure that "name" exists..
		if elem.Labels[componentlabels.ComponentLabel] == "" {
			return ServiceList{}, errors.New(fmt.Sprintf("element %v returned blank name", elem))
		}

		services = append(services,
			Service{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Service",
					APIVersion: "odo.openshift.io/v1alpha1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: elem.Labels[componentlabels.ComponentLabel],
				},
				Spec:   ServiceSpec{Type: elem.Labels[componentlabels.ComponentTypeLabel], Plan: elem.Spec.ClusterServicePlanExternalName},
				Status: ServiceStatus{Status: status},
			})
	}

	return ServiceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceList",
			APIVersion: "odo.openshift.io/v1alpha1",
		},
		Items: services,
	}, nil
}

// ListWithDetailedStatus lists all the deployed services and additionally provides a "smart" status for each one of them
// The smart status takes into account how Services are used in odo.
// So when a secret has been created as a result of the created ServiceBinding, we set the appropriate status
// Same for when the secret has been "linked" into the deploymentconfig
func ListWithDetailedStatus(client *occlient.Client, applicationName string) (ServiceList, error) {

	services, err := List(client, applicationName)
	if err != nil {
		return ServiceList{}, err
	}

	// retrieve secrets in order to set status
	secrets, err := client.ListSecrets("")
	if err != nil {
		return ServiceList{}, errors.Wrapf(err, "unable to list secrets as part of the bindings check")
	}

	// use the standard selector to retrieve DeploymentConfigs
	// these are used in order to update the status of a service
	// because if a DeploymentConfig contains a secret with the service name
	// then it has been successfully linked
	labels := map[string]string{
		applabels.ApplicationLabel: applicationName,
	}
	applicationSelector := util.ConvertLabelsToSelector(labels)
	deploymentConfigs, err := client.GetDeploymentConfigsFromSelector(applicationSelector)
	if err != nil {
		return ServiceList{}, err
	}

	// go through each service and see if there is a secret that has been created
	// if so, update the status of the service
	for i, service := range services.Items {
		for _, secret := range secrets {
			if secret.Name == service.ObjectMeta.Name {
				// this is the default status when the secret exists
				services.Items[i].Status.Status = provisionedAndBoundStatus

				// if we find that the dc contains a link to the secret
				// we update the status to be even more specific
				updateStatusIfMatchingDeploymentExists(deploymentConfigs, secret.Name, services.Items, i)

				break
			}
		}
	}

	return ServiceList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceList",
			APIVersion: "odo.openshift.io/v1alpha1",
		},
		Items: services.Items,
	}, nil
}

func updateStatusIfMatchingDeploymentExists(dcs []appsv1.DeploymentConfig, secretName string, services []Service, index int) {

	for _, dc := range dcs {
		foundMatchingSecret := false
		for _, env := range dc.Spec.Template.Spec.Containers[0].EnvFrom {
			if env.SecretRef.Name == secretName {
				services[index].Status.Status = provisionedAndLinkedStatus
			}
			foundMatchingSecret = true
			break
		}

		if foundMatchingSecret {
			break
		}
	}
}

// SvcExists Checks whether a service with the given name exists in the current application or not
// serviceName is the service name to perform check for
// The first returned parameter is a bool indicating if a service with the given name already exists or not
// The second returned parameter is the error that might occurs while execution
func SvcExists(client *occlient.Client, serviceName, applicationName string) (bool, error) {

	serviceList, err := List(client, applicationName)
	if err != nil {
		return false, errors.Wrap(err, "unable to get the service list")
	}
	for _, service := range serviceList.Items {
		if service.ObjectMeta.Name == serviceName {
			return true, nil
		}
	}
	return false, nil
}

// GetServiceClassAndPlans returns the service class details with the associated plans
// serviceName is the name of the service class
// the first parameter returned is the ServiceClass object
// the second parameter returned is the array of ServicePlan associated with the service class
func GetServiceClassAndPlans(client *occlient.Client, serviceName string) (ServiceClass, []ServicePlan, error) {
	result, err := client.GetClusterServiceClass(serviceName)
	if err != nil {
		return ServiceClass{}, nil, errors.Wrap(err, "unable to get the given service")
	}

	var meta map[string]interface{}
	err = json.Unmarshal(result.Spec.ExternalMetadata.Raw, &meta)
	if err != nil {
		return ServiceClass{}, nil, errors.Wrap(err, "unable to unmarshal data the given service")
	}

	service := ServiceClass{
		Name:              result.Spec.ExternalName,
		Bindable:          result.Spec.Bindable,
		ShortDescription:  result.Spec.Description,
		Tags:              result.Spec.Tags,
		ServiceBrokerName: result.Spec.ClusterServiceBrokerName,
	}

	if val, ok := meta["longDescription"]; ok {
		service.LongDescription = val.(string)
	}

	if val, ok := meta["dependencies"]; ok {
		versions := fmt.Sprint(val)
		versions = strings.Replace(versions, "[", "", -1)
		versions = strings.Replace(versions, "]", "", -1)
		service.VersionsAvailable = strings.Split(versions, " ")
	}

	// get the plans according to the service name
	planResults, err := client.GetClusterPlansFromServiceName(result.Name)
	if err != nil {
		return ServiceClass{}, nil, errors.Wrap(err, "unable to get plans for the given service")
	}

	var plans []ServicePlan
	for _, result := range planResults {
		plan, err := NewServicePlan(result)
		if err != nil {
			return ServiceClass{}, nil, err
		}

		plans = append(plans, plan)
	}

	return service, plans, nil
}

type InstanceCreateParameterSchema struct {
	Required   []string
	Properties map[string]ServicePlanParameter
}

// NewServicePlan creates a new ServicePlan based on the specified ClusterServicePlan
func NewServicePlan(result scv1beta1.ClusterServicePlan) (plan ServicePlan, err error) {
	plan = ServicePlan{
		Name:        result.Spec.ExternalName,
		Description: result.Spec.Description,
	}

	// get the display name from the external meta data
	var externalMetaData map[string]interface{}
	err = json.Unmarshal(result.Spec.ExternalMetadata.Raw, &externalMetaData)
	if err != nil {
		return plan, errors.Wrap(err, "unable to unmarshal data the given service")
	}

	if val, ok := externalMetaData["displayName"]; ok {
		plan.DisplayName = val.(string)
	}

	// get the create parameters
	schema := InstanceCreateParameterSchema{}
	paramBytes := result.Spec.InstanceCreateParameterSchema.Raw
	err = json.Unmarshal(paramBytes, &schema)
	if err != nil {
		return plan, errors.Wrapf(err, "unable to unmarshal data the given service: %s", string(paramBytes[:]))
	}

	plan.Parameters = make([]ServicePlanParameter, 0, len(schema.Properties))
	for k, v := range schema.Properties {
		v.Name = k
		// we set the Required flag if the name of parameter
		// is one of the parameters indicated as required
		// these parameters are not strictly required since they might have default values
		v.Required = isRequired(schema.Required, k)

		plan.Parameters = append(plan.Parameters, v)
	}

	return
}

// isRequired checks whether the parameter with the specified name is among the given list of required ones
func isRequired(required []string, name string) bool {
	for _, n := range required {
		if n == name {
			return true
		}
	}
	return false
}
