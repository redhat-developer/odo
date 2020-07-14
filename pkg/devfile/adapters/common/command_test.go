package common

import (
	"reflect"
	"testing"

	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/testingutil"
)

var buildGroup = common.BuildCommandGroupType
var runGroup = common.RunCommandGroupType
var testGroup = common.TestCommandGroupType
var debugGroup = common.DebugCommandGroupType
var initGroup = common.InitCommandGroupType

func TestGetCommand(t *testing.T) {

	commands := [...]string{"ls -la", "pwd"}
	components := [...]string{"alias1", "alias2"}
	invalidComponent := "garbagealias"
	workDir := [...]string{"/", "/root"}

	emptyString := ""

	tests := []struct {
		name           string
		requestedType  []common.DevfileCommandGroupType
		execCommands   []common.Exec
		reqCommandName string
		retCommandName string
		wantErr        bool
	}{
		{
			name: "Case 1: Valid devfile",
			execCommands: []versionsCommon.Exec{
				getExecCommand("", buildGroup),
				getExecCommand("", runGroup),
			},
			requestedType: []common.DevfileCommandGroupType{buildGroup, runGroup},
			wantErr:       false,
		},
		{
			name: "Case 2: Valid devfile with devinit and devbuild",
			execCommands: []versionsCommon.Exec{
				getExecCommand("", buildGroup),
				getExecCommand("", runGroup),
			},
			requestedType: []common.DevfileCommandGroupType{initGroup, buildGroup, runGroup},
			wantErr:       false,
		},
		{
			name: "Case 3: Valid devfile with devinit and devrun",
			execCommands: []versionsCommon.Exec{
				getExecCommand("", initGroup),
				getExecCommand("", runGroup),
			},
			requestedType: []common.DevfileCommandGroupType{initGroup, runGroup},
			wantErr:       false,
		},
		{
			name: "Case 4: Invalid devfile with empty component",
			execCommands: []versionsCommon.Exec{
				{
					CommandLine: commands[0],
					Component:   emptyString,
					WorkingDir:  workDir[0],
					Group:       &versionsCommon.Group{Kind: initGroup},
				},
			},
			requestedType: []common.DevfileCommandGroupType{initGroup},
			wantErr:       true,
		},
		{
			name: "Case 5: Invalid devfile with empty devinit command",
			execCommands: []versionsCommon.Exec{
				{
					CommandLine: emptyString,
					Component:   components[0],
					WorkingDir:  workDir[0],
					Group:       &versionsCommon.Group{Kind: initGroup},
				},
			},
			requestedType: []common.DevfileCommandGroupType{initGroup},
			wantErr:       true,
		},
		{
			name: "Case 6: Valid devfile with empty workdir",
			execCommands: []common.Exec{
				{
					CommandLine: commands[0],
					Component:   components[0],
					Group:       &versionsCommon.Group{Kind: runGroup},
				},
			},
			requestedType: []common.DevfileCommandGroupType{runGroup},
			wantErr:       false,
		},
		{
			name: "Case 7: Invalid command referencing an absent component",
			execCommands: []common.Exec{
				{
					CommandLine: commands[0],
					Component:   invalidComponent,
					Group:       &versionsCommon.Group{Kind: runGroup},
				},
			},
			requestedType: []common.DevfileCommandGroupType{runGroup},
			wantErr:       true,
		},
		{
			name: "Case 8: Mismatched command type",
			execCommands: []common.Exec{
				{
					Id:          "build command",
					CommandLine: commands[0],
					Component:   components[0],
					Group:       &versionsCommon.Group{Kind: runGroup},
				},
			},
			reqCommandName: "build command",
			requestedType:  []common.DevfileCommandGroupType{buildGroup},
			wantErr:        true,
		},
		{
			name: "Case 9: Default command is returned",
			execCommands: []common.Exec{
				{
					Id:          "defaultRunCommand",
					CommandLine: commands[0],
					Component:   components[0],
					Group:       &versionsCommon.Group{Kind: runGroup, IsDefault: true},
				},
				{
					Id:          "runCommand",
					CommandLine: commands[0],
					Component:   components[0],
					Group:       &versionsCommon.Group{Kind: runGroup},
				},
			},
			retCommandName: "defaultRunCommand",
			requestedType:  []common.DevfileCommandGroupType{runGroup},
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := []common.DevfileComponent{testingutil.GetFakeComponent(tt.execCommands[0].Component)}
			if tt.execCommands[0].Component == invalidComponent {
				components = []common.DevfileComponent{testingutil.GetFakeComponent("randomComponent")}
			}
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ExecCommands: tt.execCommands,
					Components:   components,
				},
			}

			for _, gtype := range tt.requestedType {
				cmd, err := getCommand(devObj.Data, tt.reqCommandName, gtype)
				if !tt.wantErr == (err != nil) {
					t.Errorf("TestGetCommand unexpected error for command: %v wantErr: %v err: %v", gtype, tt.wantErr, err)
					return
				} else if tt.wantErr {
					return
				}

				if cmd.Exec != nil {
					if cmd.Exec.Id != tt.retCommandName {
						t.Errorf("TestGetCommand error: command names do not match expected: %v actual: %v", tt.retCommandName, cmd.Exec.Id)
					}
				}
			}
		})
	}

}

