package lclient

import (
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

// GenerateContainerConfig creates a containerConfig resource that can be used to create a local Docker container
func (dc *Client) GenerateContainerConfig(image string, entrypoint []string, args []string, envVars []string, labels map[string]string) container.Config {
	containerConfig := container.Config{
		Image:      image,
		Entrypoint: entrypoint,
		Cmd:        args,
		Env:        envVars,
		Labels:     labels,
	}
	return containerConfig
}

func (dc *Client) GenerateHostConfig(isPrivileged bool, publishPorts bool) container.HostConfig {
	hostConfig := container.HostConfig{
		Privileged:      isPrivileged,
		PublishAllPorts: publishPorts,
	}
	return hostConfig
}
