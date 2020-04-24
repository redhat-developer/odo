package lclient

import (
	"reflect"
	"testing"

	"github.com/docker/docker/api/types"
)

func TestCreateVolume(t *testing.T) {
	fakeClient := FakeNew()
	fakeErrorClient := FakeErrorNew()
	tests := []struct {
		name       string
		client     *Client
		labels     map[string]string
		wantErr    bool
		wantVolume types.Volume
	}{
		{
			name:    "Case 1: Create volume, no labels",
			client:  fakeClient,
			labels:  map[string]string{},
			wantErr: false,
			wantVolume: types.Volume{
				Driver: "local",
				Labels: map[string]string{},
			},
		},
		{
			name:   "Case 2: Create volume, multiple labels",
			client: fakeClient,
			labels: map[string]string{
				"component": "golang",
				"type":      "project",
			},
			wantErr: false,
			wantVolume: types.Volume{
				Driver: "local",
				Labels: map[string]string{
					"component": "golang",
					"type":      "project",
				},
			},
		},
		{
			name:   "Case 3: Unable to create volume",
			client: fakeErrorClient,
			labels: map[string]string{
				"component": "golang",
				"type":      "project",
			},
			wantErr:    true,
			wantVolume: types.Volume{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volume, err := tt.client.CreateVolume("", tt.labels)
			if !tt.wantErr == (err != nil) {
				t.Errorf("expected %v, wanted %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(volume, tt.wantVolume) {
				t.Errorf("expected %v, wanted %v", volume, tt.wantVolume)
			}
		})
	}
}

func TestGetVolumesByLabel(t *testing.T) {
	fakeClient := FakeNew()
	fakeErrorClient := FakeErrorNew()
	tests := []struct {
		name        string
		client      *Client
		labels      map[string]string
		wantErr     bool
		wantVolumes []types.Volume
	}{
		{
			name:   "Case 1: Only one volume with label",
			client: fakeClient,
			labels: map[string]string{
				"component": "java",
			},
			wantErr: false,
			wantVolumes: []types.Volume{
				{
					Labels: map[string]string{
						"component": "java",
					},
				},
			},
		},
		{
			name:   "Case 2: Multiple volumes with label",
			client: fakeClient,
			labels: map[string]string{
				"component": "golang",
			},
			wantErr: false,
			wantVolumes: []types.Volume{
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
			},
		},
		{
			name:   "Case 3: No volumes with label",
			client: fakeClient,
			labels: map[string]string{
				"fakecomponent": "test",
			},
			wantErr:     false,
			wantVolumes: nil,
		},
		{
			name:   "Case 4: Docker client error",
			client: fakeErrorClient,
			labels: map[string]string{
				"fakecomponent": "test",
			},
			wantErr:     true,
			wantVolumes: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volumes, err := tt.client.GetVolumesByLabel(tt.labels)
			if !tt.wantErr == (err != nil) {
				t.Errorf("expected %v, wanted %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(volumes, tt.wantVolumes) {
				t.Errorf("expected %v, wanted %v", volumes, tt.wantVolumes)
			}
		})
	}
}
