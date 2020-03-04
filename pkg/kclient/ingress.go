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
	ingress, err := c.KubeClient.ExtensionsV1beta1().Ingresses(c.Namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "unable to get ingress")
	}
	ingressTLSArray := ingress.Spec.TLS
	for _, elem := range ingressTLSArray {
		err = c.KubeClient.CoreV1().Secrets(c.Namespace).Delete(elem.SecretName, &metav1.DeleteOptions{})
		if err != nil {
			return errors.Wrap(err, "unable to delete tls secret")
		}
	}

	err = c.KubeClient.ExtensionsV1beta1().Ingresses(c.Namespace).Delete(name, &metav1.DeleteOptions{})
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

// ListIngressNames lists all the names of the ingresses based on the given label
// selector
func (c *Client) ListIngressNames(labelSelector string) ([]string, error) {
	ingresses, err := c.ListIngresses(labelSelector)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list ingresses")
	}

	var ingressNames []string
	for _, i := range ingresses {
		ingressNames = append(ingressNames, i.Name)
	}

	return ingressNames, nil
}
