package storage

import (
	"github.com/pkg/errors"
	"github.com/redhat-developer/ocdev/pkg/occlient"
)

func Add(client *occlient.Client, config *occlient.VolumeConfig) (string, error) {
	output, err := client.SetVolumes(config,
		&occlient.VolumeOperations{
			Add: true,
		})
	if err != nil {
		return "", errors.Wrap(err, "unable to create storage")
	}
	return output, nil
}

func Remove(client *occlient.Client, config *occlient.VolumeConfig) (string, error) {
	output, err := client.SetVolumes(config,
		&occlient.VolumeOperations{
			Remove: true,
		})
	if err != nil {
		return "", errors.Wrap(err, "unable to remove storage")
	}
	return output, nil
}

func List(client *occlient.Client, config *occlient.VolumeConfig) (string, error) {
	output, err := client.SetVolumes(config,
		&occlient.VolumeOperations{
			List: true,
		})
	if err != nil {
		return "", errors.Wrap(err, "unable to list storage")
	}
	return output, nil
}
