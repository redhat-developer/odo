package pipelines

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func TestCreateNamespace(t *testing.T) {
	ns := CreateNamespace("test-environment")
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

func TestNamespaceNames(t *testing.T) {
	ns := NamespaceNames("test-")
	want := map[string]string{
		"dev":   "test-dev",
		"stage": "test-stage",
		"cicd":  "test-cicd",
	}
	if diff := cmp.Diff(want, ns); diff != "" {
		t.Fatalf("namespaceNames() failed got\n%s", diff)
	}
}

func TestCreateNamespaces(t *testing.T) {
	ns := CreateNamespaces([]string{
		"test-dev",
		"test-stage",
		"test-cicd",
	})
	want := []*corev1.Namespace{
		CreateNamespace("test-dev"),
		CreateNamespace("test-stage"),
		CreateNamespace("test-cicd"),
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
	validNamespace := CreateNamespace("sample")
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			cs := testclient.NewSimpleClientset(validNamespace)
			namespaceExists, _ := CheckNamespace(cs, test.namespace)
			if diff := cmp.Diff(namespaceExists, test.valid); diff != "" {
				t.Fatalf("checkNamespace() failed:\n%v", diff)
			}
		})
	}
}
