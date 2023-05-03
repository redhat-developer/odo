package image

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/fatih/color"
	"k8s.io/klog"

	dfutil "github.com/devfile/library/v2/pkg/util"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

// DockerCompatibleBackend uses a CLI compatible with the docker CLI (at least docker itself and podman)
type DockerCompatibleBackend struct {
	name                string
	globalExtraArgs     []string
	imageBuildExtraArgs []string
}

var _ Backend = (*DockerCompatibleBackend)(nil)

func NewDockerCompatibleBackend(name string, globalExtraArgs, imageBuildExtraArgs []string) *DockerCompatibleBackend {
	return &DockerCompatibleBackend{
		name:                name,
		globalExtraArgs:     globalExtraArgs,
		imageBuildExtraArgs: imageBuildExtraArgs,
	}
}

// Build an image, as defined in devfile, using a Docker compatible CLI
func (o *DockerCompatibleBackend) Build(fs filesystem.Filesystem, image *devfile.ImageComponent, devfilePath string) error {

	dockerfile, isTemp, err := resolveAndDownloadDockerfile(fs, image.Dockerfile.Uri)
	if isTemp {
		defer func(path string) {
			if e := fs.Remove(path); e != nil {
				klog.V(3).Infof("could not remove temporary Dockerfile at path %q: %v", path, err)
			}
		}(dockerfile)
	}
	if err != nil {
		return err
	}

	// We use a "No Spin" since we are outputting to stdout / stderr
	buildSpinner := log.SpinnerNoSpin("Building image locally")
	defer buildSpinner.End(false)

	err = os.Setenv("PROJECTS_ROOT", devfilePath)
	if err != nil {
		return err
	}

	err = os.Setenv("PROJECT_SOURCE", devfilePath)
	if err != nil {
		return err
	}

	shellCmd := getShellCommand(o.name, o.globalExtraArgs, o.imageBuildExtraArgs, image, devfilePath, dockerfile)
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

// resolveAndDownloadDockerfile resolves and downloads (if needed) the specified Dockerfile URI.
// For now, it only supports resolving HTTP(S) URIs, in which case it downloads the remote file
// to a temporary file. The path to that temporary file is then returned.
//
// In all other cases, the specified URI path is returned as is.
// This means that non-HTTP(S) URIs will *not* get resolved, but will be returned as is.
//
// In addition to the path, a boolean and a potential error are returned. The boolean indicates whether
// the returned path is a temporary one; in such case, it is the caller's responsibility to delete this file
// once it is done working with it.
func resolveAndDownloadDockerfile(fs filesystem.Filesystem, uri string) (string, bool, error) {
	uriLower := strings.ToLower(uri)
	if strings.HasPrefix(uriLower, "http://") || strings.HasPrefix(uriLower, "https://") {
		s := log.Spinner("Downloading Dockerfile")
		defer s.End(false)
		tempFile, err := fs.TempFile("", "odo_*.dockerfile")
		if err != nil {
			return "", false, err
		}
		dockerfile := tempFile.Name()
		err = dfutil.DownloadFile(dfutil.DownloadParams{
			Request: dfutil.HTTPRequestParams{
				URL: uri,
			},
			Filepath: dockerfile,
		})
		s.End(err == nil)
		return dockerfile, true, err
	}
	return uri, false, nil
}

// getShellCommand creates the docker compatible build command from detected backend,
// container image and devfile path
func getShellCommand(cmdName string, globalExtraArgs []string, buildExtraArgs []string, image *devfile.ImageComponent, devfilePath string, dockerfilePath string) []string {
	imageName := image.ImageName
	dockerfile := dockerfilePath
	if !filepath.IsAbs(dockerfile) {
		dockerfile = filepath.Join(devfilePath, dockerfilePath)
	}
	buildpath := image.Dockerfile.BuildContext
	if buildpath == "" {
		buildpath = devfilePath
	}

	// +7 because of the other args
	shellCmd := make([]string, 0, len(globalExtraArgs)+len(buildExtraArgs)+len(image.Dockerfile.Args)+7)
	shellCmd = append(shellCmd, cmdName)
	shellCmd = append(shellCmd, globalExtraArgs...)
	shellCmd = append(shellCmd, "build")
	shellCmd = append(shellCmd, buildExtraArgs...)
	shellCmd = append(shellCmd, "-t", imageName, "-f", dockerfile, buildpath)

	if len(image.Dockerfile.Args) != 0 {
		shellCmd = append(shellCmd, image.Dockerfile.Args...)
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
