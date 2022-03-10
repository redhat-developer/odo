package delete

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

type Client interface {
	// ListResourcesToDelete lists Kubernetes resources from cluster in namespace for a given odo component
	ListResourcesToDelete(componentName string, namespace string) ([]unstructured.Unstructured, error)
	// DeleteResources deletes the unstuctured resources and return the resources that failed to be deleted
	DeleteResources([]unstructured.Unstructured) []unstructured.Unstructured
}
