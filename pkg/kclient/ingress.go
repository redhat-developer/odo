package kclient

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateIngress creates an ingress object for the given service and with the given labels
func (c *Client) CreateIngress(ingress extensionsv1.Ingress) (*extensionsv1.Ingress, error) {
	if ingress.GetName() == "" {
		return nil, fmt.Errorf("ingress name is empty")
	}
	ingressObj, err := c.KubeClient.ExtensionsV1beta1().Ingresses(c.Namespace).Create(context.TODO(), &ingress, metav1.CreateOptions{FieldManager: FieldManager})
	if err != nil {
		return nil, errors.Wrap(err, "error creating ingress")
	}
	return ingressObj, nil
}

// DeleteIngress deletes the given ingress
func (c *Client) DeleteIngress(name string) error {
	err := c.KubeClient.ExtensionsV1beta1().Ingresses(c.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrap(err, "unable to delete ingress")
	}
	return nil
}

// ListIngresses lists all the ingresses based on the given label selector
func (c *Client) ListIngresses(labelSelector string) ([]extensionsv1.Ingress, error) {
	ingressList, err := c.KubeClient.ExtensionsV1beta1().Ingresses(c.Namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ingress list")
	}

	return ingressList.Items, nil
}

// GetOneIngressFromSelector gets one ingress with the given selector
// if no or multiple ingresses are found with the given selector, it throws an error
func (c *Client) GetOneIngressFromSelector(selector string) (*extensionsv1.Ingress, error) {
	ingresses, err := c.ListIngresses(selector)
	if err != nil {
		return nil, err
	}

	num := len(ingresses)
	if num == 0 {
		return nil, fmt.Errorf("no ingress was found for the selector: %v", selector)
	} else if num > 1 {
		return nil, fmt.Errorf("multiple ingresses exist for the selector: %v. Only one must be present", selector)
	}

	return &ingresses[0], nil
}

// GetIngress gets an ingress based on the given name
func (c *Client) GetIngress(name string) (*extensionsv1.Ingress, error) {
	ingress, err := c.KubeClient.ExtensionsV1beta1().Ingresses(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	return ingress, err
}
