package component

import (
	"reflect"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/openshift/odo/pkg/devfile"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	versionsCommon "github.com/openshift/odo/pkg/devfile/versions/common"
	"github.com/openshift/odo/pkg/lclient"
	"github.com/openshift/odo/pkg/testingutil"
)

func TestCreateComponent(t *testing.T) {

	testComponentName := "test"
	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name          string
		componentType versionsCommon.DevfileComponentType
		client        *lclient.Client
		wantErr       bool
	}{
		{
			name:          "Case 1: Invalid devfile",
			componentType: "",
			client:        fakeClient,
			wantErr:       true,
		},
		{
			name:          "Case 2: Valid devfile",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			client:        fakeClient,
			wantErr:       false,
		},
		{
			name:          "Case 3: Valid devfile, docker client error",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			client:        fakeErrorClient,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfile.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: tt.componentType,
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			componentAdapter := New(adapterCtx, *tt.client)
			err := componentAdapter.createComponent()

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter create unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestUpdateComponent(t *testing.T) {

	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name          string
		componentType versionsCommon.DevfileComponentType
		componentName string
		client        *lclient.Client
		wantErr       bool
	}{
		{
			name:          "Case 1: Invalid devfile",
			componentType: "",
			componentName: "",
			client:        fakeClient,
			wantErr:       true,
		},
		{
			name:          "Case 2: Valid devfile",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			componentName: "node",
			client:        fakeClient,
			wantErr:       false,
		},
		{
			name:          "Case 3: Valid devfile, docker client error",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			componentName: "",
			client:        fakeErrorClient,
			wantErr:       true,
		},
		{
			name:          "Case 3: Valid devfile, missing component",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			componentName: "fakecomponent",
			client:        fakeClient,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfile.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: tt.componentType,
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: tt.componentName,
				Devfile:       devObj,
			}

			componentAdapter := New(adapterCtx, *tt.client)
			err := componentAdapter.updateComponent()

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter update unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestStartContainer(t *testing.T) {

	testComponentName := "test"
	testVolumeName := "projects"

	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name          string
		componentType versionsCommon.DevfileComponentType
		client        *lclient.Client
		wantErr       bool
	}{
		{
			name:          "Case 1: Successfully start container",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			client:        fakeClient,
			wantErr:       false,
		},
		{
			name:          "Case 2: Docker client error",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			client:        fakeErrorClient,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfile.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: tt.componentType,
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			componentAdapter := New(adapterCtx, *tt.client)
			err := componentAdapter.startContainer(testComponentName, testVolumeName, adapterCtx.Devfile.Data.GetAliasedComponents()[0])

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter create unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestAddProjectVolumeToComp(t *testing.T) {
	projectVolumeName := "projects"

	tests := []struct {
		name   string
		mounts []mount.Mount
		want   container.HostConfig
	}{
		{
			name:   "Case 1: No existing mounts",
			mounts: []mount.Mount{},
			want: container.HostConfig{
				Mounts: []mount.Mount{
					{
						Type:   mount.TypeVolume,
						Source: projectVolumeName,
						Target: lclient.OdoSourceVolumeMount,
					},
				},
			},
		},
		{
			name: "Case 2: One existing mount",
			mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: "/my/local/folder",
					Target: "/test",
				},
			},
			want: container.HostConfig{
				Mounts: []mount.Mount{
					{
						Type:   mount.TypeBind,
						Source: "/my/local/folder",
						Target: "/test",
					},
					{
						Type:   mount.TypeVolume,
						Source: projectVolumeName,
						Target: lclient.OdoSourceVolumeMount,
					},
				},
			},
		},
		{
			name: "Case 3: Multiple mounts",
			mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: "/my/local/folder",
					Target: "/test",
				},
				{
					Type:   mount.TypeBind,
					Source: "/my/second/folder",
					Target: "/two",
				},
			},
			want: container.HostConfig{
				Mounts: []mount.Mount{
					{
						Type:   mount.TypeBind,
						Source: "/my/local/folder",
						Target: "/test",
					},
					{
						Type:   mount.TypeBind,
						Source: "/my/second/folder",
						Target: "/two",
					},
					{
						Type:   mount.TypeVolume,
						Source: projectVolumeName,
						Target: lclient.OdoSourceVolumeMount,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		hostConfig := container.HostConfig{
			Mounts: tt.mounts,
		}
		addProjectVolumeToComp(projectVolumeName, &hostConfig)
		if !reflect.DeepEqual(tt.want, hostConfig) {
			t.Errorf("expected %v, actual %v", tt.want, hostConfig)
		}
	}

}
