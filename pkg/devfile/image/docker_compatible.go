package image

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/openshift/odo/pkg/log"
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

	log.Infof("\nBuilding image %s", image.ImageName)

	shell := getShellCommand(o.name, image, devfilePath)

	cmd := exec.Command("bash", "-c", shell)
	cmd.Env = append(os.Environ(), "PROJECT_ROOT="+devfilePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running %s command: %w", o.name, err)
	}
	return nil
}

func getShellCommand(cmdName string, image *devfile.ImageComponent, devfilePath string) string {
	imageName := image.ImageName
	dockerfile := filepath.Join(devfilePath, image.Dockerfile.Uri)
	buildpath := image.Dockerfile.BuildContext
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
	log.Infof("\nPushing image %s", image)
	klog.V(4).Infof("Running command: %s push %s", o.name, image)
	cmd := exec.Command(o.name, "push", image)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running %s command: %w", o.name, err)
	}
	return nil
}

// String return the name of the docker compatible CLI used
func (o *DockerCompatibleBackend) String() string {
	return o.name
}
