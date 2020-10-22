package kclient

import (
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/util"
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

// GetPVCVol gets a pvc type volume with the given volume name and pvc name
func GetPVCVol(volumeName, pvcName string) corev1.Volume {

	return corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvcName,
			},
		},
	}
}

// AddVolumeMountToContainers adds the Volume Mounts in containerNameToMountPaths to the podTemplateSpec containers for a given pvc and volumeName
// containerNameToMountPaths is a map of a container name to an array of its Mount Paths
func AddVolumeMountToContainers(containers []corev1.Container, volumeName string, containerNameToMountPaths map[string][]string) []corev1.Container {

	// // Validating podTemplateSpec.Spec.Containers[] is present before dereferencing
	// if len(podTemplateSpec.Spec.Containers) == 0 {
	// 	return fmt.Errorf("podTemplateSpec %s doesn't have any Containers defined", podTemplateSpec.ObjectMeta.Name)
	// }

	for containerName, mountPaths := range containerNameToMountPaths {
		for i, container := range containers {
			if container.Name == containerName {
				for _, mountPath := range mountPaths {
					container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
						Name:      volumeName,
						MountPath: mountPath,
						SubPath:   "",
					},
					)
				}
				containers[i] = container
			}
		}
	}

	return containers
}

// GetPVCVolAndVolMount adds PVC and volume mount to the pod template spec
// volumeNameToPVCName is a map of volume name to the PVC created
// containerNameToVolumes is a map of the Devfile container names to the Devfile Volumes
func GetPVCVolAndVolMount(containers []corev1.Container, volumeNameToPVCName map[string]string, containerNameToVolumes map[string][]common.DevfileVolume) ([]corev1.Container, []corev1.Volume, error) {
	var pvcVols []corev1.Volume
	for volName, pvcName := range volumeNameToPVCName {
		generatedVolumeName, err := generateVolumeNameFromPVC(pvcName)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "Unable to generate volume name from pvc name")
		}
		pvcVols = append(pvcVols, GetPVCVol(generatedVolumeName, pvcName))

		// containerNameToMountPaths is a map of the Devfile container name to their Devfile Volume Mount Paths for a given Volume Name
		containerNameToMountPaths := make(map[string][]string)
		for containerName, volumes := range containerNameToVolumes {
			for _, volume := range volumes {
				if volName == volume.Name {
					containerNameToMountPaths[containerName] = append(containerNameToMountPaths[containerName], volume.ContainerPath)
				}
			}
		}

		containers = AddVolumeMountToContainers(containers, generatedVolumeName, containerNameToMountPaths)
	}
	return containers, pvcVols, nil
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

// generateVolumeNameFromPVC generates a volume name based on the pvc name
func generateVolumeNameFromPVC(pvc string) (volumeName string, err error) {
	volumeName, err = util.NamespaceOpenShiftObject(pvc, "vol")
	if err != nil {
		return "", err
	}
	return
}
