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
		wantContainers []types.Container
		wantErr        bool
	}{
		{
			name:      "Case 1: Successfully retrieve one container and have proper component",
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
			},
			wantErr: false,
		},
		{
			name:           "Case 2: Invalid component name",
			client:         fakeClient,
			component:      "fake-component",
			wantContainers: nil,
			wantErr:        false,
		},
		{
			name:           "Case 3: Error listing containers",
			client:         fakeErrorClient,
			component:      "fake-component",
			wantContainers: nil,
			wantErr:        true,
		},
		{
			name:      "Case 4: Multiple components returned",
			client:    fakeClient,
			component: "golang",
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
			wantErr: false,
		},
	}
	for _, tt := range tests {
		containers, err := tt.client.GetContainersByComponent(tt.component)

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
		err := tt.client.StartContainer(&fakeContainer, nil, nil, "golang")
		if !tt.wantErr == (err != nil) {
			t.Errorf("expected %v, wanted %v", err, tt.wantErr)
		}
	}
}
