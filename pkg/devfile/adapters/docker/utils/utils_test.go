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
		t.Run(tt.name, func(t *testing.T) {
			cmpExists := ComponentExists(*tt.client, tt.componentName)
			if tt.want != cmpExists {
				t.Errorf("expected %v, wanted %v", cmpExists, tt.want)
			}
		})
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
		envVars         []common.DockerimageEnv
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
		{
			name: "Case 5: Update required, port changed",
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
				DevfileComponentDockerimage: common.DevfileComponentDockerimage{
					Image: &tt.image,
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

	tests := []struct {
		name        string
		customImage bool
		want        map[string]string
	}{
		{
			name:        "Case 1: Default supervisord image",
			customImage: false,
			want: map[string]string{
				"name": adaptersCommon.SupervisordVolumeName,
				"type": supervisordVolume,
			},
		},
		{
			name:        "Case 2: Custom supervisord image",
			customImage: true,
			want: map[string]string{
				"name": adaptersCommon.SupervisordVolumeName,
				"type": supervisordVolume,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.customImage {
				os.Setenv("ODO_BOOTSTRAPPER_IMAGE", "customimage:customtag")
			}
			image := adaptersCommon.GetBootstrapperImage()
			_, _, _, imageTag := util.ParseComponentImageName(image)

			tt.want["version"] = imageTag

			labels := GetSupervisordVolumeLabels()
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
	validCommandType := common.DevfileCommandTypeExec
	supervisordVolumeName := "supervisordVolumeName"
	defaultWorkDirEnv := adaptersCommon.EnvOdoCommandRunWorkingDir
	defaultCommandEnv := adaptersCommon.EnvOdoCommandRun

	tests := []struct {
		name                  string
		commandActions        []common.DevfileCommandAction
		commandName           string
		comp                  common.DevfileComponent
		supervisordVolumeName string
		hostConfig            container.HostConfig
		wantHostConfig        container.HostConfig
		wantCommand           []string
		wantArgs              []string
		wantEnv               []common.DockerimageEnv
	}{
		{
			name: "Case 1: No component commands, args, env",
			commandActions: []common.DevfileCommandAction{
				{
					Command:   &command,
					Component: &component,
					Workdir:   &workDir,
					Type:      &validCommandType,
				},
			},
			commandName: emptyString,
			comp: common.DevfileComponent{
				Alias: &component,
				DevfileComponentDockerimage: common.DevfileComponentDockerimage{
					Command: []string{},
					Args:    []string{},
					Env:     []common.DockerimageEnv{},
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
			wantEnv: []common.DockerimageEnv{
				{
					Name:  &defaultWorkDirEnv,
					Value: &workDir,
				},
				{
					Name:  &defaultCommandEnv,
					Value: &command,
				},
			},
		},
		{
			name: "Case 2: Existing component command and no args, env",
			commandActions: []common.DevfileCommandAction{
				{
					Command:   &command,
					Component: &component,
					Workdir:   &workDir,
					Type:      &validCommandType,
				},
			},
			commandName: emptyString,
			comp: common.DevfileComponent{
				Alias: &component,
				DevfileComponentDockerimage: common.DevfileComponentDockerimage{
					Command: []string{"some", "command"},
					Args:    []string{},
					Env:     []common.DockerimageEnv{},
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
			wantEnv: []common.DockerimageEnv{
				{
					Name:  &defaultWorkDirEnv,
					Value: &workDir,
				},
				{
					Name:  &defaultCommandEnv,
					Value: &command,
				},
			},
		},
		{
			name: "Case 3: Existing component command and args and no env",
			commandActions: []common.DevfileCommandAction{
				{
					Command:   &command,
					Component: &component,
					Workdir:   &workDir,
					Type:      &validCommandType,
				},
			},
			commandName: emptyString,
			comp: common.DevfileComponent{
				Alias: &component,
				DevfileComponentDockerimage: common.DevfileComponentDockerimage{
					Command: []string{"some", "command"},
					Args:    []string{"some", "args"},
					Env:     []common.DockerimageEnv{},
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
			wantEnv: []common.DockerimageEnv{
				{
					Name:  &defaultWorkDirEnv,
					Value: &workDir,
				},
				{
					Name:  &defaultCommandEnv,
					Value: &command,
				},
			},
		},
		{
			name: "Case 4: Existing component command, args and env",
			commandActions: []common.DevfileCommandAction{
				{
					Command:   &command,
					Component: &component,
					Workdir:   &workDir,
					Type:      &validCommandType,
				},
			},
			commandName: emptyString,
			comp: common.DevfileComponent{
				Alias: &component,
				DevfileComponentDockerimage: common.DevfileComponentDockerimage{
					Command: []string{"some", "command"},
					Args:    []string{"some", "args"},
					Env: []common.DockerimageEnv{
						{
							Name:  &defaultWorkDirEnv,
							Value: &garbageString,
						},
						{
							Name:  &defaultCommandEnv,
							Value: &garbageString,
						},
					},
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
			wantEnv: []common.DockerimageEnv{
				{
					Name:  &defaultWorkDirEnv,
					Value: &garbageString,
				},
				{
					Name:  &defaultCommandEnv,
					Value: &garbageString,
				},
			},
		},
		{
			name: "Case 5: Existing host config, should append to it",
			commandActions: []common.DevfileCommandAction{
				{
					Command:   &command,
					Component: &component,
					Workdir:   &workDir,
					Type:      &validCommandType,
				},
			},
			commandName: emptyString,
			comp: common.DevfileComponent{
				Alias: &component,
				DevfileComponentDockerimage: common.DevfileComponentDockerimage{
					Command: []string{"some", "command"},
					Args:    []string{"some", "args"},
					Env: []common.DockerimageEnv{
						{
							Name:  &defaultWorkDirEnv,
							Value: &garbageString,
						},
						{
							Name:  &defaultCommandEnv,
							Value: &garbageString,
						},
					},
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
			wantEnv: []common.DockerimageEnv{
				{
					Name:  &defaultWorkDirEnv,
					Value: &garbageString,
				},
				{
					Name:  &defaultCommandEnv,
					Value: &garbageString,
				},
			},
		},
		{
			name: "Case 6: Not a run command component",
			commandActions: []common.DevfileCommandAction{
				{
					Command:   &command,
					Component: &component,
					Workdir:   &workDir,
					Type:      &validCommandType,
				},
			},
			commandName: emptyString,
			comp: common.DevfileComponent{
				Alias: &garbageString,
				DevfileComponentDockerimage: common.DevfileComponentDockerimage{
					Command: []string{},
					Args:    []string{},
					Env:     []common.DockerimageEnv{},
				},
			},
			supervisordVolumeName: supervisordVolumeName,
			hostConfig:            container.HostConfig{},
			wantHostConfig: container.HostConfig{
				Mounts: []mount.Mount{},
			},
			wantCommand: []string{},
			wantArgs:    []string{},
			wantEnv:     []common.DockerimageEnv{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					CommandActions: tt.commandActions,
					ComponentType:  common.DevfileComponentTypeDockerimage,
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
			if !reflect.DeepEqual(tt.comp.Command, tt.wantCommand) {
				t.Errorf("TestUpdateComponentWithSupervisord: component commands dont match actual: %v wanted: %v", tt.comp.Command, tt.wantCommand)
			}

			// Check the component args
			if !reflect.DeepEqual(tt.comp.Args, tt.wantArgs) {
				t.Errorf("TestUpdateComponentWithSupervisord: component args dont match actual: %v wanted: %v", tt.comp.Args, tt.wantArgs)
			}

			// Check the component env
			for _, compEnv := range tt.comp.Env {
				matched := false
				for _, wantEnv := range tt.wantEnv {
					if reflect.DeepEqual(wantEnv, compEnv) {
						matched = true
					}
				}

				if !matched {
					t.Errorf("TestUpdateComponentWithSupervisord: component env dont match env: %v:%v not present in wanted list", *compEnv.Name, *compEnv.Value)
				}
			}

		})
	}

}
