package common

import (
	"os"
	"reflect"
	"testing"

	devfilev1 "github.com/devfile/api/pkg/apis/workspaces/v1alpha2"
	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/testingutil"
)

func TestGetVolumes(t *testing.T) {

	size := "4Gi"

	tests := []struct {
		name                       string
		component                  []devfilev1.Component
		wantContainerNameToVolumes map[string][]DevfileVolume
	}{
		{
			name:      "Case 1: Valid devfile with container referencing a volume component",
			component: []devfilev1.Component{testingutil.GetFakeContainerComponent("comp1"), testingutil.GetFakeVolumeComponent("myvolume1", size)},
			wantContainerNameToVolumes: map[string][]DevfileVolume{
				"comp1": {
					{
						Name:          "myvolume1",
						Size:          size,
						ContainerPath: "/my/volume/mount/path1",
					},
				},
			},
		},
		{
			name: "Case 2: Valid devfile with container referencing multiple volume components",
			component: []devfilev1.Component{
				testingutil.GetFakeVolumeComponent("myvolume1", size),
				testingutil.GetFakeVolumeComponent("myvolume2", size),
				testingutil.GetFakeVolumeComponent("myvolume3", size),
				{
					Name: "mycontainer",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								Image: "image",
								VolumeMounts: []devfilev1.VolumeMount{
									{
										Name: "myvolume1",
										Path: "/myvolume1",
									},
									{
										Name: "myvolume2",
										Path: "/myvolume2",
									},
								},
							},
						},
					},
				},
			},
			wantContainerNameToVolumes: map[string][]DevfileVolume{
				"mycontainer": {
					{
						Name:          "myvolume1",
						Size:          size,
						ContainerPath: "/myvolume1",
					},
					{
						Name:          "myvolume2",
						Size:          size,
						ContainerPath: "/myvolume2",
					},
				},
			},
		},
		{
			name:      "Case 3: Valid devfile with container referencing no volume component",
			component: []devfilev1.Component{testingutil.GetFakeContainerComponent("comp1"), testingutil.GetFakeVolumeComponent("myvolume2", size)},
			wantContainerNameToVolumes: map[string][]DevfileVolume{
				"comp1": {
					{
						Name:          "myvolume1",
						Size:          "1Gi",
						ContainerPath: "/my/volume/mount/path1",
					},
				},
			},
		},
		{
			name: "Case 4: Valid devfile with no container volume mounts",
			component: []devfilev1.Component{
				testingutil.GetFakeVolumeComponent("myvolume2", size),
				{
					Name: "mycontainer",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								Image: "image",
							},
						},
					},
				},
			},
			wantContainerNameToVolumes: map[string][]DevfileVolume{},
		},
		{
			name: "Case 5: Valid devfile with container referencing no volume mount path",
			component: []devfilev1.Component{
				testingutil.GetFakeVolumeComponent("myvolume1", size),
				{
					Name: "mycontainer",
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{
								Image: "image",
								VolumeMounts: []devfilev1.VolumeMount{
									{
										Name: "myvolume1",
									},
								},
							},
						},
					},
				},
			},
			wantContainerNameToVolumes: map[string][]DevfileVolume{
				"mycontainer": {
					{
						Name:          "myvolume1",
						Size:          "4Gi",
						ContainerPath: "/myvolume1",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: tt.component,
				},
			}

			containerNameToVolumes := GetVolumes(devObj)

			if !reflect.DeepEqual(containerNameToVolumes, tt.wantContainerNameToVolumes) {
				t.Errorf("TestGetVolumes error - got %v wanted %v", containerNameToVolumes, tt.wantContainerNameToVolumes)
			}
		})
	}

}

