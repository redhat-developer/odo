package lclient

import (
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/pkg/errors"
)

// GetContainersByComponent returns the list of Docker containers that matches the specified component label
// If no container with that component exists, it returns an empty list
func (dc *Client) GetContainersByComponent(componentName string, containers []types.Container) []types.Container {
	var containerList []types.Container

	for _, container := range containers {
		if container.Labels["component"] == componentName {
			containerList = append(containerList, container)
		}
	}
	return containerList
}

// GetContainersByComponentAndAlias returns the list of Docker containers that have the same component and alias labeled
func (dc *Client) GetContainersByComponentAndAlias(componentName string, alias string) ([]types.Container, error) {
	containerList, err := dc.GetContainerList()
	if err != nil {
		return nil, err
	}
	var labeledContainers []types.Container
	for _, container := range containerList {
		if container.Labels["component"] == componentName && container.Labels["alias"] == alias {
			labeledContainers = append(labeledContainers, container)
		}
	}
	return labeledContainers, nil
}

// GetContainerList returns a list of all of the running containers on the user's system
func (dc *Client) GetContainerList() ([]types.Container, error) {
	containers, err := dc.Client.ContainerList(dc.Context, types.ContainerListOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to retrieve Docker containers")
	}
	return containers, nil
}

// StartContainer takes in a Docker container object and starts it.
// containerConfig - configurations for the container itself (image name, command, ports, etc) (if needed)
// hostConfig - configurations related to the host (volume mounts, exposed ports, etc) (if needed)
// networkingConfig - endpoints to expose (if needed)
// Returns an error if the container couldn't be started.
func (dc *Client) StartContainer(containerConfig *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig) error {
	resp, err := dc.Client.ContainerCreate(dc.Context, containerConfig, hostConfig, networkingConfig, "")
	if err != nil {
		return err
	}

	// Start the container
	if err := dc.Client.ContainerStart(dc.Context, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	return nil
}

// RemoveContainer takes in a given container ID and kills it, then removes it.
func (dc *Client) RemoveContainer(containerID string) error {
	err := dc.Client.ContainerRemove(dc.Context, containerID, types.ContainerRemoveOptions{
		Force: true,
	})
	if err != nil {
		return errors.Wrapf(err, "unable to remove container %s", containerID)
	}
	return nil
}

// RemoveVolume removes a volume with the specified volume ID
func (dc *Client) RemoveVolume(volumeID string) error {
	if len(strings.TrimSpace(volumeID)) == 0 {
		return errors.Errorf("A valid volume ID must be specified \"%s\"", volumeID)
	}

	err := dc.Client.VolumeRemove(dc.Context, volumeID, true)
	if err != nil {
		return errors.Wrapf(err, "unable to remove volume %s", volumeID)
	}
	return nil
}

// GetContainerConfigHostConfigAndMounts takes in a given container ID and retrieves its corresponding container config, host config and mounts
func (dc *Client) GetContainerConfigHostConfigAndMounts(containerID string) (*container.Config, *container.HostConfig, []types.MountPoint, error) {
	containerJSON, err := dc.Client.ContainerInspect(dc.Context, containerID)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "unable to inspect container %s", containerID)
	}
	return containerJSON.Config, containerJSON.HostConfig, containerJSON.Mounts, err
}
