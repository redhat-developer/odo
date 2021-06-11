package utils

import (
	"os"
	"reflect"
	"testing"

	"github.com/docker/go-connections/nat"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/lclient"
	"github.com/openshift/odo/pkg/util"
)

func TestGetComponentContainers(t *testing.T) {
	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name          string
		componentName string
		client        *lclient.Client
		wantContainer types.Container
		wantErr       bool
	}{
		{
			name:          "Case 1: Component exists",
			componentName: "test",
			client:        fakeClient,
			wantContainer: types.Container{
				Names: []string{"/node"},
				ID:    "1",
				Image: "node",
				Labels: map[string]string{
					"component": "test",
					"alias":     "alias1",
				},
				Mounts: []types.MountPoint{
					{
						Name:        lclient.ProjectSourceVolumeName,
						Destination: lclient.OdoSourceVolumeMount,
					},
				},
			},
			wantErr: false,
		},
		{
			name:          "Case 2: Error client",
			componentName: "test",
			client:        fakeErrorClient,
			wantContainer: types.Container{},
			wantErr:       true,
		},
		{
			name:          "Case 3: Component does not exist",
			componentName: "somerandomcomponent",
			client:        fakeClient,
			wantContainer: types.Container{},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			containers, err := GetComponentContainers(*tt.client, tt.componentName)
			if !tt.wantErr && err != nil {
				t.Errorf("TestGetComponentContainers error: unexpected error %v", err)
			} else if !tt.wantErr {
				matched := false
				for _, container := range containers {
					if reflect.DeepEqual(tt.wantContainer, container) {
						matched = true
					}
				}
				if !matched && len(containers) > 0 {
					t.Errorf("TestGetComponentContainers error: did not match wanted container %v", tt.wantContainer.Names)
				}
			}
		})
	}
}

func TestConvertEnvs(t *testing.T) {
	envVarsNames := []string{"test", "sample-var", "myvar"}
	envVarsValues := []string{"value1", "value2", "value3"}
	tests := []struct {
		name    string
		envVars []devfilev1.EnvVar
		want    []string
	}{
		{
			name: "Case 1: One env var",
			envVars: []devfilev1.EnvVar{
				{
					Name:  envVarsNames[0],
					Value: envVarsValues[0],
				},
			},
			want: []string{"test=value1"},
		},
		{
			name: "Case 2: Multiple env vars",
			envVars: []devfilev1.EnvVar{
				{
					Name:  envVarsNames[0],
					Value: envVarsValues[0],
				},
				{
					Name:  envVarsNames[1],
					Value: envVarsValues[1],
				},
				{
					Name:  envVarsNames[2],
					Value: envVarsValues[2],
				},
			},
			want: []string{"test=value1", "sample-var=value2", "myvar=value3"},
		},
		{
			name:    "Case 3: No env vars",
			envVars: []devfilev1.EnvVar{},
			want:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVars := ConvertEnvs(tt.envVars)
			if !reflect.DeepEqual(tt.want, envVars) {
				t.Errorf("expected %v, wanted %v", envVars, tt.want)
			}
		})
	}
}

