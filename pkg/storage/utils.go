package storage

import (
	"fmt"

	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/openshift/odo/pkg/occlient"
	storagelabels "github.com/openshift/odo/pkg/storage/labels"
	"github.com/openshift/odo/pkg/util"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getPVCNameFromStorageName returns the PVC associated with the given storage
func getPVCNameFromStorageName(client *occlient.Client, storageName string) (string, error) {
	var labels = make(map[string]string)
	labels[storagelabels.StorageLabel] = storageName

	selector := util.ConvertLabelsToSelector(labels)
	pvcs, err := client.GetKubeClient().ListPVCNames(selector)
	if err != nil {
		return "", errors.Wrapf(err, "unable to get PVC names for selector %v", selector)
	}
	numPVCs := len(pvcs)
	if numPVCs != 1 {
		return "", fmt.Errorf("expected exactly one PVC attached to storage %v, but got %v, %v", storageName, numPVCs, pvcs)
	}
	return pvcs[0], nil
}

// GetMachineReadableFormatForList gives machine-readable StorageList definition
func GetMachineReadableFormatForList(storage []Storage) StorageList {
	return StorageList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "List",
			APIVersion: apiVersion,
		},
		ListMeta: metav1.ListMeta{},
		Items:    storage,
	}
}

// GetMachineReadableFormat gives machine-readable Storage definition
// storagePath indicates the path to which the storage is mounted to, "" if not mounted
func GetMachineReadableFormat(storageName, storageSize, storagePath string) Storage {
	return Storage{
		TypeMeta:   metav1.TypeMeta{Kind: "Storage", APIVersion: apiVersion},
		ObjectMeta: metav1.ObjectMeta{Name: storageName},
		Spec: StorageSpec{
			Size: storageSize,
			Path: storagePath,
		},
	}
}

// generatePVCName generates a PVC name from the Devfile volume name, component name and app name
func generatePVCName(volName, componentName, appName string) (string, error) {

	pvcName, err := util.NamespaceKubernetesObject(volName, componentName)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create namespaced name")
	}

	pvcName, err = util.NamespaceKubernetesObject(pvcName, appName)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create namespaced name")
	}

	return pvcName, nil
}

// ConvertListLocalToMachine It converts storage config list to StorageList type
func ConvertListLocalToMachine(storageListConfig []localConfigProvider.LocalStorage) StorageList {

	var storageListLocal []Storage

	for _, storeLocal := range storageListConfig {
		s := GetMachineReadableFormat(storeLocal.Name, storeLocal.Size, storeLocal.Path)
		s.Spec.ContainerName = storeLocal.Container
		storageListLocal = append(storageListLocal, s)
	}

	return GetMachineReadableFormatForList(storageListLocal)
}
