package lclient

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	gomock "github.com/golang/mock/gomock"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
)

func TestGetContainersByComponentName(t *testing.T) {
	fakeClient := FakeNew()
	fakeErrorClient := FakeErrorNew()

	tests := []struct {
		name           string
		client         *Client
		component      string
		containers     []types.Container
		wantContainers []types.Container
		wantErr        bool
	}{
		{
			name:      "Case 1: Successfully retrieve one container and have proper component",
			client:    fakeClient,
			component: "node",
			containers: []types.Container{
				{
					Names: []string{"/tester"},
					Image: "tester",
					Labels: map[string]string{
						"component": "tester",
					},
				},
				{
					Names: []string{"/node"},
					Image: "node",
					Labels: map[string]string{
						"component": "node",
					},
				},
			},
			wantContainers: []types.Container{
				{
					Names: []string{"/node"},
					Image: "node",
					Labels: map[string]string{
						"component": "node",
					},
				},
			},
		},
		{
			name:      "Case 2: Invalid component name",
			client:    fakeClient,
			component: "fake-component",
			containers: []types.Container{
				{
					Names: []string{"/go-test"},
					Image: "golang",
					Labels: map[string]string{
						"component": "golang",
					},
				},
				{
					Names: []string{"/node-build"},
					Image: "node",
					Labels: map[string]string{
						"component": "node",
					},
				},
				{
					Names: []string{"/go-test-build"},
					Image: "golang",
					Labels: map[string]string{
						"component": "golang",
					},
				},
			},
			wantContainers: nil,
		},
		{
			name:           "Case 3: Error listing containers",
			client:         fakeErrorClient,
			component:      "fake-component",
			containers:     nil,
			wantContainers: nil,
		},
		{
			name:      "Case 4: Multiple components returned",
			client:    fakeClient,
			component: "golang",
			containers: []types.Container{
				{
					Names: []string{"/go-test"},
					Image: "golang",
					Labels: map[string]string{
						"component": "golang",
					},
				},
				{
					Names: []string{"/node-build"},
					Image: "node",
					Labels: map[string]string{
						"component": "node",
					},
				},
				{
					Names: []string{"/go-test-build"},
					Image: "golang",
					Labels: map[string]string{
						"component": "golang",
					},
				},
			},
			wantContainers: []types.Container{
				{
					Names: []string{"/go-test"},
					Image: "golang",
					Labels: map[string]string{
						"component": "golang",
					},
				},
				{
					Names: []string{"/go-test-build"},
					Image: "golang",
					Labels: map[string]string{
						"component": "golang",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		containers := tt.client.GetContainersByComponent(tt.component, tt.containers)

		if !reflect.DeepEqual(tt.wantContainers, containers) {
			t.Errorf("Expected %v, got %v", tt.wantContainers, containers)
		}
	}
}

func TestGetContainersList(t *testing.T) {
	fakeClient := FakeNew()
	fakeErrorClient := FakeErrorNew()

	tests := []struct {
		name           string
		client         *Client
		component      string
		wantContainers []types.Container
		wantErr        bool
	}{
		{
			name:      "Case 1: Successfully retrieve the container list",
			client:    fakeClient,
			component: "node",
			wantContainers: []types.Container{
				{
					Names: []string{"/node"},
					Image: "node",
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
				{
					Names: []string{"/go-test"},
					Image: "golang",
					Labels: map[string]string{
						"component": "golang",
					},
				},
				{
					Names: []string{"/go-test-build"},
					Image: "golang",
					Labels: map[string]string{
						"component": "golang",
					},
				},
			},
			wantErr: false,
		},
		{
			name:           "Case 2: Error retrieving the container list",
			client:         fakeErrorClient,
			wantContainers: nil,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			containers, err := tt.client.GetContainerList()

			if !tt.wantErr == (err != nil) {
				t.Errorf("expected %v, wanted %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(tt.wantContainers, containers) {
				t.Errorf("Expected %v, got %v", tt.wantContainers, containers)
			}
		})
	}
}

func TestStartContainer(t *testing.T) {
	fakeClient := FakeNew()
	fakeErrorClient := FakeErrorNew()

	fakeContainer := container.Config{}
	tests := []struct {
		name    string
		client  *Client
		wantErr bool
	}{
		{
			name:    "Case 1: Successfully start container",
			client:  fakeClient,
			wantErr: false,
		},
		{
			name:    "Case 2: Fail to start",
			client:  fakeErrorClient,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.StartContainer(&fakeContainer, nil, nil)
			if !tt.wantErr == (err != nil) {
				t.Errorf("expected %v, wanted %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveContainer(t *testing.T) {
	fakeClient := FakeNew()
	fakeErrorClient := FakeErrorNew()

	fakeContainerID := "golang"
	tests := []struct {
		name    string
		client  *Client
		wantErr bool
	}{
		{
			name:    "Case 1: Successfully remove container",
			client:  fakeClient,
			wantErr: false,
		},
		{
			name:    "Case 2: Fail to remove container",
			client:  fakeErrorClient,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.RemoveContainer(fakeContainerID)
			if !tt.wantErr == (err != nil) {
				t.Errorf("expected %v, wanted %v", err, tt.wantErr)
			}
		})
	}
}

func TestRemoveVolume(t *testing.T) {
	tests := []struct {
		name           string
		volumeToRemove string
		wantErr        bool
	}{
		{
			name:           "Case 1: Remove a volume and ensure the correct remove parameter is passed to the docker client",
			volumeToRemove: "volume1",
			wantErr:        false,
		},
		{
			name:           "Case 2: Pass an invalid volume parameter to remove, and verify error",
			volumeToRemove: "",
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			client, mockDockerClient := FakeNewMockClient(ctrl)

			if !tt.wantErr {
				mockDockerClient.EXPECT().VolumeRemove(gomock.Any(), gomock.Eq(tt.volumeToRemove), gomock.Eq(true)).Return(nil)
			}

			err := client.RemoveVolume(tt.volumeToRemove)

			if !tt.wantErr == (err != nil) {
				t.Errorf("expected %v but wanted %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractProjectToComponent(t *testing.T) {
	fakeClient := FakeNew()
	fakeErrorClient := FakeErrorNew()

	compInfo := common.ComponentInfo{
		ContainerName: "container",
	}
	targetPath := "/tmp"
	r := strings.NewReader("Hello!")

	tests := []struct {
		name    string
		client  *Client
		wantErr bool
	}{
		{
			name:    "Case 1: Successfully extract project to container",
			client:  fakeClient,
			wantErr: false,
		},
		{
			name:    "Case 2: Fail to extract project to container",
			client:  fakeErrorClient,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.ExtractProjectToComponent(compInfo, targetPath, r)
			if !tt.wantErr == (err != nil) {
				t.Errorf("got %v, wanted %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecCMDInContainer(t *testing.T) {
	fakeClient := FakeNew()
	fakeErrorClient := FakeErrorNew()

	compInfo := common.ComponentInfo{
		ContainerName: "container",
	}
	cmd := []string{"echo", "hello"}
	_, writer := io.Pipe()

	tests := []struct {
		name    string
		client  *Client
		wantErr bool
	}{
		{
			name:    "Case 1: Successfully execute command in the container",
			client:  fakeClient,
			wantErr: false,
		},
		{
			name:    "Case 2: Fail to execute command in the container",
			client:  fakeErrorClient,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.ExecCMDInContainer(compInfo, cmd, writer, writer, nil, false)
			if !tt.wantErr == (err != nil) {
				t.Errorf("got %v, wanted %v", err, tt.wantErr)
			}
		})
	}
}
