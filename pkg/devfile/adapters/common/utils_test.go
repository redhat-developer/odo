package common

import (
	"reflect"
	"testing"

	"github.com/openshift/odo/pkg/devfile"
	"github.com/openshift/odo/pkg/devfile/versions/common"
	versionsCommon "github.com/openshift/odo/pkg/devfile/versions/common"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/openshift/odo/pkg/util"
)

func TestGetSupportedComponents(t *testing.T) {

	tests := []struct {
		name                 string
		componentType        versionsCommon.DevfileComponentType
		alias                []string
		expectedMatchesCount int
	}{
		{
			name:                 "Case: Invalid devfile",
			componentType:        "",
			expectedMatchesCount: 0,
		},
		{
			name:                 "Case: Valid devfile with wrong component type (CheEditor)",
			componentType:        versionsCommon.DevfileComponentTypeCheEditor,
			alias:                []string{"alias1", "alias2"},
			expectedMatchesCount: 0,
		},
		{
			name:                 "Case: Valid devfile with wrong component type (ChePlugin)",
			componentType:        versionsCommon.DevfileComponentTypeChePlugin,
			alias:                []string{"alias1", "alias2"},
			expectedMatchesCount: 0,
		},
		{
			name:                 "Case: Valid devfile with wrong component type (Kubernetes)",
			componentType:        versionsCommon.DevfileComponentTypeKubernetes,
			alias:                []string{"alias1", "alias2"},
			expectedMatchesCount: 0,
		},
		{
			name:                 "Case: Valid devfile with wrong component type (Openshift)",
			componentType:        versionsCommon.DevfileComponentTypeOpenshift,
			alias:                []string{"alias1", "alias2"},
			expectedMatchesCount: 0,
		},
		{
			name:                 "Case: Valid devfile with correct component type (Dockerimage)",
			componentType:        versionsCommon.DevfileComponentTypeDockerimage,
			alias:                []string{"alias1", "alias2"},
			expectedMatchesCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfile.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: tt.componentType,
				},
			}

			devfileComponents := GetSupportedComponents(devObj.Data)

			componentsMatched := 0
			for _, component := range devfileComponents {
				if component.Type != versionsCommon.DevfileComponentTypeDockerimage {
					t.Errorf("TestGetSupportedComponents error: wrong component type expected %v, actual %v", versionsCommon.DevfileComponentTypeDockerimage, component.Type)
				}
				if util.In(tt.alias, *component.Alias) {
					componentsMatched++
				}
			}

			if componentsMatched != tt.expectedMatchesCount {
				t.Errorf("TestGetSupportedComponents error: wrong number of components matched: expected %v, actual %v", tt.expectedMatchesCount, componentsMatched)
			}
		})
	}

}

