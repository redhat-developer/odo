package validate

import (
	"testing"

	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/testingutil"
)

var buildGroup = common.BuildCommandGroupType
var runGroup = common.RunCommandGroupType

func TestValidateCommands(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	invalidComponent := "garbagealias"
	emptyString := ""

	tests := []struct {
		name    string
		exec    []common.DevfileCommand
		comp    []common.DevfileCommand
		wantErr bool
	}{
		{
			name: "Case 1: Valid Exec Command",
			exec: []common.DevfileCommand{
				{
					Id: "somecommand",
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
			name: "Case 2: Invalid Exec Command with empty command",
			exec: []common.DevfileCommand{
				{
					Id: "somecommand",
					Exec: &common.Exec{
						CommandLine: emptyString,
						Component:   component,
						WorkingDir:  workDir,
						Group:       &common.Group{Kind: runGroup},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 3: Invalid Exec Command with missing component",
			exec: []common.DevfileCommand{
				{
					Id: "somecommand",
					Exec: &common.Exec{
						CommandLine: command,
						WorkingDir:  workDir,
						Group:       &common.Group{Kind: runGroup},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 4: Valid Exec Command with invalid component",
			exec: []common.DevfileCommand{
				{
					Id: "somecommand",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   invalidComponent,
						WorkingDir:  workDir,
						Group:       &common.Group{Kind: runGroup},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 5: valid Exec Command with Group nil",
			exec: []common.DevfileCommand{
				{
					Id: "somecommand",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 6: valid Composite Command",
			exec: []common.DevfileCommand{
				{
					Id: "somecommand1",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
					},
				},
				{
					Id: "somecommand2",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
					},
				},
			},
			comp: []common.DevfileCommand{
				{
					Id: "composite1",
					Composite: &common.Composite{
						Commands: []string{"somecommand1", "somecommand2"},
						Group:    &common.Group{Kind: buildGroup, IsDefault: true},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 7: Invalid Composite Command",
			exec: []common.DevfileCommand{
				{
					Id: "somecommand1",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
					},
				},
				{
					Id: "somecommand2",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
					},
				},
			},
			comp: []common.DevfileCommand{
				{
					Id: "composite1",
					Composite: &common.Composite{
						Commands: []string{"fakecommand"},
						Group:    &common.Group{Kind: buildGroup, IsDefault: true},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 9: Duplicate commands",
			exec: []common.DevfileCommand{
				{
					Id: "somecommand1",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
					},
				},
				{
					Id: "somecommand1",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
					},
				},
				{
					Id: "somecommand2",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 9: Duplicate commands, different types",
			exec: []common.DevfileCommand{
				{
					Id: "somecommand1",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
					},
				},
				{
					Id: "somecommand2",
					Exec: &common.Exec{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
					},
				},
			},
			comp: []common.DevfileCommand{
				{
					Id: "somecommand1",
					Composite: &common.Composite{
						Commands: []string{"fakecommand"},
						Group:    &common.Group{Kind: buildGroup, IsDefault: true},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		devfileData := testingutil.TestDevfileData{
			Commands:   append(tt.comp, tt.exec...),
			Components: []common.DevfileComponent{testingutil.GetFakeContainerComponent(component)},
		}
		devObj := devfileParser.DevfileObj{
			Data: &devfileData,
		}

		commands := devObj.Data.GetCommands()
		components := devObj.Data.GetComponents()

		t.Run(tt.name, func(t *testing.T) {
			err := validateCommands(devfileData.Commands, commands, components)
			if !tt.wantErr == (err != nil) {
				t.Errorf("TestValidateAction unexpected error: %v", err)
				return
			}
		})
	}

}

func TestValidateCompositeCommand(t *testing.T) {

	command := []string{"ls -la", "ps", "ls /"}
	component := "alias1"
	workDir := []string{"/", "/dev", "/etc"}
	id := []string{"command1", "command2", "command3", "command4", "command5"}

	tests := []struct {
		name              string
		compositeCommands []common.DevfileCommand
		execCommands      []common.DevfileCommand
		wantErr           bool
	}{
		{
			name: "Case 1: Valid Composite Command",
			compositeCommands: []common.DevfileCommand{
				{
					Id: id[3],
					Composite: &common.Composite{
						Commands: []string{id[0], id[1], id[2]},
						Group:    &common.Group{Kind: buildGroup},
					},
				},
			},
			execCommands: []common.DevfileCommand{
				{
					Id: id[0],
					Exec: &common.Exec{
						CommandLine: command[0],
						Component:   component,
						Group:       &common.Group{Kind: runGroup},
						WorkingDir:  workDir[0],
					},
				},
				{
					Id: id[1],
					Exec: &common.Exec{
						CommandLine: command[1],
						Component:   component,
						Group:       &common.Group{Kind: buildGroup},
						WorkingDir:  workDir[1],
					},
				},
				{
					Id: id[2],
					Exec: &common.Exec{
						CommandLine: command[2],
						Component:   component,
						Group:       &common.Group{Kind: runGroup},
						WorkingDir:  workDir[2],
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 2: Invalid composite command, references non-existent command",
			compositeCommands: []common.DevfileCommand{
				{
					Id: id[3],
					Composite: &common.Composite{
						Commands: []string{id[0], "fakecommand", id[2]},
						Group:    &common.Group{Kind: buildGroup},
					},
				},
			},
			execCommands: []common.DevfileCommand{
				{
					Id: id[0],
					Exec: &common.Exec{
						CommandLine: command[0],
						Component:   component,
						Group:       &common.Group{Kind: runGroup},
						WorkingDir:  workDir[0],
					},
				},
				{
					Id: id[1],
					Exec: &common.Exec{
						CommandLine: command[1],
						Component:   component,
						Group:       &common.Group{Kind: buildGroup},
						WorkingDir:  workDir[1],
					},
				},
				{
					Id: id[2],
					Exec: &common.Exec{
						CommandLine: command[2],
						Component:   component,
						Group:       &common.Group{Kind: runGroup},
						WorkingDir:  workDir[2],
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 3: Invalid composite command, references itself",
			compositeCommands: []common.DevfileCommand{
				{
					Id: id[3],
					Composite: &common.Composite{
						Commands: []string{id[0], id[3], id[2]},
						Group:    &common.Group{Kind: buildGroup},
					},
				},
			},
			execCommands: []common.DevfileCommand{
				{
					Id: id[0],
					Exec: &common.Exec{
						CommandLine: command[0],
						Component:   component,
						Group:       &common.Group{Kind: runGroup},
						WorkingDir:  workDir[0],
					},
				},
				{
					Id: id[1],
					Exec: &common.Exec{
						CommandLine: command[1],
						Component:   component,
						Group:       &common.Group{Kind: buildGroup},
						WorkingDir:  workDir[1],
					},
				},
				{
					Id: id[2],
					Exec: &common.Exec{
						CommandLine: command[2],
						Component:   component,
						Group:       &common.Group{Kind: runGroup},
						WorkingDir:  workDir[2],
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 4: Invalid composite run command",
			compositeCommands: []common.DevfileCommand{
				{
					Id: id[3],
					Composite: &common.Composite{
						Commands: []string{id[0], id[3], id[2]},
						Group:    &common.Group{Kind: runGroup},
					},
				},
			},
			execCommands: []common.DevfileCommand{
				{
					Id: id[0],
					Exec: &common.Exec{
						CommandLine: command[0],
						Component:   component,
						Group:       &common.Group{Kind: runGroup},
						WorkingDir:  workDir[0],
					},
				},
				{
					Id: id[1],
					Exec: &common.Exec{
						CommandLine: command[1],
						Component:   component,
						Group:       &common.Group{Kind: buildGroup},
						WorkingDir:  workDir[1],
					},
				},
				{
					Id: id[2],
					Exec: &common.Exec{
						CommandLine: command[2],
						Component:   component,
						Group:       &common.Group{Kind: runGroup},
						WorkingDir:  workDir[2],
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 5: Invalid composite command, indirectly references itself",
			compositeCommands: []common.DevfileCommand{
				{
					Id: id[3],
					Composite: &common.Composite{
						Commands: []string{id[4], id[3], id[2]},
						Group:    &common.Group{Kind: buildGroup},
					},
				},
				{
					Id: id[4],
					Composite: &common.Composite{
						Commands: []string{id[0], id[3], id[2]},
						Group:    &common.Group{Kind: buildGroup},
					},
				},
			},
			execCommands: []common.DevfileCommand{
				{
					Id: id[0],
					Exec: &common.Exec{
						CommandLine: command[0],
						Component:   component,
						Group:       &common.Group{Kind: runGroup},
						WorkingDir:  workDir[0],
					},
				},
				{
					Id: id[1],
					Exec: &common.Exec{
						CommandLine: command[1],
						Component:   component,
						Group:       &common.Group{Kind: buildGroup},
						WorkingDir:  workDir[1],
					},
				},
				{
					Id: id[2],
					Exec: &common.Exec{
						CommandLine: command[2],
						Component:   component,
						Group:       &common.Group{Kind: runGroup},
						WorkingDir:  workDir[2],
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 6: Invalid composite command, points to invalid exec command",
			compositeCommands: []common.DevfileCommand{
				{
					Id: id[3],
					Composite: &common.Composite{
						Commands: []string{id[0], id[1]},
						Group:    &common.Group{Kind: buildGroup},
					},
				},
			},
			execCommands: []common.DevfileCommand{
				{
					Id: id[0],
					Exec: &common.Exec{
						CommandLine: command[0],
						Component:   component,
						Group:       &common.Group{Kind: runGroup},
						WorkingDir:  workDir[0],
					},
				},
				{
					Id: id[1],
					Exec: &common.Exec{
						CommandLine: command[1],
						Component:   "some-fake-component",
						Group:       &common.Group{Kind: runGroup},
						WorkingDir:  workDir[1],
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		devObj := devfileParser.DevfileObj{
			Data: &testingutil.TestDevfileData{
				Commands:   append(tt.execCommands, tt.compositeCommands...),
				Components: []common.DevfileComponent{testingutil.GetFakeContainerComponent(component)},
			},
		}

		commands := devObj.Data.GetCommands()
		components := devObj.Data.GetComponents()

		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.compositeCommands[0]
			parentCommands := make(map[string]string)

			err := validateCompositeCommand(&cmd, parentCommands, commands, components)
			if !tt.wantErr == (err != nil) {
				t.Errorf("TestValidateAction unexpected error: %v", err)
				return
			}
		})
	}
}
