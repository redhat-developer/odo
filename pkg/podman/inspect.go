package podman

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"k8s.io/klog"
)

// PodInspectData originates from From https://github.com/containers/podman/blob/main/libpod/define/pod_inspect.go
type PodInspectData struct {
	// ID is the ID of the pod.
	ID string `json:"Id"`
	// Name is the name of the pod.
	Name string
	// Namespace is the Libpod namespace the pod is placed in.
	Namespace string `json:"Namespace,omitempty"`
	// State represents the current state of the pod.
	State string `json:"State"`
	// Labels is a set of key-value labels that have been applied to the
	// pod.
	Labels map[string]string `json:"Labels,omitempty"`
}

func (o *PodmanCli) PodInspect(podname string) (PodInspectData, error) {
	cmd := exec.Command(o.podmanCmd, "pod", "inspect", podname, "--format", "json")
	klog.V(3).Infof("executing %v", cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		return PodInspectData{}, err
	}

	var result PodInspectData

	err = json.Unmarshal(out, &result)
	return result, err
}