func TestGetCommand(t *testing.T) {

	commands := [...]string{"ls -la", "pwd"}
	components := [...]string{"alias1", "alias2"}
	invalidComponent := "garbagealias"
	workDir := [...]string{"/", "/root"}
	validCommandType := common.DevfileCommandTypeExec
	invalidCommandType := common.DevfileCommandType("garbage")
	emptyString := ""

	tests := []struct {
		name              string
		requestedCommands []string
		commandActions    []common.DevfileCommandAction
		isCommandRequired []bool
		wantErr           bool
	}{
		{
			name:              "Case: Valid devfile",
			requestedCommands: []string{"devbuild", "devrun"},
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &commands[0],
					Component: &components[0],
					Workdir:   &workDir[0],
					Type:      &validCommandType,
				},
			},
			isCommandRequired: []bool{false, true},
			wantErr:           false,
		},
		{
			name:              "Case: Wrong command requested",
			requestedCommands: []string{"garbage1"},
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &commands[0],
					Component: &components[0],
					Workdir:   &workDir[0],
					Type:      &validCommandType,
				},
			},
			isCommandRequired: []bool{true},
			wantErr:           true,
		},
		{
			name:              "Case: Invalid devfile with wrong command type",
			requestedCommands: []string{"devbuild"},
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &commands[0],
					Component: &components[0],
					Workdir:   &workDir[0],
					Type:      &invalidCommandType,
				},
			},
			isCommandRequired: []bool{true},
			wantErr:           true,
		},
		{
			name:              "Case: Invalid devfile with empty component",
			requestedCommands: []string{"devbuild"},
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &commands[0],
					Component: &emptyString,
					Workdir:   &workDir[0],
					Type:      &validCommandType,
				},
			},
			isCommandRequired: []bool{false},
			wantErr:           true,
		},
		{
			name:              "Case: Invalid devfile with empty command",
			requestedCommands: []string{"devbuild"},
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &emptyString,
					Component: &components[0],
					Workdir:   &workDir[0],
					Type:      &validCommandType,
				},
			},
			isCommandRequired: []bool{false},
			wantErr:           true,
		},
		{
			name:              "Case: Valid devfile with empty workdir",
			requestedCommands: []string{"devrun"},
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &commands[0],
					Component: &components[0],
					Type:      &validCommandType,
				},
			},
			isCommandRequired: []bool{true},
			wantErr:           false,
		},
		{
			name:              "Case: Invalid command referencing an absent component",
			requestedCommands: []string{"devrun"},
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &commands[0],
					Component: &invalidComponent,
					Type:      &validCommandType,
				},
			},
			isCommandRequired: []bool{true},
			wantErr:           true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfile.DevfileObj{
				Data: testingutil.TestDevfileData{
					CommandActions: tt.commandActions,
					ComponentType:  versionsCommon.DevfileComponentTypeDockerimage,
				},
			}

			for i, commandName := range tt.requestedCommands {
				command, err := getCommand(devObj.Data, commandName, tt.isCommandRequired[i])
				if !tt.wantErr == (err != nil) {
					t.Errorf("TestGetCommand unexpected error for command: %v wantErr: %v err: %v", commandName, tt.wantErr, err)
					return
				} else if tt.wantErr {
					return
				}

				if command.Name != commandName {
					t.Errorf("TestGetCommand error: command names do not match expected: %v actual: %v", commandName, command.Name)
				}

				if len(command.Actions) != 1 {
					t.Errorf("TestGetCommand error: command %v do not have the correct number of actions actual: %v", commandName, len(command.Actions))
				}
			}
		})
	}

}