func TestGetCommandFromDevfile(t *testing.T) {

	commands := [...]string{"ls -la", "pwd"}
	components := [...]string{"alias1", "alias2"}
	invalidComponent := "garbagealias"
	workDir := [...]string{"/", "/root"}

	emptyString := ""

	tests := []struct {
		name           string
		requestedType  []common.DevfileCommandGroupType
		execCommands   []common.Exec
		retCommandName string
		wantErr        bool
	}{
		{
			name: "Case 1: Valid devfile",
			execCommands: []versionsCommon.Exec{
				getExecCommand("", buildGroup),
				getExecCommand("", runGroup),
			},
			requestedType: []common.DevfileCommandGroupType{buildGroup, runGroup},
			wantErr:       false,
		},
		{
			name: "Case 2: Valid devfile with devinit and devbuild",
			execCommands: []versionsCommon.Exec{
				getExecCommand("", buildGroup),
				getExecCommand("", runGroup),
			},
			requestedType: []common.DevfileCommandGroupType{initGroup, buildGroup, runGroup},
			wantErr:       false,
		},
		{
			name: "Case 3: Valid devfile with devinit and devrun",
			execCommands: []versionsCommon.Exec{
				getExecCommand("", initGroup),
				getExecCommand("", runGroup),
			},
			requestedType: []common.DevfileCommandGroupType{initGroup, runGroup},
			wantErr:       false,
		},
		{
			name: "Case 4: Invalid devfile with empty component",
			execCommands: []versionsCommon.Exec{
				{
					CommandLine: commands[0],
					Component:   emptyString,
					WorkingDir:  workDir[0],
					Group:       &versionsCommon.Group{Kind: initGroup},
				},
			},
			requestedType: []common.DevfileCommandGroupType{initGroup},
			wantErr:       true,
		},
		{
			name: "Case 5: Invalid devfile with empty devinit command",
			execCommands: []versionsCommon.Exec{
				{
					CommandLine: emptyString,
					Component:   components[0],
					WorkingDir:  workDir[0],
					Group:       &versionsCommon.Group{Kind: initGroup},
				},
			},
			requestedType: []common.DevfileCommandGroupType{initGroup},
			wantErr:       true,
		},
		{
			name: "Case 6: Valid devfile with empty workdir",
			execCommands: []common.Exec{
				{
					CommandLine: commands[0],
					Component:   components[0],
					Group:       &versionsCommon.Group{Kind: runGroup},
				},
			},
			requestedType: []common.DevfileCommandGroupType{runGroup},
			wantErr:       false,
		},
		{
			name: "Case 7: Invalid command referencing an absent component",
			execCommands: []common.Exec{
				{
					CommandLine: commands[0],
					Component:   invalidComponent,
					Group:       &versionsCommon.Group{Kind: runGroup},
				},
			},
			requestedType: []common.DevfileCommandGroupType{runGroup},
			wantErr:       true,
		},
		{
			name: "Case 8: Default command is returned",
			execCommands: []common.Exec{
				{
					Id:          "defaultRunCommand",
					CommandLine: commands[0],
					Component:   components[0],
					Group:       &versionsCommon.Group{Kind: runGroup, IsDefault: true},
				},
				{
					Id:          "runCommand",
					CommandLine: commands[0],
					Component:   components[0],
					Group:       &versionsCommon.Group{Kind: runGroup},
				},
			},
			retCommandName: "defaultRunCommand",
			requestedType:  []common.DevfileCommandGroupType{runGroup},
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := []common.DevfileComponent{testingutil.GetFakeComponent(tt.execCommands[0].Component)}
			if tt.execCommands[0].Component == invalidComponent {
				components = []common.DevfileComponent{testingutil.GetFakeComponent("randomComponent")}
			}
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ExecCommands: tt.execCommands,
					Components:   components,
				},
			}

			for _, gtype := range tt.requestedType {
				cmd, err := getCommandFromDevfile(devObj.Data, gtype)
				if !tt.wantErr == (err != nil) {
					t.Errorf("TestGetCommandFromDevfile unexpected error for command: %v wantErr: %v err: %v", gtype, tt.wantErr, err)
					return
				} else if tt.wantErr {
					return
				}

				if cmd.Exec != nil {
					if cmd.Exec.Id != tt.retCommandName {
						t.Errorf("TestGetCommandFromDevfile error: command names do not match expected: %v actual: %v", tt.retCommandName, cmd.Exec.Id)
					}
				}
			}
		})
	}

}

