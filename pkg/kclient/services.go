package kclient

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetService retrieves the service with the given name
func (c *Client) GetService(name string) (*corev1.Service, error) {
	service, err := c.KubeClient.CoreV1().Services(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get the Service with name %q", name)
	}
	return service, err
}

// CreateService generates and creates the service
// commonObjectMeta is the ObjectMeta for the service
func (c *Client) CreateService(svc corev1.Service) (*corev1.Service, error) {
	service, err := c.KubeClient.CoreV1().Services(c.Namespace).Create(context.TODO(), &svc, metav1.CreateOptions{FieldManager: FieldManager})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create Service for %s", svc.Name)
	}
	return service, err
}

// UpdateService updates a service based on the given service spec
func (c *Client) UpdateService(svc corev1.Service) (*corev1.Service, error) {
	service, err := c.KubeClient.CoreV1().Services(c.Namespace).Update(context.TODO(), &svc, metav1.UpdateOptions{FieldManager: FieldManager})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to update Service for %s", svc.Name)
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
