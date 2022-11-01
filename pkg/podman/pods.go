package podman

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// GetPodsMatchingSelector returns all pods matching the given label selector.
func (o *PodmanCli) GetPodsMatchingSelector(selector string) (*corev1.PodList, error) {
	return nil, nil
}

// GetAllResourcesFromSelector returns all resources of any kind matching the given label selector.
func (o *PodmanCli) GetAllResourcesFromSelector(selector string, ns string) ([]unstructured.Unstructured, error) {
	return nil, nil
}

// GetAllPodsInNamespaceMatchingSelector returns all pods matching the given label selector and in the specified namespace.
func (o *PodmanCli) GetAllPodsInNamespaceMatchingSelector(selector string, ns string) (*corev1.PodList, error) {
	return nil, nil
}

// GetRunningPodFromSelector returns any pod matching the given label selector.
// If multiple pods are found, implementations might have different behavior, by either returning an error or returning any element.
func (o *PodmanCli) GetRunningPodFromSelector(selector string) (*corev1.Pod, error) {
	return nil, nil
}
