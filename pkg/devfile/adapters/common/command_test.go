package common

import (
	"reflect"
	"testing"

	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/testingutil"
)

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
			name:              "Case 1: Valid devfile",
			requestedCommands: []string{"devbuild", "devrun"},
			commandActions: []common.DevfileCommandAction{
				{
					Command:   &commands[0],
					Component: &components[0],
					Workdir:   &workDir[0],
					Type:      &validCommandType,
				},
			},
			isCommandRequired: []bool{false, false, true},
			wantErr:           false,
		},
		{
			name:              "Case 2: Valid devfile with devinit and devbuild",
			requestedCommands: []string{"devinit", "devbuild", "devrun"},
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &commands[0],
					Component: &components[0],
					Workdir:   &workDir[0],
					Type:      &validCommandType,
				},
			},
			isCommandRequired: []bool{false, false, true},
			wantErr:           false,
		},
		{
			name:              "Case 3: Valid devfile with devinit and devrun",
			requestedCommands: []string{"devinit", "devrun"},
			commandActions: []versionsCommon.DevfileCommandAction{
				{
					Command:   &commands[0],
					Component: &components[0],
					Workdir:   &workDir[0],
					Type:      &validCommandType,
				},
			},
			isCommandRequired: []bool{false, false, true},
			wantErr:           false,
		},
		{
			name:              "Case 4: Wrong command requested",
			requestedCommands: []string{"garbage1"},
			commandActions: []common.DevfileCommandAction{
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
			name:              "Case 5: Invalid devfile with wrong devinit command type",
			requestedCommands: []string{"devinit"},
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
			name:              "Case 6: Invalid devfile with empty devinit component",
			requestedCommands: []string{"devinit"},
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
			name:              "Case 7: Invalid devfile with empty devinit command",
			requestedCommands: []string{"devinit"},
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
			name:              "Case 8: Invalid devfile with wrong devbuild command type",
			requestedCommands: []string{"devbuild"},
			commandActions: []common.DevfileCommandAction{
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
			name:              "Case 9: Invalid devfile with empty devbuild component",
			requestedCommands: []string{"devbuild"},
			commandActions: []common.DevfileCommandAction{
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
			name:              "Case 10: Invalid devfile with empty devbuild command",
			requestedCommands: []string{"devbuild"},
			commandActions: []common.DevfileCommandAction{
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
			name:              "Case 11: Valid devfile with empty workdir",
			requestedCommands: []string{"devrun"},
			commandActions: []common.DevfileCommandAction{
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
			name:              "Case 12: Invalid command referencing an absent component",
			requestedCommands: []string{"devrun"},
			commandActions: []common.DevfileCommandAction{
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
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					CommandActions: tt.commandActions,
					ComponentType:  common.DevfileComponentTypeDockerimage,
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
			action: common.DevfileCommandAction{
				Command:   &command,
				Component: &component,
				Workdir:   &workDir,
				Type:      &validCommandType,
			},
			wantErr: false,
		},
		{
			name: "Case: Invalid Command Action with empty command",
			action: common.DevfileCommandAction{
				Command:   &emptyString,
				Component: &component,
				Workdir:   &workDir,
				Type:      &validCommandType,
			},
			wantErr: true,
		},
		{
			name: "Case: Invalid Command Action with missing component",
			action: common.DevfileCommandAction{
				Command: &command,
				Workdir: &workDir,
				Type:    &validCommandType,
			},
			wantErr: true,
		},
		{
			name: "Case: Invalid Command Action with wrong type",
			action: common.DevfileCommandAction{
				Command:   &command,
				Component: &component,
				Workdir:   &workDir,
				Type:      &invalidCommandType,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		devObj := devfileParser.DevfileObj{
			Data: testingutil.TestDevfileData{
				CommandActions: []common.DevfileCommandAction{
					{
						Command:   &command,
						Component: &component,
						Type:      &validCommandType,
					},
				},
				ComponentType: common.DevfileComponentTypeDockerimage,
			},
		}
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommand(devObj.Data, tt.action)
			if !tt.wantErr == (err != nil) {
				t.Errorf("TestValidateAction unexpected error: %v", err)
				return
			}
		})
	}

}

func TestGetInitCommand(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	validCommandType := common.ExecCommandType
	emptyString := ""

	var emptyCommand common.DevfileCommand

	tests := []struct {
		name           string
		commandName    string
		commandActions []common.DevfileCommandAction
		wantErr        bool
	}{
		{
			name:        "Case: Default Init Command",
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
			name:        "Case: Custom Init Command",
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
			name:        "Case: Missing Init Command",
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
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					CommandActions: tt.commandActions,
					ComponentType:  versionsCommon.DevfileComponentTypeDockerimage,
				},
			}

			command, err := GetInitCommand(devObj.Data, tt.commandName)

			if !tt.wantErr == (err != nil) {
				t.Errorf("TestGetInitCommand: unexpected error for command \"%v\" expected: %v actual: %v", tt.commandName, tt.wantErr, err)
			} else if !tt.wantErr && reflect.DeepEqual(emptyCommand, command) {
				t.Errorf("TestGetInitCommand: unexpected empty command returned for command: %v", tt.commandName)
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
			commandActions: []common.DevfileCommandAction{
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
			commandActions: []common.DevfileCommandAction{
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
			commandActions: []common.DevfileCommandAction{
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
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					CommandActions: tt.commandActions,
					ComponentType:  common.DevfileComponentTypeDockerimage,
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
			commandActions: []common.DevfileCommandAction{
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
			commandActions: []common.DevfileCommandAction{
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
			commandActions: []common.DevfileCommandAction{
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
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					CommandActions: tt.commandActions,
					ComponentType:  common.DevfileComponentTypeDockerimage,
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

	actions := []common.DevfileCommandAction{
		{
			Command:   &command,
			Component: &component,
			Workdir:   &workDir,
			Type:      &validCommandType,
		},
	}

	tests := []struct {
		name                string
		initCommand         string
		buildCommand        string
		runCommand          string
		numberOfCommands    int
		componentType       common.DevfileComponentType
		missingInitCommand  bool
		missingBuildCommand bool
		wantErr             bool
	}{
		{
			name:             "Case: Default Devfile Commands",
			initCommand:      emptyString,
			buildCommand:     emptyString,
			runCommand:       emptyString,
			numberOfCommands: 3,
			componentType:    common.DevfileComponentTypeDockerimage,
			wantErr:          false,
		},
		{
			name:             "Case: Default Init and Build Command, and Provided Run Command",
			initCommand:      emptyString,
			buildCommand:     emptyString,
			runCommand:       "customcommand",
			numberOfCommands: 3,
			componentType:    common.DevfileComponentTypeDockerimage,
			wantErr:          false,
		},
		{
			name:             "Case: No Dockerimage Component",
			initCommand:      emptyString,
			buildCommand:     "customcommand",
			runCommand:       "customcommand",
			numberOfCommands: 0,
			componentType:    "",
			wantErr:          true,
		},
		{
			name:             "Case: Provided Wrong Build Command and Provided Run Command",
			initCommand:      emptyString,
			buildCommand:     "customcommand123",
			runCommand:       "customcommand",
			numberOfCommands: 1,
			componentType:    common.DevfileComponentTypeDockerimage,
			wantErr:          true,
		},
		{
			name:             "Case: Provided Wrong Init Command and Provided Build and Run Command",
			initCommand:      "customcommand123",
			buildCommand:     emptyString,
			runCommand:       "customcommand",
			numberOfCommands: 1,
			componentType:    versionsCommon.DevfileComponentTypeDockerimage,
			wantErr:          true,
		},
		{
			name:                "Case: Missing Init and Build Command, and Provided Run Command",
			initCommand:         emptyString,
			buildCommand:        emptyString,
			runCommand:          "customcommand",
			numberOfCommands:    1,
			componentType:       common.DevfileComponentTypeDockerimage,
			missingInitCommand:  true,
			missingBuildCommand: true,
			wantErr:             false,
		},
		{
			name:               "Case: Missing Init Command with provided Build and Run Command",
			initCommand:        emptyString,
			buildCommand:       "customcommand",
			runCommand:         "customcommand",
			numberOfCommands:   2,
			componentType:      versionsCommon.DevfileComponentTypeDockerimage,
			missingInitCommand: true,
			wantErr:            false,
		},
		{
			name:                "Case: Missing Build Command with provided Init and Run Command",
			initCommand:         "customcommand",
			buildCommand:        emptyString,
			runCommand:          "customcommand",
			numberOfCommands:    2,
			componentType:       versionsCommon.DevfileComponentTypeDockerimage,
			missingBuildCommand: true,
			wantErr:             false,
		},
		{
			name:             "Case: Optional Init Command with provided Build and Run Command",
			initCommand:      "customcommand",
			buildCommand:     "customcommand",
			runCommand:       "customcommand",
			numberOfCommands: 3,
			componentType:    versionsCommon.DevfileComponentTypeDockerimage,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					CommandActions:      actions,
					ComponentType:       tt.componentType,
					MissingInitCommand:  tt.missingInitCommand,
					MissingBuildCommand: tt.missingBuildCommand,
				},
			}

			pushCommands, err := ValidateAndGetPushDevfileCommands(devObj.Data, tt.initCommand, tt.buildCommand, tt.runCommand)
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
