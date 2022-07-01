package image

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/fatih/color"
	"github.com/redhat-developer/odo/pkg/log"
	"k8s.io/klog"
)

// This backend uses a CLI compatible with the docker CLI (at least docker itself and podman)
type DockerCompatibleBackend struct {
	name string
}

var _ Backend = (*DockerCompatibleBackend)(nil)

func NewDockerCompatibleBackend(name string) *DockerCompatibleBackend {
	return &DockerCompatibleBackend{name: name}
}

// Build an image, as defined in devfile, using a Docker compatible CLI
func (o *DockerCompatibleBackend) Build(image *devfile.ImageComponent, devfilePath string) error {

	if strings.HasPrefix(image.Dockerfile.Uri, "http") {
		return errors.New("HTTP URL for uri is not supported")
	}

	// We use a "No Spin" since we are outputting to stdout / stderr
	buildSpinner := log.SpinnerNoSpin("Building image locally")
	defer buildSpinner.End(false)

	err := os.Setenv("PROJECTS_ROOT", devfilePath)
	if err != nil {
		return err
	}

	err = os.Setenv("PROJECT_SOURCE", devfilePath)
	if err != nil {
		return err
	}

	shellCmd := getShellCommand(o.name, image, devfilePath)
	klog.V(4).Infof("Running command: %v", shellCmd)
	for i, cmd := range shellCmd {
		shellCmd[i] = os.ExpandEnv(cmd)
	}
	cmd := exec.Command(shellCmd[0], shellCmd[1:]...)
	cmdEnv := []string{
		"PROJECTS_ROOT=" + devfilePath,
		"PROJECT_SOURCE=" + devfilePath,
	}
	cmd.Env = append(os.Environ(), cmdEnv...)
	cmd.Stdout = log.GetStdout()
	cmd.Stderr = log.GetStderr()

	// Set all output as italic when doing a push, then return to normal at the end
	color.Set(color.Italic)
	defer color.Unset()
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("error running %s command: %w", o.name, err)
	}

	buildSpinner.End(true)
	return nil
}

//getShellCommand creates the docker compatible build command from detected backend,
//container image and devfile path
func getShellCommand(cmdName string, image *devfile.ImageComponent, devfilePath string) []string {
	var shellCmd []string
	imageName := image.ImageName
	dockerfile := filepath.Join(devfilePath, image.Dockerfile.Uri)
	buildpath := image.Dockerfile.BuildContext
	if buildpath == "" {
		buildpath = devfilePath
	}
	args := image.Dockerfile.Args
	shellCmd = []string{
		cmdName,
		"build",
		"-t",
		imageName,
		"-f",
		dockerfile,
		buildpath,
	}
	if len(args) > 0 {
		shellCmd = append(shellCmd, args...)
	}
	return shellCmd
}

// Push an image to its registry using a Docker compatible CLI
func (o *DockerCompatibleBackend) Push(image string) error {

	// We use a "No Spin" since we are outputting to stdout / stderr
	pushSpinner := log.SpinnerNoSpin("Pushing image to container registry")
	defer pushSpinner.End(false)
	klog.V(4).Infof("Running command: %s push %s", o.name, image)

	cmd := exec.Command(o.name, "push", image)

	cmd.Stdout = log.GetStdout()
	cmd.Stderr = log.GetStderr()

	// Set all output as italic when doing a push, then return to normal at the end
	color.Set(color.Italic)
	defer color.Unset()
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running %s command: %w", o.name, err)
	}

	pushSpinner.End(true)
	return nil
}

// String return the name of the docker compatible CLI used
func (o *DockerCompatibleBackend) String() string {
	return o.name
}
