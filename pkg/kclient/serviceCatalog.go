package kclient

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/ghodss/yaml"
	scv1beta1 "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
)

// serviceInstanceParameters converts a map of variable assignments to a byte encoded json document,
// which is what the ServiceCatalog API consumes.
func serviceInstanceParameters(params map[string]string) (*runtime.RawExtension, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	return &runtime.RawExtension{Raw: paramsJSON}, nil
}

// CreateServiceInstance creates service instance from service catalog
func (c *Client) CreateServiceInstance(serviceName string, serviceType string, servicePlan string, parameters map[string]string, labels map[string]string) (string, error) {
	serviceInstanceParameters, err := serviceInstanceParameters(parameters)
	if err != nil {
		return "", errors.Wrap(err, "unable to create the service instance parameters")
	}

	si := &scv1beta1.ServiceInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceInstance",
			APIVersion: "servicecatalog.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: c.Namespace,
			Labels:    labels,
		},
		Spec: scv1beta1.ServiceInstanceSpec{
			PlanReference: scv1beta1.PlanReference{
				ClusterServiceClassExternalName: serviceType,
				ClusterServicePlanExternalName:  servicePlan,
			},
			Parameters: serviceInstanceParameters,
		},
	}

	_, err = c.serviceCatalogClient.ServiceInstances(c.Namespace).Create(context.TODO(), si, metav1.CreateOptions{FieldManager: FieldManager})

	if err != nil {
		return "", errors.Wrapf(err, "unable to create the service instance %s for the service type %s and plan %s", serviceName, serviceType, servicePlan)
	}

	// Create the secret containing the parameters of the plan selected.
	err = c.CreateServiceBinding(serviceName, c.Namespace, labels)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create the secret %s for the service instance", serviceName)
	}

	siString, err := yaml.Marshal(si)
	if err != nil {
		return "", errors.Wrapf(err, "unable to marshal service instance to string")
	}

	return string(siString), nil
}

// ListServiceInstances returns list service instances
func (c *Client) ListServiceInstances(selector string) ([]scv1beta1.ServiceInstance, error) {
	// List ServiceInstance according to given selectors
	svcList, err := c.serviceCatalogClient.ServiceInstances(c.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list ServiceInstances")
	}

	return svcList.Items, nil
}

// DeleteServiceInstance takes labels as a input and based on it, deletes respective service instance
func (c *Client) DeleteServiceInstance(labels map[string]string) error {
	klog.V(3).Infof("Deleting Service Instance")

	// convert labels to selector
	selector := util.ConvertLabelsToSelector(labels)
	klog.V(3).Infof("Selectors used for deletion: %s", selector)

	// Listing out serviceInstance because `DeleteCollection` method don't work on serviceInstance
	serviceInstances, err := c.ListServiceInstances(selector)
	if err != nil {
		return errors.Wrap(err, "unable to list service instance")
	}

	// Iterating over serviceInstance List and deleting one by one
	for _, serviceInstance := range serviceInstances {
		// we need to delete the ServiceBinding before deleting the ServiceInstance
		err = c.serviceCatalogClient.ServiceBindings(c.Namespace).Delete(context.TODO(), serviceInstance.Name, metav1.DeleteOptions{})
		if err != nil {
			return errors.Wrap(err, "unable to delete serviceBinding")
		}
		// now we perform the actual deletion
		err = c.serviceCatalogClient.ServiceInstances(c.Namespace).Delete(context.TODO(), serviceInstance.Name, metav1.DeleteOptions{})
		if err != nil {
			return errors.Wrap(err, "unable to delete serviceInstance")
		}
	}

	return nil
}

// GetClusterServiceClass returns the required service class from the service name
// serviceName is the name of the service
// returns the required service class and the error
func (c *Client) GetClusterServiceClass(serviceName string) (*scv1beta1.ClusterServiceClass, error) {
	opts := metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("spec.externalName", serviceName).String(),
	}
	searchResults, err := c.serviceCatalogClient.ClusterServiceClasses().List(context.TODO(), opts)
	if err != nil {
		return nil, fmt.Errorf("unable to search classes by name (%s)", err)
	}
	if len(searchResults.Items) == 0 {
		return nil, fmt.Errorf("class '%s' not found", serviceName)
	}
	if len(searchResults.Items) > 1 {
		return nil, fmt.Errorf("more than one matching class found for '%s'", serviceName)
	}
	return &searchResults.Items[0], nil
}

