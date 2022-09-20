package podman

import corev1 "k8s.io/api/core/v1"

type Client interface {
	// PlayKube creates the Pod with Podman
	PlayKube(pod *corev1.Pod) error

	// PodStop stops the pod with given podname
	PodStop(podname string) error

	// PodRm deletes the pod with given podname
	PodRm(podname string) error

	// VolumeRm deletes the volume with given volumeName
	VolumeRm(volumeName string) error
}
