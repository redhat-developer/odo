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

	shell := getShellCommand(o.name, image, devfilePath)

	cmd := exec.Command("bash", "-c", shell)
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
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running %s command: %w", o.name, err)
	}

	buildSpinner.End(true)
	return nil
}

func getShellCommand(cmdName string, image *devfile.ImageComponent, devfilePath string) string {
	imageName := image.ImageName
	dockerfile := filepath.Join(devfilePath, image.Dockerfile.Uri)
	buildpath := image.Dockerfile.BuildContext
	if buildpath == "" {
		buildpath = devfilePath
	}
	args := image.Dockerfile.Args

	shell := fmt.Sprintf(`%s build -t "%s" -f "%s" %s`, cmdName, imageName, dockerfile, buildpath)
	if len(args) > 0 {
		shell = shell + " " + strings.Join(args, " ")
	}
	klog.V(4).Infof("Running command: %s", shell)
	return shell
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
