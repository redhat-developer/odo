package podman

import "io"

// GetPodLogs returns the logs of the specified pod container.
// All logs for all containers part of the pod are returned if an empty string is provided as container name.
func (o *PodmanCli) GetPodLogs(podName, containerName string, followLog bool) (io.ReadCloser, error) {
	return nil, nil
}
