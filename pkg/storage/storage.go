package storage

import (
	"github.com/pkg/errors"
	"github.com/redhat-developer/ocdev/pkg/occlient"
)

func Add(config *occlient.VolumeConfig) (string, error) {
	output, err := occlient.SetVolumes(config,
		&occlient.VolumeOpertaions{
			Add: true,
		})
	if err != nil {
		return "", errors.Wrap(err, "unable to create storage")
	}
	return output, nil
}

func Remove(config *occlient.VolumeConfig) (string, error) {
	output, err := occlient.SetVolumes(config,
		&occlient.VolumeOpertaions{
			Remove: true,
		})
	if err != nil {
		return "", errors.Wrap(err, "unable to remove storage")
	}
	return output, nil
}
