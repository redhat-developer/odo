package podman

import (
	"io"
	"os/exec"

	"k8s.io/klog"
)

// GetPodLogs returns the logs of the specified pod container.
// All logs for all containers part of the pod are returned if an empty string is provided as container name.
func (o *PodmanCli) GetPodLogs(podName, containerName string, followLog bool) (io.ReadCloser, error) {
	// TODO(feloy) implement followLog = true
	args := []string{"pod", "logs"}
	if containerName != "" {
		args = append(args, "--container", podName+"-"+containerName)
	}
	args = append(args, podName)

	cmd := exec.Command(o.podmanCmd, args...)
	klog.V(3).Infof("executing %v", cmd.Args)

	out, _ := cmd.StdoutPipe()
	// We get the commbined output as podman logs outputs logs in stdout && stderr (when kubectl outputs all logs on stdout)
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return out, nil
}
