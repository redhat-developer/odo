package lclient

import (
	"reflect"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	volumeTypes "github.com/docker/docker/api/types/volume"
	"github.com/pkg/errors"
)

// CreateVolume creates a Docker volume with the given labels and the default Docker storage driver
func (dc *Client) CreateVolume(name string, labels map[string]string) (types.Volume, error) {
	volume, err := dc.Client.VolumeCreate(dc.Context, volumeTypes.VolumeCreateBody{
		Name:   name,
		Driver: DockerStorageDriver,
		Labels: labels,
	})

	if err != nil {
		return volume, errors.Wrapf(err, "error creating docker volume")
	}

	return volume, nil
}

// GetVolumesByLabel returns the list of all volumes matching the given label.
func (dc *Client) GetVolumesByLabel(labels map[string]string) ([]types.Volume, error) {
	var volumes []types.Volume
	volumeList, err := dc.Client.VolumeList(dc.Context, filters.Args{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get list of docker volumes")
	}

	for _, vol := range volumeList.Volumes {
		if reflect.DeepEqual(vol.Labels, labels) {
			volumes = append(volumes, *vol)
		}
	}

	return volumes, nil
}

// GetVolumes returns the list of all volumes
func (dc *Client) GetVolumes() ([]types.Volume, error) {
	var volumes []types.Volume
	volumeList, err := dc.Client.VolumeList(dc.Context, filters.Args{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get list of docker volumes")
	}

	for _, vol := range volumeList.Volumes {
		volumes = append(volumes, *vol)
	}

	return volumes, nil
}
