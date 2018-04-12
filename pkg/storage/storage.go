package storage

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/occlient"
	"github.com/redhat-developer/odo/pkg/util"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

const StorageLabel = "app.kubernetes.io/storage-name"

type StorageInfo struct {
	Name string
	Size string
	Path string
}

// Create adds storage to given component of given application
func Create(client *occlient.Client, name string, size string, path string, componentName string, applicationName string) (string, error) {

	labels := GetLabels(name, componentName, applicationName, true)

	log.Debugf("Got labels for PVC: %v", labels)

	// Create PVC
	pvc, err := client.CreatePVC(generatePVCNameFromStorageName(name), size, labels)
	if err != nil {
		return "", errors.Wrap(err, "unable to create PVC")
	}

	// Get DeploymentConfig for the given component
	dc, err := component.GetComponentDeploymentConfig(client, componentName, applicationName)
	if err != nil {
		return "", errors.Wrapf(err, "unable to get Deployment Config for component: %v in application: %v", componentName, applicationName)
	}
	log.Debugf("Deployment Config: %v is associated with the component: %v", dc.Name, componentName)

	// Add PVC to DeploymentConfig
	if err := client.AddPVCToDeploymentConfig(dc, pvc.Name, path); err != nil {
		return "", errors.Wrap(err, "unable to add PVC to DeploymentConfig")
	}

	return dc.Name, nil
}

// Remove removes storage from given component of the given application.
// If no storage name is provided, all storage for the given component is removed
func Remove(client *occlient.Client, name string, applicationName string, componentName string) error {

	// Get DeploymentConfig for the given component
	dc, err := component.GetComponentDeploymentConfig(client, componentName, applicationName)
	if err != nil {
		return errors.Wrapf(err, "unable to get Deployment Config for component: %v in application: %v", componentName, applicationName)
	}

	if name == "" {
		labels := GetLabels("", componentName, applicationName, false)
		selector := util.ConvertLabelsToSelector(labels)
		pvcs, err := client.GetPVCNamesFromSelector(selector)
		log.Debugf("Found PVC names\n%v\nfor selector\n%v", pvcs, selector)
		if err != nil {
			return errors.Wrapf(err, "error getting PVC names from selector: %v", selector)
		}

		for _, pvc := range pvcs {
			log.Debugf("Removing storage for PVC %v from Deployment Config %v", pvc, dc.Name)
			if err := removeStorage(client, pvc, dc.Name); err != nil {
				return errors.Wrap(err, "unable to remove storage")
			}
			log.Debugf("Removed storage for pvc: %v", pvc)
		}
	} else {
		pvc := generatePVCNameFromStorageName(name)
		if err := removeStorage(client, pvc, dc.Name); err != nil {
			return errors.Wrap(err, "unable to remove storage")
		}
	}

	return nil
}

// List lists all the storage associated with the given component of the given
// application
func List(client *occlient.Client, applicationName string, componentName string) ([]StorageInfo, error) {
	labels := GetLabels("", componentName, applicationName, false)
	selector := util.ConvertLabelsToSelector(labels)

	log.Debugf("Looking for PVCs with the selector: %v", selector)
	pvcs, err := client.GetPVCsFromSelector(selector)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get PVC names")
	}

	var storageList []StorageInfo
	for _, pvc := range pvcs {
		storage := getStorageFromPVC(pvc)
		if storage != "" {
			size := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
			storageList = append(storageList, StorageInfo{
				Name: storage,
				Size: size.String(),
			})
		}
	}
	return storageList, nil
}

// generatePVCNameFromStorageName generates a PVC name from the given storage
// name
func generatePVCNameFromStorageName(storage string) string {
	return fmt.Sprintf("%v-pvc", storage)
}

// GetLabels gets the labels to be applied to the given storage besides the
// component labels and application labels.
func GetLabels(storageName string, componentName string, applicationName string, additional bool) map[string]string {
	labels := component.GetLabels(componentName, applicationName, additional)
	labels[StorageLabel] = storageName
	return labels
}

// getStorageFromPVC returns the storage assocaited with the given PVC
func getStorageFromPVC(pvc corev1.PersistentVolumeClaim) string {
	if _, ok := pvc.Labels[StorageLabel]; !ok {
		return ""
	}
	return pvc.Labels[StorageLabel]
}

// removeStorage removes the given PVC from the given Deployment Config and also
// deletes the PVC
func removeStorage(client *occlient.Client, pvc string, dc string) error {
	// Remove PVC from Deployment Config
	if err := client.RemoveVolumeFromDeploymentConfig(pvc, dc); err != nil {
		return errors.Wrapf(err, "unable to remove volume: %v from Deployment Config: %v", pvc, dc)
	}

	// Delete PVC
	if err := client.DeletePVC(pvc); err != nil {
		return errors.Wrapf(err, "unable to delete PVC: %v", pvc)
	}

	return nil
}
