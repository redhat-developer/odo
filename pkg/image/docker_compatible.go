package image

import (
	"fmt"
	"os/exec"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
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
func (o *DockerCompatibleBackend) Build(image *devfile.ImageComponent) error {
	imageName := image.ImageName
	dockerfile := image.Dockerfile.Uri
	buildpath := image.Dockerfile.BuildContext
	shell := fmt.Sprintf("%s build -t %s -f %s %s", o.name, imageName, dockerfile, buildpath)
	klog.V(4).Infof("Running command: %s", shell)
	cmd := exec.Command("bash", "-c", shell)
	output, err := cmd.CombinedOutput()
	klog.V(4).Infoln(string(output))
	if err != nil {
		return fmt.Errorf("error running %s command: %w", o.name, err)
	}
	return nil
}

// Push an image to its registry using a Docker compatible CLI
func (o *DockerCompatibleBackend) Push(image string) error {
	klog.V(4).Infof("Running command: %s push %s", o.name, image)
	cmd := exec.Command(o.name, "push", image)
	output, err := cmd.CombinedOutput()
	klog.V(4).Infoln(string(output))
	if err != nil {
		return fmt.Errorf("error running %s command: %w", o.name, err)
	}
	return nil
}

// String return the name of the docker compatible CLI used
func (o *DockerCompatibleBackend) String() string {
	return o.name
}
