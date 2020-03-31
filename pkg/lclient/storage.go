package lclient

import (
	volumeTypes "github.com/docker/docker/api/types/volume"
)

func (dc *Client) CreateVolume(name string, labels map[string]string) {
	dc.Client.VolumeCreate(dc.Context, volumeTypes.VolumeCreateBody{
		Driver: DockerStorageDriver,
		Name:   name,
		Labels: labels,
	})
}
