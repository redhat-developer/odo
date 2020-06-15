package lclient

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	volumeTypes "github.com/docker/docker/api/types/volume"
	gomock "github.com/golang/mock/gomock"
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
		ID:    "1",
		Labels: map[string]string{
			"component": "test",
			"alias":     "alias1",
		},
		Mounts: []types.MountPoint{
			{
				Destination: OdoSourceVolumeMount,
			},
		},
	},
	types.Container{
		Names: []string{"/go-test"},
		Image: "golang",
		ID:    "2",
		Labels: map[string]string{
			"component": "golang",
			"8080":      "testurl2",
		},
		HostConfig: container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port("8080/tcp"): []nat.PortBinding{
					nat.PortBinding{
						HostIP:   "127.0.0.1",
						HostPort: "54321",
					},
				},
			},
		},
	},
	types.Container{
		Names: []string{"/go-test-build"},
		Image: "golang",
		ID:    "3",
		Labels: map[string]string{
			"component": "golang",
			"8080":      "testurl3",
			"alias":     "alias1",
		},
		Mounts: []types.MountPoint{
			{
				Destination: OdoSourceVolumeMount,
			},
		},
		HostConfig: container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port("8080/tcp"): []nat.PortBinding{
					nat.PortBinding{
						HostIP:   "127.0.0.1",
						HostPort: "65432",
					},
				},
			},
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
	for _, containerElement := range mockContainerList {
		if containerElement.ID == containerID {
			containerConfig := container.Config{
				Image:  containerElement.Image,
				Labels: containerElement.Labels,
			}
			return types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					HostConfig: &containerElement.HostConfig,
				},
				Config: &containerConfig,
			}, nil
		}
	}
	return types.ContainerJSON{}, nil

}

func (m *mockDockerClient) ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (<-chan container.ContainerWaitOKBody, <-chan error) {
	resultC := make(chan container.ContainerWaitOKBody)
	go func() {
		res := container.ContainerWaitOKBody{
			StatusCode: 0,
			Error: &container.ContainerWaitOKBodyError{
				Message: "",
			},
		}
		resultC <- res
	}()
	return resultC, nil
}

func (m *mockDockerClient) DistributionInspect(ctx context.Context, image, encodedRegistryAuth string) (registry.DistributionInspect, error) {
	return registry.DistributionInspect{}, nil
}

func (m *mockDockerClient) VolumeCreate(ctx context.Context, options volumeTypes.VolumeCreateBody) (types.Volume, error) {
	return types.Volume{
		Name:   options.Name,
		Driver: "local",
		Labels: options.Labels,
	}, nil
}

func (m *mockDockerClient) VolumeList(ctx context.Context, filter filters.Args) (volumeTypes.VolumeListOKBody, error) {
	return volumeTypes.VolumeListOKBody{
		Volumes: []*types.Volume{
			{
				Labels: map[string]string{
					"component": "golang",
				},
			},
			{
				Labels: map[string]string{
					"component": "golang",
				},
			},
			{
				Labels: map[string]string{
					"component": "java",
				},
			},
			{
				Labels: map[string]string{
					"component": "test",
					"type":      "projects",
				},
				Name: "odo-project-source-test",
			},
			{
				Labels: map[string]string{
					"component": "duplicate",
					"type":      "projects",
				},
				Name: "odo-project-source-duplicate1",
			},
			{
				Labels: map[string]string{
					"component": "duplicate",
					"type":      "projects",
				},
				Name: "odo-project-source-duplicate2",
			},
		},
	}, nil
}

func (m *mockDockerClient) VolumeRemove(ctx context.Context, volumeID string, force bool) error {
	return nil
}

func (m *mockDockerClient) ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error) {
	return types.IDResponse{
		ID: "someid",
	}, nil
}

func (m *mockDockerClient) ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error) {
	s1 := strings.NewReader("hello")
	r := bufio.NewReader(s1)
	server, _ := net.Pipe()

	return types.HijackedResponse{
		Reader: r,
		Conn:   server,
	}, nil
}

func (m *mockDockerClient) CopyToContainer(ctx context.Context, container, path string, content io.Reader, options types.CopyToContainerOptions) error {
	return nil
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
var errContainerWait = errors.New("error waiting for container")
var errDistributionInspect = errors.New("error inspecting distribution")
var errVolumeCreate = errors.New("error creating volume")
var errVolumeList = errors.New("error listing volume")
var errRemoveVolume = errors.New("error removing volume")
var errContainerExecCreate = errors.New("error creating container exec")
var errContainerExecAttach = errors.New("error attach container exec")
var errCopyToContainer = errors.New("error copying to container")

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
	resultC := make(chan container.ContainerWaitOKBody)
	err := make(chan error)
	go func() {
		if condition == container.WaitConditionNextExit {
			res := container.ContainerWaitOKBody{
				StatusCode: 1,
				Error: &container.ContainerWaitOKBodyError{
					Message: "bad status code",
				},
			}
			resultC <- res
		}
		err <- errContainerWait
	}()
	if condition == container.WaitConditionNextExit {
		return resultC, nil
	}
	return nil, err
}

func (m *mockDockerErrorClient) DistributionInspect(ctx context.Context, image, encodedRegistryAuth string) (registry.DistributionInspect, error) {
	return registry.DistributionInspect{}, errDistributionInspect
}

func (m *mockDockerErrorClient) VolumeCreate(ctx context.Context, options volumeTypes.VolumeCreateBody) (types.Volume, error) {
	return types.Volume{}, errVolumeCreate
}

func (m *mockDockerErrorClient) VolumeList(ctx context.Context, filter filters.Args) (volumeTypes.VolumeListOKBody, error) {
	return volumeTypes.VolumeListOKBody{}, errVolumeList
}

func (m *mockDockerErrorClient) VolumeRemove(ctx context.Context, volumeID string, force bool) error {
	return errRemoveVolume
}

// FakeNewMockClient returns a fake local client instance that can be used in unit tests
// To regenerate the mock file, in the same directory as mock_client.go, run:
// 'mockgen -source=client.go -package=lclient DockerClient > /tmp/mock_client.go ; cp /tmp/mock_client.go ./mock_client.go'
func FakeNewMockClient(ctrl *gomock.Controller) (*Client, *MockDockerClient) {

	dockerClient := NewMockDockerClient(ctrl)

	localClient := Client{
		Client: dockerClient,
	}
	return &localClient, dockerClient
}

func (m *mockDockerErrorClient) ContainerExecCreate(ctx context.Context, container string, config types.ExecConfig) (types.IDResponse, error) {
	return types.IDResponse{}, errContainerExecCreate
}

func (m *mockDockerErrorClient) ContainerExecAttach(ctx context.Context, execID string, config types.ExecStartCheck) (types.HijackedResponse, error) {
	return types.HijackedResponse{}, errContainerExecAttach
}

func (m *mockDockerErrorClient) CopyToContainer(ctx context.Context, container, path string, content io.Reader, options types.CopyToContainerOptions) error {
	return errCopyToContainer
}
