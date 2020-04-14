package manifest

import (
	"fmt"

	"github.com/openshift/odo/pkg/manifest/clientconfig"
	"github.com/openshift/odo/pkg/manifest/meta"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	namespaceBaseNames = map[string]string{
		"dev":   "dev-environment",
		"stage": "stage-environment",
		"cicd":  "cicd-environment",
	}

	namespaceTypeMeta = meta.TypeMeta("Namespace", "v1")
)

// CreateNamespaces create namespaces for the given names
func CreateNamespaces(names []string) []*corev1.Namespace {
	ns := []*corev1.Namespace{}
	for _, n := range names {
		ns = append(ns, CreateNamespace(n))
	}
	return ns
}

// NamespaceNames returns namespaces of all environments
func NamespaceNames(prefix string) map[string]string {
	prefixedNames := make(map[string]string)
	for k, v := range namespaceBaseNames {
		prefixedNames[k] = fmt.Sprintf("%s%s", prefix, v)
	}
	return prefixedNames
}

// CreateNamespace creates a Namespace object from a string
func CreateNamespace(name string) *corev1.Namespace {
	ns := &corev1.Namespace{
		TypeMeta: namespaceTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return ns
}

// GetClientSet returns clientset
func GetClientSet() (*kubernetes.Clientset, error) {
	clientConfig, err := clientconfig.GetRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get client config due to %w", err)
	}
	clientSet, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get APIs client due to %w", err)
	}
	return clientSet, nil
}

// CheckNamespace returns true if the given namespace exists
func CheckNamespace(clientSet kubernetes.Interface, name string) (bool, error) {
	_, err := clientSet.CoreV1().Namespaces().Get(name, metav1.GetOptions{})
	if err != nil {
		return false, nil
	}
	return true, nil
}

// AddPrefix add namespace prefix
func AddPrefix(prefix, name string) string {
	if prefix != "" {
		return prefix + name
	}
	return name
}
