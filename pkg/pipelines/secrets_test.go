package pipelines

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateOpaqueSecret(t *testing.T) {
	data := []byte(`abcdefghijklmnop`)
	secret, err := CreateOpaqueSecret("github-auth", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	want := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "github-auth",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"token": data,
		},
	}

	if diff := cmp.Diff(want, secret); diff != "" {
		t.Fatalf("CreateOpaqueSecret() failed got\n%s", diff)
	}
}
