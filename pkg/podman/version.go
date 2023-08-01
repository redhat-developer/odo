package podman

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

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

// Version returns the version of the Podman client.
func (o *PodmanCli) Version(ctx context.Context) (SystemVersionReport, error) {
	// Because Version is used at the very beginning of odo, when resolving and injecting dependencies (for commands that might require the Podman client),
	// it is expected to return in a timely manner (hence this configurable timeout).
	// This is to avoid situations like the one described in https://github.com/redhat-developer/odo/issues/6575
	// (where a podman CLI that takes too long to respond affects the "odo dev" command, even if the user did not intend to use the Podman platform).

	cmd := exec.CommandContext(ctx, o.podmanCmd, "version", "--format", "json")
	klog.V(3).Infof("executing %v", cmd.Args)

	outbuf, errbuf := new(bytes.Buffer), new(bytes.Buffer)
	cmd.Stdout, cmd.Stderr = outbuf, errbuf

	err := cmd.Start()
	if err != nil {
		return SystemVersionReport{}, err
	}

	// Use a channel to signal completion so we can use a select statement
	done := make(chan error)
	go func() { done <- cmd.Wait() }()

	var timeoutErr error
	select {
	case <-time.After(o.podmanCmdInitTimeout):
		err = cmd.Process.Kill()
		if err != nil {
			klog.V(3).Infof("unable to kill podman version process: %s", err)
		}
		timeoutErr = fmt.Errorf("timeout (%s) while waiting for Podman version", o.podmanCmdInitTimeout.Round(time.Second).String())
		klog.V(3).Infof(timeoutErr.Error())

	case err = <-done:
		if err != nil {
			klog.V(3).Infof("Non-zero exit code for podman version: %v", err)

			stderr := errbuf.String()
			if len(stderr) > 0 {
				klog.V(3).Infof("podman version stderr: %v", stderr)
			}

			return SystemVersionReport{}, err
		}
	}

	var result SystemVersionReport
	err = json.NewDecoder(outbuf).Decode(&result)
	if err != nil {
		klog.V(3).Infof("unable to decode output: %v", err)
		if timeoutErr != nil {
			return SystemVersionReport{}, timeoutErr
		}
		return SystemVersionReport{}, err
	}

	return result, nil
}
