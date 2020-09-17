package common

import (
	"fmt"
	"reflect"
	"testing"

	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/testingutil"
	"github.com/openshift/odo/pkg/util"
)

var buildGroup = common.BuildCommandGroupType
var runGroup = common.RunCommandGroupType
var testGroup = common.TestCommandGroupType
var debugGroup = common.DebugCommandGroupType

func TestGetCommand(t *testing.T) {

	commands := [...]string{"ls -la", "pwd"}
	components := [...]string{"alias1", "alias2"}

	tests := []struct {
		name           string
		requestedType  []common.DevfileCommandGroupType
		execCommands   []common.DevfileCommand
		compCommands   []common.DevfileCommand
		reqCommandName string
		retCommandName string
		wantErr        bool
	}{
		{
			name: "Case 1: Valid devfile",
			execCommands: []common.DevfileCommand{
				getExecCommand("build", buildGroup),
				getExecCommand("run", runGroup),
			},
			requestedType: []common.DevfileCommandGroupType{buildGroup, runGroup},
			wantErr:       false,
		},
		{
			name: "Case 2: Valid devfile with devrun and devbuild",
			execCommands: []common.DevfileCommand{
				getExecCommand("build", buildGroup),
				getExecCommand("run", runGroup),
			},
			requestedType: []common.DevfileCommandGroupType{buildGroup, runGroup},
			wantErr:       false,
		},
		{
			name: "Case 3: Valid devfile with empty workdir",
			execCommands: []common.DevfileCommand{
				{
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: runGroup},
					},
				},
			},
			requestedType: []common.DevfileCommandGroupType{runGroup},
			wantErr:       false,
		},
		{
			name: "Case 4: Mismatched command type",
			execCommands: []common.DevfileCommand{
				{
					Id: "build command",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: runGroup},
					},
				},
			},
			reqCommandName: "build command",
			requestedType:  []common.DevfileCommandGroupType{buildGroup},
			wantErr:        true,
		},
		{
			name: "Case 5: Default command is returned",
			execCommands: []common.DevfileCommand{
				{
					Id: "defaultRunCommand",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: runGroup, IsDefault: true},
					},
				},
				{
					Id: "runCommand",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: runGroup},
					},
				},
			},
			retCommandName: "defaultruncommand",
			requestedType:  []common.DevfileCommandGroupType{runGroup},
			wantErr:        false,
		},
		{
			name: "Case 6: Composite command is returned",
			execCommands: []common.DevfileCommand{
				{
					Id: "build",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: buildGroup, IsDefault: false},
					},
				},
				{
					Id: "run",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: runGroup},
					},
				},
			},
			compCommands: []common.DevfileCommand{
				{
					Id: "myComposite",
					Composite: &common.Composite{
						Commands: []string{"build", "run"},
						Group:    &common.Group{Kind: buildGroup, IsDefault: true},
					},
				},
			},
			retCommandName: "mycomposite",
			requestedType:  []common.DevfileCommandGroupType{buildGroup},
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := []common.DevfileComponent{testingutil.GetFakeContainerComponent(tt.execCommands[0].Exec.Component)}
			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Commands:   append(tt.execCommands, tt.compCommands...),
					Components: components,
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

				if len(tt.retCommandName) > 0 && cmd.GetID() != tt.retCommandName {
					t.Errorf("TestGetCommand error: command names do not match expected: %v actual: %v", tt.retCommandName, cmd.Id)
				}
			}
		})
	}

}

