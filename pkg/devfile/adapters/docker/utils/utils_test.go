package utils

import (
	"os"
	"reflect"
	"testing"

	"github.com/docker/go-connections/nat"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/lclient"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/openshift/odo/pkg/util"
)

func TestComponentExists(t *testing.T) {
	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name          string
		componentName string
		client        *lclient.Client
		componentType common.DevfileComponentType
		want          bool
		wantErr       bool
	}{
		{
			name:          "Case 1: Component exists",
			componentName: "golang",
			client:        fakeClient,
			componentType: common.DevfileComponentTypeDockerimage,
			want:          true,
			wantErr:       false,
		},
		{
			name:          "Case 2: Component doesn't exist",
			componentName: "fakecomponent",
			client:        fakeClient,
			componentType: common.DevfileComponentTypeDockerimage,
			want:          false,
			wantErr:       false,
		},
		{
			name:          "Case 3: Error with docker client",
			componentName: "golang",
			client:        fakeErrorClient,
			componentType: common.DevfileComponentTypeDockerimage,
			want:          false,
			wantErr:       true,
		},
		{
			name:          "Case 4: Container and devfile component mismatch",
			componentName: "test",
			client:        fakeClient,
			componentType: common.DevfileComponentTypeDockerimage,
			want:          false,
			wantErr:       true,
		},
		{
			name:          "Case 5: Devfile does not have supported components",
			componentName: "golang",
			client:        fakeClient,
			componentType: common.DevfileComponentTypeCheEditor,
			want:          false,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: tt.componentType,
				},
			}
			cmpExists, err := ComponentExists(*tt.client, devObj.Data, tt.componentName)
			if !tt.wantErr && err != nil {
				t.Errorf("TestComponentExists error, unexpected error - %v", err)
			} else if !tt.wantErr && tt.want != cmpExists {
				t.Errorf("expected %v, wanted %v", cmpExists, tt.want)
			}
		})
	}
}

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
		envVars []common.Env
		want    []string
	}{
		{
			name: "Case 1: One env var",
			envVars: []common.Env{
				{
					Name:  envVarsNames[0],
					Value: envVarsValues[0],
				},
			},
			want: []string{"test=value1"},
		},
		{
			name: "Case 2: Multiple env vars",
			envVars: []common.Env{
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
			envVars: []common.Env{},
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
		envVars         []common.Env
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
			envVars: []common.Env{
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
			envVars: []common.Env{
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
			envVars: []common.Env{
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
			envVars: []common.Env{
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
			envVars: []common.Env{
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
			envVars: []common.Env{
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
			envVars: []common.Env{
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
			component := common.DevfileComponent{
				Container: &common.Container{
					Image: tt.image,
					Env:   tt.envVars,
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

func TestUpdateComponentWithSupervisord(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""
	garbageString := "garbageString"
	supervisordVolumeName := "supervisordVolumeName"
	defaultWorkDirEnv := adaptersCommon.EnvOdoCommandRunWorkingDir
	defaultCommandEnv := adaptersCommon.EnvOdoCommandRun

	tests := []struct {
		name                  string
		commandExecs          []common.Exec
		commandName           string
		comp                  common.DevfileComponent
		supervisordVolumeName string
		hostConfig            container.HostConfig
		wantHostConfig        container.HostConfig
		wantCommand           []string
		wantArgs              []string
		wantEnv               []common.Env
	}{
		{
			name: "Case 1: No component commands, args, env",
			commandExecs: []common.Exec{
				{
					CommandLine: command,
					Component:   component,
					Group: &common.Group{
						Kind: common.RunCommandGroupType,
					},
					WorkingDir: workDir,
				},
			},
			commandName: emptyString,
			comp: common.DevfileComponent{
				Container: &common.Container{
					Command: []string{},
					Args:    []string{},
					Env:     []common.Env{},
					Name:    component,
				},
			},
			supervisordVolumeName: supervisordVolumeName,
			hostConfig:            container.HostConfig{},
			wantHostConfig: container.HostConfig{
				Mounts: []mount.Mount{
					{
						Type:   mount.TypeVolume,
						Source: supervisordVolumeName,
						Target: adaptersCommon.SupervisordMountPath,
					},
				},
			},
			wantCommand: []string{adaptersCommon.SupervisordBinaryPath},
			wantArgs:    []string{"-c", adaptersCommon.SupervisordConfFile},
			wantEnv: []common.Env{
				{
					Name:  defaultWorkDirEnv,
					Value: workDir,
				},
				{
					Name:  defaultCommandEnv,
					Value: command,
				},
			},
		},
		{
			name: "Case 2: Existing component command and no args, env",
			commandExecs: []common.Exec{
				{
					CommandLine: command,
					Component:   component,
					Group: &common.Group{
						Kind: common.RunCommandGroupType,
					},
					WorkingDir: workDir,
				},
			},
			commandName: emptyString,
			comp: common.DevfileComponent{
				Container: &common.Container{
					Command: []string{"some", "command"},
					Args:    []string{},
					Env:     []common.Env{},
					Name:    component,
				},
			},
			supervisordVolumeName: supervisordVolumeName,
			hostConfig:            container.HostConfig{},
			wantHostConfig: container.HostConfig{
				Mounts: []mount.Mount{
					{
						Type:   mount.TypeVolume,
						Source: supervisordVolumeName,
						Target: adaptersCommon.SupervisordMountPath,
					},
				},
			},
			wantCommand: []string{"some", "command"},
			wantArgs:    []string{},
			wantEnv: []common.Env{
				{
					Name:  defaultWorkDirEnv,
					Value: workDir,
				},
				{
					Name:  defaultCommandEnv,
					Value: command,
				},
			},
		},
		{
			name: "Case 3: Existing component command and args and no env",
			commandExecs: []common.Exec{
				{
					CommandLine: command,
					Component:   component,
					Group: &common.Group{
						Kind: common.RunCommandGroupType,
					},
					WorkingDir: workDir,
				},
			},
			commandName: emptyString,
			comp: common.DevfileComponent{
				Container: &common.Container{
					Command: []string{"some", "command"},
					Args:    []string{"some", "args"},
					Env:     []common.Env{},
					Name:    component,
				},
			},
			supervisordVolumeName: supervisordVolumeName,
			hostConfig:            container.HostConfig{},
			wantHostConfig: container.HostConfig{
				Mounts: []mount.Mount{
					{
						Type:   mount.TypeVolume,
						Source: supervisordVolumeName,
						Target: adaptersCommon.SupervisordMountPath,
					},
				},
			},
			wantCommand: []string{"some", "command"},
			wantArgs:    []string{"some", "args"},
			wantEnv: []common.Env{
				{
					Name:  defaultWorkDirEnv,
					Value: workDir,
				},
				{
					Name:  defaultCommandEnv,
					Value: command,
				},
			},
		},
		{
			name: "Case 4: Existing component command, args and env",
			commandExecs: []common.Exec{
				{
					CommandLine: command,
					Component:   component,
					Group: &common.Group{
						Kind: common.RunCommandGroupType,
					},
					WorkingDir: workDir,
				},
			},
			commandName: emptyString,
			comp: common.DevfileComponent{
				Container: &common.Container{
					Command: []string{"some", "command"},
					Args:    []string{"some", "args"},
					Env: []common.Env{
						{
							Name:  defaultWorkDirEnv,
							Value: garbageString,
						},
						{
							Name:  defaultCommandEnv,
							Value: garbageString,
						},
					},
					Name: component,
				},
			},
			supervisordVolumeName: supervisordVolumeName,
			hostConfig:            container.HostConfig{},
			wantHostConfig: container.HostConfig{
				Mounts: []mount.Mount{
					{
						Type:   mount.TypeVolume,
						Source: supervisordVolumeName,
						Target: adaptersCommon.SupervisordMountPath,
					},
				},
			},
			wantCommand: []string{"some", "command"},
			wantArgs:    []string{"some", "args"},
			wantEnv: []common.Env{
				{
					Name:  defaultWorkDirEnv,
					Value: garbageString,
				},
				{
					Name:  defaultCommandEnv,
					Value: garbageString,
				},
			},
		},
		{
			name: "Case 5: Existing host config, should append to it",
			commandExecs: []common.Exec{
				{
					CommandLine: command,
					Component:   component,
					Group: &common.Group{
						Kind: common.RunCommandGroupType,
					},
					WorkingDir: workDir,
				},
			},
			commandName: emptyString,
			comp: common.DevfileComponent{
				Container: &common.Container{
					Command: []string{"some", "command"},
					Args:    []string{"some", "args"},
					Env: []common.Env{
						{
							Name:  defaultWorkDirEnv,
							Value: garbageString,
						},
						{
							Name:  defaultCommandEnv,
							Value: garbageString,
						},
					},
					Name: component,
				},
			},
			supervisordVolumeName: supervisordVolumeName,
			hostConfig: container.HostConfig{
				Mounts: []mount.Mount{
					{
						Type:   mount.TypeVolume,
						Source: garbageString,
						Target: garbageString,
					},
				},
			},
			wantHostConfig: container.HostConfig{
				Mounts: []mount.Mount{
					{
						Type:   mount.TypeVolume,
						Source: supervisordVolumeName,
						Target: adaptersCommon.SupervisordMountPath,
					},
					{
						Type:   mount.TypeVolume,
						Source: garbageString,
						Target: garbageString,
					},
				},
			},
			wantCommand: []string{"some", "command"},
			wantArgs:    []string{"some", "args"},
			wantEnv: []common.Env{
				{
					Name:  defaultWorkDirEnv,
					Value: garbageString,
				},
				{
					Name:  defaultCommandEnv,
					Value: garbageString,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ExecCommands: tt.commandExecs,
					Components: []common.DevfileComponent{
						{
							Container: &common.Container{
								Name: tt.comp.Container.Name,
							},
						},
					},
				},
			}

			runCommand, err := adaptersCommon.GetRunCommand(devObj.Data, tt.commandName)
			if err != nil {
				t.Errorf("TestUpdateComponentWithSupervisord: error getting the run command")
			}

			UpdateComponentWithSupervisord(&tt.comp, runCommand, tt.supervisordVolumeName, &tt.hostConfig)

			// Check the container host config
			for _, containerHostConfigMount := range tt.hostConfig.Mounts {
				matched := false
				for _, wantHostConfigMount := range tt.wantHostConfig.Mounts {
					if reflect.DeepEqual(wantHostConfigMount, containerHostConfigMount) {
						matched = true
					}
				}

				if !matched {
					t.Errorf("TestUpdateComponentWithSupervisord: host configs source: %v target:%v do not match wanted host config", containerHostConfigMount.Source, containerHostConfigMount.Target)
				}
			}

			// Check the component command
			if !reflect.DeepEqual(tt.comp.Container.Command, tt.wantCommand) {
				t.Errorf("TestUpdateComponentWithSupervisord: component commands dont match actual: %v wanted: %v", tt.comp.Container.Command, tt.wantCommand)
			}

			// Check the component args
			if !reflect.DeepEqual(tt.comp.Container.Args, tt.wantArgs) {
				t.Errorf("TestUpdateComponentWithSupervisord: component args dont match actual: %v wanted: %v", tt.comp.Container.Args, tt.wantArgs)
			}

			// Check the component env
			for _, compEnv := range tt.comp.Container.Env {
				matched := false
				for _, wantEnv := range tt.wantEnv {
					if reflect.DeepEqual(wantEnv, compEnv) {
						matched = true
					}
				}

				if !matched {
					t.Errorf("TestUpdateComponentWithSupervisord: component env dont match env: %v:%v not present in wanted list", compEnv.Name, compEnv.Value)
				}
			}

		})
	}

}

func TestStartBootstrapSupervisordInitContainer(t *testing.T) {

	supervisordVolumeName := supervisordVolume

	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name    string
		client  *lclient.Client
		wantErr bool
	}{
		{
			name:    "Case 1: Successfully create a bootstrap container",
			client:  fakeClient,
			wantErr: false,
		},
		{
			name:    "Case 2: Failed to create a bootstrap container ",
			client:  fakeErrorClient,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := StartBootstrapSupervisordInitContainer(*tt.client, supervisordVolumeName)
			if !tt.wantErr && err != nil {
				t.Errorf("TestStartBootstrapSupervisordInitContainer: unexpected error got: %v wanted: %v", err, tt.wantErr)
			}
		})
	}

}

func TestCreateAndInitSupervisordVolume(t *testing.T) {

	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name    string
		client  *lclient.Client
		wantErr bool
	}{
		{
			name:    "Case 1: Successfully create a bootstrap vol and container",
			client:  fakeClient,
			wantErr: false,
		},
		{
			name:    "Case 2: Failed to create a bootstrap vol and container ",
			client:  fakeErrorClient,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volName, err := CreateAndInitSupervisordVolume(*tt.client)
			if !tt.wantErr && err != nil {
				t.Logf("TestCreateAndInitSupervisordVolume: unexpected error %v, wanted %v", err, tt.wantErr)
			} else if !tt.wantErr && volName != adaptersCommon.SupervisordVolumeName {
				t.Logf("TestCreateAndInitSupervisordVolume: unexpected supervisord vol name, expected: %v got: %v", adaptersCommon.SupervisordVolumeName, volName)
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
