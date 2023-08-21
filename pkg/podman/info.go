package podman

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"k8s.io/klog"
)

// podmanInfo originates from https://github.com/containers/podman/blob/main/libpod/define/info.go
type podmanInfo struct {
	Host *HostInfo `json:"host"`
}

type HostInfo struct {
	CgroupsVersion string `json:"cgroupVersion"`
}

func (o *PodmanCli) getInfo() (podmanInfo, error) {
	cmd := exec.Command(o.podmanCmd, append(o.containerRunGlobalExtraArgs, "info", "--format", "json")...)
	klog.V(3).Infof("executing %v", cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		return podmanInfo{}, err
	}

	var result podmanInfo

	err = json.Unmarshal(out, &result)
	return result, err
}
