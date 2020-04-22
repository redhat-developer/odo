package lclient

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
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

// GenerateHostConfig creates a HostConfig resource that can be used to create a local Docker container
func (dc *Client) GenerateHostConfig(isPrivileged bool, publishPorts bool, portmap nat.PortMap) container.HostConfig {
	hostConfig := container.HostConfig{
		Privileged:      isPrivileged,
		PublishAllPorts: publishPorts,
		PortBindings:    portmap,
	}
	return hostConfig
}
