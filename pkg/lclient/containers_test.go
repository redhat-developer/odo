package lclient

import (
	"reflect"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
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
						"component": "node",
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
		containers, err := tt.client.GetContainerList()

		if !tt.wantErr == (err != nil) {
			t.Errorf("expected %v, wanted %v", err, tt.wantErr)
		}

		if !reflect.DeepEqual(tt.wantContainers, containers) {
			t.Errorf("Expected %v, got %v", tt.wantContainers, containers)
		}
	}
}

func TestStartStartContainer(t *testing.T) {
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
		err := tt.client.StartContainer(&fakeContainer, nil, nil)
		if !tt.wantErr == (err != nil) {
			t.Errorf("expected %v, wanted %v", err, tt.wantErr)
		}
	}
}
