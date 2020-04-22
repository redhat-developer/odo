package secrets

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/openshift/client-go/route/clientset/versioned/scheme"
	"github.com/openshift/odo/pkg/manifest/clientconfig"
	"github.com/openshift/odo/pkg/manifest/meta"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/util/cert"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
)

var (
	secretTypeMeta       = meta.TypeMeta("Secret", "v1")
	sealedSecretTypeMeta = meta.TypeMeta("SealedSecret", "bitnami.com/v1alpha1")
)

// DefaultPublicKeyFunc is the func used to get the key from Bitnami.
var DefaultPublicKeyFunc = getClusterPublicKey

type PublicKeyFunc func() (*rsa.PublicKey, error)

// CreateSealedDockerConfigSecret creates a SealedSecret with the given name and reader
func CreateSealedDockerConfigSecret(name types.NamespacedName, in io.Reader) (*ssv1alpha1.SealedSecret, error) {
	secret, err := createDockerConfigSecret(name, in)
	if err != nil {
		return nil, err
	}

	return seal(secret, DefaultPublicKeyFunc)
}

// CreateSealedSecret creates a SealedSecret with the provided name and body/data and type
func CreateSealedSecret(name types.NamespacedName, data, secretKey string) (*ssv1alpha1.SealedSecret, error) {
	secret, err := createOpaqueSecret(name, data, secretKey)
	if err != nil {
		return nil, err
	}

	return seal(secret, DefaultPublicKeyFunc)
}

// Returns a sealed secret
func seal(secret *corev1.Secret, pubKey PublicKeyFunc) (*ssv1alpha1.SealedSecret, error) {
	// Strip read-only server-side ObjectMeta (if present)
	secret.SetSelfLink("")
	secret.SetUID("")
	secret.SetResourceVersion("")
	secret.Generation = 0
	secret.SetCreationTimestamp(metav1.Time{})
	secret.SetDeletionTimestamp(nil)
	secret.DeletionGracePeriodSeconds = nil

	key, err := pubKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get public key from cluster (is sealed-secrets installed?): %w", err)
	}

	sealedSecret, err := ssv1alpha1.NewSealedSecret(scheme.Codecs, key, secret)
	if err != nil {
		return nil, err
	}

	// NewSealedSecret() doesn't add TypeMeta to SealedSecret
	sealedSecret.TypeMeta = sealedSecretTypeMeta
	return sealedSecret, err
}

// Retrieves a public key from sealed-secrets-controller
func getClusterPublicKey() (*rsa.PublicKey, error) {
	client, err := getRESTClient()
	if err != nil {
		return nil, err
	}

	f, err := openCertCluster(client)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseKey(f)
}

// Returns a reader of public key from sealed-secrets-controller
func openCertCluster(c clientv1.CoreV1Interface) (io.ReadCloser, error) {
	f, err := c.
		Services("kube-system").
		ProxyGet("http", "sealed-secrets-controller", "", "/v1/cert.pem", nil).
		Stream()
	if err != nil {
		return nil, fmt.Errorf("cannot fetch certificate: %v", err)
	}
	return f, nil
}

// Reads and parses a public key from a reader
func parseKey(r io.Reader) (*rsa.PublicKey, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	certs, err := cert.ParseCertsPEM(data)
	if err != nil {
		return nil, err
	}

	// ParseCertsPem returns error if len(certs) == 0, but best to be sure...
	if len(certs) == 0 {
		return nil, errors.New("Failed to read any certificates")
	}

	cert, ok := certs[0].PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("Expected RSA public key but found %v", certs[0].PublicKey)
	}

	return cert, nil
}

// Gets a REST client
func getRESTClient() (*clientv1.CoreV1Client, error) {
	config, err := clientconfig.GetRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get client config due to %w", err)
	}

	config.AcceptContentTypes = "application/x-pem-file, */*"
	return clientv1.NewForConfig(config)
}

// createOpaqueSecret creates a Kubernetes v1/Secret with the provided name and
// body, and type Opaque.
func createOpaqueSecret(name types.NamespacedName, data, secretKey string) (*corev1.Secret, error) {
	r := strings.NewReader(data)

	return createSecret(name, secretKey, corev1.SecretTypeOpaque, r)
}

// createDockerConfigSecret creates a Kubernetes v1/Secret with the provided name and
// body, and type DockerConfigJson.
func createDockerConfigSecret(name types.NamespacedName, in io.Reader) (*corev1.Secret, error) {
	return createSecret(name, ".dockerconfigjson", corev1.SecretTypeDockerConfigJson, in)
}

func createSecret(name types.NamespacedName, key string, st corev1.SecretType, in io.Reader) (*corev1.Secret, error) {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret data: %w", err)
	}
	secret := &corev1.Secret{
		TypeMeta:   secretTypeMeta,
		ObjectMeta: meta.ObjectMeta(name),
		Type:       st,
		Data: map[string][]byte{
			key: data,
		},
	}
	return secret, nil
}