func TestIsEnvPresent(t *testing.T) {

	envName := "myenv"
	envValue := "myenvvalue"

	envVars := []devfilev1.EnvVar{
		{
			Name:  envName,
			Value: envValue,
		},
	}

	tests := []struct {
		name          string
		envVarName    string
		wantIsPresent bool
	}{
		{
			name:          "Case 1: Env var present",
			envVarName:    envName,
			wantIsPresent: true,
		},
		{
			name:          "Case 2: Env var absent",
			envVarName:    "someenv",
			wantIsPresent: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isPresent := IsEnvPresent(envVars, tt.envVarName)
			if isPresent != tt.wantIsPresent {
				t.Errorf("TestIsEnvPresent error: env var expectation mismatch, want: %v got: %v", tt.wantIsPresent, isPresent)
			}
		})
	}

}

func TestIsPortPresent(t *testing.T) {

	endpointName := "8080/tcp"
	var endpointPort int = 8080

	endpoints := []devfilev1.Endpoint{
		{
			Name:       endpointName,
			TargetPort: endpointPort,
		},
	}

	tests := []struct {
		name          string
		port          int
		wantIsPresent bool
	}{
		{
			name:          "Case 1: Endpoint port present",
			port:          8080,
			wantIsPresent: true,
		},
		{
			name:          "Case 2: Endpoint port absent",
			port:          1234,
			wantIsPresent: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isPresent := IsPortPresent(endpoints, tt.port)
			if isPresent != tt.wantIsPresent {
				t.Errorf("TestIsPortPresent error: endpoint port expectation mismatch, want: %v got: %v", tt.wantIsPresent, isPresent)
			}
		})
	}

}

func TestGetBootstrapperImage(t *testing.T) {

	customImage := "customimage:customtag"

	tests := []struct {
		name        string
		customImage bool
		wantImage   string
	}{
		{
			name:        "Case 1: Default bootstrap image",
			customImage: false,
			wantImage:   defaultBootstrapperImage,
		},
		{
			name:        "Case 2: Custom bootstrap image",
			customImage: true,
			wantImage:   customImage,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.customImage {
				os.Setenv(bootstrapperImageEnvName, customImage)
			}
			image := GetBootstrapperImage()

			if image != tt.wantImage {
				t.Errorf("TestGetBootstrapperImage error: bootstrap image mismatch, expected: %v got: %v", tt.wantImage, image)
			}
		})
	}

}

func TestGetVolumeMountPath(t *testing.T) {

	tests := []struct {
		name        string
		volumeMount devfilev1.VolumeMount
		wantPath    string
	}{
		{
			name: "Case 1: Mount Path is present",
			volumeMount: devfilev1.VolumeMount{
				Name: "name1",
				Path: "/path1",
			},
			wantPath: "/path1",
		},
		{
			name: "Case 2: Mount Path is absent",
			volumeMount: devfilev1.VolumeMount{
				Name: "name1",
			},
			wantPath: "/name1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := GetVolumeMountPath(tt.volumeMount)

			if path != tt.wantPath {
				t.Errorf("TestGetVolumeMountPath error: mount path mismatch, expected: %v got: %v", tt.wantPath, path)
			}
		})
	}

}