func TestDoesContainerNeedUpdating(t *testing.T) {
	envVarsNames := []string{"test", "sample-var", "myvar"}
	envVarsValues := []string{"value1", "value2", "value3"}

	volNames := []string{"vol1", "vol2", "vol3"}
	volPaths := []string{"/path1", "/path2", "/path3"}

	tests := []struct {
		name            string
		envVars         []devfilev1.EnvVar
		mounts          []mount.Mount
		image           string
		containerConfig container.Config
		containerMounts []types.MountPoint
		portmap         nat.PortMap
		hostConfig      container.HostConfig
		want            bool
	}{
		{
			name: "Case 1: No changes",
			envVars: []devfilev1.EnvVar{
				{
					Name:  envVarsNames[0],
					Value: envVarsValues[0],
				},
				{
					Name:  envVarsNames[1],
					Value: envVarsValues[1],
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
			envVars: []devfilev1.EnvVar{
				{
					Name:  envVarsNames[2],
					Value: envVarsValues[2],
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
			envVars: []devfilev1.EnvVar{
				{
					Name:  envVarsNames[2],
					Value: envVarsValues[2],
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
			envVars: []devfilev1.EnvVar{
				{
					Name:  envVarsNames[0],
					Value: envVarsValues[0],
				},
				{
					Name:  envVarsNames[1],
					Value: envVarsValues[1],
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
		{
			name: "Case 5: Update required, port changed",
			envVars: []devfilev1.EnvVar{
				{
					Name:  envVarsNames[0],
					Value: envVarsValues[0],
				},
				{
					Name:  envVarsNames[1],
					Value: envVarsValues[1],
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
				ExposedPorts: nat.PortSet{
					"8080/tcp": struct{}{},
				},
			},
			containerMounts: []types.MountPoint{
				{
					Name:        volNames[0],
					Destination: volPaths[0],
				},
			},
			want: true,
		},
		{
			name: "Case 6: Update required, exposed port changed",
			envVars: []devfilev1.EnvVar{
				{
					Name:  envVarsNames[0],
					Value: envVarsValues[0],
				},
				{
					Name:  envVarsNames[1],
					Value: envVarsValues[1],
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
				ExposedPorts: nat.PortSet{
					"8080/tcp": struct{}{},
				},
			},
			hostConfig: container.HostConfig{
				PortBindings: nat.PortMap{
					"8080/tcp": []nat.PortBinding{
						{
							HostIP:   "127.0.0.1",
							HostPort: "55555",
						},
					},
				},
			},
			containerMounts: []types.MountPoint{
				{
					Name:        volNames[0],
					Destination: volPaths[0],
				},
			},
			portmap: nat.PortMap{
				"8080/tcp": []nat.PortBinding{
					{
						HostIP:   "127.0.0.1",
						HostPort: "66666",
					},
				},
			},
			want: true,
		},
		{
			name: "Case 7: Update not required, exposed port unchanged",
			envVars: []devfilev1.EnvVar{
				{
					Name:  envVarsNames[0],
					Value: envVarsValues[0],
				},
				{
					Name:  envVarsNames[1],
					Value: envVarsValues[1],
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
				ExposedPorts: nat.PortSet{
					"8080/tcp": struct{}{},
				},
			},
			hostConfig: container.HostConfig{
				PortBindings: nat.PortMap{
					"8080/tcp": []nat.PortBinding{
						{
							HostIP:   "127.0.0.1",
							HostPort: "55555",
						},
					},
				},
			},
			containerMounts: []types.MountPoint{
				{
					Name:        volNames[0],
					Destination: volPaths[0],
				},
			},
			portmap: nat.PortMap{
				"8080/tcp": []nat.PortBinding{
					{
						HostIP:   "127.0.0.1",
						HostPort: "55555",
					},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := devfilev1.Component{
				ComponentUnion: devfilev1.ComponentUnion{
					Container: &devfilev1.ContainerComponent{
						Container: devfilev1.Container{
							Image: tt.image,
							Env:   tt.envVars,
						},
					},
				},
			}
			needsUpdating := DoesContainerNeedUpdating(component, &tt.containerConfig, &tt.hostConfig, tt.mounts, tt.containerMounts, tt.portmap)
			if needsUpdating != tt.want {
				t.Errorf("expected %v, wanted %v", needsUpdating, tt.want)
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
		t.Run(tt.name, func(t *testing.T) {
			hostConfig := container.HostConfig{
				Mounts: tt.mounts,
			}
			AddVolumeToContainer(projectVolumeName, lclient.OdoSourceVolumeMount, &hostConfig)
			if !reflect.DeepEqual(tt.want, hostConfig) {
				t.Errorf("expected %v, actual %v", tt.want, hostConfig)
			}
		})
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
				"type":      ProjectsVolume,
			},
		},
		{
			name:          "Case 1: Empty component name",
			componentName: "",
			want: map[string]string{
				"component": "",
				"type":      ProjectsVolume,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labels := GetProjectVolumeLabels(tt.componentName)
			if !reflect.DeepEqual(tt.want, labels) {
				t.Errorf("expected %v, actual %v", tt.want, labels)
			}
		})
	}

}

func TestGetContainerLabels(t *testing.T) {

	tests := []struct {
		name          string
		componentName string
		alias         string
		want          map[string]string
	}{
		{
			name:          "Case 1: Regular component name and alias",
			componentName: "some-component",
			alias:         "some-alias",
			want: map[string]string{
				"component": "some-component",
				"alias":     "some-alias",
			},
		},
		{
			name:          "Case 1: Empty component name and alias",
			componentName: "",
			want: map[string]string{
				"component": "",
				"alias":     "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labels := GetContainerLabels(tt.componentName, tt.alias)
			if !reflect.DeepEqual(tt.want, labels) {
				t.Errorf("expected %v, actual %v", tt.want, labels)
			}
		})
	}

}

func TestGetSupervisordVolumeLabels(t *testing.T) {

	componentNameArr := []string{"myComponent1", "myComponent2"}

	tests := []struct {
		name          string
		componentName string
		customImage   bool
		want          map[string]string
	}{
		{
			name:          "Case 1: Default supervisord image",
			componentName: componentNameArr[0],
			customImage:   false,
			want: map[string]string{
				"component": componentNameArr[0],
				"type":      SupervisordVolume,
			},
		},
		{
			name:          "Case 2: Custom supervisord image",
			componentName: componentNameArr[1],
			customImage:   true,
			want: map[string]string{
				"component": componentNameArr[1],
				"type":      SupervisordVolume,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.customImage {
				os.Setenv("ODO_BOOTSTRAPPER_IMAGE", "customimage:customtag")
			}
			image := adaptersCommon.GetBootstrapperImage()
			_, imageWithoutTag, _, imageTag := util.ParseComponentImageName(image)

			tt.want["version"] = imageTag
			tt.want["image"] = imageWithoutTag

			labels := GetSupervisordVolumeLabels(tt.componentName)
			if !reflect.DeepEqual(tt.want, labels) {
				t.Errorf("expected %v, actual %v", tt.want, labels)
			}
		})
	}

}

func TestGetContainerIDForAlias(t *testing.T) {

	containers := []types.Container{
		{
			ID: "someid",
			Labels: map[string]string{
				"alias": "somealias",
			},
		},
		{
			ID: "someid2",
			Labels: map[string]string{
				"alias": "somealias2",
			},
		},
	}

	tests := []struct {
		name            string
		alias           string
		wantContainerID string
	}{
		{
			name:            "Case 1: Get a container id for the a label match",
			alias:           "somealias",
			wantContainerID: "someid",
		},
		{
			name:            "Case 2: No container id for a label mismatch",
			alias:           "garbagealias",
			wantContainerID: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			containerID := GetContainerIDForAlias(containers, tt.alias)
			if containerID != tt.wantContainerID {
				t.Logf("TestGetContainerIDForAlias error: container id %v does not match the expected container id %v", containerID, tt.wantContainerID)
			}
		})
	}

}