func TestGetCommandFromDevfile(t *testing.T) {

	commands := [...]string{"ls -la", "pwd"}
	components := [...]string{"alias1", "alias2"}

	tests := []struct {
		name           string
		requestedType  []common.DevfileCommandGroupType
		execCommands   []common.DevfileCommand
		compCommands   []common.DevfileCommand
		retCommandName string
		wantErr        bool
	}{
		{
			name: "Case 1: Valid devfile",
			execCommands: []common.DevfileCommand{
				getExecCommand("", buildGroup),
				getExecCommand("", runGroup),
			},
			requestedType: []common.DevfileCommandGroupType{buildGroup, runGroup},
			wantErr:       false,
		},
		{
			name: "Case 2: Valid devfile with devrun and devbuild",
			execCommands: []common.DevfileCommand{
				getExecCommand("", buildGroup),
				getExecCommand("", runGroup),
			},
			requestedType: []common.DevfileCommandGroupType{buildGroup, runGroup},
			wantErr:       false,
		},
		{
			name: "Case 3: Valid devfile with empty workdir",
			execCommands: []common.DevfileCommand{
				{
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: runGroup},
					},
				},
			},
			requestedType: []common.DevfileCommandGroupType{runGroup},
			wantErr:       false,
		},
		{
			name: "Case 4: Default command is returned",
			execCommands: []common.DevfileCommand{
				{
					Id: "defaultruncommand",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: runGroup, IsDefault: true},
					},
				},
				{
					Id: "runcommand",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: runGroup},
					},
				},
			},
			retCommandName: "defaultruncommand",
			requestedType:  []common.DevfileCommandGroupType{runGroup},
			wantErr:        false,
		},
		{
			name: "Case 5: Valid devfile, has composite command",
			execCommands: []common.DevfileCommand{
				{
					Id: "build1",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: buildGroup},
					},
				},
				{
					Id: "build2",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: buildGroup},
					},
				},
				{
					Id: "run",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: runGroup},
					},
				},
			},
			compCommands: []common.DevfileCommand{
				{
					Id: "mycomp",
					Composite: &common.Composite{
						Commands: []string{"build1", "run"},
						Group:    &common.Group{Kind: buildGroup, IsDefault: true},
					},
				},
			},
			retCommandName: "mycomp",
			requestedType:  []common.DevfileCommandGroupType{buildGroup},
			wantErr:        false,
		},
		{
			name: "Case 6: Default composite command",
			execCommands: []common.DevfileCommand{
				{
					Id: "build",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: buildGroup, IsDefault: false},
					},
				},
				{
					Id: "run",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: runGroup},
					},
				},
			},
			compCommands: []common.DevfileCommand{
				{
					Id: "mycomp",
					Composite: &common.Composite{
						Commands: []string{"build", "run"},
						Group:    &common.Group{Kind: buildGroup, IsDefault: true},
					},
				},
				{
					Id: "mycomp2",
					Composite: &common.Composite{
						Commands: []string{"build", "run"},
						Group:    &common.Group{Kind: buildGroup, IsDefault: false},
					},
				},
			},
			retCommandName: "mycomp",
			requestedType:  []common.DevfileCommandGroupType{buildGroup},
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := []common.DevfileComponent{testingutil.GetFakeContainerComponent(tt.execCommands[0].Exec.Component)}
			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Commands:   append(tt.execCommands, tt.compCommands...),
					Components: components,
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

				if len(tt.retCommandName) > 0 && cmd.GetID() != tt.retCommandName {
					t.Errorf("TestGetCommandFromDevfile error: command names do not match expected: %v actual: %v", tt.retCommandName, cmd.GetID())
				}
			}
		})
	}

}

