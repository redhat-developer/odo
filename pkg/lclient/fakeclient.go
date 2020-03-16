/*******************************************************************************
 * Copyright (c) 2020 IBM Corporation and others.
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * which accompanies this distribution, and is available at
 * http://www.eclipse.org/legal/epl-v20.html
 *
 * Contributors:
 *     IBM Corporation - initial API and implementation
 *******************************************************************************/

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
	return container.ContainerCreateCreatedBody{}, nil
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

/*func TestPullImage(t *testing.T) {
	t.Run("does not error when docker ImagePull succeeds", func(t *testing.T) {
		client := &mockDockerClient{}
		err := PullImage(client, "dummyImage", true)
		assert.Nil(t, err)
	})

	t.Run("returns DockerError when docker ImagePull errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		err := PullImage(client, "dummyImage", true)
		wantErr := &DockerError{errOpImagePull, errImagePull, errImagePull.Error()}
		assert.Equal(t, wantErr, err)
	})
}*/

/*func TestGetImageList(t *testing.T) {
	t.Run("gets the image list that is returned by the docker client", func(t *testing.T) {
		client := &mockDockerClientWithCw{}

		imageList, err := GetImageList(client)
		assert.Nil(t, err)
		assert.Equal(t, imageList, mockImageSummaryWithCwImages)
	})

	t.Run("returns DockerError when docker ImageList errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		_, err := GetImageList(client)
		wantErr := &DockerError{errOpImageList, errImageList, errImageList.Error()}
		assert.Equal(t, wantErr, err)
	})
}

func TestGetContainerList(t *testing.T) {
	t.Run("gets the container list that is returned by the docker client", func(t *testing.T) {
		client := &mockDockerClientWithCw{}

		containerList, err := GetContainerList(client)
		assert.Nil(t, err)
		assert.Equal(t, containerList, mockContainerListWithCwContainers)
	})

	t.Run("returns error when docker ContainerList returns error", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		_, err := GetContainerList(client)
		wantErr := &DockerError{errOpContainerList, errContainerList, errContainerList.Error()}
		assert.Equal(t, wantErr, err)
	})
}

func TestCheckImageStatus(t *testing.T) {
	t.Run("returns true when correct images are returned by the docker client", func(t *testing.T) {
		client := &mockDockerClientWithCw{}

		imageStatus, err := CheckImageStatus(client)
		assert.Nil(t, err)
		assert.True(t, imageStatus)
	})

	t.Run("returns false when codewind images are not returned by the docker client", func(t *testing.T) {
		client := &mockDockerClientWithoutCw{}

		imageStatus, err := CheckImageStatus(client)
		assert.Nil(t, err)
		assert.False(t, imageStatus)
	})

	t.Run("returns DockerError when docker ImageList errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		_, err := CheckImageStatus(client)
		wantErr := &DockerError{errOpImageList, errImageList, errImageList.Error()}
		assert.Equal(t, wantErr, err)
	})
}

func TestCheckContainerStatus(t *testing.T) {
	t.Run("returns true when correct containers are returned by the docker client", func(t *testing.T) {
		client := &mockDockerClientWithCw{}

		containerStatus, err := CheckContainerStatus(client)
		assert.Nil(t, err)
		assert.True(t, containerStatus)
	})

	t.Run("returns false when correct codewind containers are not returned by the docker client", func(t *testing.T) {
		client := &mockDockerClientWithoutCw{}

		containerStatus, err := CheckContainerStatus(client)
		assert.Nil(t, err)
		assert.False(t, containerStatus)
	})

	t.Run("returns DockerError when docker ContainerList errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		_, err := CheckContainerStatus(client)
		wantErr := &DockerError{errOpContainerList, errContainerList, errContainerList.Error()}
		assert.Equal(t, wantErr, err)
	})
}

func TestGetImageTags(t *testing.T) {
	t.Run("returns the image tags set in the ImageList mock", func(t *testing.T) {
		client := &mockDockerClientWithCw{}

		imageTags, err := GetImageTags(client)
		assert.Nil(t, err)
		assert.Equal(t, []string{"0.0.9"}, imageTags)
	})

	t.Run("returns DockerError when docker ImageList errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		_, err := GetImageTags(client)
		wantErr := &DockerError{errOpImageList, errImageList, errImageList.Error()}
		assert.Equal(t, wantErr, err)
	})
}

func TestGetContainerTags(t *testing.T) {
	t.Run("returns the container tags set in the ContainerList mock", func(t *testing.T) {
		client := &mockDockerClientWithCw{}

		imageTags, err := GetContainerTags(client)
		assert.Nil(t, err)
		assert.Equal(t, []string{"0.0.9"}, imageTags)
	})

	t.Run("returns DockerError when docker ContainerList errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		_, err := GetContainerTags(client)
		wantErr := &DockerError{errOpContainerList, errContainerList, errContainerList.Error()}
		assert.Equal(t, err, wantErr)
	})
}

func TestGetPFEHostAndPort(t *testing.T) {
	t.Run("returns the PFE host and port set in the ContainerList mock", func(t *testing.T) {
		client := &mockDockerClientWithCw{}

		host, port, err := GetPFEHostAndPort(client)
		assert.Nil(t, err)
		assert.Equal(t, "pfe", host)
		assert.Equal(t, "1000", port)
	})

	t.Run("returns DockerError when docker ContainerList errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		_, _, err := GetPFEHostAndPort(client)
		wantErr := &DockerError{errOpContainerList, errContainerList, errContainerList.Error()}
		assert.Equal(t, wantErr, err)
	})
}

func TestValidateImageDigest(t *testing.T) {
	t.Run("no error returned when image digests match those from dockerhub", func(t *testing.T) {
		client := &mockDockerClientWithCw{}
		_, err := ValidateImageDigest(client, "test:0.0.9")
		assert.Nil(t, err)
	})

	t.Run("returns DockerError when docker ImageList errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		_, err := ValidateImageDigest(client, "test:0.0.9")
		wantErr := &DockerError{errOpImageList, errImageList, errImageList.Error()}
		assert.Equal(t, wantErr, err)
	})
}

func TestGetAutoRemovePolicy(t *testing.T) {
	t.Run("no error returned when image digests match those from dockerhub", func(t *testing.T) {
		client := &mockDockerClientWithCw{}
		autoremovePolicy, err := getContainerAutoRemovePolicy(client, "pfe")
		assert.Nil(t, err)
		assert.True(t, autoremovePolicy)
	})

	t.Run("returns DockerError when docker ImageList errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		_, err := getContainerAutoRemovePolicy(client, "pfe")
		wantErr := &DockerError{errOpContainerInspect, errContainerInspect, errContainerInspect.Error()}
		assert.Equal(t, wantErr, err)
	})
}

func TestStopContainer(t *testing.T) {
	t.Run("no error returned when container is stopped", func(t *testing.T) {
		client := &mockDockerClientWithCw{}
		err := StopContainer(client, types.Container{
			Names: []string{"/codewind-pfe"},
			ID:    "pfe",
			Image: "eclipse/codewind-pfe:0.0.9",
			Ports: []types.Port{types.Port{PrivatePort: 9090, PublicPort: 1000, IP: "pfe"}}})
		assert.Nil(t, err)
	})

	t.Run("returns DockerError when docker ContainerStop errors", func(t *testing.T) {
		client := &mockDockerErrorClient{}
		err := StopContainer(client, types.Container{})
		containerInspectErr := &DockerError{errOpContainerInspect, errContainerInspect, errContainerInspect.Error()}
		wantErr := &DockerError{errOpStopContainer, containerInspectErr, containerInspectErr.Desc}
		assert.Equal(t, wantErr, err)
	})
}

func TestGetContainersToRemove(t *testing.T) {
	tests := map[string]struct {
		containerList      []types.Container
		expectedContainers []string
	}{
		"Returns project containers (cw-)": {
			containerList: []types.Container{
				types.Container{
					Names: []string{"/cw-nodejsexpress"},
				},
				types.Container{
					Names: []string{"/cw-springboot"},
				},
			},
			expectedContainers: []string{
				"/cw-nodejsexpress",
				"/cw-springboot",
			},
		},
		"Ignores a non-codewind container": {
			containerList: []types.Container{
				types.Container{
					Names: []string{"/cw-valid-container"},
				},
				types.Container{
					Names: []string{"invalid-container"},
				},
			},
			expectedContainers: []string{
				"/cw-valid-container",
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			containersToRemove := GetContainersToRemove(test.containerList)
			assert.Equal(t, len(test.expectedContainers), len(containersToRemove))
			for _, container := range containersToRemove {
				assert.Contains(t, test.expectedContainers, container.Names[0])
			}
		})
	}
}

func TestRemoveDuplicateEntries(t *testing.T) {
	dupArr := []string{"test", "test", "test"}
	result := RemoveDuplicateEntries(dupArr)

	if len(result) != 1 {
		log.Fatal("Test 1: Failed to delete duplicate array index")
	}

	dupArr = []string{"", "test", "test"}
	result = RemoveDuplicateEntries(dupArr)
	if len(result) != 1 {
		log.Fatal("Test 2: Failed to delete duplicate array index")
	}

	dupArr = []string{"", "", ""}
	result = RemoveDuplicateEntries(dupArr)
	if len(result) != 0 {
		log.Fatal("Test 3: Failed to identify empty array values")
	}
}*/
