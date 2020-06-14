package lclient

import (
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
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
// Returns containerID of the started container, an error if the container couldn't be started
func (dc *Client) StartContainer(containerConfig *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig) (string, error) {
	resp, err := dc.Client.ContainerCreate(dc.Context, containerConfig, hostConfig, networkingConfig, "")
	if err != nil {
		return "", err
	}

	// Start the container
	if err := dc.Client.ContainerStart(dc.Context, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
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

//ExecCMDInContainer executes the command in the container with containerID
func (dc *Client) ExecCMDInContainer(compInfo common.ComponentInfo, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {

	execConfig := types.ExecConfig{
		AttachStdin:  stdin != nil,
		AttachStdout: stdout != nil,
		AttachStderr: stderr != nil,
		Cmd:          cmd,
		Tty:          tty,
	}

	resp, err := dc.Client.ContainerExecCreate(dc.Context, compInfo.ContainerName, execConfig)
	if err != nil {
		return err
	}

	hresp, err := dc.Client.ContainerExecAttach(dc.Context, resp.ID, types.ExecStartCheck{})
	if err != nil {
		return err
	}
	defer hresp.Close()

	errorCh := make(chan error)

	// read the output
	go func() {
		_, err = stdcopy.StdCopy(stdout, stderr, hresp.Reader)
		errorCh <- err
	}()

	err = <-errorCh
	if err != nil {
		return err
	}

	hresp.Close()

	return nil
}

// ExtractProjectToComponent extracts the project archive(tar) to the target path from the reader stdin
func (dc *Client) ExtractProjectToComponent(compInfo common.ComponentInfo, targetPath string, stdin io.Reader) error {

	err := dc.Client.CopyToContainer(dc.Context, compInfo.ContainerName, targetPath, stdin, types.CopyToContainerOptions{})
	if err != nil {
		return err
	}
	return nil
}

// WaitForContainer waits for the container until the condition is reached
func (dc *Client) WaitForContainer(containerID string, condition container.WaitCondition) error {

	containerWaitCh, errCh := dc.Client.ContainerWait(dc.Context, containerID, condition)
	for {
		select {
		case containerWait := <-containerWaitCh:
			if containerWait.StatusCode != 0 {
				return errors.Errorf("error waiting on container %s until condition %s; status code: %v, error message: %v", containerID, string(condition), containerWait.StatusCode, containerWait.Error.Message)
			}
			return nil
		case err := <-errCh:
			return errors.Wrapf(err, "unable to wait on container %s until condition %s", containerID, string(condition))
		case <-time.After(2 * time.Minute):
			return errors.Errorf("timeout while waiting for container %s to reach condition %s", containerID, string(condition))
		}
	}
}