func TestGetSupportedCommandActions(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	validCommandType := common.DevfileCommandTypeExec
	invalidCommandType := common.DevfileCommandType("garbage")
	emptyString := ""

	tests := []struct {
		name    string
		command common.DevfileCommand
		wantErr bool
	}{
		{
			name: "Case: Valid Command Action",
			command: common.DevfileCommand{
				Name: "testCommand",
				Actions: []versionsCommon.DevfileCommandAction{
					{
						Command:   &command,
						Component: &component,
						Workdir:   &workDir,
						Type:      &validCommandType,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case: Invalid Command Action with empty command",
			command: common.DevfileCommand{
				Name: "testCommand",
				Actions: []versionsCommon.DevfileCommandAction{
					{
						Command:   &emptyString,
						Component: &component,
						Workdir:   &workDir,
						Type:      &validCommandType,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case: Invalid Command Action with missing component",
			command: common.DevfileCommand{
				Name: "testCommand",
				Actions: []versionsCommon.DevfileCommandAction{
					{
						Command: &command,
						Workdir: &workDir,
						Type:    &validCommandType,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case: Invalid Command Action with wrong type",
			command: common.DevfileCommand{
				Name: "testCommand",
				Actions: []versionsCommon.DevfileCommandAction{
					{
						Command:   &command,
						Component: &component,
						Workdir:   &workDir,
						Type:      &invalidCommandType,
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		devObj := devfile.DevfileObj{
			Data: testingutil.TestDevfileData{
				CommandActions: []versionsCommon.DevfileCommandAction{
					{
						Command:   &command,
						Component: &component,
						Type:      &validCommandType,
					},
				},
				ComponentType: versionsCommon.DevfileComponentTypeDockerimage,
			},
		}
		t.Run(tt.name, func(t *testing.T) {
			supportedCommandActions, _ := getSupportedCommandActions(devObj.Data, tt.command)
			if !tt.wantErr && len(supportedCommandActions) != len(tt.command.Actions) {
				t.Errorf("TestGetSupportedCommandActions error: incorrect number of command actions expected: %v actual: %v", len(tt.command.Actions), len(supportedCommandActions))
			} else if tt.wantErr && len(supportedCommandActions) != 0 {
				t.Errorf("TestGetSupportedCommandActions error: incorrect number of command actions expected: %v actual: %v", 0, len(supportedCommandActions))
			}
		})
	}

}

func TestValidateAction(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	validCommandType := common.DevfileCommandTypeExec
	invalidCommandType := common.DevfileCommandType("garbage")
	emptyString := ""

	tests := []struct {
		name    string
		action  common.DevfileCommandAction
		wantErr bool
	}{
		{
			name: "Case: Valid Command Action",
			action: versionsCommon.DevfileCommandAction{
				Command:   &command,
				Component: &component,
				Workdir:   &workDir,
				Type:      &validCommandType,
			},
			wantErr: false,
		},
		{
			name: "Case: Invalid Command Action with empty command",
			action: versionsCommon.DevfileCommandAction{
				Command:   &emptyString,
				Component: &component,
				Workdir:   &workDir,
				Type:      &validCommandType,
			},
			wantErr: true,
		},
		{
			name: "Case: Invalid Command Action with missing component",
			action: versionsCommon.DevfileCommandAction{
				Command: &command,
				Workdir: &workDir,
				Type:    &validCommandType,
			},
			wantErr: true,
		},
		{
			name: "Case: Invalid Command Action with wrong type",
			action: versionsCommon.DevfileCommandAction{
				Command:   &command,
				Component: &component,
				Workdir:   &workDir,
				Type:      &invalidCommandType,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		devObj := devfile.DevfileObj{
			Data: testingutil.TestDevfileData{
				CommandActions: []versionsCommon.DevfileCommandAction{
					{
						Command:   &command,
						Component: &component,
						Type:      &validCommandType,
					},
				},
				ComponentType: versionsCommon.DevfileComponentTypeDockerimage,
			},
		}
		t.Run(tt.name, func(t *testing.T) {
			err := validateAction(devObj.Data, tt.action)
			if !tt.wantErr == (err != nil) {
				t.Errorf("TestValidateAction unexpected error: %v", err)
				return
			}
		})
	}

}

func TestGetBuildCommand(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	validCommandType := common.DevfileCommandTypeExec
	emptyString := ""

	var emptyCommand common.DevfileCommand

	tests := []struct {
		name           string
		commandName    string
		commandActions []common.DevfileCommandAction
		wantErr        bool
	}{
		{
			name:        "Case: Default Build Command",
			commandName: emptyString,
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &command,
					Component: &component,
					Workdir:   &workDir,
					Type:      &validCommandType,
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Custom Build Command",
			commandName: "customcommand",
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &command,
					Component: &component,
					Workdir:   &workDir,
					Type:      &validCommandType,
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Missing Build Command",
			commandName: "customcommand123",
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &command,
					Component: &component,
					Workdir:   &workDir,
					Type:      &validCommandType,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfile.DevfileObj{
				Data: testingutil.TestDevfileData{
					CommandActions: tt.commandActions,
					ComponentType:  versionsCommon.DevfileComponentTypeDockerimage,
				},
			}

			command, err := GetBuildCommand(devObj.Data, tt.commandName)

			if !tt.wantErr == (err != nil) {
				t.Errorf("TestGetBuildCommand: unexpected error for command \"%v\" expected: %v actual: %v", tt.commandName, tt.wantErr, err)
			} else if !tt.wantErr && reflect.DeepEqual(emptyCommand, command) {
				t.Errorf("TestGetBuildCommand: unexpected empty command returned for command: %v", tt.commandName)
			}

		})
	}

}

func TestGetRunCommand(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	validCommandType := common.DevfileCommandTypeExec
	emptyString := ""

	var emptyCommand common.DevfileCommand

	tests := []struct {
		name           string
		commandName    string
		commandActions []common.DevfileCommandAction
		wantErr        bool
	}{
		{
			name:        "Case: Default Run Command",
			commandName: emptyString,
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &command,
					Component: &component,
					Workdir:   &workDir,
					Type:      &validCommandType,
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Custom Run Command",
			commandName: "customcommand",
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &command,
					Component: &component,
					Workdir:   &workDir,
					Type:      &validCommandType,
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Missing Run Command",
			commandName: "customcommand123",
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &command,
					Component: &component,
					Workdir:   &workDir,
					Type:      &validCommandType,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfile.DevfileObj{
				Data: testingutil.TestDevfileData{
					CommandActions: tt.commandActions,
					ComponentType:  versionsCommon.DevfileComponentTypeDockerimage,
				},
			}

			command, err := GetRunCommand(devObj.Data, tt.commandName)

			if !tt.wantErr == (err != nil) {
				t.Errorf("TestGetRunCommand: unexpected error for command \"%v\" expected: %v actual: %v", tt.commandName, tt.wantErr, err)
			} else if !tt.wantErr && reflect.DeepEqual(emptyCommand, command) {
				t.Errorf("TestGetRunCommand: unexpected empty command returned for command: %v", tt.commandName)
			}
		})
	}

}

func TestValidateAndGetPushDevfileCommands(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	validCommandType := common.DevfileCommandTypeExec
	emptyString := ""

	actions := []versionsCommon.DevfileCommandAction{
		{
			Command:   &command,
			Component: &component,
			Workdir:   &workDir,
			Type:      &validCommandType,
		},
	}

	tests := []struct {
		name                string
		buildCommand        string
		runCommand          string
		numberOfCommands    int
		componentType       versionsCommon.DevfileComponentType
		missingBuildCommand bool
		wantErr             bool
	}{
		{
			name:             "Case: Default Devfile Commands",
			buildCommand:     emptyString,
			runCommand:       emptyString,
			numberOfCommands: 2,
			componentType:    versionsCommon.DevfileComponentTypeDockerimage,
			wantErr:          false,
		},
		{
			name:             "Case: Default Build Command and Provided Run Command",
			buildCommand:     emptyString,
			runCommand:       "customcommand",
			numberOfCommands: 2,
			componentType:    versionsCommon.DevfileComponentTypeDockerimage,
			wantErr:          false,
		},
		{
			name:             "Case: Provided Build Command and Provided Run Command",
			buildCommand:     "customcommand",
			runCommand:       "customcommand",
			numberOfCommands: 2,
			componentType:    versionsCommon.DevfileComponentTypeDockerimage,
			wantErr:          false,
		},
		{
			name:             "Case: No Dockerimage Component",
			buildCommand:     "customcommand",
			runCommand:       "customcommand",
			numberOfCommands: 0,
			componentType:    "",
			wantErr:          true,
		},
		{
			name:             "Case: Provided Wrong Build Command and Provided Run Command",
			buildCommand:     "customcommand123",
			runCommand:       "customcommand",
			numberOfCommands: 1,
			componentType:    versionsCommon.DevfileComponentTypeDockerimage,
			wantErr:          true,
		},
		{
			name:                "Case: Missing Build Command and Provided Run Command",
			buildCommand:        emptyString,
			runCommand:          "customcommand",
			numberOfCommands:    1,
			componentType:       versionsCommon.DevfileComponentTypeDockerimage,
			missingBuildCommand: true,
			wantErr:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfile.DevfileObj{
				Data: testingutil.TestDevfileData{
					CommandActions:      actions,
					ComponentType:       tt.componentType,
					MissingBuildCommand: tt.missingBuildCommand,
				},
			}

			pushCommands, err := ValidateAndGetPushDevfileCommands(devObj.Data, tt.buildCommand, tt.runCommand)
			if !tt.wantErr == (err != nil) {
				t.Errorf("TestValidateAndGetPushDevfileCommands unexpected error when validating commands wantErr: %v err: %v", tt.wantErr, err)
			} else if tt.wantErr && err != nil {
				return
			}

			if len(pushCommands) != tt.numberOfCommands {
				t.Errorf("TestValidateAndGetPushDevfileCommands error: wrong number of validated commands expected: %v actual :%v", tt.numberOfCommands, len(pushCommands))
			}
		})
	}

}
