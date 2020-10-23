package common

import (
	"os"
	"reflect"
	"testing"

	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/testingutil"
)

// func TestGetDevfileContainerComponents(t *testing.T) {

// 	tests := []struct {
// 		name                 string
// 		component            []versionsCommon.DevfileComponent
// 		alias                []string
// 		expectedMatchesCount int
// 	}{
// 		{
// 			name:                 "Case 1: Invalid devfile",
// 			component:            []versionsCommon.DevfileComponent{},
// 			expectedMatchesCount: 0,
// 		},
// 		{
// 			name:                 "Case 2: Valid devfile with wrong component type (Openshift)",
// 			component:            []versionsCommon.DevfileComponent{{Openshift: &versionsCommon.Openshift{}}},
// 			expectedMatchesCount: 0,
// 		},
// 		{
// 			name:                 "Case 3: Valid devfile with wrong component type (Kubernetes)",
// 			component:            []versionsCommon.DevfileComponent{{Kubernetes: &versionsCommon.Kubernetes{}}},
// 			expectedMatchesCount: 0,
// 		},

// 		{
// 			name:                 "Case 4 : Valid devfile with correct component type (Container)",
// 			component:            []versionsCommon.DevfileComponent{testingutil.GetFakeContainerComponent("comp1"), testingutil.GetFakeContainerComponent("comp2")},
// 			expectedMatchesCount: 2,
// 		},

// 		{
// 			name:                 "Case 5: Valid devfile with correct component type (Container) without name",
// 			component:            []versionsCommon.DevfileComponent{testingutil.GetFakeContainerComponent("comp1"), testingutil.GetFakeContainerComponent("")},
// 			expectedMatchesCount: 1,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			devObj := devfileParser.DevfileObj{
// 				Data: &testingutil.TestDevfileData{
// 					Components: tt.component,
// 				},
// 			}

// 			devfileComponents := GetDevfileContainerComponents(devObj.Data)

// 			if len(devfileComponents) != tt.expectedMatchesCount {
// 				t.Errorf("TestGetDevfileContainerComponents error: wrong number of components matched: expected %v, actual %v", tt.expectedMatchesCount, len(devfileComponents))
// 			}
// 		})
// 	}

// }

func TestGetDevfileVolumeComponents(t *testing.T) {

	tests := []struct {
		name                 string
		component            []versionsCommon.DevfileComponent
		alias                []string
		expectedMatchesCount int
	}{
		{
			name:                 "Case 1: Invalid devfile",
			component:            []versionsCommon.DevfileComponent{},
			expectedMatchesCount: 0,
		},
		{
			name:                 "Case 2: Valid devfile with wrong component type (Openshift)",
			component:            []versionsCommon.DevfileComponent{{Openshift: &versionsCommon.Openshift{}}},
			expectedMatchesCount: 0,
		},
		{
			name:                 "Case 3: Valid devfile with wrong component type (Kubernetes)",
			component:            []versionsCommon.DevfileComponent{{Kubernetes: &versionsCommon.Kubernetes{}}},
			expectedMatchesCount: 0,
		},

		{
			name:                 "Case 4 : Valid devfile with wrong component type (Container)",
			component:            []versionsCommon.DevfileComponent{testingutil.GetFakeContainerComponent("comp1"), testingutil.GetFakeContainerComponent("comp2")},
			expectedMatchesCount: 0,
		},

		{
			name:                 "Case 5: Valid devfile with correct component type (Volume)",
			component:            []versionsCommon.DevfileComponent{testingutil.GetFakeContainerComponent("comp1"), testingutil.GetFakeVolumeComponent("myvol", "4Gi")},
			expectedMatchesCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: tt.component,
				},
			}

			devfileComponents := GetDevfileVolumeComponents(devObj.Data)

			if len(devfileComponents) != tt.expectedMatchesCount {
				t.Errorf("TestGetDevfileVolumeComponents error: wrong number of components matched: expected %v, actual %v", tt.expectedMatchesCount, len(devfileComponents))
			}
		})
	}

}

