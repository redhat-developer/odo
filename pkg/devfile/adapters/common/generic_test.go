package common

import (
	"fmt"
	"io"
	"testing"

	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/redhat-developer/odo/pkg/util"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfileParser "github.com/devfile/library/pkg/devfile/parser"
)

// Create a simple mock client for the ExecClient interface for the devfile exec unit tests.
type mockExecClient struct {
}

type mockExecErrorClient struct {
}

func (fc mockExecClient) ExecCMDInContainer(compInfo ComponentInfo, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
	return nil
}

func (fc mockExecErrorClient) ExecCMDInContainer(compInfo ComponentInfo, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
	return fmt.Errorf("exec error in container %s", compInfo.ContainerName)
}

func TestExecuteDevfileCommand(t *testing.T) {
	var fakeExecClient mockExecClient
	var fakeExecErrorClient mockExecErrorClient
	compInfo := ComponentInfo{
		ContainerName: "some-container",
	}
	cif := func(command devfilev1.Command) (ComponentInfo, error) {
		return compInfo, nil
	}

	commands := []string{"command1", "command2", "command3", "command4"}
	tests := []struct {
		name       string
		commands   []devfilev1.Command
		cmd        devfilev1.Command
		execClient ExecClient
		wantErr    bool
	}{
		{
			name: "Case 1: Non-parallel, successful exec",
			commands: []devfilev1.Command{
				{
					Id: commands[0],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{HotReloadCapable: util.GetBoolPtr(false)},
					},
				},
				{
					Id: commands[1],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{HotReloadCapable: util.GetBoolPtr(false)},
					},
				},
				{
					Id: commands[2],
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{Commands: []string{""}},
					},
				},
			},
			cmd: createCommandFrom(commands[2],
				devfilev1.CompositeCommand{
					Commands: []string{commands[0], commands[1]},
					Parallel: util.GetBoolPtr(false),
				}),
			execClient: fakeExecClient,
			wantErr:    false,
		},
		{
			name: "Case 2: Non-parallel, failed exec",
			commands: []devfilev1.Command{
				{
					Id: commands[0],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{HotReloadCapable: util.GetBoolPtr(false)},
					},
				},
				{
					Id: commands[1],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{HotReloadCapable: util.GetBoolPtr(false)},
					},
				},
				{
					Id: commands[2],
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{Commands: []string{""}},
					},
				},
			},
			cmd: createCommandFrom(commands[2], devfilev1.CompositeCommand{
				Commands: []string{commands[0], commands[1]},
				Parallel: util.GetBoolPtr(false),
			}),
			execClient: fakeExecErrorClient,
			wantErr:    true,
		},
		{
			name: "Case 3: Parallel, successful exec",
			commands: []devfilev1.Command{
				{
					Id: commands[0],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{HotReloadCapable: util.GetBoolPtr(false)},
					},
				},
				{
					Id: commands[1],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{HotReloadCapable: util.GetBoolPtr(false)},
					},
				},
				{
					Id: commands[2],
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{Commands: []string{""}},
					},
				},
			},
			cmd: createCommandFrom(commands[2], devfilev1.CompositeCommand{
				Commands: []string{commands[0], commands[1]},
				Parallel: util.GetBoolPtr(true),
			}),
			execClient: fakeExecClient,
			wantErr:    false,
		},
		{
			name: "Case 4: Parallel, failed exec",
			commands: []devfilev1.Command{
				{
					Id: commands[0],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{HotReloadCapable: util.GetBoolPtr(false)},
					},
				},
				{
					Id: commands[1],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{HotReloadCapable: util.GetBoolPtr(false)},
					},
				},
				{
					Id: commands[2],
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{Commands: []string{""}},
					},
				},
			},
			cmd: createCommandFrom(commands[2], devfilev1.CompositeCommand{
				Commands: []string{commands[0], commands[1]},
				Parallel: util.GetBoolPtr(true),
			}),
			execClient: fakeExecErrorClient,
			wantErr:    true,
		},
		{
			name: "Case 5: Non-Parallel, command not found",
			commands: []devfilev1.Command{
				{
					Id: commands[0],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{HotReloadCapable: util.GetBoolPtr(false)},
					},
				},
				{
					Id: commands[1],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{HotReloadCapable: util.GetBoolPtr(false)},
					},
				},
				{
					Id: commands[2],
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{Commands: []string{""}},
					},
				},
			},
			cmd: createCommandFrom(commands[2], devfilev1.CompositeCommand{
				Commands: []string{commands[0], "fake-command"},
				Parallel: util.GetBoolPtr(false),
			}),
			execClient: fakeExecClient,
			wantErr:    true,
		},
		{
			name: "Case 6: Parallel, command not found",
			commands: []devfilev1.Command{
				{
					Id: commands[0],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{HotReloadCapable: util.GetBoolPtr(false)},
					},
				},
				{
					Id: commands[1],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{HotReloadCapable: util.GetBoolPtr(false)},
					},
				},
				{
					Id: commands[2],
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{Commands: []string{""}},
					},
				},
			},
			cmd: createCommandFrom(commands[2], devfilev1.CompositeCommand{
				Commands: []string{commands[0], "fake-command"},
				Parallel: util.GetBoolPtr(true),
			}),
			execClient: fakeExecClient,
			wantErr:    true,
		},
		{
			name: "Case 7: Nested composite commands",
			commands: []devfilev1.Command{
				{
					Id: commands[0],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{HotReloadCapable: util.GetBoolPtr(false)},
					},
				},
				{
					Id: commands[1],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{HotReloadCapable: util.GetBoolPtr(false)},
					},
				},
				{
					Id: commands[2],
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{Commands: []string{commands[0], commands[1]}},
					},
				},
				{
					Id: commands[3],
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{Commands: []string{""}},
					},
				},
			},
			cmd: createCommandFrom(commands[3], devfilev1.CompositeCommand{
				Commands: []string{commands[0], commands[2]},
				Parallel: util.GetBoolPtr(false),
			}),
			execClient: fakeExecClient,
			wantErr:    false,
		},
		{
			name: "Case 8: Nested parallel composite commands",
			commands: []devfilev1.Command{
				{
					Id: commands[0],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{HotReloadCapable: util.GetBoolPtr(false)},
					},
				},
				{
					Id: commands[1],
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{HotReloadCapable: util.GetBoolPtr(false)},
					},
				},
				{
					Id: commands[2],
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{Commands: []string{commands[0], commands[1]}},
					},
				},
				{
					Id: commands[3],
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{Commands: []string{""}},
					},
				},
			},
			cmd: createCommandFrom(commands[3], devfilev1.CompositeCommand{
				Commands: []string{commands[0], commands[2]},
				Parallel: util.GetBoolPtr(true),
			}),
			execClient: fakeExecClient,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter(tt.execClient, tt.commands, cif).ExecuteDevfileCommand(tt.cmd, false)
			if !tt.wantErr == (err != nil) {
				t.Errorf("expected %v, wanted %v", err, tt.wantErr)
			}
		})
	}
}

func adapter(fakeExecClient ExecClient, commands []devfilev1.Command, cif func(command devfilev1.Command) (ComponentInfo, error)) *GenericAdapter {
	data := func() data.DevfileData {
		devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
		return devfileData
	}()
	_ = data.AddCommands(commands)
	devObj := devfileParser.DevfileObj{
		Data: data,
	}
	ctx := AdapterContext{
		Devfile: devObj,
	}
	a := NewGenericAdapter(fakeExecClient, ctx)
	a.supervisordComponentInfo = cif
	a.componentInfo = cif
	return a
}

func createCommandFrom(id string, composite devfilev1.CompositeCommand) devfilev1.Command {
	return devfilev1.Command{CommandUnion: devfilev1.CommandUnion{Composite: &composite}}
}