func TestGetCommandsForGroup(t *testing.T) {

	component := []devfilev1.Component{
		testingutil.GetFakeContainerComponent("alias1"),
	}
	componentName := "alias1"
	command := "ls -la"
	workDir := "/"
	execCommands := []devfilev1.Command{
		{
			Id: "run command",
			CommandUnion: devfilev1.CommandUnion{
				Exec: &devfilev1.ExecCommand{
					LabeledCommand: devfilev1.LabeledCommand{
						BaseCommand: devfilev1.BaseCommand{
							Group: &devfilev1.CommandGroup{
								Kind:      runGroup,
								IsDefault: true,
							},
						},
					},
					CommandLine: command,
					Component:   componentName,
					WorkingDir:  workDir,
				},
			},
		},
		{
			Id: "build command",
			CommandUnion: devfilev1.CommandUnion{
				Exec: &devfilev1.ExecCommand{
					LabeledCommand: devfilev1.LabeledCommand{
						BaseCommand: devfilev1.BaseCommand{
							Group: &devfilev1.CommandGroup{Kind: buildGroup},
						},
					},
					CommandLine: command,
					Component:   componentName,
					WorkingDir:  workDir,
				},
			},
		},
		{
			Id: "test command",
			CommandUnion: devfilev1.CommandUnion{
				Exec: &devfilev1.ExecCommand{
					LabeledCommand: devfilev1.LabeledCommand{
						BaseCommand: devfilev1.BaseCommand{
							Group: &devfilev1.CommandGroup{Kind: testGroup},
						},
					},
					CommandLine: command,
					Component:   componentName,
					WorkingDir:  workDir,
				},
			},
		},
		{
			Id: "debug command",
			CommandUnion: devfilev1.CommandUnion{
				Exec: &devfilev1.ExecCommand{
					LabeledCommand: devfilev1.LabeledCommand{
						BaseCommand: devfilev1.BaseCommand{
							Group: &devfilev1.CommandGroup{Kind: debugGroup},
						},
					},
					CommandLine: command,
					Component:   componentName,
					WorkingDir:  workDir,
				},
			},
		},
		{
			Id: "customcommand",
			CommandUnion: devfilev1.CommandUnion{
				Exec: &devfilev1.ExecCommand{
					LabeledCommand: devfilev1.LabeledCommand{
						BaseCommand: devfilev1.BaseCommand{
							Group: &devfilev1.CommandGroup{Kind: runGroup},
						},
					},
					CommandLine: command,
					Component:   componentName,
					WorkingDir:  workDir,
				},
			},
		},
	}

	devObj := devfileParser.DevfileObj{
		Data: &testingutil.TestDevfileData{
			Components: component,
			Commands:   execCommands,
		},
	}

	tests := []struct {
		name             string
		groupType        devfilev1.CommandGroupKind
		numberOfCommands int
	}{
		{
			name:             "Case 1: Build Group Command",
			groupType:        buildGroup,
			numberOfCommands: 1,
		},
		{
			name:             "Case 2: Run Group Command",
			groupType:        runGroup,
			numberOfCommands: 2,
		},
		{
			name:             "Case 3: Test Group Command",
			groupType:        testGroup,
			numberOfCommands: 1,
		},
		{
			name:             "Case 4: Debug Group Command",
			groupType:        debugGroup,
			numberOfCommands: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commands := getCommandsByGroup(devObj.Data, tt.groupType)

			if len(commands) != tt.numberOfCommands {
				t.Errorf("TestGetCommandsForGroup error: number of commands mismatch for group %v, expected: %v got: %v", string(tt.groupType), tt.numberOfCommands, len(commands))
			}

			for _, command := range commands {
				if command.Exec.Group.Kind != tt.groupType {
					t.Errorf("TestGetCommandsForGroup error: command group mismatch, expected: %v got: %v", string(tt.groupType), string(command.Exec.Group.Kind))
				}
			}
		})
	}

}

