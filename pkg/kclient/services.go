package kclient

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateService generates and creates the service
// commonObjectMeta is the ObjectMeta for the service
func (c *Client) CreateService(commonObjectMeta metav1.ObjectMeta, svcSpec corev1.ServiceSpec) (*corev1.Service, error) {
	svc := corev1.Service{
		ObjectMeta: commonObjectMeta,
		Spec:       svcSpec,
	}

	service, err := c.KubeClient.CoreV1().Services(c.Namespace).Create(&svc)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create Service for %s", commonObjectMeta.Name)
	}
	return service, err
}

// UpdateService updates a service based on the given service spec
func (c *Client) UpdateService(commonObjectMeta metav1.ObjectMeta, svcSpec corev1.ServiceSpec) (*corev1.Service, error) {
	svc := corev1.Service{
		ObjectMeta: commonObjectMeta,
		Spec:       svcSpec,
	}

	service, err := c.KubeClient.CoreV1().Services(c.Namespace).Update(&svc)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to update Service for %s", commonObjectMeta.Name)
	}
	return service, err
}

// ListServices returns an array of Service resources which match the
// given selector
func (c *Client) ListServices(selector string) ([]corev1.Service, error) {
	serviceList, err := c.KubeClient.CoreV1().Services(c.Namespace).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to list Services")
	}
	return serviceList.Items, nil
}
