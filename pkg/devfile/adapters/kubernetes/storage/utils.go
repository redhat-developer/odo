package storage

import (
	"fmt"
	"strings"

	"github.com/devfile/library/pkg/devfile/generator"
	"github.com/openshift/odo/pkg/storage/labels"
	"github.com/pkg/errors"
	"k8s.io/klog"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/kclient"
	"github.com/openshift/odo/pkg/storage"
	"github.com/openshift/odo/pkg/util"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const pvcNameMaxLen = 45

// CreateComponentStorage creates PVCs with the given list of storages if it does not exist, else it uses the existing PVC
func CreateComponentStorage(Client *kclient.Client, storages []common.Storage, componentName string) (err error) {

	for _, storage := range storages {
		volumeName := storage.Volume.Name
		volumeSize := storage.Volume.Size
		pvcName := storage.Name

		existingPVCName, err := GetExistingPVC(Client, volumeName, componentName)
		if err != nil {
			return err
		}

		if len(existingPVCName) == 0 {
			klog.V(2).Infof("Creating a PVC for %v", volumeName)
			_, err := Create(Client, volumeName, volumeSize, componentName, pvcName)
			if err != nil {
				return errors.Wrapf(err, "Error creating PVC for "+volumeName)
			}
		}
	}

	return
}

// Create creates the pvc for the given pvc name, volume name, volume size and component name
func Create(Client *kclient.Client, name, size, componentName, pvcName string) (*corev1.PersistentVolumeClaim, error) {

	label := map[string]string{
		"component":                componentName,
		labels.DevfileStorageLabel: name,
	}

	if strings.Contains(pvcName, storage.OdoSourceVolume) {
		// Add label for source pvc
		label[labels.SourcePVCLabel] = name
	}

	objectMeta := generator.GetObjectMeta(pvcName, Client.Namespace, label, nil)

	// Get the deployment
	deployment, err := Client.GetDeploymentByName(componentName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get deployment")
	}

	// Generate owner reference for the deployment and update objectMeta
	ownerReference := generator.GetOwnerReference(deployment)
	objectMeta.OwnerReferences = append(objectMeta.OwnerReferences, ownerReference)

	quantity, err := resource.ParseQuantity(size)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse size: %v", size)
	}

	pvcParams := generator.PVCParams{
		ObjectMeta: objectMeta,
		Quantity:   quantity,
	}
	pvc := generator.GetPVC(pvcParams)

	// Create PVC
	klog.V(2).Infof("Creating a PVC with name %v and labels %v", pvcName, label)
	pvc, err = Client.CreatePVC(*pvc)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create PVC")
	}
	return pvc, nil
}

// GetExistingPVC checks if a PVC is present and return the name if it exists
func GetExistingPVC(Client *kclient.Client, volumeName, componentName string) (string, error) {

	label := fmt.Sprintf("component=%s,%s=%s", componentName, labels.DevfileStorageLabel, volumeName)

	klog.V(2).Infof("Checking PVC for volume %v and label %v\n", volumeName, label)

	PVCs, err := Client.ListPVCs(label)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to get PVC with selectors "+label)
	}
	if len(PVCs) == 1 {
		klog.V(2).Infof("Found an existing PVC for volume %v and label %v\n", volumeName, label)
		existingPVC := &PVCs[0]
		return existingPVC.Name, nil
	} else if len(PVCs) == 0 {
		return "", nil
	} else {
		err = errors.New("More than 1 PVC found with the label " + label)
		return "", err
	}
}

// DeleteOldPVCs deletes all the old PVCs which are not in the processedVolumes map
func DeleteOldPVCs(Client *kclient.Client, componentName string, processedVolumes map[string]bool) error {
	label := fmt.Sprintf("component=%s", componentName)
	PVCs, err := Client.ListPVCs(label)
	if err != nil {
		return errors.Wrapf(err, "unable to get PVC with selectors "+label)
	}
	for _, pvc := range PVCs {
		storageName, ok := pvc.GetLabels()[labels.DevfileStorageLabel]
		if ok && !processedVolumes[storageName] {
			// the pvc is not in the processedVolumes map
			// thus deleting those PVCs
			err := Client.DeletePVC(pvc.GetName())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

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
