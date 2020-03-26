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
		name          string
		componentType versionsCommon.DevfileComponentType
		alias         []string
	}{
		{
			name:          "Case: Invalid devfile",
			componentType: "",
		},
		{
			name:          "Case: Valid devfile",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			alias:         []string{"alias1", "alias2"},
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

			if componentsMatched != len(tt.alias) {
				t.Errorf("TestGetSupportedComponents error: wrong number of components matched: expected %v, actual %v", len(tt.alias), componentsMatched)
			}
		})
	}

}

func TestGetCommand(t *testing.T) {

	var emptyCommand common.DevfileCommand

	commands := [...]string{"ls -la", "pwd"}
	components := [...]string{"alias1", "alias2"}
	workDir := [...]string{"/", "/root"}
	validCommandType := common.DevfileCommandTypeExec
	invalidCommandType := common.DevfileCommandType("garbage")
	emptyString := ""

	tests := []struct {
		name           string
		commandNames   []string
		commandActions []common.DevfileCommandAction
		wantErr        bool
	}{
		{
			name:         "Case: Valid devfile",
			commandNames: []string{"devbuild", "devrun"},
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &commands[0],
					Component: &components[0],
					Workdir:   &workDir[0],
					Type:      &validCommandType,
				},
			},
			wantErr: false,
		},
		{
			name:         "Case: Wrong Command",
			commandNames: []string{"garbage1"},
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &commands[0],
					Component: &components[0],
					Workdir:   &workDir[0],
					Type:      &validCommandType,
				},
			},
			wantErr: true,
		},
		{
			name:         "Case: Invalid devfile with wrong command type",
			commandNames: []string{"devbuild"},
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &commands[0],
					Component: &components[0],
					Workdir:   &workDir[0],
					Type:      &invalidCommandType,
				},
			},
			wantErr: true,
		},
		{
			name:         "Case: Invalid devfile with empty component",
			commandNames: []string{"devbuild"},
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &commands[0],
					Component: &emptyString,
					Workdir:   &workDir[0],
					Type:      &validCommandType,
				},
			},
			wantErr: true,
		},
		{
			name:         "Case: Invalid devfile with empty command",
			commandNames: []string{"devbuild"},
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &emptyString,
					Component: &components[0],
					Workdir:   &workDir[0],
					Type:      &validCommandType,
				},
			},
			wantErr: true,
		},
		{
			name:         "Case: Valid devfile with empty workdir",
			commandNames: []string{"devrun"},
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &commands[0],
					Component: &components[0],
					Type:      &validCommandType,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfile.DevfileObj{
				Data: testingutil.TestDevfileData{
					CommandActions: tt.commandActions,
				},
			}

			for _, commandName := range tt.commandNames {
				command := GetCommand(devObj.Data, commandName)

				if !tt.wantErr && reflect.DeepEqual(emptyCommand, command) {
					t.Errorf("TestGetCommand error: could not find devfile command for %v", commandName)
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
		t.Run(tt.name, func(t *testing.T) {
			supportedCommandActions := getSupportedCommandActions(tt.command)
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
		t.Run(tt.name, func(t *testing.T) {
			isValid := validateAction(tt.action)
			if tt.wantErr == isValid {
				t.Errorf("TestValidateAction error: command action validation failed expected: %v actual: %v", !isValid, isValid)
			} else if tt.wantErr && isValid {
				t.Errorf("TestValidateAction error: command action validation failed expected: %v actual: %v", !isValid, isValid)
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
				},
			}

			command := GetBuildCommand(devObj.Data, tt.commandName)

			if !tt.wantErr && reflect.DeepEqual(emptyCommand, command) {
				t.Errorf("TestGetBuildCommand error: command not found expected: %v actual: <EMPTY>", tt.commandName)
			} else if tt.wantErr && !reflect.DeepEqual(emptyCommand, command) {
				t.Errorf("TestGetBuildCommand error: command found expected: <EMPTY> actual: %v", command.Name)
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
				},
			}

			command := GetRunCommand(devObj.Data, tt.commandName)

			if !tt.wantErr && reflect.DeepEqual(emptyCommand, command) {
				t.Errorf("TestGetRunCommand error: command not found expected: %v actual: <EMPTY>", tt.commandName)
			} else if tt.wantErr && !reflect.DeepEqual(emptyCommand, command) {
				t.Errorf("TestGetRunCommand error: command found expected: <EMPTY> actual: %v", command.Name)
			}

		})
	}

}

func TestIsCommandPresent(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	validCommandType := common.DevfileCommandTypeExec

	var emptyCommand common.DevfileCommand

	tests := []struct {
		name    string
		command common.DevfileCommand
		wantErr bool
	}{
		{
			name: "Case: Valid Command",
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
			name:    "Case: Valid Command",
			command: emptyCommand,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			isCommandPresent := IsCommandPresent(tt.command)

			if isCommandPresent != !tt.wantErr {
				t.Errorf("TestIsCommandPresent error: expected: %v actual: %v", !tt.wantErr, isCommandPresent)
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
			if !tt.wantErr && err != nil {
				t.Errorf("TestValidateAndGetPushDevfileCommands unexpected error when validating commands: %v", err.Error())
			} else if tt.wantErr && err != nil {
				return
			}

			if len(pushCommands) != tt.numberOfCommands {
				t.Errorf("TestValidateAndGetPushDevfileCommands error: wrong number of validated commands expected: %v actual :%v", tt.numberOfCommands, len(pushCommands))
			}
		})
	}

}

func TestIsCommandValid(t *testing.T) {

	command := "ls -la"
	component := []string{"alias1", "garbagealias"}
	workDir := "/"
	validCommandType := common.DevfileCommandTypeExec

	tests := []struct {
		name          string
		command       common.DevfileCommand
		componentType versionsCommon.DevfileComponentType
		wantErr       bool
	}{
		{
			name: "Case: Valid Command",
			command: common.DevfileCommand{
				Name: "testCommand",
				Actions: []versionsCommon.DevfileCommandAction{
					{
						Command:   &command,
						Component: &component[0],
						Workdir:   &workDir,
						Type:      &validCommandType,
					},
				},
			},
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			wantErr:       false,
		},
		{
			name: "Case: Invalid Command Referencing Component With Wrong Type",
			command: common.DevfileCommand{
				Name: "testCommand",
				Actions: []versionsCommon.DevfileCommandAction{
					{
						Command:   &command,
						Component: &component[0],
						Workdir:   &workDir,
						Type:      &validCommandType,
					},
				},
			},
			componentType: "",
			wantErr:       true,
		},
		{
			name: "Case: Invalid Command Referencing An Absent Component",
			command: common.DevfileCommand{
				Name: "testCommand",
				Actions: []versionsCommon.DevfileCommandAction{
					{
						Command:   &command,
						Component: &component[1],
						Workdir:   &workDir,
						Type:      &validCommandType,
					},
				},
			},
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfile.DevfileObj{
				Data: testingutil.TestDevfileData{
					CommandActions: tt.command.Actions,
					ComponentType:  tt.componentType,
				},
			}

			isCommandValid := IsCommandValid(devObj.Data, tt.command)

			if tt.wantErr == isCommandValid {
				t.Errorf("TestIsCommandValid unexpected error when validating command %v expected: %v actual: %v", tt.command.Name, !tt.wantErr, isCommandValid)
			} else if tt.wantErr && isCommandValid {
				t.Errorf("TestIsCommandValid unexpected error when validating command %v expected: %v actual: %v", tt.command.Name, !tt.wantErr, isCommandValid)
			}

		})
	}

}