func TestGetCommandFromFlag(t *testing.T) {

	commands := [...]string{"ls -la", "pwd"}
	components := [...]string{"alias1", "alias2"}
	invalidComponent := "garbagealias"
	workDir := [...]string{"/", "/root"}

	emptyString := ""

	tests := []struct {
		name           string
		requestedType  common.DevfileCommandGroupType
		execCommands   []common.Exec
		reqCommandName string
		retCommandName string
		wantErr        bool
	}{
		{
			name: "Case 1: Valid devfile",
			execCommands: []versionsCommon.Exec{
				getExecCommand("a", buildGroup),
				getExecCommand("b", runGroup),
			},
			reqCommandName: "b",
			retCommandName: "b",
			requestedType:  runGroup,
			wantErr:        false,
		},
		{
			name: "Case 2: Invalid devfile with empty component",
			execCommands: []versionsCommon.Exec{
				{
					Id:          "build command",
					CommandLine: commands[0],
					Component:   emptyString,
					WorkingDir:  workDir[0],
					Group:       &versionsCommon.Group{Kind: buildGroup},
				},
			},
			reqCommandName: "build command",
			requestedType:  buildGroup,
			wantErr:        true,
		},
		{
			name: "Case 3: Valid devfile with empty workdir",
			execCommands: []common.Exec{
				{
					Id:          "build command",
					CommandLine: commands[0],
					Component:   components[0],
					Group:       &versionsCommon.Group{Kind: runGroup},
				},
			},
			reqCommandName: "build command",
			retCommandName: "build command",
			requestedType:  runGroup,
			wantErr:        false,
		},
		{
			name: "Case 4: Invalid command",
			execCommands: []common.Exec{
				{
					Id:          "build command",
					CommandLine: commands[0],
					Component:   invalidComponent,
					Group:       &versionsCommon.Group{Kind: runGroup},
				},
			},
			reqCommandName: "build command wrong",
			requestedType:  runGroup,
			wantErr:        true,
		},
		{
			name: "Case 5: Mismatched command type",
			execCommands: []common.Exec{
				{
					Id:          "build command",
					CommandLine: commands[0],
					Component:   components[0],
					Group:       &versionsCommon.Group{Kind: runGroup},
				},
			},
			reqCommandName: "build command",
			requestedType:  buildGroup,
			wantErr:        true,
		},
		{
			name: "Case 6: Multiple default commands but should be with the flag",
			execCommands: []common.Exec{
				{
					Id:          "defaultRunCommand",
					CommandLine: commands[0],
					Component:   components[0],
					Group:       &versionsCommon.Group{Kind: runGroup, IsDefault: true},
				},
				{
					Id:          "runCommand",
					CommandLine: commands[0],
					Component:   components[0],
					Group:       &versionsCommon.Group{Kind: runGroup, IsDefault: true},
				},
			},
			reqCommandName: "defaultRunCommand",
			retCommandName: "defaultRunCommand",
			requestedType:  runGroup,
			wantErr:        false,
		},
		{
			name: "Case 7: No default command but should be with the flag",
			execCommands: []common.Exec{
				{
					Id:          "defaultRunCommand",
					CommandLine: commands[0],
					Component:   components[0],
					Group:       &versionsCommon.Group{Kind: runGroup},
				},
				{
					Id:          "runCommand",
					CommandLine: commands[0],
					Component:   components[0],
					Group:       &versionsCommon.Group{Kind: runGroup},
				},
			},
			reqCommandName: "defaultRunCommand",
			retCommandName: "defaultRunCommand",
			requestedType:  runGroup,
			wantErr:        false,
		},
		{
			name: "Case 8: No Command Group",
			execCommands: []common.Exec{
				{
					Id:          "defaultRunCommand",
					CommandLine: commands[0],
					Component:   components[0],
				},
			},
			reqCommandName: "defaultRunCommand",
			retCommandName: "defaultRunCommand",
			requestedType:  runGroup,
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := []common.DevfileComponent{testingutil.GetFakeComponent(tt.execCommands[0].Component)}
			if tt.execCommands[0].Component == invalidComponent {
				components = []common.DevfileComponent{testingutil.GetFakeComponent("randomComponent")}
			}
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ExecCommands: tt.execCommands,
					Components:   components,
				},
			}

			cmd, err := getCommandFromFlag(devObj.Data, tt.requestedType, tt.reqCommandName)
			if !tt.wantErr == (err != nil) {
				t.Errorf("TestGetCommand unexpected error for command: %v wantErr: %v err: %v", tt.requestedType, tt.wantErr, err)
				return
			} else if tt.wantErr {
				return
			}

			if cmd.Exec != nil {
				if cmd.Exec.Id != tt.retCommandName {
					t.Errorf("TestGetCommand error: command names do not match expected: %v actual: %v", tt.retCommandName, cmd.Exec.Id)
				}
			}
		})
	}

}

