package delete

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

type Client interface {
	ListResourcesToDelete(name string, namespace string) ([]unstructured.Unstructured, error)
	DeleteResources([]unstructured.Unstructured) error
}
