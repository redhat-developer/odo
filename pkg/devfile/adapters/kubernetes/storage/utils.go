package storage

import (
	"fmt"

	componentlabels "github.com/openshift/odo/pkg/component/labels"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/preference"
	"github.com/openshift/odo/pkg/storage"
	storagelabels "github.com/openshift/odo/pkg/storage/labels"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
)

// GetPVCAndVolumeMount gets the PVC and updates the containers with the volume mount
// volumeNameToPVCName is a map of volume name to the PVC created
// containerNameToVolumes is a map of the Devfile container names to the Devfile Volumes
func GetPVCAndVolumeMount(containers []corev1.Container, volumeNameToPVCName map[string]string, containerNameToVolumes map[string][]common.DevfileVolume) ([]corev1.Container, []corev1.Volume, error) {
	var pvcVols []corev1.Volume
	for volName, pvcName := range volumeNameToPVCName {
		generatedVolumeName, err := generateVolumeNameFromPVC(pvcName)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "Unable to generate volume name from pvc name")
		}
		pvcVols = append(pvcVols, getPVC(generatedVolumeName, pvcName))

		// containerNameToMountPaths is a map of the Devfile container name to their Devfile Volume Mount Paths for a given Volume Name
		containerNameToMountPaths := make(map[string][]string)
		for containerName, volumes := range containerNameToVolumes {
			for _, volume := range volumes {
				if volName == volume.Name {
					containerNameToMountPaths[containerName] = append(containerNameToMountPaths[containerName], volume.ContainerPath)
				}
			}
		}

		containers = addVolumeMountToContainers(containers, generatedVolumeName, containerNameToMountPaths)
	}
	return containers, pvcVols, nil
}

// getPVC gets a pvc type volume with the given volume name and pvc name
func getPVC(volumeName, pvcName string) corev1.Volume {

	return corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: pvcName,
			},
		},
	}
}

// generateVolumeNameFromPVC generates a volume name based on the pvc name
func generateVolumeNameFromPVC(pvc string) (volumeName string, err error) {
	volumeName, err = util.NamespaceOpenShiftObject(pvc, "vol")
	if err != nil {
		return "", err
	}
	return
}

// addVolumeMountToContainers adds the Volume Mounts in containerNameToMountPaths to the containers for a given pvc and volumeName
// containerNameToMountPaths is a map of a container name to an array of its Mount Paths
func addVolumeMountToContainers(containers []corev1.Container, volumeName string, containerNameToMountPaths map[string][]string) []corev1.Container {

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

// HandleEphemeralStorage creates or deletes the ephemeral volume based on the preference setting
func HandleEphemeralStorage(client kclient.Client, storageClient storage.Client, componentName string) error {
	pref, err := preference.New()
	if err != nil {
		return err
	}

	selector := fmt.Sprintf("%v=%s,%s=%s", componentlabels.ComponentLabel, componentName, storagelabels.SourcePVCLabel, storage.OdoSourceVolume)

	pvcs, err := client.ListPVCs(selector)
	if err != nil && !kerrors.IsNotFound(err) {
		return err
	}

	if !pref.GetEphemeralSourceVolume() {
		if len(pvcs) == 0 {
			err := storageClient.Create(storage.Storage{
				ObjectMeta: metav1.ObjectMeta{
					Name: storage.OdoSourceVolume,
				},
				Spec: storage.StorageSpec{
					Size: storage.OdoSourceVolumeSize,
				},
			})

			if err != nil {
				return err
			}
		} else if len(pvcs) > 1 {
			return fmt.Errorf("number of source volumes shouldn't be greater than 1")
		}
	} else {
		if len(pvcs) > 0 {
			for _, pvc := range pvcs {
				err := client.DeletePVC(pvc.Name)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
