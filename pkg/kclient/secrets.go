package kclient

import (
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateTLSSecret creates a TLS Secret with the given certificate and private key
// serviceName is the name of the service for the target reference
// ingressDomain is the ingress domain to use for the ingress
func (c *Client) CreateTLSSecret(tlsCertificate []byte, tlsPrivKey []byte, objectMeta metav1.ObjectMeta) (*corev1.Secret, error) {
	if objectMeta.Name == "" {
		return nil, fmt.Errorf("tlsSecret name is empty")
	}
	data := make(map[string][]byte)
	data["tls.crt"] = tlsCertificate
	data["tls.key"] = tlsPrivKey
	secretTemplate := corev1.Secret{
		ObjectMeta: objectMeta,
		Type:       corev1.SecretTypeTLS,
		Data:       data,
	}

	secret, err := c.KubeClient.CoreV1().Secrets(c.Namespace).Create(&secretTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create secret %s", objectMeta.Name)
	}
	return secret, nil
}