// ListClusterServiceClasses queries the service service catalog to get available clusterServiceClasses
func (c *Client) ListClusterServiceClasses() ([]scv1beta1.ClusterServiceClass, error) {
	classList, err := c.serviceCatalogClient.ClusterServiceClasses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list cluster service classes")
	}
	return classList.Items, nil
}

// ListServiceClassesByCategory retrieves a map associating category name to ClusterServiceClasses matching the category
func (c *Client) ListServiceClassesByCategory() (categories map[string][]scv1beta1.ClusterServiceClass, err error) {
	categories = make(map[string][]scv1beta1.ClusterServiceClass)
	classes, err := c.ListClusterServiceClasses()
	if err != nil {
		return nil, err
	}

	// TODO: Should we replicate the classification performed in
	// https://github.com/openshift/console/blob/master/frontend/public/components/catalog/catalog-items.jsx?
	for _, class := range classes {
		tags := class.Spec.Tags
		category := "other"
		if len(tags) > 0 && len(tags[0]) > 0 {
			category = tags[0]
		}
		categories[category] = append(categories[category], class)
	}

	return categories, err
}

// ListClusterServicePlans returns list of available plans
func (c *Client) ListClusterServicePlans() ([]scv1beta1.ClusterServicePlan, error) {
	planList, err := c.serviceCatalogClient.ClusterServicePlans().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster service plan")
	}

	return planList.Items, nil
}

// ListClusterServicePlansByServiceName returns the plans associated with a service class
// serviceName is the name (the actual id, NOT the external name) of the service class whose plans are required
// returns array of ClusterServicePlans or error
func (c *Client) ListClusterServicePlansByServiceName(serviceName string) ([]scv1beta1.ClusterServicePlan, error) {
	opts := metav1.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("spec.clusterServiceClassRef.name", serviceName).String(),
	}

	searchResults, err := c.serviceCatalogClient.ClusterServicePlans().List(context.TODO(), opts)
	if err != nil {
		return nil, fmt.Errorf("unable to search plans for service name '%s', (%s)", serviceName, err)
	}
	return searchResults.Items, nil
}

// GetServiceBinding returns the ServiceBinding named serviceName in the namespace namespace
func (c *Client) GetServiceBinding(serviceName string, namespace string) (*scv1beta1.ServiceBinding, error) {
	return c.serviceCatalogClient.ServiceBindings(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
}

// CreateServiceBinding creates a ServiceBinding (essentially a secret) within the namespace of the
// service instance created using the service's parameters.
func (c *Client) CreateServiceBinding(bindingName string, namespace string, labels map[string]string) error {
	_, err := c.serviceCatalogClient.ServiceBindings(namespace).Create(context.TODO(),
		&scv1beta1.ServiceBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      bindingName,
				Namespace: namespace,
				Labels:    labels,
			},
			Spec: scv1beta1.ServiceBindingSpec{
				InstanceRef: scv1beta1.LocalObjectReference{
					Name: bindingName,
				},
				SecretName: bindingName,
			},
		}, metav1.CreateOptions{FieldManager: FieldManager})

	if err != nil {
		return errors.Wrap(err, "Creation of the secret failed")
	}

	return nil
}

// ListMatchingPlans retrieves a map associating service plan name to service plan instance associated with the specified service
// class
func (c *Client) ListMatchingPlans(class scv1beta1.ClusterServiceClass) (plans map[string]scv1beta1.ClusterServicePlan, err error) {
	planList, err := c.serviceCatalogClient.ClusterServicePlans().List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.clusterServiceClassRef.name==" + class.Spec.ExternalID,
	})

	plans = make(map[string]scv1beta1.ClusterServicePlan)
	for _, v := range planList.Items {
		plans[v.Spec.ExternalName] = v
	}
	return plans, err
}

// ListServiceInstanceLabelValues get label values of given label from objects in project that match the selector
func (c *Client) ListServiceInstanceLabelValues(label string, selector string) ([]string, error) {

	// List ServiceInstance according to given selectors
	svcList, err := c.serviceCatalogClient.ServiceInstances(c.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list ServiceInstances")
	}

	// Grab all the matched strings
	var values []string
	for _, elem := range svcList.Items {
		val, ok := elem.Labels[label]
		if ok {
			values = append(values, val)
		}
	}

	// Sort alphabetically
	sort.Strings(values)

	return values, nil
}
