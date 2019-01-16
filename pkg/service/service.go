package service

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"strings"

	scv1beta1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	appsv1 "github.com/openshift/api/apps/v1"
	"sort"

	"github.com/pkg/errors"
	applabels "github.com/redhat-developer/odo/pkg/application/labels"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/util"
)

const provisionedAndBoundStatus = "ProvisionedAndBound"
const provisionedAndLinkedStatus = "ProvisionedAndLinked"

// ServiceInfo holds all important information about one service
type ServiceInfo struct {
	Name   string
	Type   string
	Status string
}

type ServiceClass struct {
	Name              string
	Bindable          bool
	ShortDescription  string
	LongDescription   string
	Tags              []string
	VersionsAvailable []string
	ServiceBrokerName string
}

type ServicePlanParameter struct {
	Name            string
	Title           string
	Description     string
	HasDefaultValue bool
	Default         string
	Type            string
	Required        bool
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

type ServicePlan struct {
	Name        string
	DisplayName string
	Description string
	Parameters  servicePlanParameters
}

// ListCatalog lists all the available service types
func ListCatalog(client *occlient.Client) ([]occlient.Service, error) {

	clusterServiceClasses, err := client.GetClusterServiceClassExternalNamesAndPlans()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get cluster serviceClassExternalName")
	}

	// Sorting service classes alphabetically
	// Reference: https://golang.org/pkg/sort/#example_Slice
	sort.Slice(clusterServiceClasses, func(i, j int) bool {
		return clusterServiceClasses[i].Name < clusterServiceClasses[j].Name
	})

	return clusterServiceClasses, nil
}

// Search searches for the services
func Search(client *occlient.Client, name string) ([]occlient.Service, error) {
	var result []occlient.Service
	serviceList, err := ListCatalog(client)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list services")
	}

	// do a partial search in all the services
	for _, service := range serviceList {
		if strings.Contains(service.Name, name) {
			result = append(result, service)
		}
	}

	return result, nil
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
						glog.V(2).Infof("Component %s was succesfully unlinked from service", componentName)
					}
				}
			}
		}
	}

	return nil
}

// List lists all the deployed services
func List(client *occlient.Client, applicationName string) ([]ServiceInfo, error) {
	labels := map[string]string{
		applabels.ApplicationLabel: applicationName,
	}

	//since, service is associated with application, it consist of application label as well
	// which we can give as a selector
	applicationSelector := util.ConvertLabelsToSelector(labels)

	// get service instance list based on given selector
	serviceInstanceList, err := client.GetServiceInstanceList(applicationSelector)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list services")
	}

	var services []ServiceInfo
	// Iterate through serviceInstanceList and add to service
	for _, elem := range serviceInstanceList {
		services = append(services, ServiceInfo{Name: elem.Labels[componentlabels.ComponentLabel], Type: elem.Labels[componentlabels.ComponentTypeLabel], Status: elem.Status.Conditions[0].Reason})
	}

	return services, nil
}

// ListWithDetailedStatus lists all the deployed services and additionally provides a "smart" status for each one of them
// The smart status takes into account how Services are used in odo.
// So when a secret has been created as a result of the created ServiceBinding, we set the appropriate status
// Same for when the secret has been "linked" into the deploymentconfig
func ListWithDetailedStatus(client *occlient.Client, applicationName string) ([]ServiceInfo, error) {

	services, err := List(client, applicationName)
	if err != nil {
		return nil, err
	}

	// retrieve secrets in order to set status
	secrets, err := client.ListSecrets("")
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list secrets as part of the bindings check")
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
		return nil, err
	}

	// go through each service and see if there is a secret that has been created
	// if so, update the status of the service
	for i, service := range services {
		for _, secret := range secrets {
			if secret.Name == service.Name {
				// this is the default status when the secret exists
				services[i].Status = provisionedAndBoundStatus

				// if we find that the dc contains a link to the secret
				// we update the status to be even more specific
				updateStatusIfMatchingDeploymentExists(deploymentConfigs, secret.Name, services, i)

				break
			}
		}
	}

	return services, nil
}

func updateStatusIfMatchingDeploymentExists(dcs []appsv1.DeploymentConfig, secretName string,
	services []ServiceInfo, index int) {

	for _, dc := range dcs {
		foundMatchingSecret := false
		for _, env := range dc.Spec.Template.Spec.Containers[0].EnvFrom {
			if env.SecretRef.Name == secretName {
				services[index].Status = provisionedAndLinkedStatus
			}
			foundMatchingSecret = true
			break
		}

		if foundMatchingSecret {
			break
		}
	}
}

// GetSvcByType returns the matching (by type) service or nil of there are no matches
func GetSvcByType(client *occlient.Client, serviceType string) (*occlient.Service, error) {
	catalogList, err := ListCatalog(client)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list catalog")
	}

	for _, supported := range catalogList {
		if serviceType == supported.Name {
			return &supported, nil
		}
	}
	return nil, nil
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
	for _, service := range serviceList {
		if service.Name == serviceName {
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

		}

		plans = append(plans, plan)
	}

	return service, plans, nil
}

type serviceInstanceCreateParameterSchema struct {
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
	schema := serviceInstanceCreateParameterSchema{}
	paramBytes := result.Spec.ServiceInstanceCreateParameterSchema.Raw
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
		if len(v.Default) > 0 {
			v.HasDefaultValue = true
		}

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
