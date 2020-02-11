package pipelines

import (
	"fmt"
	"io"
	"io/ioutil"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// createOpaqueSecret creates a Kubernetes v1/Secret with the provided name and
// body, and type Opaque.
func createOpaqueSecret(name string, in io.Reader) (*corev1.Secret, error) {
	return createSecret(name, "token", corev1.SecretTypeOpaque, in)
}

// createDockerConfigSecret creates a Kubernetes v1/Secret with the provided name and
// body, and type DockerConfigJson.
func createDockerConfigSecret(name string, in io.Reader) (*corev1.Secret, error) {
	return createSecret(name, ".dockerconfigjson", corev1.SecretTypeDockerConfigJson, in)
}

func createSecret(name, key string, st corev1.SecretType, in io.Reader) (*corev1.Secret, error) {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret data: %w", err)
	}
	secret := &corev1.Secret{
		TypeMeta:   createTypeMeta("Secret", "v1"),
		ObjectMeta: createObjectMeta(name),
		Type:       st,
		Data: map[string][]byte{
			key: data,
		},
	}
	return secret, nil
}

func createTypeMeta(kind, apiVersion string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       kind,
		APIVersion: apiVersion,
	}
}

func createObjectMeta(name string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name: name,
	}
}
