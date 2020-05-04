package namespaces

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func TestCreate(t *testing.T) {
	ns := Create("test-environment")
	want := &corev1.Namespace{
		TypeMeta: namespaceTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-environment",
		},
	}

	if diff := cmp.Diff(want, ns); diff != "" {
		t.Fatalf("createNamespace() failed got\n%s", diff)
	}
}

func TestNamesWithPrefix(t *testing.T) {
	ns := NamesWithPrefix("test-")
	want := map[string]string{
		"dev":   "test-dev",
		"stage": "test-stage",
		"cicd":  "test-cicd",
	}
	if diff := cmp.Diff(want, ns); diff != "" {
		t.Fatalf("NamesWithPrefix() failed got\n%s", diff)
	}
}

func TestNamespaces(t *testing.T) {
	ns := Namespaces([]string{
		"test-dev",
		"test-stage",
		"test-cicd",
	})
	want := []*corev1.Namespace{
		Create("test-dev"),
		Create("test-stage"),
		Create("test-cicd"),
	}
	if diff := cmp.Diff(want, ns); diff != "" {
		t.Fatalf("Namespaces() failed got\n%s", diff)
	}
}

func TestExists(t *testing.T) {
	tests := []struct {
		desc      string
		namespace string
		valid     bool
	}{
		{
			"Namespace sample already exists",
			"sample",
			true,
		},
		{
			"Namespace test doesn't exist",
			"test",
			false,
		},
	}
	validNamespace := Create("sample")
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			cs := testclient.NewSimpleClientset(validNamespace)
			namespaceExists, _ := Exists(cs, test.namespace)
			if diff := cmp.Diff(namespaceExists, test.valid); diff != "" {
				t.Fatalf("Exists() failed:\n%v", diff)
			}
		})
	}
}
