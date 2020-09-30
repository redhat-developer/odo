package kclient

import (
	"fmt"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// constants for volumes
const (
	PersistentVolumeClaimKind       = "PersistentVolumeClaim"
	PersistentVolumeClaimAPIVersion = "v1"
)

// CreatePVC creates a PVC resource in the cluster with the given name, size and labels
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

// DeletePVC deletes the required PVC resource from the cluster
func (c *Client) DeletePVC(pvcName string) error {
	return c.KubeClient.CoreV1().PersistentVolumeClaims(c.Namespace).Delete(pvcName, &metav1.DeleteOptions{})
}

// AddPVCAndVolumeMount adds PVC and volume mount to the pod template spec
// containerNameToVolumes is a map of the Devfile container names to the Devfile Volumes
func AddPVCAndVolumeMount(podTemplateSpec *corev1.PodTemplateSpec, uniqueStorage []common.DevfileVolume, containerNameToVolumes map[string][]common.DevfileVolume) error {

	addPVCToPodTemplateSpec(podTemplateSpec, uniqueStorage)

	err := addVolumeMountToPodTemplateSpec(podTemplateSpec, containerNameToVolumes)
	if err != nil {
		return errors.Wrapf(err, "Unable to add volumes mounts to the pod")
	}

	return nil
}

// GetPVCsFromSelector returns the PVCs based on the given selector
func (c *Client) GetPVCsFromSelector(selector string) ([]corev1.PersistentVolumeClaim, error) {
	pvcList, err := c.KubeClient.CoreV1().PersistentVolumeClaims(c.Namespace).List(metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get PVCs for selector: %v", selector)
	}

	return pvcList.Items, nil
}

// AddPVCToPodTemplateSpec adds the given PVC to the podTemplateSpec
func addPVCToPodTemplateSpec(podTemplateSpec *corev1.PodTemplateSpec, uniqueStorage []common.DevfileVolume) {
	for _, vol := range uniqueStorage {
		podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, corev1.Volume{
			Name: vol.Name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: vol.GeneratedName,
				},
			},
		})
	}
}

// AddVolumeMountToPodTemplateSpec adds the Volume Mounts in containerNameToMountPaths to the podTemplateSpec containers for a given pvc and volumeName
// containerNameToMountPaths is a map of a container name to an array of its Mount Paths
func addVolumeMountToPodTemplateSpec(podTemplateSpec *corev1.PodTemplateSpec, containerNameToVolumes map[string][]common.DevfileVolume) error {

	// Validating podTemplateSpec.Spec.Containers[] is present before dereferencing
	if len(podTemplateSpec.Spec.Containers) == 0 {
		return fmt.Errorf("podTemplateSpec %s doesn't have any Containers defined", podTemplateSpec.ObjectMeta.Name)
	}

	for _, container := range podTemplateSpec.Spec.Containers {
		vols := containerNameToVolumes[container.Name]
		for _, vol := range vols {
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      vol.Name,
				MountPath: vol.ContainerPath,
			})
		}
	}

	return nil
}