func TestGetVolumes(t *testing.T) {

	size := "4Gi"

	tests := []struct {
		name                       string
		component                  []versionsCommon.DevfileComponent
		wantContainerNameToVolumes map[string][]DevfileVolume
	}{
		{
			name:      "Case 1: Valid devfile with container referencing a volume component",
			component: []versionsCommon.DevfileComponent{testingutil.GetFakeContainerComponent("comp1"), testingutil.GetFakeVolumeComponent("myvolume1", size)},
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
			component: []versionsCommon.DevfileComponent{
				testingutil.GetFakeVolumeComponent("myvolume1", size),
				testingutil.GetFakeVolumeComponent("myvolume2", size),
				testingutil.GetFakeVolumeComponent("myvolume3", size),
				{
					Name: "mycontainer",
					Container: &versionsCommon.Container{
						Image: "image",
						VolumeMounts: []versionsCommon.VolumeMount{
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
			component: []versionsCommon.DevfileComponent{testingutil.GetFakeContainerComponent("comp1"), testingutil.GetFakeVolumeComponent("myvolume2", size)},
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
			component: []versionsCommon.DevfileComponent{
				testingutil.GetFakeVolumeComponent("myvolume2", size),
				{
					Name: "mycontainer",
					Container: &versionsCommon.Container{
						Image: "image",
					},
				},
			},
			wantContainerNameToVolumes: map[string][]DevfileVolume{},
		},
		{
			name: "Case 5: Valid devfile with container referencing no volume mount path",
			component: []versionsCommon.DevfileComponent{
				testingutil.GetFakeVolumeComponent("myvolume1", size),
				{
					Name: "mycontainer",
					Container: &versionsCommon.Container{
						Image: "image",
						VolumeMounts: []versionsCommon.VolumeMount{
							{
								Name: "myvolume1",
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

	envVars := []common.Env{
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
	var endpointPort int32 = 8080

	endpoints := []common.Endpoint{
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
		volumeMount common.VolumeMount
		wantPath    string
	}{
		{
			name: "Case 1: Mount Path is present",
			volumeMount: common.VolumeMount{
				Name: "name1",
				Path: "/path1",
			},
			wantPath: "/path1",
		},
		{
			name: "Case 2: Mount Path is absent",
			volumeMount: common.VolumeMount{
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

	component := []versionsCommon.DevfileComponent{
		testingutil.GetFakeContainerComponent("alias1"),
	}
	componentName := "alias1"
	command := "ls -la"
	workDir := "/"
	execCommands := []common.DevfileCommand{
		{
			Id: "run command",
			Exec: &common.Exec{
				CommandLine: command,
				Component:   componentName,
				WorkingDir:  workDir,
				Group: &versionsCommon.Group{
					Kind:      runGroup,
					IsDefault: true,
				},
			},
		},
		{
			Id: "build command",
			Exec: &common.Exec{
				CommandLine: command,
				Component:   componentName,
				WorkingDir:  workDir,
				Group:       &versionsCommon.Group{Kind: buildGroup},
			},
		},
		{
			Id: "test command",
			Exec: &common.Exec{
				CommandLine: command,
				Component:   componentName,
				WorkingDir:  workDir,
				Group:       &versionsCommon.Group{Kind: testGroup},
			},
		},
		{
			Id: "debug command",
			Exec: &common.Exec{
				CommandLine: command,
				Component:   componentName,
				WorkingDir:  workDir,
				Group:       &versionsCommon.Group{Kind: debugGroup},
			},
		},
		{
			Id: "customcommand",
			Exec: &common.Exec{
				CommandLine: command,
				Component:   componentName,
				WorkingDir:  workDir,
				Group:       &versionsCommon.Group{Kind: runGroup},
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
		groupType        common.DevfileCommandGroupType
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

	component := []versionsCommon.DevfileComponent{
		testingutil.GetFakeContainerComponent("alias1"),
	}

	tests := []struct {
		name             string
		execCommands     []common.DevfileCommand
		compCommands     []common.DevfileCommand
		expectedCommands []versionsCommon.DevfileCommand
	}{
		{
			name: "Case 1: One command",
			execCommands: []common.DevfileCommand{
				{
					Id: "somecommand",
					Exec: &common.Exec{
						HotReloadCapable: false,
					},
				},
			},
			expectedCommands: []versionsCommon.DevfileCommand{
				{
					Id: "somecommand",
					Exec: &common.Exec{
						HotReloadCapable: false,
					},
				},
			},
		},
		{
			name: "Case 2: Multiple commands",
			execCommands: []common.DevfileCommand{
				{
					Id: "somecommand",
					Exec: &common.Exec{
						HotReloadCapable: false,
					},
				},
				{
					Id: "somecommand2",
					Exec: &common.Exec{
						HotReloadCapable: false,
					},
				},
			},
			compCommands: []common.DevfileCommand{
				{
					Id: "mycomposite",
					Composite: &common.Composite{
						Commands: []string{},
					},
				},
			},
			expectedCommands: []versionsCommon.DevfileCommand{
				{
					Id: "somecommand",
					Exec: &common.Exec{
						HotReloadCapable: false,
					},
				},
				{
					Id: "somecommand2",
					Exec: &common.Exec{
						HotReloadCapable: false,
					},
				},
				{
					Id: "mycomposite",
					Composite: &common.Composite{
						Commands: []string{},
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
		envs []common.Env
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
			envs: []common.Env{
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
			envs: []common.Env{
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

	execCommands := []versionsCommon.DevfileCommand{
		{
			Id:   "exec1",
			Exec: &versionsCommon.Exec{},
		},
		{
			Id:   "exec2",
			Exec: &versionsCommon.Exec{},
		},
		{
			Id:   "exec3",
			Exec: &versionsCommon.Exec{},
		},
	}

	compCommands := []versionsCommon.DevfileCommand{
		{
			Id: "comp1",
			Composite: &common.Composite{
				Commands: []string{
					"exec1",
					"exec3",
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
