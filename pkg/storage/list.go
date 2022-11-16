package storage

import (
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/generator"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
)

const (
	// DefaultVolumeSize Default volume size for volumes defined in a devfile
	DefaultVolumeSize = "1Gi"
)

// LocalStorage holds storage related information
type LocalStorage struct {
	// Name of the storage
	Name string `yaml:"Name,omitempty"`
	// Size of the storage
	Size string `yaml:"Size,omitempty"`
	// Boolean indicating if the volume should be ephemeral. A nil pointer indicates to use the default behaviour
	Ephemeral *bool `yaml:"Ephemeral,omitempty"`
	// Path of the storage to which it will be mounted on the container
	Path string `yaml:"Path,omitempty"`
	// Container is the container name on which this storage is mounted
	Container string `yaml:"-" json:"-"`
}

// ListStorage gets all the storage from the devfile.yaml
func ListStorage(devfileObj parser.DevfileObj) ([]LocalStorage, error) {
	var storageList []LocalStorage

	volumeMap := make(map[string]devfilev1.Volume)
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{})
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
				storageList = append(storageList, LocalStorage{
					Name:      volumeMount.Name,
					Size:      vol.Size,
					Ephemeral: vol.Ephemeral,
					Path:      generator.GetVolumeMountPath(volumeMount),
					Container: component.Name,
				})
			}
		}
	}

	return storageList, nil
}
