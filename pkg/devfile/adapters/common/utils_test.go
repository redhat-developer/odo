package common

import (
	"os"
	"testing"

	devfileParser "github.com/cli-playground/devfile-parser/pkg/devfile/parser"
	"github.com/cli-playground/devfile-parser/pkg/devfile/parser/data/common"
	versionsCommon "github.com/cli-playground/devfile-parser/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/testingutil"
)

func TestGetSupportedComponents(t *testing.T) {

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
			name:                 "Case 4 : Valid devfile with correct component type (Container)",
			component:            []versionsCommon.DevfileComponent{testingutil.GetFakeComponent("comp1"), testingutil.GetFakeComponent("comp2")},
			expectedMatchesCount: 2,
		},

		{
			name:                 "Case 5: Valid devfile with correct component type (Container) without name",
			component:            []versionsCommon.DevfileComponent{testingutil.GetFakeComponent("comp1"), testingutil.GetFakeComponent("")},
			expectedMatchesCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					Components: tt.component,
				},
			}

			devfileComponents := GetSupportedComponents(devObj.Data)

			if len(devfileComponents) != tt.expectedMatchesCount {
				t.Errorf("TestGetSupportedComponents error: wrong number of components matched: expected %v, actual %v", tt.expectedMatchesCount, len(devfileComponents))
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

func TestIsComponentSupported(t *testing.T) {

	tests := []struct {
		name            string
		component       common.DevfileComponent
		wantIsSupported bool
	}{
		{
			name:            "Case 1: Supported component",
			component:       testingutil.GetFakeComponent("comp1"),
			wantIsSupported: true,
		},
		{
			name: "Case 2: Unsupported component",
			component: common.DevfileComponent{
				Openshift: &versionsCommon.Openshift{},
			},
			wantIsSupported: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isSupported := isComponentSupported(tt.component)
			if isSupported != tt.wantIsSupported {
				t.Errorf("TestIsComponentSupported error: component support mismatch, expected: %v got: %v", tt.wantIsSupported, isSupported)
			}
		})
	}

}

func TestGetCommandsForGroup(t *testing.T) {

	component := []versionsCommon.DevfileComponent{
		testingutil.GetFakeComponent("alias1"),
	}
	componentName := "alias1"
	command := "ls -la"
	workDir := "/"
	execCommands := []common.Exec{
		{
			Id:          "run command",
			CommandLine: command,
			Component:   componentName,
			WorkingDir:  workDir,
			Group: &versionsCommon.Group{
				Kind:      runGroup,
				IsDefault: true,
			},
		},
		{
			Id:          "build command",
			CommandLine: command,
			Component:   componentName,
			WorkingDir:  workDir,
			Group:       &versionsCommon.Group{Kind: buildGroup},
		},
		{
			Id:          "test command",
			CommandLine: command,
			Component:   componentName,
			WorkingDir:  workDir,
			Group:       &versionsCommon.Group{Kind: testGroup},
		},
		{
			Id:          "debug command",
			CommandLine: command,
			Component:   componentName,
			WorkingDir:  workDir,
			Group:       &versionsCommon.Group{Kind: debugGroup},
		},
		{
			Id:          "customcommand",
			CommandLine: command,
			Component:   componentName,
			WorkingDir:  workDir,
			Group:       &versionsCommon.Group{Kind: runGroup},
		},
	}

	devObj := devfileParser.DevfileObj{
		Data: testingutil.TestDevfileData{
			Components:   component,
			ExecCommands: execCommands,
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
		{
			name:             "Case 5: Wrong Group Command",
			groupType:        initGroup,
			numberOfCommands: 0,
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
