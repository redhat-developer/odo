package platform

import (
	"io"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Client is the interface that wraps operations that can be performed on any supported platform.
type Client interface {

	// ExecCMDInContainer executes the specified command in the container of a pod.
	// If an empty string is passed as container name, the command will be executed in the first container found in the pod.
	ExecCMDInContainer(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error

	// GetPodLogs returns the logs of the specified pod container.
	// All logs for all containers part of the pod are returned if an empty string is provided as container name.
	GetPodLogs(podName, containerName string, followLog bool) (io.ReadCloser, error)

	// GetPodsMatchingSelector returns all pods matching the given label selector.
	GetPodsMatchingSelector(selector string) (*corev1.PodList, error)

	// GetAllResourcesFromSelector returns all resources of any kind matching the given label selector.
	GetAllResourcesFromSelector(selector string, ns string) ([]unstructured.Unstructured, error)

	// GetAllPodsInNamespaceMatchingSelector returns all pods matching the given label selector and in the specified namespace.
	GetAllPodsInNamespaceMatchingSelector(selector string, ns string) (*corev1.PodList, error)

	// GetRunningPodFromSelector returns any pod matching the given label selector.
	// If multiple pods are found, implementations might have different behavior, by either returning an error or returning any element.
	GetRunningPodFromSelector(selector string) (*corev1.Pod, error)
}