func TestValidateCommandsForGroup(t *testing.T) {

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
			Id:          "build command 2",
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
			Group: &versionsCommon.Group{
				Kind:      debugGroup,
				IsDefault: true,
			},
		},
		{
			Id:          "customcommand",
			CommandLine: command,
			Component:   componentName,
			WorkingDir:  workDir,
			Group: &versionsCommon.Group{
				Kind:      runGroup,
				IsDefault: true,
			},
		},
	}

	devObj := devfileParser.DevfileObj{
		Data: testingutil.TestDevfileData{
			Components: []versionsCommon.DevfileComponent{
				testingutil.GetFakeComponent("alias1"),
			},
			ExecCommands: execCommands,
		},
	}

	tests := []struct {
		name      string
		groupType common.DevfileCommandGroupType
		wantErr   bool
	}{
		{
			name:      "Case 1: Two default run commands",
			groupType: runGroup,
			wantErr:   true,
		},
		{
			name:      "Case 2: No default for more than one build commands",
			groupType: buildGroup,
			wantErr:   true,
		},
		{
			name:      "Case 3: One command does not need default",
			groupType: testGroup,
			wantErr:   false,
		},
		{
			name:      "Case 4: One command can have default",
			groupType: debugGroup,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := validateCommandsForGroup(devObj.Data, tt.groupType)
			if !tt.wantErr && err != nil {
				t.Errorf("TestValidateCommandsForGroup unexpected error: %v", err)
			}
		})
	}

}

