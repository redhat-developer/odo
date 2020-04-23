package utils

import (
	"reflect"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/lclient"
)

func TestComponentExists(t *testing.T) {
	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name          string
		componentName string
		client        *lclient.Client
		want          bool
	}{
		{
			name:          "Case 1: Component exists",
			componentName: "golang",
			client:        fakeClient,
			want:          true,
		},
		{
			name:          "Case 2: Component doesn't exist",
			componentName: "fakecomponent",
			client:        fakeClient,
			want:          false,
		},
		{
			name:          "Case 3: Error with docker client",
			componentName: "golang",
			client:        fakeErrorClient,
			want:          false,
		},
	}

	for _, tt := range tests {
		cmpExists := ComponentExists(*tt.client, tt.componentName)
		if tt.want != cmpExists {
			t.Errorf("expected %v, wanted %v", cmpExists, tt.want)
		}
	}
}

func TestConvertEnvs(t *testing.T) {
	envVarsNames := []string{"test", "sample-var", "myvar"}
	envVarsValues := []string{"value1", "value2", "value3"}
	tests := []struct {
		name    string
		envVars []common.DockerimageEnv
		want    []string
	}{
		{
			name: "Case 1: One env var",
			envVars: []common.DockerimageEnv{
				{
					Name:  &envVarsNames[0],
					Value: &envVarsValues[0],
				},
			},
			want: []string{"test=value1"},
		},
		{
			name: "Case 2: Multiple env vars",
			envVars: []common.DockerimageEnv{
				{
					Name:  &envVarsNames[0],
					Value: &envVarsValues[0],
				},
				{
					Name:  &envVarsNames[1],
					Value: &envVarsValues[1],
				},
				{
					Name:  &envVarsNames[2],
					Value: &envVarsValues[2],
				},
			},
			want: []string{"test=value1", "sample-var=value2", "myvar=value3"},
		},
		{
			name:    "Case 3: No env vars",
			envVars: []common.DockerimageEnv{},
			want:    []string{},
		},
	}

	for _, tt := range tests {
		envVars := ConvertEnvs(tt.envVars)
		if !reflect.DeepEqual(tt.want, envVars) {
			t.Errorf("expected %v, wanted %v", envVars, tt.want)
		}
	}
}

func TestDoesContainerNeedUpdating(t *testing.T) {
	envVarsNames := []string{"test", "sample-var", "myvar"}
	envVarsValues := []string{"value1", "value2", "value3"}

	volNames := []string{"vol1", "vol2", "vol3"}
	volPaths := []string{"/path1", "/path2", "/path3"}

	tests := []struct {
		name            string
		envVars         []common.DockerimageEnv
		mounts          []mount.Mount
		image           string
		containerConfig container.Config
		containerMounts []types.MountPoint
		want            bool
	}{
		{
			name: "Case 1: No changes",
			envVars: []common.DockerimageEnv{
				{
					Name:  &envVarsNames[0],
					Value: &envVarsValues[0],
				},
				{
					Name:  &envVarsNames[1],
					Value: &envVarsValues[1],
				},
			},
			mounts: []mount.Mount{
				{
					Source: volNames[0],
					Target: volPaths[0],
				},
			},
			image: "golang",
			containerConfig: container.Config{
				Image: "golang",
				Env:   []string{"test=value1", "sample-var=value2"},
			},
			containerMounts: []types.MountPoint{
				{
					Name:        volNames[0],
					Destination: volPaths[0],
				},
			},
			want: false,
		},
		{
			name: "Case 2: Update required, env var changed",
			envVars: []common.DockerimageEnv{
				{
					Name:  &envVarsNames[2],
					Value: &envVarsValues[2],
				},
			},
			image: "golang",
			containerConfig: container.Config{
				Image: "golang",
				Env:   []string{"test=value1", "sample-var=value2"},
			},
			want: true,
		},
		{
			name: "Case 3: Update required, image changed",
			envVars: []common.DockerimageEnv{
				{
					Name:  &envVarsNames[2],
					Value: &envVarsValues[2],
				},
			},
			image: "node",
			containerConfig: container.Config{
				Image: "golang",
				Env:   []string{"test=value1", "sample-var=value2"},
			},
			want: true,
		},
		{
			name: "Case 4: Update required, volumes changed",
			envVars: []common.DockerimageEnv{
				{
					Name:  &envVarsNames[0],
					Value: &envVarsValues[0],
				},
				{
					Name:  &envVarsNames[1],
					Value: &envVarsValues[1],
				},
			},
			mounts: []mount.Mount{
				{
					Source: volNames[0],
					Target: volPaths[0],
				},
				{
					Source: volNames[1],
					Target: volPaths[1],
				},
			},
			image: "golang",
			containerConfig: container.Config{
				Image: "golang",
				Env:   []string{"test=value1", "sample-var=value2"},
			},
			containerMounts: []types.MountPoint{
				{
					Name:        volNames[0],
					Destination: volPaths[0],
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		component := common.DevfileComponent{
			DevfileComponentDockerimage: common.DevfileComponentDockerimage{
				Image: &tt.image,
				Env:   tt.envVars,
			},
		}
		needsUpdating := DoesContainerNeedUpdating(component, &tt.containerConfig, tt.mounts, tt.containerMounts)
		if needsUpdating != tt.want {
			t.Errorf("expected %v, wanted %v", needsUpdating, tt.want)
		}
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
		AddProjectVolumeToComp(projectVolumeName, &hostConfig)
		if !reflect.DeepEqual(tt.want, hostConfig) {
			t.Errorf("expected %v, actual %v", tt.want, hostConfig)
		}
	}

}

func TestGetProjectVolumeLabels(t *testing.T) {

	tests := []struct {
		name          string
		componentName string
		want          map[string]string
	}{
		{
			name:          "Case 1: Regular component name",
			componentName: "some-component",
			want: map[string]string{
				"component": "some-component",
				"type":      "projects",
			},
		},
		{
			name:          "Case 1: Empty component name",
			componentName: "",
			want: map[string]string{
				"component": "",
				"type":      "projects",
			},
		},
	}
	for _, tt := range tests {
		labels := GetProjectVolumeLabels(tt.componentName)
		if !reflect.DeepEqual(tt.want, labels) {
			t.Errorf("expected %v, actual %v", tt.want, labels)
		}
	}

}
