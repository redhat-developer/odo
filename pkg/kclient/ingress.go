package kclient

import (
	"context"
	"fmt"

	"github.com/redhat-developer/odo/pkg/unions"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetOneIngressFromSelector gets one ingress with the given selector
// if no or multiple ingresses are found with the given selector, it throws an error
func (c *Client) GetOneIngressFromSelector(selector string) (*unions.KubernetesIngress, error) {
	ingresses, err := c.ListIngresses(selector)
	if err != nil {
		return nil, err
	}

	if num := len(ingresses.Items); num == 0 {
		return nil, fmt.Errorf("no ingress was found for the selector: %v", selector)
	} else if num > 1 {
		return nil, fmt.Errorf("multiple ingresses exist for the selector: %v. Only one must be present", selector)
	}

	return ingresses.Items[0], nil
}

//CreateIngress creates a specified Kubernetes Ingress as a networking v1 or extensions v1beta1 ingress depending on what
//is supported, with preference for networking v1 ingress. The passed ingress MUST be a generated one, i.e it must
//have been created using unions.NewKubernetesIngressFromParams
func (c *Client) CreateIngress(ingress unions.KubernetesIngress) (*unions.KubernetesIngress, error) {
	var err error
	if !ingress.IsGenerated() {
		return nil, fmt.Errorf("create ingress should get a generated ingress. If you are hiting this, contact the developer")
	}
	if ingress.GetName() == "" {
		return nil, fmt.Errorf("cannot create an ingress without a name")
	}
	err = c.checkIngressSupport()
	if err != nil {
		return nil, err
	}
	created := false
	kubernetesIngress := unions.NewNonGeneratedKubernetesIngress()
	if c.isNetworkingV1IngressSupported {
		kubernetesIngress.NetworkingV1Ingress, err = c.KubeClient.NetworkingV1().Ingresses(c.Namespace).Create(context.TODO(), ingress.NetworkingV1Ingress, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create networking v1 ingress %w", err)
		}
		created = true
	} else if c.isExtensionV1Beta1IngressSupported {
		kubernetesIngress.ExtensionV1Beta1Ingress, err = c.KubeClient.ExtensionsV1beta1().Ingresses(c.Namespace).Create(context.TODO(), ingress.ExtensionV1Beta1Ingress, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create extension v1 beta 1 ingress %w", err)
		}
		created = true
	}
	if !created {
		return nil, fmt.Errorf("failed to create ingress as none of the options are supported")
	}
	return kubernetesIngress, nil
}

func (c *Client) DeleteIngress(name string) error {
	var err error
	err = c.checkIngressSupport()
	if err != nil {
		return err
	}
	if c.isNetworkingV1IngressSupported {
		err = c.KubeClient.NetworkingV1().Ingresses(c.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete networking v1 ingress %w", err)
		}
	} else if c.isExtensionV1Beta1IngressSupported {
		err = c.KubeClient.ExtensionsV1beta1().Ingresses(c.Namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete extensionv v1beta1 ingress %w", err)
		}
	}
	return nil
}

//ListIngresses lists all the ingresses based on given label selector
func (c *Client) ListIngresses(labelSelector string) (*unions.KubernetesIngressList, error) {
	kubernetesIngressList := unions.NewEmptyKubernetesIngressList()
	err := c.checkIngressSupport()
	if err != nil {
		return nil, err
	}
	// if networking v1 ingress is supported then extension v1 ingress are automatically wrapped
	// by net v1 ingresses
	if c.isNetworkingV1IngressSupported {
		ingressList, err := c.KubeClient.NetworkingV1().Ingresses(c.Namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to list ingresses as networking v1 ingresses %w", err)
		} else {
			for k := range ingressList.Items {
				ki := unions.NewNonGeneratedKubernetesIngress()
				ki.NetworkingV1Ingress = &ingressList.Items[k]
				kubernetesIngressList.Items = append(kubernetesIngressList.Items, ki)
			}
		}
		return kubernetesIngressList, nil
	} else if c.isExtensionV1Beta1IngressSupported {
		ingressList, err := c.KubeClient.ExtensionsV1beta1().Ingresses(c.Namespace).List(context.TODO(), metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to list ingresses as extensions v1 ingress %w", err)
		} else {
			for k := range ingressList.Items {
				ki := unions.NewNonGeneratedKubernetesIngress()
				ki.ExtensionV1Beta1Ingress = &ingressList.Items[k]
				kubernetesIngressList.Items = append(kubernetesIngressList.Items, ki)
			}
		}
		return kubernetesIngressList, nil
	}
	return kubernetesIngressList, fmt.Errorf("ingresses on cluster are not supported")
}

func (c *Client) GetIngress(name string) (*unions.KubernetesIngress, error) {
	ki := unions.NewNonGeneratedKubernetesIngress()
	err := c.checkIngressSupport()
	if err != nil {
		return nil, err
	}
	if c.isNetworkingV1IngressSupported {
		ki.NetworkingV1Ingress, err = c.KubeClient.NetworkingV1().Ingresses(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		return ki, nil
	}
	if c.isExtensionV1Beta1IngressSupported {
		ki.ExtensionV1Beta1Ingress, err = c.KubeClient.ExtensionsV1beta1().Ingresses(c.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		return ki, nil
	}
	return nil, fmt.Errorf("could not get supported type of ingress")
}
