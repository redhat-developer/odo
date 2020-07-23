package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func SecretMock(ns, name string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
		Data: map[string][]byte{
			"user":     []byte("user"),
			"password": []byte("password"),
		},
	}
}

func TestUnstructuredToUnstructured(t *testing.T) {
	ns := "ns"
	name := "name"

	secret := SecretMock(ns, name)

	u, err := ToUnstructured(secret)
	assert.NoError(t, err)
	t.Logf("Unstructured: '%#v'", u)

	assert.Equal(t, ns, u.GetNamespace())
	assert.Equal(t, name, u.GetName())
}

func TestUnstructuredToUnstructuredAsGVK(t *testing.T) {
	ns := "ns"
	name := "name"

	secret := SecretMock(ns, name)
	gvk := schema.GroupVersion{Group: "", Version: "v1"}.WithKind("Secret")

	u, err := ToUnstructuredAsGVK(secret, gvk)
	assert.NoError(t, err)
	t.Logf("Unstructured: '%#v'", u)

	assert.Equal(t, "Secret", u.GetKind())
	assert.Equal(t, "v1", u.GetAPIVersion())
}
