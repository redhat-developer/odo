package kclient

import (
	"fmt"

	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// CreateTLSSecret creates a TLS Secret with the given certificate and private key
// serviceName is the name of the service for the target reference
// ingressDomain is the ingress domain to use for the ingress
// portNumber is the target port of the ingress
func (c *Client) CreateTLSSecret(tlsCertificate []byte, tlsPrivKey []byte, componentName string, applicationName string, portNumber int) (*corev1.Secret, error) {
	labels := componentlabels.GetLabels(componentName, applicationName, true)
	portAsString := fmt.Sprintf("%v", portNumber)
	tlsSecretName := componentName + "-" + portAsString + "-tlssecret"
	data := make(map[string][]byte)
	data["tls.crt"] = tlsCertificate
	data["tls.key"] = tlsPrivKey
	secretTemplate := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   tlsSecretName,
			Labels: labels,
		},
		Type: corev1.SecretTypeTLS,
		Data: data,
	}

	secret, err := c.KubeClient.CoreV1().Secrets(c.Namespace).Create(&secretTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create secret %s", tlsSecretName)
	}
	return secret, nil
}

// IngressParamater struct for function createIngress
// Name is the Ingress name
// serviceName is the name of the service for the target reference
// ingressDomain is the ingress domain to use for the ingress
// portNumber is the target port of the ingress
// TLSSecretName is the target TLS Secret name of the ingress
type IngressParamater struct {
	Name          string
	ServiceName   string
	IngressDomain string
	PortNumber    intstr.IntOrString
	TLSSecretName string
}

// CreateIngress creates an ingress object for the given service and with the given labels
func (c *Client) CreateIngress(ingressParam IngressParamater, labels map[string]string) (*extensionsv1.Ingress, error) {
	ingress := &extensionsv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ingressParam.Name,
			Labels: labels,
		},
		Spec: extensionsv1.IngressSpec{
			Rules: []extensionsv1.IngressRule{
				{
					Host: ingressParam.IngressDomain,
					IngressRuleValue: extensionsv1.IngressRuleValue{
						HTTP: &extensionsv1.HTTPIngressRuleValue{
							Paths: []extensionsv1.HTTPIngressPath{
								{
									Path: "/",
									Backend: extensionsv1.IngressBackend{
										ServiceName: ingressParam.ServiceName,
										ServicePort: ingressParam.PortNumber,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	secretNameLength := len(ingressParam.TLSSecretName)
	if secretNameLength != 0 {
		ingress.Spec.TLS = []extensionsv1.IngressTLS{
			{
				Hosts: []string{
					ingressParam.IngressDomain,
				},
				SecretName: ingressParam.TLSSecretName,
			},
		}
	}

	r, err := c.KubeClient.ExtensionsV1beta1().Ingresses(c.Namespace).Create(ingress)
	if err != nil {
		return nil, errors.Wrap(err, "error creating ingress")
	}
	return r, nil
}

// DeleteIngress deleted the given route
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

// ListSecrets lists all the secrets based on the given label selector
func (c *Client) ListSecrets(labelSelector string) ([]corev1.Secret, error) {
	listOptions := metav1.ListOptions{}
	if len(labelSelector) > 0 {
		listOptions = metav1.ListOptions{
			LabelSelector: labelSelector,
		}
	}

	secretList, err := c.KubeClient.CoreV1().Secrets(c.Namespace).List(listOptions)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get secret list")
	}

	return secretList.Items, nil
}
