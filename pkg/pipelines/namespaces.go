package pipelines

import (
	"fmt"

	"github.com/openshift/odo/pkg/pipelines/meta"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var namespaceBaseNames = map[string]string{
	"dev":   "dev-environment",
	"stage": "stage-environment",
	"cicd":  "cicd-environment",
}

func createNamespaces(names []string) []*corev1.Namespace {
	ns := []*corev1.Namespace{}
	for _, n := range names {
		ns = append(ns, createNamespace(n))
	}
	return ns
}

func namespaceNames(prefix string) map[string]string {
	prefixedNames := make(map[string]string)
	for k, v := range namespaceBaseNames {
		prefixedNames[k] = fmt.Sprintf("%s%s", prefix, v)
	}
	return prefixedNames
}

func createNamespace(name string) *corev1.Namespace {
	ns := &corev1.Namespace{
		TypeMeta: meta.TypeMeta("Namespace", "v1"),
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return ns
}
