package podman

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	// it is expected to return in a timely manny (hence this timeout of 1 second).
	// This is to avoid situations like the one described in https://github.com/redhat-developer/odo/issues/6575
	// (where a podman that takes too long to respond affects the "odo dev" command, even if the user did not intend to use the Podman platform).
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctxWithTimeout, o.podmanCmd, "version", "--format", "json")
	klog.V(3).Infof("executing %v", cmd.Args)

	// Because cmd.Output() does not respect the context timeout (see https://github.com/golang/go/issues/57129),
	// we are reading from the connected pipes instead.
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return SystemVersionReport{}, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return SystemVersionReport{}, err
	}

	err = cmd.Start()
	if err != nil {
		return SystemVersionReport{}, err
	}

	var result SystemVersionReport
	go func() {
		// Reading from the pipe is a blocking call, hence this goroutine.
		// The goroutine will exit after the pipe is closed or the command exits;
		// these will be triggered by cmd.Wait() either after the timeout expires or the command finished.
		err = json.NewDecoder(stdoutPipe).Decode(&result)
		if err != nil {
			klog.V(3).Infof("unable to decode output: %v", err)
		}
	}()

	var stderr string
	go func() {
		var buf bytes.Buffer
		_, rErr := buf.ReadFrom(stderrPipe)
		if rErr != nil {
			klog.V(7).Infof("unable to read from stderr pipe: %v", rErr)
		}
		stderr = buf.String()
	}()

	// Wait will block until the timeout expires or the command exits. It will then close all resources associated with cmd,
	// including the stdout and stderr pipes above, which will in turn terminate the goroutines spawned above.
	wErr := cmd.Wait()
	if wErr != nil {
		ctxErr := ctxWithTimeout.Err()
		if ctxErr != nil {
			msg := "error"
			if errors.Is(ctxErr, context.DeadlineExceeded) {
				msg = "timeout"
			}
			wErr = fmt.Errorf("%s while waiting for Podman version: %s: %w", msg, ctxErr, wErr)
		}
		if exitErr, ok := wErr.(*exec.ExitError); ok {
			wErr = fmt.Errorf("%s: %s", wErr, string(exitErr.Stderr))
		}
		if err != nil {
			wErr = fmt.Errorf("%s: (%w)", wErr, err)
		}
		if stderr != "" {
			wErr = fmt.Errorf("%w: %s", wErr, stderr)
		}
		return SystemVersionReport{}, fmt.Errorf("%v. Please check the output of the following command: %v", wErr, cmd.Args)
	}

	return result, nil
}
