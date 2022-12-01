package podman

import (
	"encoding/json"
	"fmt"
	"os/exec"
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
	out, err := exec.Command(o.podmanCmd, "version", "--format", "json").Output()
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
