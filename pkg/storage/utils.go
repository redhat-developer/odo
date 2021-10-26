package storage

import (
	"fmt"

	"github.com/openshift/odo/v2/pkg/localConfigProvider"
	"github.com/openshift/odo/v2/pkg/occlient"
	storagelabels "github.com/openshift/odo/v2/pkg/storage/labels"
	"github.com/openshift/odo/v2/pkg/util"
	"github.com/pkg/errors"
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

// ConvertListLocalToMachine converts storage config list to StorageList type
func ConvertListLocalToMachine(storageListConfig []localConfigProvider.LocalStorage) StorageList {

	var storageListLocal []Storage

	for _, storeLocal := range storageListConfig {
		s := NewStorage(storeLocal.Name, storeLocal.Size, storeLocal.Path)
		s.Spec.ContainerName = storeLocal.Container
		storageListLocal = append(storageListLocal, s)
	}

	return NewStorageList(storageListLocal)
}
