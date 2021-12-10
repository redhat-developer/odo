package envinfo

import (
	"fmt"

	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"k8s.io/klog/v2"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/redhat-developer/odo/pkg/localConfigProvider"
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

	if storage.Container == "" {
		return nil
	}

	//check if specified container exists or not (if specified)
	containerExists := false
	containers, err := ei.GetContainers()
	if err != nil {
		return err
	}

	for _, c := range containers {
		if c.Name == storage.Container {
			containerExists = true
		}
	}
	if !containerExists {
		return fmt.Errorf("specified container %s does not exist", storage.Container)
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
	//initialize volume mount and volume container
	vm := []devfilev1.VolumeMount{
		{
			Name: storage.Name,
			Path: storage.Path,
		},
	}
	vc := []devfilev1.Component{{
		Name: storage.Name,
		ComponentUnion: devfilev1.ComponentUnion{
			Volume: &devfilev1.VolumeComponent{
				Volume: devfilev1.Volume{
					Size: storage.Size,
				},
			},
		},
	}}
	volumeExists := false
	// Get all the containers in the devfile
	containers, err := ei.GetContainers()
	if err != nil {
		return err
	}

	// Add volumeMount to all containers if no container is specified else to specified container(s) in the devfile
	for _, c := range containers {
		if storage.Container == "" || (storage.Container != "" && c.Name == storage.Container) {
			if e := ei.devfileObj.Data.AddVolumeMounts(c.Name, vm); e != nil {
				return e
			}
		}
	}

	// Get all the components to check if volume component exists
	components, err := ei.devfileObj.Data.GetComponents(common.DevfileOptions{})
	if err != nil {
		return err
	}

	// check if volume component already exists
	for _, component := range components {
		if component.Volume != nil && component.Name == storage.Name {
			volumeExists = true
		}
	}

	// Add volume component to devfile. Think along the lines of a k8s pod spec's volumeMount and volume.
	// Add if volume does not exist, otherwise update
	if !volumeExists {
		err = ei.devfileObj.Data.AddComponents(vc)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("volume with name %s already exists", storage.Name)
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

	volumeMap := make(map[string]devfilev1.Volume)
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
		volumeMap[component.Name] = component.Volume.Volume
	}

	for _, component := range components {
		if component.Container == nil {
			continue
		}
		for _, volumeMount := range component.Container.VolumeMounts {
			vol, ok := volumeMap[volumeMount.Name]
			if ok {
				storageList = append(storageList, localConfigProvider.LocalStorage{
					Name:      volumeMount.Name,
					Size:      vol.Size,
					Ephemeral: vol.Ephemeral,
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
	if len(containers) == 0 {
		return "", fmt.Errorf("invalid devfile: components.container: required value")
	}

	// go over all containers
	for _, c := range containers {
		// get all volume mount paths in current container
		pt, err := ei.devfileObj.Data.GetVolumeMountPaths(storageName, c.Name)
		if err != nil {
			klog.V(2).Infof("Failed to get volume mount paths for storage %s in container %s: %s", storageName, c.Name, err.Error())
		}
		if len(pt) > 0 {
			return pt[0], nil
		}
	}
	// TODO: Below "if" storage needs to be mounted on multiple containers, then the return will have to be an array.
	return "", nil
}

// GetVolumeMountPath gets the volume mount's path.
// To be moved to devfile/library.
func GetVolumeMountPath(volumeMount devfilev1.VolumeMount) string {
	// if there is no volume mount path, default to volume mount name as per devfile schema
	if volumeMount.Path == "" {
		volumeMount.Path = "/" + volumeMount.Name
	}

	return volumeMount.Path
}
