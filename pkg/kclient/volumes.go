package kclient

import (
	"fmt"

	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// The length of the string to be generated for names of resources
	nameLength = 5
)

// CreatePVC creates a PVC resource in the cluster with the given name, size and
// labels
func (c *Client) CreatePVC(objectMeta metav1.ObjectMeta, pvcSpec corev1.PersistentVolumeClaimSpec) (*corev1.PersistentVolumeClaim, error) {

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: objectMeta,
		Spec:       pvcSpec,
	}

	createdPvc, err := c.KubeClient.CoreV1().PersistentVolumeClaims(c.Namespace).Create(pvc)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create PVC")
	}
	return createdPvc, nil
}

// AddPVCToPodSpec adds the given PVC to the given pod
func AddPVCToPodSpec(pod *corev1.Pod, pvc, volumeName string) {

	pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvc,
			},
		},
	})
}

// AddVolumeMountToPodContainerSpec adds the Volume Mounts for the containers for a given PVC
func AddVolumeMountToPodContainerSpec(pod *corev1.Pod, volumeName, pvc string, containerMountPathsMap map[string][]string) error {

	// Validating pod.Spec.Containers[] is present before dereferencing
	if len(pod.Spec.Containers) == 0 {
		return fmt.Errorf("Pod %s doesn't have any Containers defined", pod.Name)
	}

	for containerName, mountPaths := range containerMountPathsMap {
		for i, container := range pod.Spec.Containers {
			if container.Name == containerName {
				for _, mountPath := range mountPaths {
					container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
						Name:      volumeName,
						MountPath: mountPath,
						SubPath:   "",
					},
					)
				}
				pod.Spec.Containers[i] = container
			}
		}
	}

	return nil
}

// generateVolumeNameFromPVC generates a random volume name based on the name
// of the given PVC
func generateVolumeNameFromPVC(pvc string) string {
	return fmt.Sprintf("%v-%v-volume", pvc, util.GenerateRandomString(nameLength))
}
