package common

import (
	"fmt"
	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/testingutil"
	"io"
	"testing"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
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
	cif := func(command common.DevfileCommand) (ComponentInfo, error) {
		return compInfo, nil
	}

	commands := []string{"command1", "command2", "command3", "command4"}
	tests := []struct {
		name       string
		commands   []common.DevfileCommand
		cmd        common.DevfileCommand
		execClient ExecClient
		wantErr    bool
	}{
		{
			name: "Case 1: Non-parallel, successful exec",
			commands: []common.DevfileCommand{
				{
					Exec: &common.Exec{Id: commands[0]},
				},
				{
					Exec: &common.Exec{Id: commands[1]},
				},
				{
					Composite: &common.Composite{Id: commands[2]},
				},
			},
			cmd: createFrom(common.Composite{
				Id:       commands[2],
				Commands: []string{commands[0], commands[1]},
				Parallel: false,
			}),
			execClient: fakeExecClient,
			wantErr:    false,
		},
		{
			name: "Case 2: Non-parallel, failed exec",
			commands: []common.DevfileCommand{
				{
					Exec: &common.Exec{Id: commands[0]},
				},
				{
					Exec: &common.Exec{Id: commands[1]},
				},
				{
					Composite: &common.Composite{Id: commands[2]},
				},
			},
			cmd: createFrom(common.Composite{
				Id:       commands[2],
				Commands: []string{commands[0], commands[1]},
				Parallel: false,
			}),
			execClient: fakeExecErrorClient,
			wantErr:    true,
		},
		{
			name: "Case 3: Parallel, successful exec",
			commands: []common.DevfileCommand{
				{
					Exec: &common.Exec{Id: commands[0]},
				},
				{
					Exec: &common.Exec{Id: commands[1]},
				},
				{
					Composite: &common.Composite{Id: commands[2]},
				},
			},
			cmd: createFrom(common.Composite{
				Id:       commands[2],
				Commands: []string{commands[0], commands[1]},
				Parallel: true,
			}),
			execClient: fakeExecClient,
			wantErr:    false,
		},
		{
			name: "Case 4: Parallel, failed exec",
			commands: []common.DevfileCommand{
				{
					Exec: &common.Exec{Id: commands[0]},
				},
				{
					Exec: &common.Exec{Id: commands[1]},
				},
				{
					Composite: &common.Composite{Id: commands[2]},
				},
			},
			cmd: createFrom(common.Composite{
				Id:       commands[2],
				Commands: []string{commands[0], commands[1]},
				Parallel: true,
			}),
			execClient: fakeExecErrorClient,
			wantErr:    true,
		},
		{
			name: "Case 5: Non-Parallel, command not found",
			commands: []common.DevfileCommand{
				{
					Exec: &common.Exec{Id: commands[0]},
				},
				{
					Exec: &common.Exec{Id: commands[1]},
				},
				{
					Composite: &common.Composite{Id: commands[2]},
				},
			},
			cmd: createFrom(common.Composite{
				Id:       commands[2],
				Commands: []string{commands[0], "fake-command"},
				Parallel: false,
			}),
			execClient: fakeExecClient,
			wantErr:    true,
		},
		{
			name: "Case 6: Parallel, command not found",
			commands: []common.DevfileCommand{
				{
					Exec: &common.Exec{Id: commands[0]},
				},
				{
					Exec: &common.Exec{Id: commands[1]},
				},
				{
					Composite: &common.Composite{Id: commands[2]},
				},
			},
			cmd: createFrom(common.Composite{
				Id:       commands[2],
				Commands: []string{commands[0], "fake-command"},
				Parallel: true,
			}),
			execClient: fakeExecClient,
			wantErr:    true,
		},
		{
			name: "Case 7: Nested composite commands",
			commands: []common.DevfileCommand{
				{
					Exec: &common.Exec{Id: commands[0]},
				},
				{
					Exec: &common.Exec{Id: commands[1]},
				},
				{
					Composite: &common.Composite{Id: commands[2], Commands: []string{commands[0], commands[1]}},
				},
				{
					Composite: &common.Composite{Id: commands[3]},
				},
			},
			cmd: createFrom(common.Composite{
				Id:       commands[3],
				Commands: []string{commands[0], commands[2]},
				Parallel: false,
			}),
			execClient: fakeExecClient,
			wantErr:    false,
		},
		{
			name: "Case 8: Nested parallel composite commands",
			commands: []common.DevfileCommand{
				{
					Exec: &common.Exec{Id: commands[0]},
				},
				{
					Exec: &common.Exec{Id: commands[1]},
				},
				{
					Composite: &common.Composite{Id: commands[2], Commands: []string{commands[0], commands[1]}},
				},
				{
					Composite: &common.Composite{Id: commands[3]},
				},
			},
			cmd: createFrom(common.Composite{
				Id:       commands[3],
				Commands: []string{commands[0], commands[2]},
				Parallel: true,
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

func adapter(fakeExecClient ExecClient, commands []common.DevfileCommand, cif func(command common.DevfileCommand) (ComponentInfo, error)) GenericAdapter {
	devObj := devfileParser.DevfileObj{
		Data: testingutil.TestDevfileData{
			Commands: commands,
		},
	}
	ctx := AdapterContext{
		Devfile: devObj,
	}
	return NewGenericAdapter(fakeExecClient, ctx, cif, cif)
}

func createFrom(composite common.Composite) common.DevfileCommand {
	return common.DevfileCommand{Composite: &composite}
}
