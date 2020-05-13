package conditions

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	watchtools "k8s.io/client-go/tools/watch"
	krun "k8s.io/kubectl/pkg/cmd/run"
)

// ErrContainerTerminated is returned by PodContainerRunning in the intermediate
// state where the pod indicates it's still running, but its container is already terminated
var ErrContainerTerminated = fmt.Errorf("container terminated")
var ErrNonZeroExitCode = fmt.Errorf("non-zero exit code from debug container")

// PodContainerRunning returns false until the named container has ContainerStatus running (at least once),
// and will return an error if the pod is deleted, runs to completion, or the container pod is not available.
func PodContainerRunning(containerName string, coreClient corev1client.CoreV1Interface) watchtools.ConditionFunc {
	return func(event watch.Event) (bool, error) {
		switch event.Type {
		case watch.Deleted:
			return false, errors.NewNotFound(schema.GroupResource{Resource: "pods"}, "")
		}
		switch t := event.Object.(type) {
		case *corev1.Pod:
			switch t.Status.Phase {
			case corev1.PodRunning, corev1.PodPending:
				for _, s := range t.Status.ContainerStatuses {
					if s.State.Waiting != nil {
						// Return error here if pod is pending and container status indicates a failure
						// otherwise, user would have to wait the timeout period (15 min)
						// for pod to exit.
						if s.State.Waiting.Reason == "CreateContainerError" || s.State.Waiting.Reason == "ImagePullBackOff" {
							return false, fmt.Errorf(s.State.Waiting.Message)
						}
					}
				}
			case corev1.PodFailed, corev1.PodSucceeded:
				for _, s := range t.Status.ContainerStatuses {
					if s.State.Terminated != nil {
						exitCode := s.State.Terminated.ExitCode
						if exitCode != 0 {
							// User will get more information about non-zero exit code from pod logs retrieval
							// in debug.go.  Here we mark the non-zero exit to separate success logs
							// from failed container logs.
							return false, ErrNonZeroExitCode
						}
					}
				}
				return false, krun.ErrPodCompleted
			default:
				return false, nil
			}
			for _, s := range t.Status.ContainerStatuses {
				if s.Name != containerName {
					continue
				}
				if s.State.Terminated != nil {
					return false, ErrContainerTerminated
				}
				return s.State.Running != nil, nil
			}
			for _, s := range t.Status.InitContainerStatuses {
				if s.Name != containerName {
					continue
				}
				if s.State.Terminated != nil {
					return false, ErrContainerTerminated
				}
				return s.State.Running != nil, nil
			}
			return false, nil
		}
		return false, nil
	}
}
