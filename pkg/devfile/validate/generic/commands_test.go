package generic

import (
	"testing"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/devfile/library/pkg/testingutil"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
)

var buildGroup = devfilev1.BuildCommandGroupKind
var runGroup = devfilev1.RunCommandGroupKind

func TestValidateCommands(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"

	tests := []struct {
		name    string
		exec    []devfilev1.Command
		comp    []devfilev1.Command
		wantErr bool
	}{
		{
			name: "Case 1: Valid Exec Command",
			exec: []devfilev1.Command{
				{
					Id: "somecommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 6: Valid Composite Command",
			exec: []devfilev1.Command{
				{
					Id: "somecommand1",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
				{
					Id: "somecommand2",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			comp: []devfilev1.Command{
				{
					Id: "composite1",
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup, IsDefault: true},
								},
							},
							Commands: []string{"somecommand1", "somecommand2"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 8: Duplicate commands",
			exec: []devfilev1.Command{
				{
					Id: "somecommand1",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
				{
					Id: "somecommand1",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
				{
					Id: "somecommand2",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 9: Duplicate commands, different types",
			exec: []devfilev1.Command{
				{
					Id: "somecommand1",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
				{
					Id: "somecommand2",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			comp: []devfilev1.Command{
				{
					Id: "somecommand1",
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup, IsDefault: true},
								},
							},
							Commands: []string{"fakecommand"},
						},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		devfileData := testingutil.TestDevfileData{
			Commands:   append(tt.comp, tt.exec...),
			Components: []devfilev1.Component{testingutil.GetFakeContainerComponent(component)},
		}
		devObj := devfileParser.DevfileObj{
			Data: &devfileData,
		}

		commands, err := devObj.Data.GetCommands(parsercommon.DevfileOptions{})
		if err != nil {
			t.Errorf("unexpected error occured: %v", err)
		}

		components, err := devObj.Data.GetComponents(parsercommon.DevfileOptions{})
		if err != nil {
			t.Errorf("unexpected error occured: %v", err)
		}

		commandsMap := common.GetCommandsMap(commands)

		t.Run(tt.name, func(t *testing.T) {
			err := validateCommands(devfileData.Commands, commandsMap, components)
			if !tt.wantErr == (err != nil) {
				t.Errorf("TestValidateAction unexpected error: %v", err)
				return
			}
		})
	}

}

func TestValidateExecCommand(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	invalidComponent := "garbagealias"

	tests := []struct {
		name    string
		exec    devfilev1.Command
		wantErr bool
	}{
		{
			name: "Case 1: Valid Exec Command",
			exec: devfilev1.Command{

				Id: "somecommand",
				CommandUnion: devfilev1.CommandUnion{
					Exec: &devfilev1.ExecCommand{
						LabeledCommand: devfilev1.LabeledCommand{
							BaseCommand: devfilev1.BaseCommand{
								Group: &devfilev1.CommandGroup{Kind: runGroup},
							},
						},
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 2: Invalid Exec Command with empty command",
			exec: devfilev1.Command{
				Id: "somecommand",
				CommandUnion: devfilev1.CommandUnion{
					Exec: &devfilev1.ExecCommand{
						LabeledCommand: devfilev1.LabeledCommand{
							BaseCommand: devfilev1.BaseCommand{
								Group: &devfilev1.CommandGroup{Kind: runGroup},
							},
						},
						CommandLine: "",
						Component:   component,
						WorkingDir:  workDir,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 3: Invalid Exec Command with missing component",
			exec: devfilev1.Command{
				Id: "somecommand",
				CommandUnion: devfilev1.CommandUnion{
					Exec: &devfilev1.ExecCommand{
						LabeledCommand: devfilev1.LabeledCommand{
							BaseCommand: devfilev1.BaseCommand{
								Group: &devfilev1.CommandGroup{Kind: runGroup},
							},
						},
						CommandLine: command,
						WorkingDir:  workDir,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 4: Valid Exec Command with invalid component",
			exec: devfilev1.Command{
				Id: "somecommand",
				CommandUnion: devfilev1.CommandUnion{
					Exec: &devfilev1.ExecCommand{
						LabeledCommand: devfilev1.LabeledCommand{
							BaseCommand: devfilev1.BaseCommand{
								Group: &devfilev1.CommandGroup{Kind: runGroup},
							},
						},
						CommandLine: command,
						Component:   invalidComponent,
						WorkingDir:  workDir,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 5: valid Exec Command with Group nil",
			exec: devfilev1.Command{
				Id: "somecommand",
				CommandUnion: devfilev1.CommandUnion{
					Exec: &devfilev1.ExecCommand{
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		devfileData := testingutil.TestDevfileData{
			Components: []devfilev1.Component{testingutil.GetFakeContainerComponent(component)},
		}
		devObj := devfileParser.DevfileObj{
			Data: &devfileData,
		}

		components, err := devObj.Data.GetComponents(parsercommon.DevfileOptions{})
		if err != nil {
			t.Errorf("unexpected error occured: %v", err)
		}

		t.Run(tt.name, func(t *testing.T) {
			err := validateExecCommand(tt.exec, components)
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
		compositeCommands []devfilev1.Command
		execCommands      []devfilev1.Command
		wantErr           bool
	}{
		{
			name: "Case 1: Valid Composite Command",
			compositeCommands: []devfilev1.Command{
				{
					Id: id[3],
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup},
								},
							},
							Commands: []string{id[0], id[1], id[2]},
						},
					},
				},
			},
			execCommands: []devfilev1.Command{
				{
					Id: id[0],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: command[0],
							Component:   component,
							WorkingDir:  workDir[0],
						},
					},
				},
				{
					Id: id[1],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup},
								},
							},
							CommandLine: command[1],
							Component:   component,
							WorkingDir:  workDir[1],
						},
					},
				},
				{
					Id: id[2],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: command[2],
							Component:   component,
							WorkingDir:  workDir[2],
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 2: Invalid composite command, references non-existent command",
			compositeCommands: []devfilev1.Command{
				{
					Id: id[3],
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup},
								},
							},
							Commands: []string{id[0], "fakecommand", id[2]},
						},
					},
				},
			},
			execCommands: []devfilev1.Command{
				{
					Id: id[0],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: command[0],
							Component:   component,
							WorkingDir:  workDir[0],
						},
					},
				},
				{
					Id: id[1],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup},
								},
							},
							CommandLine: command[1],
							Component:   component,
							WorkingDir:  workDir[1],
						},
					},
				},
				{
					Id: id[2],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: command[2],
							Component:   component,
							WorkingDir:  workDir[2],
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 3: Invalid composite command, references itself",
			compositeCommands: []devfilev1.Command{
				{
					Id: id[3],
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup},
								},
							},
							Commands: []string{id[0], id[3], id[2]},
						},
					},
				},
			},
			execCommands: []devfilev1.Command{
				{
					Id: id[0],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: command[0],
							Component:   component,
							WorkingDir:  workDir[0],
						},
					},
				},
				{
					Id: id[1],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup},
								},
							},
							CommandLine: command[1],
							Component:   component,
							WorkingDir:  workDir[1],
						},
					},
				},
				{
					Id: id[2],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: command[2],
							Component:   component,
							WorkingDir:  workDir[2],
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 4: Invalid composite command, indirectly references itself",
			compositeCommands: []devfilev1.Command{
				{
					Id: id[3],
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup},
								},
							},
							Commands: []string{id[4], id[3], id[2]},
						},
					},
				},
				{
					Id: id[4],
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup},
								},
							},
							Commands: []string{id[0], id[3], id[2]},
						},
					},
				},
			},
			execCommands: []devfilev1.Command{
				{
					Id: id[0],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: command[0],
							Component:   component,
							WorkingDir:  workDir[0],
						},
					},
				},
				{
					Id: id[1],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup},
								},
							},
							CommandLine: command[1],
							Component:   component,
							WorkingDir:  workDir[1],
						},
					},
				},
				{
					Id: id[2],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: command[2],
							Component:   component,
							WorkingDir:  workDir[2],
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Case 5: Invalid composite command, points to invalid exec command",
			compositeCommands: []devfilev1.Command{
				{
					Id: id[3],
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup},
								},
							},
							Commands: []string{id[0], id[1]},
						},
					},
				},
			},
			execCommands: []devfilev1.Command{
				{
					Id: id[0],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: command[0],
							Component:   component,
							WorkingDir:  workDir[0],
						},
					},
				},
				{
					Id: id[1],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: command[1],
							Component:   "some-fake-component",
							WorkingDir:  workDir[1],
						},
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
				Components: []devfilev1.Component{testingutil.GetFakeContainerComponent(component)},
			},
		}

		commands, err := devObj.Data.GetCommands(parsercommon.DevfileOptions{})
		if err != nil {
			t.Errorf("unexpected error occured: %v", err)
		}
		components, err := devObj.Data.GetComponents(parsercommon.DevfileOptions{})
		if err != nil {
			t.Errorf("unexpected error occured: %v", err)
		}

		commandsMap := common.GetCommandsMap(commands)
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.compositeCommands[0]
			parentCommands := make(map[string]string)

			err := validateCompositeCommand(&cmd, parentCommands, commandsMap, components)
			if !tt.wantErr == (err != nil) {
				t.Errorf("TestValidateAction unexpected error: %v", err)
				return
			}
		})
	}
}
