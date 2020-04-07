package kclient

import (
	"fmt"

	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateTLSSecret creates a TLS Secret with the given certificate and private key
// serviceName is the name of the service for the target reference
// ingressDomain is the ingress domain to use for the ingress
func (c *Client) CreateTLSSecret(tlsCertificate []byte, tlsPrivKey []byte, componentName string, applicationName string) (*corev1.Secret, error) {
	if componentName == "" {
		return nil, fmt.Errorf("componentName name is empty")
	}
	deployment, err := c.GetDeploymentByName(componentName)
	if err != nil {
		return nil, err
	}
	labels := componentlabels.GetLabels(componentName, applicationName, true)
	tlsSecretName := componentName + "-tlssecret"
	data := make(map[string][]byte)
	data["tls.crt"] = tlsCertificate
	data["tls.key"] = tlsPrivKey
	secretTemplate := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   tlsSecretName,
			Labels: labels,
			OwnerReferences: []v1.OwnerReference{
				GenerateOwnerReference(deployment),
			},
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
