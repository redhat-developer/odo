package pipelines

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func TestCreateNamespace(t *testing.T) {
	ns := createNamespace("test-environment")
	want := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-environment",
		},
	}

	if diff := cmp.Diff(want, ns); diff != "" {
		t.Fatalf("createNamespace() failed got\n%s", diff)
	}
}

func TestNamespaceNames(t *testing.T) {
	ns := namespaceNames("test-")
	want := map[string]string{
		"dev":   "test-dev-environment",
		"stage": "test-stage-environment",
		"cicd":  "test-cicd-environment",
	}
	if diff := cmp.Diff(want, ns); diff != "" {
		t.Fatalf("namespaceNames() failed got\n%s", diff)
	}
}

func TestCreateNamespaces(t *testing.T) {
	ns := createNamespaces([]string{
		"test-dev-environment",
		"test-stage-environment",
		"test-cicd-environment",
	})
	want := []*corev1.Namespace{
		createNamespace("test-dev-environment"),
		createNamespace("test-stage-environment"),
		createNamespace("test-cicd-environment"),
	}
	if diff := cmp.Diff(want, ns); diff != "" {
		t.Fatalf("createNamespaces() failed got\n%s", diff)
	}
}

func TestCheckNamespace(t *testing.T) {
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
	validNamespace := createNamespace("sample")
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			cs := testclient.NewSimpleClientset(validNamespace)
			namespaceExists, _ := checkNamespace(cs, test.namespace)
			if diff := cmp.Diff(namespaceExists, test.valid); diff != "" {
				t.Fatalf("checkNamespace() failed:\n%v", diff)
			}
		})
	}
}
