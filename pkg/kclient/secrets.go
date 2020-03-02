package kclient

import (
	"fmt"

	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
