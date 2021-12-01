package common

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/devfile/library/pkg/devfile/parser/data"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/testingutil"
	"github.com/kylelemons/godebug/pretty"
	"github.com/redhat-developer/odo/pkg/util"
)

var buildGroup = devfilev1.BuildCommandGroupKind
var runGroup = devfilev1.RunCommandGroupKind
var testGroup = devfilev1.TestCommandGroupKind
var debugGroup = devfilev1.DebugCommandGroupKind

func TestGetCommand(t *testing.T) {

	commands := [...]string{"ls -la", "pwd"}
	components := [...]string{"alias1", "alias2"}

	tests := []struct {
		name           string
		requestedType  []devfilev1.CommandGroupKind
		execCommands   []devfilev1.Command
		compCommands   []devfilev1.Command
		reqCommandName string
		retCommandName string
		wantErr        bool
	}{
		{
			name: "Case 1: Valid devfile",
			execCommands: []devfilev1.Command{
				getExecCommand("build", buildGroup),
				getExecCommand("run", runGroup),
			},
			requestedType: []devfilev1.CommandGroupKind{buildGroup, runGroup},
			wantErr:       false,
		},
		{
			name: "Case 2: Valid devfile with devrun and devbuild",
			execCommands: []devfilev1.Command{
				getExecCommand("build", buildGroup),
				getExecCommand("run", runGroup),
			},
			requestedType: []devfilev1.CommandGroupKind{buildGroup, runGroup},
			wantErr:       false,
		},
		{
			name: "Case 3: Valid devfile with empty workdir",
			execCommands: []devfilev1.Command{
				{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			requestedType: []devfilev1.CommandGroupKind{runGroup},
			wantErr:       false,
		},
		{
			name: "Case 4: Mismatched command type",
			execCommands: []devfilev1.Command{
				{
					Id: "build command",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			reqCommandName: "build command",
			requestedType:  []devfilev1.CommandGroupKind{buildGroup},
			wantErr:        true,
		},
		{
			name: "Case 5: Default command is returned",
			execCommands: []devfilev1.Command{
				{
					Id: "defaultRunCommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
				{
					Id: "runCommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			retCommandName: "defaultRunCommand",
			requestedType:  []devfilev1.CommandGroupKind{runGroup},
			wantErr:        false,
		},
		{
			name: "Case 6: Composite command is returned",
			execCommands: []devfilev1.Command{
				{
					Id: "build",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(false)},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
				{
					Id: "run",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			compCommands: []devfilev1.Command{
				{
					Id: "myComposite",
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							Commands: []string{"build", "run"},
						},
					},
				},
			},
			retCommandName: "myComposite",
			requestedType:  []devfilev1.CommandGroupKind{buildGroup},
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := []devfilev1.Component{testingutil.GetFakeContainerComponent(tt.execCommands[0].Exec.Component)}
			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.execCommands)
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.compCommands)
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents(components)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			for _, gtype := range tt.requestedType {
				cmd, err := getCommand(devObj.Data, tt.reqCommandName, gtype)
				if !tt.wantErr == (err != nil) {
					t.Errorf("TestGetCommand unexpected error for command: %v wantErr: %v err: %v", gtype, tt.wantErr, err)
					return
				} else if tt.wantErr {
					return
				}

				if len(tt.retCommandName) > 0 && cmd.Id != tt.retCommandName {
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
		requestedType  []devfilev1.CommandGroupKind
		execCommands   []devfilev1.Command
		compCommands   []devfilev1.Command
		retCommandName string
		wantErr        bool
	}{
		{
			name: "Case 1: Valid devfile",
			execCommands: []devfilev1.Command{
				getExecCommand("", buildGroup),
				getExecCommand("", runGroup),
			},
			requestedType: []devfilev1.CommandGroupKind{buildGroup, runGroup},
			wantErr:       false,
		},
		{
			name: "Case 2: Valid devfile with devrun and devbuild",
			execCommands: []devfilev1.Command{
				getExecCommand("", buildGroup),
				getExecCommand("", runGroup),
			},
			requestedType: []devfilev1.CommandGroupKind{buildGroup, runGroup},
			wantErr:       false,
		},
		{
			name: "Case 3: Valid devfile with empty workdir",
			execCommands: []devfilev1.Command{
				{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			requestedType: []devfilev1.CommandGroupKind{runGroup},
			wantErr:       false,
		},
		{
			name: "Case 4: Default command is returned",
			execCommands: []devfilev1.Command{
				{
					Id: "defaultruncommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
				{
					Id: "runcommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			retCommandName: "defaultruncommand",
			requestedType:  []devfilev1.CommandGroupKind{runGroup},
			wantErr:        false,
		},
		{
			name: "Case 5: Valid devfile, has composite command",
			execCommands: []devfilev1.Command{
				{
					Id: "build1",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
				{
					Id: "build2",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
				{
					Id: "run",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			compCommands: []devfilev1.Command{
				{
					Id: "mycomp",
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							Commands: []string{"build1", "run"},
						},
					},
				},
			},
			retCommandName: "mycomp",
			requestedType:  []devfilev1.CommandGroupKind{buildGroup},
			wantErr:        false,
		},
		{
			name: "Case 6: Default composite command",
			execCommands: []devfilev1.Command{
				{
					Id: "build",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(false)},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
				{
					Id: "run",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			compCommands: []devfilev1.Command{
				{
					Id: "mycomp",
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							Commands: []string{"build", "run"},
						},
					},
				},
				{
					Id: "mycomp2",
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(false)},
								},
							},
							Commands: []string{"build", "run"},
						},
					},
				},
			},
			retCommandName: "mycomp",
			requestedType:  []devfilev1.CommandGroupKind{buildGroup},
			wantErr:        false,
		},
		{
			name: "Case 7: no build and debug commands",
			execCommands: []devfilev1.Command{
				getExecCommand("", runGroup),
			},
			requestedType: []devfilev1.CommandGroupKind{buildGroup, debugGroup},
			wantErr:       false,
		},
		{
			name: "Case 8: no default build and debug commands",
			execCommands: []devfilev1.Command{
				getExecCommand("build-0", buildGroup),
				getExecCommand("build-1", buildGroup),
				getExecCommand("debug-0", debugGroup),
				getExecCommand("debug-1", debugGroup),
				getExecCommand("", runGroup),
			},
			requestedType: []devfilev1.CommandGroupKind{buildGroup, debugGroup},
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := []devfilev1.Component{testingutil.GetFakeContainerComponent(tt.execCommands[0].Exec.Component)}
			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.execCommands)
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.compCommands)
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents(components)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			for _, gtype := range tt.requestedType {
				cmd, err := getCommandFromDevfile(devObj.Data, gtype)
				if !tt.wantErr == (err != nil) {
					t.Errorf("TestGetCommandFromDevfile unexpected error for command: %v wantErr: %v err: %v", gtype, tt.wantErr, err)
					return
				} else if tt.wantErr {
					return
				}

				if len(tt.retCommandName) > 0 && cmd.Id != tt.retCommandName {
					t.Errorf("TestGetCommandFromDevfile error: command names do not match expected: %v actual: %v", tt.retCommandName, cmd.Id)
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
		requestedType  devfilev1.CommandGroupKind
		execCommands   []devfilev1.Command
		compCommands   []devfilev1.Command
		reqCommandName string
		retCommandName string
		wantErr        bool
	}{
		{
			name: "Case 1: Valid devfile",
			execCommands: []devfilev1.Command{
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
			execCommands: []devfilev1.Command{
				{
					Id: "build command",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
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
			execCommands: []devfilev1.Command{
				{
					Id: "build command",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   invalidComponent,
						},
					},
				},
			},
			reqCommandName: "build command wrong",
			requestedType:  runGroup,
			wantErr:        true,
		},
		{
			name: "Case 4: Mismatched command type",
			execCommands: []devfilev1.Command{
				{
					Id: "build command",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			reqCommandName: "build command",
			requestedType:  buildGroup,
			wantErr:        true,
		},
		{
			name: "Case 5: Multiple default commands but should be with the flag",
			execCommands: []devfilev1.Command{
				{
					Id: "defaultruncommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
				{
					Id: "runcommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
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
			execCommands: []devfilev1.Command{
				{
					Id: "defaultruncommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
				{
					Id: "runcommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
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
			execCommands: []devfilev1.Command{
				{
					Id: "defaultruncommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							CommandLine: commands[0],
							Component:   components[0],
						},
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
			execCommands: []devfilev1.Command{
				{
					Id: "build",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(false)},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
				{
					Id: "run",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			compCommands: []devfilev1.Command{
				{
					Id: "mycomp",
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							Commands: []string{"build", "run"},
						},
					},
				},
				{
					Id: "mycomp2",
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(false)},
								},
							},
							Commands: []string{"build", "run"},
						},
					},
				},
			},
			reqCommandName: "mycomp",
			retCommandName: "mycomp",
			requestedType:  buildGroup,
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := []devfilev1.Component{testingutil.GetFakeContainerComponent(tt.execCommands[0].Exec.Component)}
			if tt.execCommands[0].Exec.Component == invalidComponent {
				components = []devfilev1.Component{testingutil.GetFakeContainerComponent("randomComponent")}
			}
			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.execCommands)
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.compCommands)
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents(components)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
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

func TestGetBuildCommand(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""

	tests := []struct {
		name         string
		commandName  string
		execCommands []devfilev1.Command
		wantCommand  devfilev1.Command
		wantErr      bool
	}{
		{
			name:        "Case 1: Default Build Command",
			commandName: emptyString,
			execCommands: []devfilev1.Command{
				{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantCommand: devfilev1.Command{
				CommandUnion: devfilev1.CommandUnion{
					Exec: &devfilev1.ExecCommand{
						LabeledCommand: devfilev1.LabeledCommand{
							BaseCommand: devfilev1.BaseCommand{
								Group: &devfilev1.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(true)},
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
			name:        "Case 2: Build Command passed through the odo flag",
			commandName: "flagcommand",
			execCommands: []devfilev1.Command{
				{
					Id: "flagcommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
				{
					Id: "build command",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantCommand: devfilev1.Command{
				Id: "flagcommand",
				CommandUnion: devfilev1.CommandUnion{
					Exec: &devfilev1.ExecCommand{
						LabeledCommand: devfilev1.LabeledCommand{
							BaseCommand: devfilev1.BaseCommand{
								Group: &devfilev1.CommandGroup{Kind: buildGroup},
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
			name:        "Case 3: Build Command not found",
			commandName: "customcommand123",
			execCommands: []devfilev1.Command{
				{
					Id: "build command",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
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
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.execCommands)
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{testingutil.GetFakeContainerComponent(component)})
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			command, err := GetBuildCommand(devObj.Data, tt.commandName)

			if !tt.wantErr == (err != nil) {
				t.Errorf("TestGetBuildCommand: unexpected error for command \"%v\" expected: %v actual: %v", tt.commandName, tt.wantErr, err)
			} else if !tt.wantErr && !reflect.DeepEqual(tt.wantCommand, command) {
				t.Errorf("TestGetBuildCommand: unexpected command returned: %v", pretty.Compare(tt.wantCommand, command))
			}

		})
	}

}

func TestGetDebugCommand(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""

	var emptyCommand devfilev1.Command

	tests := []struct {
		name         string
		commandName  string
		execCommands []devfilev1.Command
		wantErr      bool
	}{
		{
			name:        "Case: Default Debug Command",
			commandName: emptyString,
			execCommands: []devfilev1.Command{
				{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{
										IsDefault: util.GetBoolPtr(true),
										Kind:      devfilev1.DebugCommandGroupKind,
									},
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
			name:        "Case: Custom Debug Command",
			commandName: "customdebugcommand",
			execCommands: []devfilev1.Command{
				{
					Id: "customdebugcommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{
										IsDefault: util.GetBoolPtr(false),
										Kind:      devfilev1.DebugCommandGroupKind,
									},
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
			name:        "Case: Missing Debug Command",
			commandName: "customcommand123",
			execCommands: []devfilev1.Command{
				{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{
										IsDefault: util.GetBoolPtr(true),
										Kind:      devfilev1.BuildCommandGroupKind,
									},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
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
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{testingutil.GetFakeContainerComponent(component)})
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.execCommands)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
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

	var emptyCommand devfilev1.Command

	tests := []struct {
		name         string
		commandName  string
		execCommands []devfilev1.Command
		wantErr      bool
	}{
		{
			name:        "Case: Default Test Command",
			commandName: emptyString,
			execCommands: []devfilev1.Command{
				{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{
										IsDefault: util.GetBoolPtr(true),
										Kind:      devfilev1.TestCommandGroupKind,
									},
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
			name:        "Case: Custom Test Command",
			commandName: "customtestcommand",
			execCommands: []devfilev1.Command{
				{
					Id: "customtestcommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{
										IsDefault: util.GetBoolPtr(false),
										Kind:      devfilev1.TestCommandGroupKind,
									},
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
			name:        "Case: Missing Test Command",
			commandName: "customcommand123",
			execCommands: []devfilev1.Command{
				{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{
										IsDefault: util.GetBoolPtr(true),
										Kind:      devfilev1.BuildCommandGroupKind,
									},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
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
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{testingutil.GetFakeContainerComponent(component)})
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.execCommands)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
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

	var emptyCommand devfilev1.Command

	tests := []struct {
		name         string
		commandName  string
		execCommands []devfilev1.Command
		wantErr      bool
	}{
		{
			name:        "Case 1: Default Run Command",
			commandName: emptyString,
			execCommands: []devfilev1.Command{
				{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup, IsDefault: util.GetBoolPtr(true)},
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
			name:        "Case 2: Run Command passed through odo flag",
			commandName: "flagcommand",
			execCommands: []devfilev1.Command{
				{
					Id: "flagcommand",
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
				{
					Id: "run command",
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
			name:        "Case 3: Missing Run Command",
			commandName: "",
			execCommands: []devfilev1.Command{
				{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: buildGroup},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
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
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{testingutil.GetFakeContainerComponent(component)})
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.execCommands)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
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

	execCommands := []devfilev1.Command{
		{
			CommandUnion: devfilev1.CommandUnion{
				Exec: &devfilev1.ExecCommand{
					LabeledCommand: devfilev1.LabeledCommand{
						BaseCommand: devfilev1.BaseCommand{
							Group: &devfilev1.CommandGroup{
								IsDefault: util.GetBoolPtr(true),
								Kind:      devfilev1.DebugCommandGroupKind,
							},
						},
					},
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
				},
			},
		},
		{
			Id: "customdebugcommand",
			CommandUnion: devfilev1.CommandUnion{
				Exec: &devfilev1.ExecCommand{
					LabeledCommand: devfilev1.LabeledCommand{
						BaseCommand: devfilev1.BaseCommand{
							Group: &devfilev1.CommandGroup{
								IsDefault: util.GetBoolPtr(false),
								Kind:      devfilev1.DebugCommandGroupKind,
							},
						},
					},
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
				},
			},
		},
	}

	tests := []struct {
		name          string
		debugCommand  string
		componentType devfilev1.ComponentType
		wantErr       bool
	}{
		{
			name:          "Case: Default Devfile Commands",
			debugCommand:  emptyString,
			componentType: devfilev1.ContainerComponentType,
			wantErr:       false,
		},
		{
			name:          "Case: provided debug Command",
			debugCommand:  "customdebugcommand",
			componentType: devfilev1.ContainerComponentType,
			wantErr:       false,
		},
		{
			name:          "Case: invalid debug Command",
			debugCommand:  "invaliddebugcommand",
			componentType: devfilev1.ContainerComponentType,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{testingutil.GetFakeContainerComponent(component)})
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(execCommands)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
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

	execCommands := []devfilev1.Command{
		{
			Id: "run command",
			CommandUnion: devfilev1.CommandUnion{
				Exec: &devfilev1.ExecCommand{
					LabeledCommand: devfilev1.LabeledCommand{
						BaseCommand: devfilev1.BaseCommand{
							Group: &devfilev1.CommandGroup{
								Kind:      runGroup,
								IsDefault: util.GetBoolPtr(true),
							},
						},
					},
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
				},
			},
		},

		{
			Id: "build command",
			CommandUnion: devfilev1.CommandUnion{
				Exec: &devfilev1.ExecCommand{
					LabeledCommand: devfilev1.LabeledCommand{
						BaseCommand: devfilev1.BaseCommand{
							Group: &devfilev1.CommandGroup{Kind: buildGroup},
						},
					},
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
				},
			},
		},

		{
			Id: "customcommand",
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
	}

	wrongCompTypeCmd := devfilev1.Command{

		Id: "wrong",
		CommandUnion: devfilev1.CommandUnion{
			Exec: &devfilev1.ExecCommand{
				LabeledCommand: devfilev1.LabeledCommand{
					BaseCommand: devfilev1.BaseCommand{
						Group: &devfilev1.CommandGroup{Kind: runGroup},
					},
				},
				CommandLine: command,
				Component:   "",
				WorkingDir:  workDir,
			},
		},
	}

	tests := []struct {
		name                string
		buildCommand        string
		runCommand          string
		execCommands        []devfilev1.Command
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
			execCommands: []devfilev1.Command{
				{
					Id: "customcommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: runGroup},
								},
							},
							Component:   component,
							CommandLine: command,
						},
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
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.execCommands)
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents(([]devfilev1.Component{testingutil.GetFakeContainerComponent(component)}))
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
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

	execCommands := []devfilev1.Command{
		{
			CommandUnion: devfilev1.CommandUnion{
				Exec: &devfilev1.ExecCommand{
					LabeledCommand: devfilev1.LabeledCommand{
						BaseCommand: devfilev1.BaseCommand{
							Group: &devfilev1.CommandGroup{
								IsDefault: util.GetBoolPtr(true),
								Kind:      devfilev1.TestCommandGroupKind,
							},
						},
					},
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
				},
			},
		},
		{
			Id: "customtestcommand",
			CommandUnion: devfilev1.CommandUnion{
				Exec: &devfilev1.ExecCommand{
					LabeledCommand: devfilev1.LabeledCommand{
						BaseCommand: devfilev1.BaseCommand{
							Group: &devfilev1.CommandGroup{
								IsDefault: util.GetBoolPtr(false),
								Kind:      devfilev1.TestCommandGroupKind,
							},
						},
					},
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
				},
			},
		},
	}

	tests := []struct {
		name          string
		testCommand   string
		componentType devfilev1.ComponentType
		wantErr       bool
	}{
		{
			name:          "Case: Default Devfile Commands",
			testCommand:   emptyString,
			componentType: devfilev1.ContainerComponentType,
			wantErr:       false,
		},
		{
			name:          "Case: provided test Command",
			testCommand:   "customtestcommand",
			componentType: devfilev1.ContainerComponentType,
			wantErr:       false,
		},
		{
			name:          "Case: invalid test Command",
			testCommand:   "invalidtestcommand",
			componentType: devfilev1.ContainerComponentType,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{testingutil.GetFakeContainerComponent(component)})
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(execCommands)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
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

func getExecCommand(id string, group devfilev1.CommandGroupKind) devfilev1.Command {
	if len(id) == 0 {
		id = fmt.Sprintf("%s-%s", "cmd", util.GenerateRandomString(10))
	}
	commands := [...]string{"ls -la", "pwd"}
	components := [...]string{"alias1", "alias2"}
	workDir := [...]string{"/", "/root"}

	return devfilev1.Command{
		Id: id,
		CommandUnion: devfilev1.CommandUnion{
			Exec: &devfilev1.ExecCommand{
				LabeledCommand: devfilev1.LabeledCommand{
					BaseCommand: devfilev1.BaseCommand{
						Group: &devfilev1.CommandGroup{Kind: group},
					},
				},
				CommandLine: commands[0],
				Component:   components[0],
				WorkingDir:  workDir[0],
			},
		},
	}

}