func TestGetCommands(t *testing.T) {

	component := []devfilev1.Component{
		testingutil.GetFakeContainerComponent("alias1"),
	}

	tests := []struct {
		name             string
		execCommands     []devfilev1.Command
		compCommands     []devfilev1.Command
		expectedCommands []devfilev1.Command
	}{
		{
			name: "Case 1: One command",
			execCommands: []devfilev1.Command{
				{
					Id: "somecommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							HotReloadCapable: false,
						},
					},
				},
			},
			expectedCommands: []devfilev1.Command{
				{
					Id: "somecommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							HotReloadCapable: false,
						},
					},
				},
			},
		},
		{
			name: "Case 2: Multiple commands",
			execCommands: []devfilev1.Command{
				{
					Id: "somecommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							HotReloadCapable: false,
						},
					},
				},
				{
					Id: "somecommand2",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							HotReloadCapable: false,
						},
					},
				},
			},
			compCommands: []devfilev1.Command{
				{
					Id: "mycomposite",
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{
							Commands: []string{},
						},
					},
				},
			},
			expectedCommands: []devfilev1.Command{
				{
					Id: "somecommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							HotReloadCapable: false,
						},
					},
				},
				{
					Id: "somecommand2",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							HotReloadCapable: false,
						},
					},
				},
				{
					Id: "mycomposite",
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{
							Commands: []string{},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: component,
					Commands:   append(tt.execCommands, tt.compCommands...),
				},
			}

			commandsMap := devObj.Data.GetCommands()
			if len(commandsMap) != len(tt.expectedCommands) {
				t.Errorf("TestGetCommands error: number of returned commands don't match: %v got: %v", len(tt.expectedCommands), len(commandsMap))
			}
			for _, command := range tt.expectedCommands {
				_, ok := commandsMap[command.Id]
				if !ok {
					t.Errorf("TestGetCommands error: command %v not found in map", command.Id)
				}
			}
		})
	}

}

func TestGetComponentEnvVar(t *testing.T) {

	tests := []struct {
		name string
		env  string
		envs []devfilev1.EnvVar
		want string
	}{
		{
			name: "Case 1: No env vars",
			env:  "test",
			envs: nil,
			want: "",
		},
		{
			name: "Case 2: Has env",
			env:  "PROJECTS_ROOT",
			envs: []devfilev1.EnvVar{
				{
					Name:  "SOME_ENV",
					Value: "test",
				},
				{
					Name:  "TESTER",
					Value: "tester",
				},
				{
					Name:  "PROJECTS_ROOT",
					Value: "/test",
				},
			},
			want: "/test",
		},
		{
			name: "Case 3: No env, multiple values",
			env:  "PROJECTS_ROOT",
			envs: []devfilev1.EnvVar{
				{
					Name:  "TESTER",
					Value: "fake",
				},
				{
					Name:  "FAKE",
					Value: "fake",
				},
				{
					Name:  "ENV",
					Value: "fake",
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := GetComponentEnvVar(tt.env, tt.envs)
			if value != tt.want {
				t.Errorf("TestGetComponentEnvVar error: env value mismatch, expected: %v got: %v", tt.want, value)
			}

		})
	}

}

func TestGetCommandsFromEvent(t *testing.T) {

	execCommands := []devfilev1.Command{
		{
			Id: "exec1",
			CommandUnion: devfilev1.CommandUnion{
				Exec: &devfilev1.ExecCommand{},
			},
		},
		{
			Id: "exec2",
			CommandUnion: devfilev1.CommandUnion{
				Exec: &devfilev1.ExecCommand{},
			},
		},
		{
			Id: "exec3",
			CommandUnion: devfilev1.CommandUnion{
				Exec: &devfilev1.ExecCommand{},
			},
		},
	}

	compCommands := []devfilev1.Command{
		{
			Id: "comp1",
			CommandUnion: devfilev1.CommandUnion{
				Composite: &devfilev1.CompositeCommand{
					Commands: []string{
						"exec1",
						"exec3",
					},
				},
			},
		},
	}

	tests := []struct {
		name         string
		eventName    string
		wantCommands []string
	}{
		{
			name:      "Case 1: composite event",
			eventName: "comp1",
			wantCommands: []string{
				"exec1",
				"exec3",
			},
		},
		{
			name:      "Case 2: exec event",
			eventName: "exec2",
			wantCommands: []string{
				"exec2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Commands: append(compCommands, execCommands...),
				},
			}

			commandsMap := devObj.Data.GetCommands()
			commands := GetCommandsFromEvent(commandsMap, tt.eventName)
			if !reflect.DeepEqual(tt.wantCommands, commands) {
				t.Errorf("TestGetCommandsFromEvent error - got %v expected %v", commands, tt.wantCommands)
			}
		})
	}

}
