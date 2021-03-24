package envinfo

import (
	"fmt"

	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/openshift/odo/pkg/localConfigProvider"
)

const (
	// DefaultVolumeSize Default volume size for volumes defined in a devfile
	DefaultVolumeSize = "1Gi"
)

// CompleteStorage completes the given storage
func (ei *EnvInfo) CompleteStorage(storage *localConfigProvider.LocalStorage) {
	if storage.Size == "" {
		storage.Size = DefaultVolumeSize
	}
	if storage.Path == "" {
		// acc to the devfile schema, if the mount path is absent; it will be mounted at the dir with the mount name
		storage.Path = "/" + storage.Name
	}
}

// ValidateStorage validates the given storage
func (ei *EnvInfo) ValidateStorage(storage localConfigProvider.LocalStorage) error {
	storageList, err := ei.ListStorage()
	if err != nil {
		return err
	}

	for _, store := range storageList {
		if store.Name == storage.Name {
			return fmt.Errorf("storage with name %s already exists", storage.Name)
		}
	}
	return nil
}

// GetStorage gets the storage with the given name
func (ei *EnvInfo) GetStorage(name string) (*localConfigProvider.LocalStorage, error) {
	storageList, err := ei.ListStorage()
	if err != nil {
		return nil, err
	}
	for _, storage := range storageList {
		if name == storage.Name {
			return &storage, nil
		}
	}
	return nil, nil
}

// CreateStorage sets the storage related information in the local configuration
func (ei *EnvInfo) CreateStorage(storage localConfigProvider.LocalStorage) error {
	// Get all the containers in the devfile
	containers, err := ei.GetContainers()
	if err != nil {
		return err
	}

	// Add volumeMount to all containers in the devfile
	for _, c := range containers {
		if err := ei.devfileObj.Data.AddVolumeMounts(c.Name, []devfilev1.VolumeMount{
			{
				Name: storage.Name,
				Path: storage.Path,
			},
		}); err != nil {
			return err
		}
	}

	// Add volume component to devfile. Think along the lines of a k8s pod spec's volumeMount and volume.
	err = ei.devfileObj.Data.AddComponents([]devfilev1.Component{{
		Name: storage.Name,
		ComponentUnion: devfilev1.ComponentUnion{
			Volume: &devfilev1.VolumeComponent{
				Volume: devfilev1.Volume{
					Size: storage.Size,
				},
			},
		},
	}})
	if err != nil {
		return err
	}

	err = ei.devfileObj.WriteYamlDevfile()
	if err != nil {
		return err
	}

	return nil
}

// ListStorage gets all the storage from the devfile.yaml
func (ei *EnvInfo) ListStorage() ([]localConfigProvider.LocalStorage, error) {
	var storageList []localConfigProvider.LocalStorage

	volumeSizeMap := make(map[string]string)
	components, err := ei.devfileObj.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return storageList, err
	}

	for _, component := range components {
		if component.Volume == nil {
			continue
		}
		if component.Volume.Size == "" {
			component.Volume.Size = DefaultVolumeSize
		}
		volumeSizeMap[component.Name] = component.Volume.Size
	}

	for _, component := range components {
		if component.Container == nil {
			continue
		}
		for _, volumeMount := range component.Container.VolumeMounts {
			size, ok := volumeSizeMap[volumeMount.Name]
			if ok {
				storageList = append(storageList, localConfigProvider.LocalStorage{
					Name:      volumeMount.Name,
					Size:      size,
					Path:      GetVolumeMountPath(volumeMount),
					Container: component.Name,
				})
			}
		}
	}

	return storageList, nil
}

// DeleteStorage deletes the storage with the given name
func (ei *EnvInfo) DeleteStorage(name string) error {
	err := ei.devfileObj.Data.DeleteVolumeMount(name)
	if err != nil {
		return err
	}
	err = ei.devfileObj.Data.DeleteComponent(name)
	if err != nil {
		return err
	}

	err = ei.devfileObj.WriteYamlDevfile()
	if err != nil {
		return err
	}

	return nil
}

// GetStorageMountPath gets the mount path of the storage with the given storage name
func (ei *EnvInfo) GetStorageMountPath(storageName string) (string, error) {
	containers, err := ei.GetContainers()
	if err != nil {
		return "", err
	}

	// since all container components have same volume mounts, we simply refer to the first container in the list
	// refer https://github.com/openshift/odo/issues/4105 for addressing "all containers have same volume mounts"
	paths, err := ei.devfileObj.Data.GetVolumeMountPaths(storageName, containers[0].Name)
	if err != nil {
		return "", err
	}
	return paths[0], nil
}

// GetVolumeMountPath gets the volume mount's path
func GetVolumeMountPath(volumeMount devfilev1.VolumeMount) string {
	// if there is no volume mount path, default to volume mount name as per devfile schema
	if volumeMount.Path == "" {
		volumeMount.Path = "/" + volumeMount.Name
	}

	return volumeMount.Path
}
