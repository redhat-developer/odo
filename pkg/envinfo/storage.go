package envinfo

import (
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/generator"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"

	"github.com/redhat-developer/odo/pkg/localConfigProvider"
)

const (
	// DefaultVolumeSize Default volume size for volumes defined in a devfile
	DefaultVolumeSize = "1Gi"
)

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
					Path:      generator.GetVolumeMountPath(volumeMount),
					Container: component.Name,
				})
			}
		}
	}

	return storageList, nil
}
