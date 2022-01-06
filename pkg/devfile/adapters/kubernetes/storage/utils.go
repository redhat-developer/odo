package storage

import (
	"fmt"
	"strings"

	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/pkg/errors"
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/storage"
	storagepkg "github.com/redhat-developer/odo/pkg/storage"
	storagelabels "github.com/redhat-developer/odo/pkg/storage/labels"
	"github.com/redhat-developer/odo/pkg/util"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
)

// VolumeInfo is a struct to hold the pvc name and the volume name to create a volume.
// To be moved to devfile/library.
type VolumeInfo struct {
	PVCName    string
	VolumeName string
}

// GetVolumeInfos returns the PVC name attached to the `odo-projects` directory and a map of other PVCs
func GetVolumeInfos(pvcs []corev1.PersistentVolumeClaim) (odoSourcePVCName string, infos map[string]VolumeInfo, _ error) {
	infos = make(map[string]VolumeInfo)
	for _, pvc := range pvcs {
		// check if the pvc is in the terminating state
		if pvc.DeletionTimestamp != nil {
			continue
		}

		generatedVolumeName, e := generateVolumeNameFromPVC(pvc.Name)
		if e != nil {
			return "", nil, errors.Wrapf(e, "Unable to generate volume name from pvc name")
		}

		if pvc.Labels[storagelabels.StorageLabel] == storagepkg.OdoSourceVolume {
			odoSourcePVCName = pvc.Name
			continue
		}

		infos[pvc.Labels[storagelabels.StorageLabel]] = VolumeInfo{
			PVCName:    pvc.Name,
			VolumeName: generatedVolumeName,
		}
	}
	return odoSourcePVCName, infos, nil
}

// GetPersistentVolumesAndVolumeMounts gets the PVC volumes and updates the containers with the volume mounts.
// volumeNameToVolInfo is a map of the devfile volume name to the volume info containing the pvc name and the volume name.
// To be moved to devfile/library.
func GetPersistentVolumesAndVolumeMounts(devfileObj devfileParser.DevfileObj, containers []corev1.Container, initContainers []corev1.Container, volumeNameToVolInfo map[string]VolumeInfo, options parsercommon.DevfileOptions) ([]corev1.Volume, error) {

	containerComponents, err := devfileObj.Data.GetDevfileContainerComponents(options)
	if err != nil {
		return nil, err
	}

	var pvcVols []corev1.Volume
	for volName, volInfo := range volumeNameToVolInfo {
		pvcVols = append(pvcVols, getPVC(volInfo.VolumeName, volInfo.PVCName))

		// containerNameToMountPaths is a map of the Devfile container name to their Devfile Volume Mount Paths for a given Volume Name
		containerNameToMountPaths := make(map[string][]string)
		for _, containerComp := range containerComponents {
			for _, volumeMount := range containerComp.Container.VolumeMounts {
				if volName == volumeMount.Name {
					containerNameToMountPaths[containerComp.Name] = append(containerNameToMountPaths[containerComp.Name], envinfo.GetVolumeMountPath(volumeMount))
				}
			}
		}

		addVolumeMountToContainers(containers, initContainers, volInfo.VolumeName, containerNameToMountPaths)
	}
	return pvcVols, nil
}

func GetEphemeralVolumesAndVolumeMounts(devfileObj devfileParser.DevfileObj, containers []corev1.Container, initContainers []corev1.Container, ephemerals map[string]storage.Storage, options parsercommon.DevfileOptions) ([]corev1.Volume, error) {
	containerComponents, err := devfileObj.Data.GetDevfileContainerComponents(options)
	if err != nil {
		return nil, err
	}
	var emptydirVols []corev1.Volume
	for volName, volInfo := range ephemerals {
		emptyDir, err := getEmptyDir(volInfo.Name, volInfo.Spec.Size)
		if err != nil {
			return nil, err
		}
		emptydirVols = append(emptydirVols, emptyDir)

		// containerNameToMountPaths is a map of the Devfile container name to their Devfile Volume Mount Paths for a given Volume Name
		containerNameToMountPaths := make(map[string][]string)
		for _, containerComp := range containerComponents {
			for _, volumeMount := range containerComp.Container.VolumeMounts {
				if volName == volumeMount.Name {
					containerNameToMountPaths[containerComp.Name] = append(containerNameToMountPaths[containerComp.Name], envinfo.GetVolumeMountPath(volumeMount))
				}
			}
		}

		addVolumeMountToContainers(containers, initContainers, volInfo.Name, containerNameToMountPaths)
	}
	return emptydirVols, nil
}

// getPVC gets a pvc type volume with the given volume name and pvc name.
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

// getEmptyDir gets an emptyDir type volume with the given volume name and size.
// size should be parseable as a Kubernetes `Quantity` or an error will be returned
func getEmptyDir(volumeName string, size string) (corev1.Volume, error) {

	emptyDir := &corev1.EmptyDirVolumeSource{}
	qty, err := resource.ParseQuantity(size)
	if err != nil {
		return corev1.Volume{}, err
	}
	emptyDir.SizeLimit = &qty
	return corev1.Volume{
		Name: volumeName,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: emptyDir,
		},
	}, nil
}

// addVolumeMountToContainers adds the Volume Mounts in containerNameToMountPaths to the containers for a given pvc and volumeName
// containerNameToMountPaths is a map of a container name to an array of its Mount Paths.
// To be moved to devfile/library.
func addVolumeMountToContainers(containers []corev1.Container, initContainers []corev1.Container, volumeName string, containerNameToMountPaths map[string][]string) {

	for containerName, mountPaths := range containerNameToMountPaths {
		for i := range containers {
			if containers[i].Name == containerName {
				for _, mountPath := range mountPaths {
					containers[i].VolumeMounts = append(containers[i].VolumeMounts, corev1.VolumeMount{
						Name:      volumeName,
						MountPath: mountPath,
						SubPath:   "",
					},
					)
				}
			}
		}
		for i := range initContainers {
			if strings.HasPrefix(initContainers[i].Name, containerName) {
				for _, mountPath := range mountPaths {
					initContainers[i].VolumeMounts = append(initContainers[i].VolumeMounts, corev1.VolumeMount{
						Name:      volumeName,
						MountPath: mountPath,
						SubPath:   "",
					},
					)
				}
			}
		}
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

// HandleEphemeralStorage creates or deletes the ephemeral volume based on the preference setting
func HandleEphemeralStorage(client kclient.ClientInterface, storageClient storage.Client, componentName string, isEphemeral bool) error {
	selector := fmt.Sprintf("%v=%s,%s=%s", componentlabels.ComponentLabel, componentName, storagelabels.SourcePVCLabel, storage.OdoSourceVolume)

	pvcs, err := client.ListPVCs(selector)
	if err != nil && !kerrors.IsNotFound(err) {
		return err
	}

	if !isEphemeral {
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