func TestGetCommandFromFlag(t *testing.T) {

	commands := [...]string{"ls -la", "pwd"}
	components := [...]string{"alias1", "alias2"}
	invalidComponent := "garbagealias"

	tests := []struct {
		name           string
		requestedType  common.DevfileCommandGroupType
		execCommands   []common.DevfileCommand
		compCommands   []common.DevfileCommand
		reqCommandName string
		retCommandName string
		wantErr        bool
	}{
		{
			name: "Case 1: Valid devfile",
			execCommands: []common.DevfileCommand{
				getExecCommand("a", buildGroup),
				getExecCommand("b", runGroup),
			},
			reqCommandName: "b",
			retCommandName: "b",
			requestedType:  runGroup,
			wantErr:        false,
		},
		{
			name: "Case 2: Valid devfile with empty workdir",
			execCommands: []common.DevfileCommand{
				{
					Id: "build command",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: runGroup},
					},
				},
			},
			reqCommandName: "build command",
			retCommandName: "build command",
			requestedType:  runGroup,
			wantErr:        false,
		},
		{
			name: "Case 3: Invalid command",
			execCommands: []common.DevfileCommand{
				{
					Id: "build command",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   invalidComponent,
						Group:       &common.Group{Kind: runGroup},
					},
				},
			},
			reqCommandName: "build command wrong",
			requestedType:  runGroup,
			wantErr:        true,
		},
		{
			name: "Case 4: Mismatched command type",
			execCommands: []common.DevfileCommand{
				{
					Id: "build command",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: runGroup},
					},
				},
			},
			reqCommandName: "build command",
			requestedType:  buildGroup,
			wantErr:        true,
		},
		{
			name: "Case 5: Multiple default commands but should be with the flag",
			execCommands: []common.DevfileCommand{
				{
					Id: "defaultruncommand",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: runGroup, IsDefault: true},
					},
				},
				{
					Id: "runcommand",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: runGroup, IsDefault: true},
					},
				},
			},
			reqCommandName: "defaultruncommand",
			retCommandName: "defaultruncommand",
			requestedType:  runGroup,
			wantErr:        false,
		},
		{
			name: "Case 6: No default command but should be with the flag",
			execCommands: []common.DevfileCommand{
				{
					Id: "defaultruncommand",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: runGroup},
					},
				},
				{
					Id: "runcommand",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: runGroup},
					},
				},
			},
			reqCommandName: "defaultruncommand",
			retCommandName: "defaultruncommand",
			requestedType:  runGroup,
			wantErr:        false,
		},
		{
			name: "Case 7: No Command Group",
			execCommands: []common.DevfileCommand{
				{
					Id: "defaultruncommand",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
					},
				},
			},
			reqCommandName: "defaultruncommand",
			retCommandName: "defaultruncommand",
			requestedType:  runGroup,
			wantErr:        false,
		},
		{
			name: "Case 8: Valid devfile with composite commands",
			execCommands: []common.DevfileCommand{
				{
					Id: "build",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: buildGroup, IsDefault: false},
					},
				},
				{
					Id: "run",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: runGroup},
					},
				},
			},
			compCommands: []common.DevfileCommand{
				{
					Id: "mycomp",
					Composite: &common.Composite{
						Commands: []string{"build", "run"},
						Group:    &common.Group{Kind: buildGroup, IsDefault: true},
					},
				},
				{
					Id: "mycomp2",
					Composite: &common.Composite{
						Commands: []string{"build", "run"},
						Group:    &common.Group{Kind: buildGroup, IsDefault: false},
					},
				},
			},
			reqCommandName: "mycomp",
			retCommandName: "mycomp",
			requestedType:  buildGroup,
			wantErr:        false,
		},
		{
			name: "Case 9: Valid devfile with invalid composite commands",
			execCommands: []common.DevfileCommand{
				{
					Id: "build",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: buildGroup, IsDefault: false},
					},
				},
				{
					Id: "run",
					Exec: &common.Exec{
						CommandLine: commands[0],
						Component:   components[0],
						Group:       &common.Group{Kind: runGroup},
					},
				},
			},
			compCommands: []common.DevfileCommand{
				{
					Id: "myComp",
					Composite: &common.Composite{
						Commands: []string{"fake"},
						Group:    &common.Group{Kind: buildGroup, IsDefault: true},
					},
				},
			},
			reqCommandName: "myComp",
			retCommandName: "myComp",
			requestedType:  buildGroup,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := []common.DevfileComponent{testingutil.GetFakeContainerComponent(tt.execCommands[0].Exec.Component)}
			if tt.execCommands[0].Exec.Component == invalidComponent {
				components = []common.DevfileComponent{testingutil.GetFakeContainerComponent("randomComponent")}
			}
			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Commands:   append(tt.compCommands, tt.execCommands...),
					Components: components,
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
				if cmd.Id != tt.retCommandName {
					t.Errorf("TestGetCommand error: command names do not match expected: %v actual: %v", tt.retCommandName, cmd.Id)
				}
			}
		})
	}

}

