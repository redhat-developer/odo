package podman

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"k8s.io/klog"
)

type Version struct {
	APIVersion string
	Version    string
	GoVersion  string
	GitCommit  string
	BuiltTime  string
	Built      int64
	OsArch     string
	Os         string
}

type SystemVersionReport struct {
	Client *Version `json:",omitempty"`
}

func (o *PodmanCli) Version() (SystemVersionReport, error) {
	cmd := exec.Command(o.podmanCmd, "version", "--format", "json")
	klog.V(3).Infof("executing %v", cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		return SystemVersionReport{}, err
	}
	var result SystemVersionReport
	err = json.Unmarshal(out, &result)
	if err != nil {
		return SystemVersionReport{}, err
	}
	return result, nil
}
