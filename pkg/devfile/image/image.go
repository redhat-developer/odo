// image package provides functions to work with Components of type Image declared in the devfile
package image

import (
	"context"
	"errors"
	"os/exec"
	"path/filepath"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"

	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

// Backend is in interface that must be implemented by container runtimes
type Backend interface {
	// Build the image as defined in the devfile.
	// The filesystem specified will be used to download and store the Dockerfile if it is referenced as a remote URL.
	Build(fs filesystem.Filesystem, image *devfile.ImageComponent, devfilePath string) error
	// Push the image to its registry as defined in the devfile
	Push(image string) error
	// Return the name of the backend
	String() string
}

var lookPathCmd = exec.LookPath

// BuildPushImages build all images defined in the devfile with the detected backend
// If push is true, also push the images to their registries
func BuildPushImages(ctx context.Context, fs filesystem.Filesystem, push bool) error {
	var (
		devfileObj  = odocontext.GetDevfileObj(ctx)
		devfilePath = odocontext.GetDevfilePath(ctx)
		path        = filepath.Dir(devfilePath)
	)

	backend, err := selectBackend(ctx)
	if err != nil {
		return err
	}

	components, err := devfileObj.Data.GetComponents(common.DevfileOptions{
		ComponentOptions: common.ComponentOptions{ComponentType: devfile.ImageComponentType},
	})
	if err != nil {
		return err
	}
	if len(components) == 0 {
		return libdevfile.NewComponentTypeNotFoundError(devfile.ImageComponentType)
	}

	for _, component := range components {
		err = buildPushImage(backend, fs, component.Image, path, push)
		if err != nil {
			return err
		}
	}
	return nil
}

// BuildPushSpecificImage build an image defined in the devfile present in devfilePath
// If push is true, also push the image to its registry
func BuildPushSpecificImage(ctx context.Context, fs filesystem.Filesystem, component devfile.Component, push bool) error {
	var (
		devfilePath = odocontext.GetDevfilePath(ctx)
		path        = filepath.Dir(devfilePath)
	)
	backend, err := selectBackend(ctx)
	if err != nil {
		return err
	}
	return buildPushImage(backend, fs, component.Image, path, push)
}

// buildPushImage build an image using the provided backend
// If push is true, also push the image to its registry
func buildPushImage(backend Backend, fs filesystem.Filesystem, image *devfile.ImageComponent, devfilePath string, push bool) error {
	if image == nil {
		return errors.New("image should not be nil")
	}
	log.Sectionf("Building & Pushing Container: %s", image.ImageName)
	err := backend.Build(fs, image, devfilePath)
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
func selectBackend(ctx context.Context) (Backend, error) {

	podmanCmd := envcontext.GetEnvConfig(ctx).PodmanCmd
	if _, err := lookPathCmd(podmanCmd); err == nil {

		// Podman does NOT build x86 images on Apple Silicon / M1 and we must *WARN* the user that this will not work.
		// There is a temporary workaround in order to build x86 images on Apple Silicon / M1 by running the following commands:
		// podman machine ssh sudo rpm-ostree install qemu-user-static
		// podman machine ssh sudo systemctl reboot
		//
		// The problem is that Fedora CoreOS does not have qemu-user-static installed by default,
		// and the workaround is to install it manually as the dependencies need to be integrated into the Fedora ecosystem
		// The open discussion is here: https://github.com/containers/podman/discussions/12899
		//
		// TODO: Remove this warning when Podman natively supports x86 images on Apple Silicon / M1.
		if log.IsAppleSilicon() {
			log.Warning("WARNING: Building images on Apple Silicon / M1 is not (yet) supported natively on Podman")
			log.Warning("There is however a temporary workaround: https://github.com/containers/podman/discussions/12899")
		}
		return NewDockerCompatibleBackend(podmanCmd), nil
	}

	dockerCmd := envcontext.GetEnvConfig(ctx).DockerCmd
	if _, err := lookPathCmd(dockerCmd); err == nil {
		return NewDockerCompatibleBackend(dockerCmd), nil
	}
	//revive:disable:error-strings This is a top-level error message displayed as is to the end user
	return nil, errors.New("odo requires either Podman or Docker to be installed in your environment. Please install one of them and try again.")
	//revive:enable:error-strings
}
