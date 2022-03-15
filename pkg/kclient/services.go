package kclient

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateService generates and creates the service
// commonObjectMeta is the ObjectMeta for the service
func (c *Client) CreateService(svc corev1.Service) (*corev1.Service, error) {
	service, err := c.KubeClient.CoreV1().Services(c.Namespace).Create(context.TODO(), &svc, metav1.CreateOptions{FieldManager: FieldManager})
	if err != nil {
		return nil, fmt.Errorf("unable to create Service for %s: %w", svc.Name, err)
	}
	return service, err
}

// UpdateService updates a service based on the given service spec
func (c *Client) UpdateService(svc corev1.Service) (*corev1.Service, error) {
	service, err := c.KubeClient.CoreV1().Services(c.Namespace).Update(context.TODO(), &svc, metav1.UpdateOptions{FieldManager: FieldManager})
	if err != nil {
		return nil, fmt.Errorf("unable to update Service for %s: %w", svc.Name, err)
	}
	return service, err
}

// ListServices returns an array of Service resources which match the
// given selector
func (c *Client) ListServices(selector string) ([]corev1.Service, error) {
	serviceList, err := c.KubeClient.CoreV1().Services(c.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Services")
	}
	return serviceList.Items, nil
}

// DeleteService deletes the service with the given service name
func (c *Client) DeleteService(serviceName string) error {
	err := c.KubeClient.CoreV1().Services(c.Namespace).Delete(context.TODO(), serviceName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

// GetOneService retrieves the service with the given component and app name
// An error is thrown when exactly one service is not found for the selector.
func (c *Client) GetOneService(componentName, appName string) (*corev1.Service, error) {
	return c.GetOneServiceFromSelector(componentlabels.GetSelector(componentName, appName))
}

// GetOneServiceFromSelector returns the service object associated with the given selector.
// An error is thrown when exactly one service is not found for the selector.
func (c *Client) GetOneServiceFromSelector(selector string) (*corev1.Service, error) {
	services, err := c.ListServices(selector)
	if err != nil {
		return nil, fmt.Errorf("unable to get services for the selector %q : %w", selector, err)
	}

	num := len(services)
	if num == 0 {
		return nil, &ServiceNotFoundError{Selector: selector}
	} else if num > 1 {
		return nil, fmt.Errorf("multiple services exist for the selector: %v. Only one must be present", selector)
	}

	return &services[0], nil
}
