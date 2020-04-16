package lclient

import (
	"github.com/docker/docker/api/types/container"
)

// GenerateContainerConfig creates a containerConfig resource that can be used to create a local Docker container
func (dc *Client) GenerateContainerConfig(image string, entrypoint []string, cmd []string, envVars []string, labels map[string]string) container.Config {
	containerConfig := container.Config{
		Image:      image,
		Entrypoint: entrypoint,
		Cmd:        cmd,
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
