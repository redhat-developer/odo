package kclient

import (
	"github.com/pkg/errors"

	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateIngress creates an ingress object for the given service and with the given labels
func (c *Client) CreateIngress(objectMeta metav1.ObjectMeta, spec extensionsv1.IngressSpec) (*extensionsv1.Ingress, error) {
	ingress := &extensionsv1.Ingress{
		ObjectMeta: objectMeta,
		Spec:       spec,
	}

	ingressObj, err := c.KubeClient.ExtensionsV1beta1().Ingresses(c.Namespace).Create(ingress)
	if err != nil {
		return nil, errors.Wrap(err, "error creating ingress")
	}
	return ingressObj, nil
}

// DeleteIngress deletes the given ingress
func (c *Client) DeleteIngress(name string) error {
	err := c.KubeClient.ExtensionsV1beta1().Ingresses(c.Namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "unable to delete ingress")
	}
	return nil
}

// ListIngresses lists all the ingresses based on the given label selector
func (c *Client) ListIngresses(labelSelector string) ([]extensionsv1.Ingress, error) {
	ingressList, err := c.KubeClient.ExtensionsV1beta1().Ingresses(c.Namespace).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ingress list")
	}

	return ingressList.Items, nil
}
