package podman

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/platform"
)

type Client interface {
	platform.Client

	// PlayKube creates the Pod with Podman
	PlayKube(pod *corev1.Pod) error

	// KubeGenerate returns a Kubernetes Pod definition of an existing Pod
	KubeGenerate(name string) (*corev1.Pod, error)

	// PodStop stops the pod with given podname
	PodStop(podname string) error

	// PodRm deletes the pod with given podname
	PodRm(podname string) error

	// PodLs lists the names of existing pods
	PodLs() (map[string]bool, error)

	// VolumeLs lists the names of existing volumes
	VolumeLs() (map[string]bool, error)

	// VolumeRm deletes the volume with given volumeName
	VolumeRm(volumeName string) error

	// CleanupPodResources stops and removes a pod and its associated resources (volumes)
	CleanupPodResources(pod *corev1.Pod, cleanVolumes bool) error

	ListAllComponents() ([]api.ComponentAbstract, error)

	GetPodUsingComponentName(componentName string) (*corev1.Pod, error)

	Version(ctx context.Context) (SystemVersionReport, error)
}
