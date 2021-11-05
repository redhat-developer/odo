package image

import (
	"errors"
	"os/exec"
	"testing"

	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	gomock "github.com/golang/mock/gomock"
)

func TestBuildPushImage(t *testing.T) {
	tests := []struct {
		name            string
		devfilePath     string
		image           *devfile.ImageComponent
		push            bool
		BuildReturns    error
		PushReturns     error
		wantErr         bool
		wantBuildCalled bool
		wantPushCalled  bool
	}{
		{
			name:            "nil image and no push should return an error",
			push:            false,
			wantErr:         true,
			wantBuildCalled: false,
			wantPushCalled:  false,
		},
		{
			name:            "nil image and push should return an error",
			push:            true,
			wantErr:         true,
			wantBuildCalled: false,
			wantPushCalled:  false,
		},
		{
			name: "image and no push should call Build and not Push",
			image: &devfile.ImageComponent{
				Image: devfile.Image{
					ImageName: "a name",
				},
			},
			push:            false,
			wantErr:         false,
			wantBuildCalled: true,
			wantPushCalled:  false,
		},
		{
			name: "image and push should call Build and Push",
			image: &devfile.ImageComponent{
				Image: devfile.Image{
					ImageName: "a name",
				},
			},
			push:            true,
			wantErr:         false,
			wantBuildCalled: true,
			wantPushCalled:  true,
		},
		{
			name: "Build returns err",
			image: &devfile.ImageComponent{
				Image: devfile.Image{
					ImageName: "a name",
				},
			},
			push:            true,
			BuildReturns:    errors.New(""),
			PushReturns:     nil,
			wantErr:         true,
			wantBuildCalled: true,
			wantPushCalled:  false,
		},
		{
			name: "Push returns err",
			image: &devfile.ImageComponent{
				Image: devfile.Image{
					ImageName: "a name",
				},
			},
			push:            true,
			BuildReturns:    nil,
			PushReturns:     errors.New(""),
			wantErr:         true,
			wantBuildCalled: true,
			wantPushCalled:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			backend := NewMockBackend(ctrl)
			if tt.wantBuildCalled {
				backend.EXPECT().Build(tt.image, tt.devfilePath).Return(tt.BuildReturns).Times(1)
			} else {
				backend.EXPECT().Build(nil, tt.devfilePath).Times(0)
			}
			if tt.wantPushCalled {
				backend.EXPECT().Push(tt.image.ImageName).Return(tt.PushReturns).Times(1)
			} else {
				backend.EXPECT().Push(nil).Times(0)
			}
			err := buildPushImage(backend, tt.image, "", tt.push)

			if tt.wantErr != (err != nil) {
				t.Errorf("%s: Error result wanted %v, got %v", tt.name, tt.wantErr, err != nil)
			}
			ctrl.Finish()
		})
	}
}

func TestSelectBackend(t *testing.T) {
	tests := []struct {
		name        string
		lookPathCmd func(string) (string, error)
		wantType    string
		wantErr     bool
	}{
		{
			name: "all backends are present",
			lookPathCmd: func(string) (string, error) {
				return "", nil
			},
			wantErr:  false,
			wantType: "podman",
		},
		{
			name: "no backend are present",
			lookPathCmd: func(string) (string, error) {
				return "", errors.New("")
			},
			wantErr: true,
		},
		{
			name: "only docker is present",
			lookPathCmd: func(name string) (string, error) {
				if name == "docker" {
					return "docker", nil
				}
				return "", errors.New("")
			},
			wantErr:  false,
			wantType: "docker",
		},
		{
			name: "only podman is present",
			lookPathCmd: func(name string) (string, error) {
				if name == "podman" {
					return "podman", nil
				}
				return "", errors.New("")
			},
			wantErr:  false,
			wantType: "podman",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookPathCmd = tt.lookPathCmd
			defer func() { lookPathCmd = exec.LookPath }()
			backend, err := selectBackend()
			if tt.wantErr != (err != nil) {
				t.Errorf("%s: Error result wanted %v, got %v", tt.name, tt.wantErr, err != nil)
			}
			if tt.wantErr == false {
				if tt.wantType != backend.String() {
					t.Errorf("%s: Error backend wanted %v, got %v", tt.name, tt.wantType, backend.String())
				}
			}
		})
	}
}
