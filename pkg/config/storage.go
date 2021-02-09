package config

import "github.com/openshift/odo/pkg/localConfigProvider"

// ListStorage gets all the storage from the config
func (lci *LocalConfigInfo) ListStorage() []localConfigProvider.LocalStorage {
	if lci.componentSettings.Storage == nil {
		return []localConfigProvider.LocalStorage{}
	}

	var storageList []localConfigProvider.LocalStorage
	for _, storage := range *lci.componentSettings.Storage {
		storageList = append(storageList, localConfigProvider.LocalStorage{
			Name: storage.Name,
			Path: storage.Path,
			Size: storage.Size,
		})
	}
	return storageList
}
