package kclient

import (
	"fmt"

	"github.com/openshift/odo/pkg/devfile/versions/common"
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

// AddPVCToPodTemplateSpec adds the given PVC to the podTemplateSpec
func AddPVCToPodTemplateSpec(podTemplateSpec *corev1.PodTemplateSpec, volumeName, pvc string) {

	podTemplateSpec.Spec.Volumes = append(podTemplateSpec.Spec.Volumes, corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvc,
			},
		},
	})
}

// AddVolumeMountToPodTemplateSpec adds the Volume Mounts in containerMountPathsMap to the podTemplateSpec containers for a given PVC pvc and volume volumeName
// componentAliasToMountPaths is a map of a container name/alias to an array of Mount Paths
func AddVolumeMountToPodTemplateSpec(podTemplateSpec *corev1.PodTemplateSpec, volumeName, pvc string, componentAliasToMountPaths map[string][]string) error {

	// Validating podTemplateSpec.Spec.Containers[] is present before dereferencing
	if len(podTemplateSpec.Spec.Containers) == 0 {
		return fmt.Errorf("podTemplateSpec %s doesn't have any Containers defined", podTemplateSpec.ObjectMeta.Name)
	}

	for containerName, mountPaths := range componentAliasToMountPaths {
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

// AddPVCAndVolumeMount adds PVC and volume mount to the pod template spec
// volumeNameToPVC is a map of volume name to the PVC created
// componentAliasToVolumes is a map of the Devfile container alias to the Devfile Volumes
func AddPVCAndVolumeMount(podTemplateSpec *corev1.PodTemplateSpec, volumeNameToPVC map[string]*corev1.PersistentVolumeClaim, componentAliasToVolumes map[string][]common.DockerimageVolume) error {
	for vol, pvc := range volumeNameToPVC {
		pvcName := pvc.Name
		generatedVolumeName := generateVolumeNameFromPVC(pvcName)
		AddPVCToPodTemplateSpec(podTemplateSpec, generatedVolumeName, pvcName)

		// componentAliasToMountPaths is a map of the Devfile container alias to their Devfile Volume Mount Paths for a given Volume Name
		componentAliasToMountPaths := make(map[string][]string)
		for containerName, volumes := range componentAliasToVolumes {
			for _, volume := range volumes {
				if vol == *volume.Name {
					componentAliasToMountPaths[containerName] = append(componentAliasToMountPaths[containerName], *volume.ContainerPath)
				}
			}
		}

		err := AddVolumeMountToPodTemplateSpec(podTemplateSpec, generatedVolumeName, pvcName, componentAliasToMountPaths)
		if err != nil {
			return errors.New("Unable to add volumes mounts to the pod: " + err.Error())
		}
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

// generateVolumeNameFromPVC generates a random volume name based on the name
// of the given PVC
func generateVolumeNameFromPVC(pvc string) string {
	return fmt.Sprintf("%v-%v-volume", pvc, util.GenerateRandomString(nameLength))
}
