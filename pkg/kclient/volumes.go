package kclient

import (
	"fmt"

	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// constants for volumes
const (
	PersistentVolumeClaimKind       = "PersistentVolumeClaim"
	PersistentVolumeClaimAPIVersion = "v1"

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

// AddPVCToPodTemplateSpec adds the given PVC to the podTemplateSpec
func AddPVCToPodTemplateSpec(podTemplateSpec *corev1.PodTemplateSpec, pvc, volumeName string) {

	podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvc,
			},
		},
	})
}

// AddVolumeMountToPodTemplateSpec adds the Volume Mounts to the podTemplateSpec containers for a given PVC
func AddVolumeMountToPodTemplateSpec(podTemplateSpec *corev1.PodTemplateSpec, volumeName, pvc string, containerMountPathsMap map[string][]string) error {

	// Validating podTemplateSpec.Spec.Containers[] is present before dereferencing
	if len(podTemplateSpec.Spec.Containers) == 0 {
		return fmt.Errorf("podTemplateSpec %s doesn't have any Containers defined", podTemplateSpec.ObjectMeta.Name)
	}

	for containerName, mountPaths := range containerMountPathsMap {
		for i, container := range podTemplateSpec.Spec.Containers {
			if container.Name == containerName {
				for _, mountPath := range mountPaths {
					container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
						Name:      volumeName,
						MountPath: mountPath,
						SubPath:   "",
					},
					)
				}
				podTemplateSpec.Spec.Containers[i] = container
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
