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

	emptyString := ""
	buildGroup := common.BuildCommandGroupType
	runGroup := common.RunCommandGroupType
	initGroup := common.InitCommandGroupType

	tests := []struct {
		name          string
		requestedType []common.DevfileCommandGroupType
		execCommands  []common.Exec
		groupType     []common.DevfileCommandGroupType
		wantErr       bool
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
			name: "Case 8: Invalid command referencing an absent component",
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
				_, err := getCommand(devObj.Data, "", gtype)
				if !tt.wantErr == (err != nil) {
					t.Errorf("TestGetCommand unexpected error for command: %v wantErr: %v err: %v", gtype, tt.wantErr, err)
					return
				} else if tt.wantErr {
					return
				}

				/*if command.Exec.Group.Kind != gtype {
					t.Errorf("TestGetCommand error: command names do not match expected: %v actual: %v", gtype, command.Exec.Id)
				}*/

			}
		})
	}

}

func TestValidateAction(t *testing.T) {

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
			name: "Case: Valid CommandLine Action",
			exec: common.Exec{
				CommandLine: command,
				Component:   component,
				WorkingDir:  workDir,
			},
			wantErr: false,
		},
		{
			name: "Case: Invalid CommandLine Action with empty command",
			exec: common.Exec{
				CommandLine: emptyString,
				Component:   component,
				WorkingDir:  workDir,
			},
			wantErr: true,
		},
		{
			name: "Case: Invalid Command Action with missing component",
			exec: common.Exec{
				CommandLine: command,
				WorkingDir:  workDir,
			},
			wantErr: true,
		},
		{
			name: "Case: Invalid Command Action with wrong type",
			exec: common.Exec{
				CommandLine: command,
				Component:   component,
				WorkingDir:  workDir,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		devObj := devfileParser.DevfileObj{
			Data: testingutil.TestDevfileData{
				ExecCommands: []common.Exec{
					{
						CommandLine: command,
						Component:   component,
					},
				},
				Components: []common.DevfileComponent{testingutil.GetFakeComponent(component)},
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
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Custom Init Command",
			commandName: "customcommand",
			execCommands: []versionsCommon.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
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
			name:        "Case: Default Build Command",
			commandName: emptyString,
			execCommands: []common.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Custom Build Command",
			commandName: "customcommand",
			execCommands: []common.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Missing Build Command",
			commandName: "customcommand123",
			execCommands: []common.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
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
			name:        "Case: Default Run Command",
			commandName: emptyString,
			execCommands: []common.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Custom Run Command",
			commandName: "customcommand",
			execCommands: []common.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Missing Run Command",
			commandName: "customcommand123",
			execCommands: []common.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
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

func TestValidateAndGetPushDevfileCommands(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""

	actions := []common.Exec{
		{
			CommandLine: command,
			Component:   component,
			WorkingDir:  workDir,
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
			componentType:    common.ContainerComponentType,
			wantErr:          false,
		},
		{
			name:             "Case: Default Init and Build Command, and Provided Run Command",
			initCommand:      emptyString,
			buildCommand:     emptyString,
			runCommand:       "customcommand",
			numberOfCommands: 3,
			componentType:    common.ContainerComponentType,
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
			componentType:    common.ContainerComponentType,
			wantErr:          true,
		},
		{
			name:             "Case: Provided Wrong Init Command and Provided Build and Run Command",
			initCommand:      "customcommand123",
			buildCommand:     emptyString,
			runCommand:       "customcommand",
			numberOfCommands: 1,
			componentType:    versionsCommon.ContainerComponentType,
			wantErr:          true,
		},
		{
			name:                "Case: Missing Init and Build Command, and Provided Run Command",
			initCommand:         emptyString,
			buildCommand:        emptyString,
			runCommand:          "customcommand",
			numberOfCommands:    1,
			componentType:       common.ContainerComponentType,
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
			componentType:      versionsCommon.ContainerComponentType,
			missingInitCommand: true,
			wantErr:            false,
		},
		{
			name:                "Case: Missing Build Command with provided Init and Run Command",
			initCommand:         "customcommand",
			buildCommand:        emptyString,
			runCommand:          "customcommand",
			numberOfCommands:    2,
			componentType:       versionsCommon.ContainerComponentType,
			missingBuildCommand: true,
			wantErr:             false,
		},
		{
			name:             "Case: Optional Init Command with provided Build and Run Command",
			initCommand:      "customcommand",
			buildCommand:     "customcommand",
			runCommand:       "customcommand",
			numberOfCommands: 3,
			componentType:    versionsCommon.ContainerComponentType,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ExecCommands:        actions,
					Components:          []common.DevfileComponent{testingutil.GetFakeComponent(component)},
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
