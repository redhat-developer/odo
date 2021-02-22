package config

import (
	"fmt"

	"github.com/openshift/odo/pkg/localConfigProvider"
	"github.com/pkg/errors"
)

// GetStorage gets the storage with the given name
func (lci *LocalConfigInfo) GetStorage(storageName string) (*localConfigProvider.LocalStorage, error) {
	configStorage, err := lci.ListStorage()
	if err != nil {
		return nil, err
	}
	for _, storage := range configStorage {
		if storageName == storage.Name {
			return &storage, nil
		}
	}
	return nil, nil
}

// CreateStorage sets the storage related information in the local configuration
func (lci *LocalConfigInfo) CreateStorage(storage localConfigProvider.LocalStorage) error {
	err := lci.SetConfiguration("storage", storage)
	if err != nil {
		return err
	}
	return err
}

// ListStorage gets all the storage from the config
func (lci *LocalConfigInfo) ListStorage() ([]localConfigProvider.LocalStorage, error) {
	if lci.componentSettings.Storage == nil {
		return []localConfigProvider.LocalStorage{}, nil
	}

	var storageList []localConfigProvider.LocalStorage
	for _, storage := range *lci.componentSettings.Storage {
		storageList = append(storageList, localConfigProvider.LocalStorage{
			Name: storage.Name,
			Path: storage.Path,
			Size: storage.Size,
		})
	}
	return storageList, nil
}

// DeleteStorage deletes the storage with the given name
func (lci *LocalConfigInfo) DeleteStorage(name string) error {
	storage, err := lci.GetStorage(name)
	if err != nil {
		return err
	}
	if storage == nil {
		return errors.Errorf("storage named %s doesn't exists", name)
	}
	return lci.DeleteFromConfigurationList("storage", name)
}

// CompleteStorage completes the given storage
func (lci *LocalConfigInfo) CompleteStorage(storage *localConfigProvider.LocalStorage) {}

// ValidateStorage validates the given storage
func (lci *LocalConfigInfo) ValidateStorage(storage localConfigProvider.LocalStorage) error {
	if storage.Size == "" || storage.Path == "" {
		return fmt.Errorf("\"size\" and \"path\" flags are required for s2i components")
	}

	configStorage, err := lci.ListStorage()
	if err != nil {
		return err
	}
	for _, store := range configStorage {
		if store.Name == storage.Name {
			return errors.Errorf("there already is a storage with the name %s", storage.Name)
		}
		if store.Path == storage.Path {
			return errors.Errorf("there already is a storage mounted at %s", storage.Path)
		}
	}
	return nil
}

// GetStorageMountPath gets the mount path of the storage with the given storage name
func (lci *LocalConfigInfo) GetStorageMountPath(storageName string) (string, error) {
	configStorage, err := lci.ListStorage()
	if err != nil {
		return "", err
	}

	var mPath string
	for _, storage := range configStorage {
		if storage.Name == storageName {
			mPath = storage.Path
		}
	}
	return mPath, nil
}
