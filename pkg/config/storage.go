package config

import (
	"github.com/pkg/errors"
)

func (lci *LocalConfigInfo) StorageCreate(name, size, path string) (ComponentStorageSettings, error) {
	storage := ComponentStorageSettings{
		Name: name,
		Size: size,
		Path: path,
	}
	err := lci.SetConfiguration("storage", storage)
	if err != nil {
		return ComponentStorageSettings{}, err
	}
	return storage, err
}

func (lci *LocalConfigInfo) StorageExists(storageName string) bool {
	for _, storage := range lci.GetStorage() {
		if storageName == storage.Name {
			return true
		}
	}
	return false
}

func (lci *LocalConfigInfo) StorageList() ([]ComponentStorageSettings, error) {
	storageConfigList := lci.GetStorage()
	var storageList []ComponentStorageSettings
	for _, storage := range storageConfigList {
		storageList = append(storageList, ComponentStorageSettings{
			Name: storage.Name,
			Path: storage.Path,
			Size: storage.Size,
		})
	}
	return storageList, nil
}

func (lci *LocalConfigInfo) ValidateStorage(storageName, storagePath string) error {
	for _, storage := range lci.GetStorage() {
		if storage.Name == storageName {
			return errors.Errorf("there already is a storage with the name %s", storageName)
		}
		if storage.Path == storagePath {
			return errors.Errorf("there already is a storage mounted at %s", storagePath)
		}
	}
	return nil
}

func (lci *LocalConfigInfo) StorageDelete(name string) error {
	exists := lci.StorageExists(name)
	if !exists {
		return errors.Errorf("storage named %s doesn't exists", name)
	}
	return lci.DeleteFromConfigurationList("storage", name)
}

func (lci *LocalConfigInfo) GetMountPath(storageName string) string {
	var mPath string
	storageList, _ := lci.StorageList()
	for _, storage := range storageList {
		if storage.Name == storageName {
			mPath = storage.Path
		}
	}
	return mPath
}