func TestValidateCommandsForGroup(t *testing.T) {

	componentName := "alias1"
	command := "ls -la"
	workDir := "/"

	tests := []struct {
		name         string
		groupType    common.DevfileCommandGroupType
		execCommands []common.DevfileCommand
		compCommands []common.DevfileCommand
		wantErr      bool
	}{
		{
			name: "Case 1: Two default run commands",
			execCommands: []common.DevfileCommand{
				{
					Id: "run command",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   componentName,
						WorkingDir:  workDir,
						Group: &common.Group{
							Kind:      runGroup,
							IsDefault: true,
						},
					},
				},
				{
					Id: "customcommand",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   componentName,
						WorkingDir:  workDir,
						Group: &common.Group{
							Kind:      runGroup,
							IsDefault: true,
						},
					},
				},
			},
			groupType: runGroup,
			wantErr:   true,
		},
		{
			name: "Case 2: No default for more than one build commands",
			execCommands: []common.DevfileCommand{
				{
					Id: "build command",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   componentName,
						WorkingDir:  workDir,
						Group:       &common.Group{Kind: buildGroup},
					},
				},
				{
					Id: "build command 2",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   componentName,
						WorkingDir:  workDir,
						Group:       &common.Group{Kind: buildGroup},
					},
				},
			},
			groupType: buildGroup,
			wantErr:   true,
		},
		{
			name: "Case 3: One command does not need default",
			execCommands: []common.DevfileCommand{
				{
					Id: "test command",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   componentName,
						WorkingDir:  workDir,
						Group:       &common.Group{Kind: testGroup},
					},
				},
			},
			groupType: testGroup,
			wantErr:   false,
		},
		{
			name: "Case 4: One command can have default",
			execCommands: []common.DevfileCommand{
				{
					Id: "debug command",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   componentName,
						WorkingDir:  workDir,
						Group: &common.Group{
							Kind:      debugGroup,
							IsDefault: true,
						},
					},
				},
			},
			groupType: debugGroup,
			wantErr:   false,
		},
		{
			name: "Case 5: Composite commands in group",
			execCommands: []common.DevfileCommand{
				{
					Id: "build command",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   componentName,
						WorkingDir:  workDir,
						Group:       &common.Group{Kind: buildGroup},
					},
				},
				{
					Id: "build command 2",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   componentName,
						WorkingDir:  workDir,
						Group:       &common.Group{Kind: buildGroup},
					},
				},
			},
			compCommands: []common.DevfileCommand{
				{
					Id: "composite1",
					Composite: &common.Composite{
						Commands: []string{"build command", "build command 2"},
						Group:    &common.Group{Kind: buildGroup, IsDefault: true},
					},
				},
			},
			groupType: buildGroup,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: []common.DevfileComponent{
						testingutil.GetFakeContainerComponent("alias1"),
					},
					Commands: append(tt.compCommands, tt.execCommands...),
				},
			}

			err := validateCommandsForGroup(devObj.Data, tt.groupType)
			if !tt.wantErr && err != nil {
				t.Errorf("TestValidateCommandsForGroup unexpected error: %v", err)
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
		execCommands []common.DevfileCommand
		wantErr      bool
	}{
		{
			name:        "Case 1: Default Build Command",
			commandName: emptyString,
			execCommands: []common.DevfileCommand{
				{
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
						Group:       &common.Group{Kind: buildGroup, IsDefault: true},
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case 2: Build Command passed through the odo flag",
			commandName: "flagcommand",
			execCommands: []common.DevfileCommand{
				{
					Id: "flagcommand",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
						Group:       &common.Group{Kind: buildGroup},
					},
				},
				{
					Id: "build command",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
						Group:       &common.Group{Kind: buildGroup},
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case 3: Missing Build Command",
			commandName: "customcommand123",
			execCommands: []common.DevfileCommand{
				{
					Id: "build command",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
						Group:       &common.Group{Kind: buildGroup},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Commands:   tt.execCommands,
					Components: []common.DevfileComponent{testingutil.GetFakeContainerComponent(component)},
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
		execCommands []common.DevfileCommand
		wantErr      bool
	}{
		{
			name:        "Case: Default Debug Command",
			commandName: emptyString,
			execCommands: []common.DevfileCommand{
				{
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
						Group: &common.Group{
							IsDefault: true,
							Kind:      common.DebugCommandGroupType,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Custom Debug Command",
			commandName: "customdebugcommand",
			execCommands: []common.DevfileCommand{
				{
					Id: "customdebugcommand",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
						Group: &common.Group{
							IsDefault: false,
							Kind:      common.DebugCommandGroupType,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Missing Debug Command",
			commandName: "customcommand123",
			execCommands: []common.DevfileCommand{
				{
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
						Group: &common.Group{
							IsDefault: true,
							Kind:      common.BuildCommandGroupType,
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: []common.DevfileComponent{testingutil.GetFakeContainerComponent(component)},
					Commands:   tt.execCommands,
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
		execCommands []common.DevfileCommand
		wantErr      bool
	}{
		{
			name:        "Case: Default Test Command",
			commandName: emptyString,
			execCommands: []common.DevfileCommand{
				{
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
						Group: &common.Group{
							IsDefault: true,
							Kind:      common.TestCommandGroupType,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Custom Test Command",
			commandName: "customtestcommand",
			execCommands: []common.DevfileCommand{
				{
					Id: "customtestcommand",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
						Group: &common.Group{
							IsDefault: false,
							Kind:      common.TestCommandGroupType,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Missing Test Command",
			commandName: "customcommand123",
			execCommands: []common.DevfileCommand{
				{
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
						Group: &common.Group{
							IsDefault: true,
							Kind:      common.BuildCommandGroupType,
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: []common.DevfileComponent{testingutil.GetFakeContainerComponent(component)},
					Commands:   tt.execCommands,
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
		execCommands []common.DevfileCommand
		wantErr      bool
	}{
		{
			name:        "Case 1: Default Run Command",
			commandName: emptyString,
			execCommands: []common.DevfileCommand{
				{
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
						Group:       &common.Group{Kind: runGroup, IsDefault: true},
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case 2: Run Command passed through odo flag",
			commandName: "flagcommand",
			execCommands: []common.DevfileCommand{
				{
					Id: "flagcommand",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
						Group:       &common.Group{Kind: runGroup},
					},
				},
				{
					Id: "run command",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
						Group:       &common.Group{Kind: runGroup},
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case 3: Missing Run Command",
			commandName: "",
			execCommands: []common.DevfileCommand{
				{
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
						Group:       &common.Group{Kind: buildGroup},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Commands:   tt.execCommands,
					Components: []common.DevfileComponent{testingutil.GetFakeContainerComponent(component)},
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

	execCommands := []common.DevfileCommand{
		{
			Exec: &common.Exec{
				CommandLine: command,
				Component:   component,
				WorkingDir:  workDir,
				Group: &common.Group{
					IsDefault: true,
					Kind:      common.DebugCommandGroupType,
				},
			},
		},
		{
			Id: "customdebugcommand",
			Exec: &common.Exec{
				CommandLine: command,
				Component:   component,
				WorkingDir:  workDir,
				Group: &common.Group{
					IsDefault: false,
					Kind:      common.DebugCommandGroupType,
				},
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
			componentType: common.ContainerComponentType,
			wantErr:       false,
		},
		{
			name:          "Case: invalid debug Command",
			debugCommand:  "invaliddebugcommand",
			componentType: common.ContainerComponentType,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: []common.DevfileComponent{testingutil.GetFakeContainerComponent(component)},
					Commands:   execCommands,
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

			if !reflect.DeepEqual(nil, debugCommand) && debugCommand.Id != tt.debugCommand {
				t.Errorf("TestValidateAndGetDebugDevfileCommands name of debug command is wrong want: %v got: %v", tt.debugCommand, debugCommand.Id)
			}
		})
	}
}

func TestValidateAndGetPushDevfileCommands(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""

	execCommands := []common.DevfileCommand{
		{
			Id: "run command",
			Exec: &common.Exec{
				CommandLine: command,
				Component:   component,
				WorkingDir:  workDir,
				Group: &common.Group{
					Kind:      runGroup,
					IsDefault: true,
				},
			},
		},

		{
			Id: "build command",
			Exec: &common.Exec{
				CommandLine: command,
				Component:   component,
				WorkingDir:  workDir,
				Group:       &common.Group{Kind: buildGroup},
			},
		},

		{
			Id: "customcommand",
			Exec: &common.Exec{
				CommandLine: command,
				Component:   component,
				WorkingDir:  workDir,
				Group:       &common.Group{Kind: runGroup},
			},
		},
	}

	wrongCompTypeCmd := common.DevfileCommand{

		Id: "wrong",
		Exec: &common.Exec{
			CommandLine: command,
			Component:   "",
			WorkingDir:  workDir,
			Group:       &common.Group{Kind: runGroup},
		},
	}

	tests := []struct {
		name                string
		buildCommand        string
		runCommand          string
		execCommands        []common.DevfileCommand
		numberOfCommands    int
		missingBuildCommand bool
		wantErr             bool
	}{
		{
			name:             "Case 1: Default Devfile Commands",
			buildCommand:     emptyString,
			runCommand:       emptyString,
			execCommands:     execCommands,
			numberOfCommands: 2,
			wantErr:          false,
		},
		{
			name:             "Case 2: Default Build Command, and Provided Run Command",
			buildCommand:     emptyString,
			runCommand:       "customcommand",
			execCommands:     execCommands,
			numberOfCommands: 2,
			wantErr:          false,
		},
		{
			name:             "Case 3: Empty Component",
			buildCommand:     "customcommand",
			runCommand:       "customcommand",
			execCommands:     append(execCommands, wrongCompTypeCmd),
			numberOfCommands: 0,
			wantErr:          true,
		},
		{
			name:             "Case 4: Provided Wrong Build Command and Provided Run Command",
			buildCommand:     "customcommand123",
			runCommand:       "customcommand",
			execCommands:     execCommands,
			numberOfCommands: 1,
			wantErr:          true,
		},
		{
			name:         "Case 5: Missing Build Command, and Provided Run Command",
			buildCommand: emptyString,
			runCommand:   "customcommand",
			execCommands: []common.DevfileCommand{
				{
					Id: "customcommand",
					Exec: &common.Exec{
						Group:       &common.Group{Kind: runGroup},
						Component:   component,
						CommandLine: command,
					},
				},
			},
			numberOfCommands: 1,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Commands:   tt.execCommands,
					Components: []common.DevfileComponent{testingutil.GetFakeContainerComponent(component)},
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

func TestValidateAndGetTestDevfileCommands(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""

	execCommands := []common.DevfileCommand{
		{
			Exec: &common.Exec{
				CommandLine: command,
				Component:   component,
				WorkingDir:  workDir,
				Group: &common.Group{
					IsDefault: true,
					Kind:      common.TestCommandGroupType,
				},
			},
		},
		{
			Id: "customtestcommand",
			Exec: &common.Exec{
				CommandLine: command,
				Component:   component,
				WorkingDir:  workDir,
				Group: &common.Group{
					IsDefault: false,
					Kind:      common.TestCommandGroupType,
				},
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
			componentType: common.ContainerComponentType,
			wantErr:       false,
		},
		{
			name:          "Case: invalid test Command",
			testCommand:   "invalidtestcommand",
			componentType: common.ContainerComponentType,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: []common.DevfileComponent{testingutil.GetFakeContainerComponent(component)},
					Commands:   execCommands,
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

			if !reflect.DeepEqual(nil, testCommand) && testCommand.Id != tt.testCommand {
				t.Errorf("TestValidateAndGetTestDevfileCommands name of test command is wrong want: %v got: %v", tt.testCommand, testCommand.Id)
			}
		})
	}
}

func getExecCommand(id string, group common.DevfileCommandGroupType) common.DevfileCommand {
	if len(id) == 0 {
		id = fmt.Sprintf("%s-%s", "cmd", util.GenerateRandomString(10))
	}
	commands := [...]string{"ls -la", "pwd"}
	components := [...]string{"alias1", "alias2"}
	workDir := [...]string{"/", "/root"}

	return common.DevfileCommand{
		Id: id,
		Exec: &common.Exec{
			CommandLine: commands[0],
			Component:   components[0],
			WorkingDir:  workDir[0],
			Group:       &common.Group{Kind: group},
		},
	}

}
