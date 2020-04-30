package storage

import (
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
	"k8s.io/klog"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/lclient"
	"github.com/openshift/odo/pkg/util"
)

const volNameMaxLength = 45

// CreateComponentStorage creates a Docker volume with the given list of storages if it does not exist, else it uses the existing volume
func CreateComponentStorage(Client *lclient.Client, storages []common.Storage, componentName string) (err error) {

	for _, storage := range storages {
		volumeName := *storage.Volume.Name
		dockerVolName := storage.Name

		existingDockerVolName, err := GetExistingVolume(Client, volumeName, componentName)
		if err != nil {
			return err
		}

		if len(existingDockerVolName) == 0 {
			klog.V(3).Infof("Creating a Docker volume for %v", volumeName)
			_, err := Create(Client, volumeName, componentName, dockerVolName)
			if err != nil {
				return errors.Wrapf(err, "Error creating Docker volume for "+volumeName)
			}
		}
	}

	return
}

// Create creates the Docker volume for the given volume name and component name
func Create(Client *lclient.Client, name, componentName, dockerVolName string) (*types.Volume, error) {

	labels := map[string]string{
		"component":    componentName,
		"storage-name": name,
	}

	klog.V(3).Infof("Creating a Docker volume with name %v and labels %v", dockerVolName, labels)
	vol, err := Client.CreateVolume(dockerVolName, labels)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create Docker volume")
	}
	return &vol, nil
}

// GenerateVolNameFromDevfileVol generates a Docker volume name from the Devfile volume name and component name
func GenerateVolNameFromDevfileVol(volName, componentName string) (string, error) {

	dockerVolName := fmt.Sprintf("%v-%v", volName, componentName)
	dockerVolName = util.TruncateString(dockerVolName, volNameMaxLength)
	randomChars := util.GenerateRandomString(4)
	dockerVolName, err := util.NamespaceOpenShiftObject(dockerVolName, randomChars)
	if err != nil {
		return "", errors.Wrapf(err, "unable to create namespaced name")
	}

	return dockerVolName, nil
}

// GetExistingVolume checks if a Docker volume is present and return the name if it exists
func GetExistingVolume(Client *lclient.Client, volumeName, componentName string) (string, error) {

	volumeLabels := map[string]string{
		"component":    componentName,
		"storage-name": volumeName,
	}

	klog.V(3).Infof("Checking Docker volume for volume %v and labels %v\n", volumeName, volumeLabels)

	vols, err := Client.GetVolumesByLabel(volumeLabels)
	if err != nil {
		return "", errors.Wrapf(err, "Unable to get Docker volume with selectors %v", volumeLabels)
	}
	if len(vols) == 1 {
		klog.V(3).Infof("Found an existing Docker volume for volume %v and labels %v\n", volumeName, volumeLabels)
		existingVolume := vols[0]
		return existingVolume.Name, nil
	} else if len(vols) == 0 {
		return "", nil
	} else {
		err = errors.New("More than 1 Docker volume found")
		return "", err
	}
}

// ProcessVolumes takes in a list of component volumes and for each unique volume in the devfile, creates a Docker volume name for it
// It returns a list of unique volumes, a mapping of devfile volume names to docker volume names, and an error if applicable
func ProcessVolumes(client *lclient.Client, componentName string, componentAliasToVolumes map[string][]common.DevfileVolume) ([]common.Storage, map[string]string, error) {
	var uniqueStorages []common.Storage
	volumeNameToDockerVolName := make(map[string]string)
	processedVolumes := make(map[string]bool)

	// Get a list of all the unique volume names and generate their Docker volume names
	for _, volumes := range componentAliasToVolumes {
		for _, vol := range volumes {
			if _, ok := processedVolumes[*vol.Name]; !ok {
				processedVolumes[*vol.Name] = true

				// Generate the volume Names
				klog.V(3).Infof("Generating Docker volumes name for %v", *vol.Name)
				generatedDockerVolName, err := GenerateVolNameFromDevfileVol(*vol.Name, componentName)
				if err != nil {
					return nil, nil, err
				}

				// Check if we have an existing volume with the labels, overwrite the generated name with the existing name if present
				existingVolName, err := GetExistingVolume(client, *vol.Name, componentName)
				if err != nil {
					return nil, nil, err
				}
				if len(existingVolName) > 0 {
					klog.V(3).Infof("Found an existing Docker volume for %v, volume %v will be re-used", *vol.Name, existingVolName)
					generatedDockerVolName = existingVolName
				}

				dockerVol := common.Storage{
					Name:   generatedDockerVolName,
					Volume: vol,
				}
				uniqueStorages = append(uniqueStorages, dockerVol)
				volumeNameToDockerVolName[*vol.Name] = generatedDockerVolName
			}
		}
	}
	return uniqueStorages, volumeNameToDockerVolName, nil
}
