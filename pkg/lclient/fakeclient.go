package lclient

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	volumeTypes "github.com/docker/docker/api/types/volume"
)

// This mock client will return container and images lists
// By creating stubs, we're able to easily mock the Docke client for both success and error scenarios
// Taken from https://github.com/eclipse/codewind-installer/blob/master/pkg/docker/docker_test.go
type mockDockerClient struct {
}

var mockImageSummary = []types.ImageSummary{
	types.ImageSummary{
		ID:       "golang",
		RepoTags: []string{"golang:0.0.9"},
	},
	types.ImageSummary{
		ID:       "node",
		RepoTags: []string{"node:0.0.9"},
	},
}

var mockContainerList = []types.Container{
	types.Container{
		Names: []string{"/node"},
		Image: "node",
		Labels: map[string]string{
			"component": "node",
		},
	},
	types.Container{
		Names: []string{"/go-test"},
		Image: "golang",
		Labels: map[string]string{
			"component": "golang",
		},
	},
	types.Container{
		Names: []string{"/go-test-build"},
		Image: "golang",
		Labels: map[string]string{
			"component": "golang",
		},
	},
}

var mockContainerCreatedBody = container.ContainerCreateCreatedBody{
	ID: "golang",
}

// FakeNew returns a fake local client instance that can be used in unit tests
func FakeNew() *Client {
	dockerClient := &mockDockerClient{}
	localClient := Client{
		Client: dockerClient,
	}
	return &localClient
}

func (m *mockDockerClient) ImagePull(ctx context.Context, image string, imagePullOptions types.ImagePullOptions) (io.ReadCloser, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte("")))
	return r, nil
}

func (m *mockDockerClient) ImageList(ctx context.Context, imageListOptions types.ImageListOptions) ([]types.ImageSummary, error) {
	return mockImageSummary, nil
}

func (m *mockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error) {
	return mockContainerCreatedBody, nil
}

func (m *mockDockerClient) ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error {
	return nil
}

func (m *mockDockerClient) ContainerList(ctx context.Context, containerListOptions types.ContainerListOptions) ([]types.Container, error) {
	return mockContainerList, nil
}

func (m *mockDockerClient) ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error {
	return nil
}

func (m *mockDockerClient) ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error {
	return nil
}

func (m *mockDockerClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return types.ContainerJSON{
		ContainerJSONBase: &types.ContainerJSONBase{
			HostConfig: &container.HostConfig{
				AutoRemove: true,
			},
		},
	}, nil
}

func (m *mockDockerClient) ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.ContainerWaitOKBody, <-chan error) {
	resultC := make(chan container.ContainerWaitOKBody)
	return resultC, nil
}

func (m *mockDockerClient) DistributionInspect(ctx context.Context, image, encodedRegistryAuth string) (registry.DistributionInspect, error) {
	return registry.DistributionInspect{}, nil
}

func (m *mockDockerClient) VolumeCreate(ctx context.Context, options volumeTypes.VolumeCreateBody) (types.Volume, error) {
	return types.Volume{
		Driver: "overlayfs2",
		Labels: map[string]string{
			"component": "golang",
		},
	}, nil
}

// This mock client will return errors for each call to a docker function
type mockDockerErrorClient struct {
}

// FakeErrorNew returns a fake local client instance that can be used in unit tests to verify errors
func FakeErrorNew() *Client {
	dockerClient := &mockDockerErrorClient{}
	localClient := Client{
		Client: dockerClient,
	}
	return &localClient
}

var errImagePull = errors.New("error pulling image")
var errImageList = errors.New("error listing images")
var errContainerCreate = errors.New("error creating containers")
var errContainerList = errors.New("error listing containers")
var errContainerStart = errors.New("error starting containers")
var errContainerStop = errors.New("error stopping container")
var errContainerRemove = errors.New("error removing container")
var errContainerInspect = errors.New("error inspecting container")
var errContainerWait = errors.New("error timeout waiting for container")
var errDistributionInspect = errors.New("error inspecting distribution")
var errVolumeCreate = errors.New("error creating volume")

func (m *mockDockerErrorClient) ImageList(ctx context.Context, imageListOptions types.ImageListOptions) ([]types.ImageSummary, error) {
	return nil, errImageList
}

func (m *mockDockerErrorClient) ImagePull(ctx context.Context, image string, imagePullOptions types.ImagePullOptions) (io.ReadCloser, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte("")))
	return r, errImagePull
}

func (m *mockDockerErrorClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error) {
	return container.ContainerCreateCreatedBody{}, errContainerCreate
}

func (m *mockDockerErrorClient) ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error {
	return errContainerStart
}

func (m *mockDockerErrorClient) ContainerList(ctx context.Context, containerListOptions types.ContainerListOptions) ([]types.Container, error) {
	return []types.Container{}, errContainerList
}

func (m *mockDockerErrorClient) ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error {
	return errContainerStop
}

func (m *mockDockerErrorClient) ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error {
	return errContainerRemove
}

func (m *mockDockerErrorClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	return types.ContainerJSON{}, errContainerInspect
}

func (m *mockDockerErrorClient) ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.ContainerWaitOKBody, <-chan error) {
	err := make(chan error)
	err <- errContainerWait
	return nil, err
}

func (m *mockDockerErrorClient) DistributionInspect(ctx context.Context, image, encodedRegistryAuth string) (registry.DistributionInspect, error) {
	return registry.DistributionInspect{}, errDistributionInspect
}

func (m *mockDockerErrorClient) VolumeCreate(ctx context.Context, options volumeTypes.VolumeCreateBody) (types.Volume, error) {
	return types.Volume{}, errVolumeCreate
}
