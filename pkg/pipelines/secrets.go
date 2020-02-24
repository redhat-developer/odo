package pipelines

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
)

// createOpaqueSecret creates a Kubernetes v1/Secret with the provided name and
// body, and type Opaque.
func createOpaqueSecret(name apitypes.NamespacedName, data string) (*corev1.Secret, error) {
	r := strings.NewReader(data)

	return createSecret(name, "token", corev1.SecretTypeOpaque, r)
}

// createDockerConfigSecret creates a Kubernetes v1/Secret with the provided name and
// body, and type DockerConfigJson.
func createDockerConfigSecret(name apitypes.NamespacedName, in io.Reader) (*corev1.Secret, error) {
	return createSecret(name, ".dockerconfigjson", corev1.SecretTypeDockerConfigJson, in)
}

func createSecret(name apitypes.NamespacedName, key string, st corev1.SecretType, in io.Reader) (*corev1.Secret, error) {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret data: %w", err)
	}
	secret := &corev1.Secret{
		TypeMeta:   typeMeta("Secret", "v1"),
		ObjectMeta: objectMeta(name),
		Type:       st,
		Data: map[string][]byte{
			key: data,
		},
	}
	return secret, nil
}
