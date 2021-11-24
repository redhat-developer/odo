// image package provides functions to work with Components of type Image declared in the devfile
package image

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/openshift/odo/pkg/odo/genericclioptions"
)

// Backend is in interface that must be implemented by container runtimes
type Backend interface {
	// Build the image as defined in the devfile
	Build(image *devfile.ImageComponent, devfilePath string) error
	// Push the image to its registry as defined in the devfile
	Push(image string) error
	// Return the name of the backend
	String() string
}

var lookPathCmd = exec.LookPath

// BuildPushImages build all images defined in the devfile with the detected backend
// If push is true, also push the images to their registries
func BuildPushImages(ctx *genericclioptions.Context, push bool) error {

	backend, err := selectBackend()
	if err != nil {
		return err
	}

	devfileObj := ctx.EnvSpecificInfo.GetDevfileObj()
	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{ComponentType: devfile.ImageComponentType},
	})
	if err != nil {
		return err
	}

	devfilePath := filepath.Dir(ctx.EnvSpecificInfo.GetDevfilePath())

	for _, component := range components {
		err = buildPushImage(backend, component.Image, devfilePath, push)
		if err != nil {
			return err
		}
	}
	return nil
}

// BuildPushSpecificImage build an image defined in the devfile
// If push is true, also push the image to its registry
func BuildPushSpecificImage(devfileObj parser.DevfileObj, devfilePath string, component devfile.Component, push bool) error {
	backend, err := selectBackend()
	if err != nil {
		return err
	}

	return buildPushImage(backend, component.Image, devfilePath, push)
}

// buildPushImage build an image using the provided backend
// If push is true, also push the image to its registry
func buildPushImage(backend Backend, image *devfile.ImageComponent, devfilePath string, push bool) error {
	if image == nil {
		return errors.New("image should not be nil")
	}
	err := backend.Build(image, devfilePath)
	if err != nil {
		return err
	}
	if push {
		err = backend.Push(image.ImageName)
		if err != nil {
			return err
		}
	}
	return nil
}

// selectBackend selects the container backend to use for building and pushing images
// It will detect podman and docker CLIs (in this order),
// or return an error if none are present locally
func selectBackend() (Backend, error) {

	podmanCmd := os.Getenv("PODMAN_CMD")
	if podmanCmd == "" {
		podmanCmd = "podman"
	}
	if _, err := lookPathCmd(podmanCmd); err == nil {
		return NewDockerCompatibleBackend(podmanCmd), nil
	}

	dockerCmd := os.Getenv("DOCKER_CMD")
	if dockerCmd == "" {
		dockerCmd = "docker"
	}
	if _, err := lookPathCmd(dockerCmd); err == nil {
		return NewDockerCompatibleBackend(dockerCmd), nil
	}
	return nil, errors.New("no backend found")
}
