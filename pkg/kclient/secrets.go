package kclient

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

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

// SelfSignedCertificate struct is the return type of function GenerateSelfSignedCertificate
// CertPem is the byte array for certificate pem encode
// KeyPem is the byte array for key pem encode
type SelfSignedCertificate struct {
	CertPem []byte
	KeyPem  []byte
}

// GenerateSelfSignedCertificate creates a self-signed SSl certificate
func GenerateSelfSignedCertificate(host string) (SelfSignedCertificate, error) {

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return SelfSignedCertificate{}, errors.Wrap(err, "unable to generate rsa key")
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:   "Odo self-signed certificate",
			Organization: []string{"Odo"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * 365 * 10),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		DNSNames:              []string{"*." + host},
	}

	certificateDerEncoding, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return SelfSignedCertificate{}, errors.Wrap(err, "unable to create certificate")
	}
	out := &bytes.Buffer{}
	err = pem.Encode(out, &pem.Block{Type: "CERTIFICATE", Bytes: certificateDerEncoding})
	if err != nil {
		return SelfSignedCertificate{}, errors.Wrap(err, "unable to encode certificate")
	}
	certPemEncode := out.String()
	certPemByteArr := []byte(certPemEncode)

	tlsPrivKeyEncoding := x509.MarshalPKCS1PrivateKey(privateKey)
	err = pem.Encode(out, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: tlsPrivKeyEncoding})
	if err != nil {
		return SelfSignedCertificate{}, errors.Wrap(err, "unable to encode rsa private key")
	}
	keyPemEncode := out.String()
	keyPemByteArr := []byte(keyPemEncode)

	return SelfSignedCertificate{CertPem: certPemByteArr, KeyPem: keyPemByteArr}, nil
}
