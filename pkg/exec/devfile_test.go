package exec

import (
	"fmt"
	"io"
	"testing"

	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/machineoutput"
)

// Create a simple mock client for the ExecClient interface for the devfile exec unit tests.
type mockExecClient struct {
}

type mockExecErrorClient struct {
}

func (fc mockExecClient) ExecCMDInContainer(compInfo adaptersCommon.ComponentInfo, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
	return nil
}

func (fc mockExecErrorClient) ExecCMDInContainer(compInfo adaptersCommon.ComponentInfo, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
	return fmt.Errorf("exec error in container %s", compInfo.ContainerName)
}

func TestExecuteCompositeDevfileAction(t *testing.T) {
	var fakeExecClient mockExecClient
	var fakeExecErrorClient mockExecErrorClient
	compInfo := adaptersCommon.ComponentInfo{
		ContainerName: "some-container",
	}

	commands := []string{"command1", "command2", "command3", "command4"}
	tests := []struct {
		name        string
		commandsMap map[string]common.DevfileCommand
		composite   common.Composite
		execClient  ExecClient
		wantErr     bool
	}{
		{
			name: "Case 1: Non-parallel, successful exec",
			commandsMap: map[string]common.DevfileCommand{
				commands[0]: {
					Exec: &common.Exec{Id: commands[0]},
				},
				commands[1]: {
					Exec: &common.Exec{Id: commands[1]},
				},
				commands[2]: {
					Composite: &common.Composite{Id: commands[2]},
				},
			},
			composite: common.Composite{
				Id:       commands[2],
				Commands: []string{commands[0], commands[1]},
				Parallel: false,
			},
			execClient: fakeExecClient,
			wantErr:    false,
		},
		{
			name: "Case 2: Non-parallel, failed exec",
			commandsMap: map[string]common.DevfileCommand{
				commands[0]: {
					Exec: &common.Exec{Id: commands[0]},
				},
				commands[1]: {
					Exec: &common.Exec{Id: commands[1]},
				},
				commands[2]: {
					Composite: &common.Composite{Id: commands[2]},
				},
			},
			composite: common.Composite{
				Id:       commands[2],
				Commands: []string{commands[0], commands[1]},
				Parallel: false,
			},
			execClient: fakeExecErrorClient,
			wantErr:    true,
		},
		{
			name: "Case 3: Parallel, successful exec",
			commandsMap: map[string]common.DevfileCommand{
				commands[0]: {
					Exec: &common.Exec{Id: commands[0]},
				},
				commands[1]: {
					Exec: &common.Exec{Id: commands[1]},
				},
				commands[2]: {
					Composite: &common.Composite{Id: commands[2]},
				},
			},
			composite: common.Composite{
				Id:       commands[2],
				Commands: []string{commands[0], commands[1]},
				Parallel: true,
			},
			execClient: fakeExecClient,
			wantErr:    false,
		},
		{
			name: "Case 4: Parallel, failed exec",
			commandsMap: map[string]common.DevfileCommand{
				commands[0]: {
					Exec: &common.Exec{Id: commands[0]},
				},
				commands[1]: {
					Exec: &common.Exec{Id: commands[1]},
				},
				commands[2]: {
					Composite: &common.Composite{Id: commands[2]},
				},
			},
			composite: common.Composite{
				Id:       commands[2],
				Commands: []string{commands[0], commands[1]},
				Parallel: true,
			},
			execClient: fakeExecErrorClient,
			wantErr:    true,
		},
		{
			name: "Case 5: Non-Parallel, command not found",
			commandsMap: map[string]common.DevfileCommand{
				commands[0]: {
					Exec: &common.Exec{Id: commands[0]},
				},
				commands[1]: {
					Exec: &common.Exec{Id: commands[1]},
				},
				commands[2]: {
					Composite: &common.Composite{Id: commands[2]},
				},
			},
			composite: common.Composite{
				Id:       commands[2],
				Commands: []string{commands[0], "fake-command"},
				Parallel: false,
			},
			execClient: fakeExecClient,
			wantErr:    true,
		},
		{
			name: "Case 6: Parallel, command not found",
			commandsMap: map[string]common.DevfileCommand{
				commands[0]: {
					Exec: &common.Exec{Id: commands[0]},
				},
				commands[1]: {
					Exec: &common.Exec{Id: commands[1]},
				},
				commands[2]: {
					Composite: &common.Composite{Id: commands[2]},
				},
			},
			composite: common.Composite{
				Id:       commands[2],
				Commands: []string{commands[0], "fake-command"},
				Parallel: true,
			},
			execClient: fakeExecClient,
			wantErr:    true,
		},
		{
			name: "Case 7: Nested composite commands",
			commandsMap: map[string]common.DevfileCommand{
				commands[0]: {
					Exec: &common.Exec{Id: commands[0]},
				},
				commands[1]: {
					Exec: &common.Exec{Id: commands[1]},
				},
				commands[2]: {
					Composite: &common.Composite{Id: commands[2], Commands: []string{commands[0], commands[1]}},
				},
				commands[3]: {
					Composite: &common.Composite{Id: commands[3]},
				},
			},
			composite: common.Composite{
				Id:       commands[3],
				Commands: []string{commands[0], commands[2]},
				Parallel: false,
			},
			execClient: fakeExecClient,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ExecuteCompositeDevfileAction(tt.execClient, tt.composite, tt.commandsMap, compInfo, true, machineoutput.NewNoOpMachineEventLoggingClient())
			if !tt.wantErr == (err != nil) {
				t.Errorf("expected %v, wanted %v", err, tt.wantErr)
			}
		})
	}
}