func TestValidateCommand(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"

	emptyString := ""

	tests := []struct {
		name    string
		exec    common.Exec
		wantErr bool
	}{
		{
			name: "Case: Valid Exec Command",
			exec: common.Exec{
				CommandLine: command,
				Component:   component,
				WorkingDir:  workDir,
				Group:       &versionsCommon.Group{Kind: runGroup},
			},
			wantErr: false,
		},
		{
			name: "Case: Invalid Exec Command with empty command",
			exec: common.Exec{
				CommandLine: emptyString,
				Component:   component,
				WorkingDir:  workDir,
				Group:       &versionsCommon.Group{Kind: runGroup},
			},
			wantErr: true,
		},
		{
			name: "Case: Invalid Exec Command with missing component",
			exec: common.Exec{
				CommandLine: command,
				WorkingDir:  workDir,
				Group:       &versionsCommon.Group{Kind: runGroup},
			},
			wantErr: true,
		},
		{
			name: "Case: valid Exec Command with Group nil",
			exec: common.Exec{
				CommandLine: command,
				Component:   component,
				WorkingDir:  workDir,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		devObj := devfileParser.DevfileObj{
			Data: testingutil.TestDevfileData{
				ExecCommands: []common.Exec{tt.exec},
				Components:   []common.DevfileComponent{testingutil.GetFakeComponent(component)},
			},
		}
		t.Run(tt.name, func(t *testing.T) {
			cmd := common.DevfileCommand{Exec: &tt.exec}
			err := validateCommand(devObj.Data, cmd)
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
	emptyString := ""

	var emptyCommand common.DevfileCommand

	tests := []struct {
		name         string
		commandName  string
		execCommands []common.Exec
		wantErr      bool
	}{
		{
			name:        "Case: Default Init Command",
			commandName: emptyString,
			execCommands: []versionsCommon.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &versionsCommon.Group{Kind: initGroup, IsDefault: true},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Init Command passed through odo flag",
			commandName: "flagcommand",
			execCommands: []versionsCommon.Exec{
				{
					Id:          "flagcommand",
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &versionsCommon.Group{Kind: initGroup},
				},
				{
					Id:          "init command",
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &versionsCommon.Group{Kind: initGroup},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Missing Init Command",
			commandName: "customcommand123",
			execCommands: []versionsCommon.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &versionsCommon.Group{Kind: initGroup},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ExecCommands: tt.execCommands,
					Components:   []common.DevfileComponent{testingutil.GetFakeComponent(component)},
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
	emptyString := ""

	var emptyCommand common.DevfileCommand

	tests := []struct {
		name         string
		commandName  string
		execCommands []common.Exec
		wantErr      bool
	}{
		{
			name:        "Case 1: Default Build Command",
			commandName: emptyString,
			execCommands: []common.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &versionsCommon.Group{Kind: buildGroup, IsDefault: true},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case 2: Build Command passed through the odo flag",
			commandName: "flagcommand",
			execCommands: []common.Exec{
				{
					Id:          "flagcommand",
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &versionsCommon.Group{Kind: buildGroup},
				},
				{
					Id:          "build command",
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &versionsCommon.Group{Kind: buildGroup},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case 3: Missing Build Command",
			commandName: "customcommand123",
			execCommands: []common.Exec{
				{
					Id:          "build command",
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &versionsCommon.Group{Kind: buildGroup},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ExecCommands: tt.execCommands,
					Components:   []common.DevfileComponent{testingutil.GetFakeComponent(component)},
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

func TestGetDebugCommand(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""

	var emptyCommand common.DevfileCommand

	tests := []struct {
		name         string
		commandName  string
		execCommands []common.Exec
		wantErr      bool
	}{
		{
			name:        "Case: Default Debug Command",
			commandName: emptyString,
			execCommands: []common.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group: &versionsCommon.Group{
						IsDefault: true,
						Kind:      versionsCommon.DebugCommandGroupType,
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Custom Debug Command",
			commandName: "customdebugcommand",
			execCommands: []common.Exec{
				{
					Id:          "customdebugcommand",
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group: &versionsCommon.Group{
						IsDefault: false,
						Kind:      versionsCommon.DebugCommandGroupType,
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Missing Debug Command",
			commandName: "customcommand123",
			execCommands: []common.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group: &versionsCommon.Group{
						IsDefault: true,
						Kind:      versionsCommon.BuildCommandGroupType,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					Components:   []common.DevfileComponent{testingutil.GetFakeComponent(component)},
					ExecCommands: tt.execCommands,
				},
			}

			command, err := GetDebugCommand(devObj.Data, tt.commandName)

			if tt.wantErr && err == nil {
				t.Errorf("Error was expected but got no error")
			} else if !tt.wantErr {
				if err != nil {
					t.Errorf("TestGetDebugCommand: unexpected error for command \"%v\" expected: %v actual: %v", tt.commandName, tt.wantErr, err)
				} else if reflect.DeepEqual(emptyCommand, command) {
					t.Errorf("TestGetDebugCommand: unexpected empty command returned for command: %v", tt.commandName)
				}
			}
		})
	}
}

func TestGetTestCommand(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""

	var emptyCommand common.DevfileCommand

	tests := []struct {
		name         string
		commandName  string
		execCommands []common.Exec
		wantErr      bool
	}{
		{
			name:        "Case: Default Test Command",
			commandName: emptyString,
			execCommands: []common.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group: &versionsCommon.Group{
						IsDefault: true,
						Kind:      versionsCommon.TestCommandGroupType,
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Custom Test Command",
			commandName: "customtestcommand",
			execCommands: []common.Exec{
				{
					Id:          "customtestcommand",
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group: &versionsCommon.Group{
						IsDefault: false,
						Kind:      versionsCommon.TestCommandGroupType,
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Missing Test Command",
			commandName: "customcommand123",
			execCommands: []common.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group: &versionsCommon.Group{
						IsDefault: true,
						Kind:      versionsCommon.BuildCommandGroupType,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					Components:   []common.DevfileComponent{testingutil.GetFakeComponent(component)},
					ExecCommands: tt.execCommands,
				},
			}

			command, err := GetTestCommand(devObj.Data, tt.commandName)

			if tt.wantErr && err == nil {
				t.Errorf("Error was expected but got no error")
			} else if !tt.wantErr {
				if err != nil {
					t.Errorf("TestGetTestCommand: unexpected error for command \"%v\" expected: %v actual: %v", tt.commandName, tt.wantErr, err)
				} else if reflect.DeepEqual(emptyCommand, command) {
					t.Errorf("TestGetTestCommand: unexpected empty command returned for command: %v", tt.commandName)
				}
			}
		})
	}
}

func TestGetRunCommand(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""

	var emptyCommand common.DevfileCommand

	tests := []struct {
		name         string
		commandName  string
		execCommands []common.Exec
		wantErr      bool
	}{
		{
			name:        "Case 1: Default Run Command",
			commandName: emptyString,
			execCommands: []common.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &versionsCommon.Group{Kind: runGroup, IsDefault: true},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case 2: Run Command passed through odo flag",
			commandName: "flagcommand",
			execCommands: []common.Exec{
				{
					Id:          "flagcommand",
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &versionsCommon.Group{Kind: runGroup},
				},
				{
					Id:          "run command",
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &versionsCommon.Group{Kind: runGroup},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case 3: Missing Run Command",
			commandName: "",
			execCommands: []common.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &versionsCommon.Group{Kind: initGroup},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ExecCommands: tt.execCommands,
					Components:   []common.DevfileComponent{testingutil.GetFakeComponent(component)},
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

func TestValidateAndGetDebugDevfileCommands(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""

	execCommands := []common.Exec{
		{
			CommandLine: command,
			Component:   component,
			WorkingDir:  workDir,
			Group: &common.Group{
				IsDefault: true,
				Kind:      common.DebugCommandGroupType,
			},
		},
		{
			Id:          "customdebugcommand",
			CommandLine: command,
			Component:   component,
			WorkingDir:  workDir,
			Group: &common.Group{
				IsDefault: false,
				Kind:      common.DebugCommandGroupType,
			},
		},
	}

	tests := []struct {
		name          string
		debugCommand  string
		componentType common.DevfileComponentType
		wantErr       bool
	}{
		{
			name:          "Case: Default Devfile Commands",
			debugCommand:  emptyString,
			componentType: common.ContainerComponentType,
			wantErr:       false,
		},
		{
			name:          "Case: provided debug Command",
			debugCommand:  "customdebugcommand",
			componentType: versionsCommon.ContainerComponentType,
			wantErr:       false,
		},
		{
			name:          "Case: invalid debug Command",
			debugCommand:  "invaliddebugcommand",
			componentType: versionsCommon.ContainerComponentType,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					Components:   []common.DevfileComponent{testingutil.GetFakeComponent(component)},
					ExecCommands: execCommands,
				},
			}

			debugCommand, err := ValidateAndGetDebugDevfileCommands(devObj.Data, tt.debugCommand)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Error was expected but got no error")
				} else {
					return
				}
			} else {
				if err != nil {
					t.Errorf("TestValidateAndGetDebugDevfileCommands: unexpected error %v", err)
				}
			}

			if !reflect.DeepEqual(nil, debugCommand) && debugCommand.Exec.Id != tt.debugCommand {
				t.Errorf("TestValidateAndGetDebugDevfileCommands name of debug command is wrong want: %v got: %v", tt.debugCommand, debugCommand.Exec.Id)
			}
		})
	}
}

func TestValidateAndGetPushDevfileCommands(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""

	execCommands := []common.Exec{
		{
			Id:          "run command",
			CommandLine: command,
			Component:   component,
			WorkingDir:  workDir,
			Group: &versionsCommon.Group{
				Kind:      runGroup,
				IsDefault: true,
			},
		},

		{
			Id:          "build command",
			CommandLine: command,
			Component:   component,
			WorkingDir:  workDir,
			Group:       &versionsCommon.Group{Kind: buildGroup},
		},

		{
			Id:          "init command",
			CommandLine: command,
			Component:   component,
			WorkingDir:  workDir,
			Group:       &versionsCommon.Group{Kind: initGroup},
		},
		{
			Id:          "customcommand",
			CommandLine: command,
			Component:   component,
			WorkingDir:  workDir,
			Group:       &versionsCommon.Group{Kind: runGroup},
		},
	}

	wrongCompTypeCmd := common.Exec{

		Id:          "run command",
		CommandLine: command,
		Component:   "",
		WorkingDir:  workDir,
		Group:       &versionsCommon.Group{Kind: runGroup},
	}

	tests := []struct {
		name                string
		initCommand         string
		buildCommand        string
		runCommand          string
		execCommands        []common.Exec
		numberOfCommands    int
		missingInitCommand  bool
		missingBuildCommand bool
		wantErr             bool
	}{
		{
			name:             "Case 1: Default Devfile Commands",
			initCommand:      emptyString,
			buildCommand:     emptyString,
			runCommand:       emptyString,
			execCommands:     execCommands,
			numberOfCommands: 3,
			wantErr:          false,
		},
		{
			name:             "Case 2: Default Init and Build Command, and Provided Run Command",
			initCommand:      emptyString,
			buildCommand:     emptyString,
			runCommand:       "customcommand",
			execCommands:     execCommands,
			numberOfCommands: 3,
			wantErr:          false,
		},
		{
			name:             "Case 3: Empty Component",
			initCommand:      emptyString,
			buildCommand:     "customcommand",
			runCommand:       "customcommand",
			execCommands:     append(execCommands, wrongCompTypeCmd),
			numberOfCommands: 0,
			wantErr:          true,
		},
		{
			name:             "Case 4: Provided Wrong Build Command and Provided Run Command",
			initCommand:      emptyString,
			buildCommand:     "customcommand123",
			runCommand:       "customcommand",
			execCommands:     execCommands,
			numberOfCommands: 1,
			wantErr:          true,
		},
		{
			name:             "Case 5: Provided Wrong Init Command and Provided Build and Run Command",
			initCommand:      "customcommand123",
			buildCommand:     emptyString,
			runCommand:       "customcommand",
			execCommands:     execCommands,
			numberOfCommands: 1,
			wantErr:          true,
		},
		{
			name:         "Case 6: Missing Init and Build Command, and Provided Run Command",
			initCommand:  emptyString,
			buildCommand: emptyString,
			runCommand:   "customcommand",
			execCommands: []common.Exec{
				{
					Id:          "customcommand",
					Group:       &common.Group{Kind: runGroup},
					Component:   component,
					CommandLine: command,
				},
			},
			numberOfCommands: 1,
			wantErr:          false,
		},
		{
			name:         "Case 7: Missing Init Command with provided Build and Run Command",
			initCommand:  emptyString,
			buildCommand: "build command",
			runCommand:   "run command",
			execCommands: []common.Exec{
				{
					Id:          "build command",
					Group:       &common.Group{Kind: buildGroup},
					Component:   component,
					CommandLine: command,
				},
				{
					Id:          "run command",
					Group:       &common.Group{Kind: runGroup},
					Component:   component,
					CommandLine: command,
				},
			},
			numberOfCommands:   2,
			missingInitCommand: true,
			wantErr:            false,
		},
		{
			name:         "Case 8: Missing Build Command with provided Init and Run Command",
			initCommand:  "init command",
			buildCommand: emptyString,
			runCommand:   "run command",
			execCommands: []common.Exec{
				{
					Id:          "init command",
					Group:       &common.Group{Kind: initGroup},
					Component:   component,
					CommandLine: command,
				},
				{
					Id:          "run command",
					Group:       &common.Group{Kind: runGroup},
					Component:   component,
					CommandLine: command,
				},
			},
			numberOfCommands: 2,
			wantErr:          false,
		},
		{
			name:             "Case 9: Optional Init Command with provided Build and Run Command",
			initCommand:      "init command",
			buildCommand:     "build command",
			runCommand:       "run command",
			execCommands:     execCommands,
			numberOfCommands: 3,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ExecCommands: tt.execCommands,
					Components:   []common.DevfileComponent{testingutil.GetFakeComponent(component)},
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

func TestValidateAndGetTestDevfileCommands(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""

	execCommands := []common.Exec{
		{
			CommandLine: command,
			Component:   component,
			WorkingDir:  workDir,
			Group: &common.Group{
				IsDefault: true,
				Kind:      common.TestCommandGroupType,
			},
		},
		{
			Id:          "customtestcommand",
			CommandLine: command,
			Component:   component,
			WorkingDir:  workDir,
			Group: &common.Group{
				IsDefault: false,
				Kind:      common.TestCommandGroupType,
			},
		},
	}

	tests := []struct {
		name          string
		testCommand   string
		componentType common.DevfileComponentType
		wantErr       bool
	}{
		{
			name:          "Case: Default Devfile Commands",
			testCommand:   emptyString,
			componentType: common.ContainerComponentType,
			wantErr:       false,
		},
		{
			name:          "Case: provided test Command",
			testCommand:   "customtestcommand",
			componentType: versionsCommon.ContainerComponentType,
			wantErr:       false,
		},
		{
			name:          "Case: invalid test Command",
			testCommand:   "invalidtestcommand",
			componentType: versionsCommon.ContainerComponentType,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					Components:   []common.DevfileComponent{testingutil.GetFakeComponent(component)},
					ExecCommands: execCommands,
				},
			}

			testCommand, err := ValidateAndGetTestDevfileCommands(devObj.Data, tt.testCommand)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Error was expected but got no error")
				} else {
					return
				}
			} else {
				if err != nil {
					t.Errorf("TestValidateAndGetTestDevfileCommands: unexpected error %v", err)
				}
			}

			if !reflect.DeepEqual(nil, testCommand) && testCommand.Exec.Id != tt.testCommand {
				t.Errorf("TestValidateAndGetTestDevfileCommands name of test command is wrong want: %v got: %v", tt.testCommand, testCommand.Exec.Id)
			}
		})
	}
}

func getExecCommand(id string, group common.DevfileCommandGroupType) versionsCommon.Exec {

	commands := [...]string{"ls -la", "pwd"}
	components := [...]string{"alias1", "alias2"}
	workDir := [...]string{"/", "/root"}

	return versionsCommon.Exec{
		Id:          id,
		CommandLine: commands[0],
		Component:   components[0],
		WorkingDir:  workDir[0],
		Group:       &common.Group{Kind: group},
	}

}
