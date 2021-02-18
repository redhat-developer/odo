package generic

import (
	parsercommon "github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"strings"
	"testing"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/testingutil"
)

func TestIsEventValid(t *testing.T) {

	containers := []string{"container1", "container2"}

	tests := []struct {
		name         string
		eventType    string
		execCommands []devfilev1.Command
		compCommands []devfilev1.Command
		eventNames   []string
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name:      "Case 1: Valid events",
			eventType: "preStart",
			execCommands: []devfilev1.Command{
				{
					Id: "command1",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							CommandLine: "/some/command1",
							Component:   containers[0],
							WorkingDir:  "workDir",
						},
					},
				},
				{
					Id: "command2",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							CommandLine: "/some/command2",
							Component:   containers[1],
							WorkingDir:  "workDir",
						},
					},
				},
			},
			compCommands: []devfilev1.Command{
				{
					Id: "composite1",
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{
							Commands: []string{"command1", "command2"},
						},
					},
				},
			},
			eventNames: []string{
				"command1",
				"composite1",
			},
			wantErr: false,
		},
		{
			name:      "Case 2: Invalid events with wrong mapping to devfile command",
			eventType: "preStart",
			execCommands: []devfilev1.Command{
				{
					Id: "command1",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							CommandLine: "/some/command1",
							Component:   containers[0],
							WorkingDir:  "workDir",
						},
					},
				},
				{
					Id: "command2",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							CommandLine: "/some/command2",
							Component:   containers[1],
							WorkingDir:  "workDir",
						},
					},
				},
			},
			compCommands: []devfilev1.Command{
				{
					Id: "composite1",
					CommandUnion: devfilev1.CommandUnion{
						Composite: &devfilev1.CompositeCommand{
							Commands: []string{"command1", "command2"},
						},
					},
				},
			},
			eventNames: []string{
				"command12345iswrong",
				"composite1",
			},
			wantErr:    true,
			wantErrMsg: "does not map to a valid devfile command",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := parser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Commands: append(tt.execCommands, tt.compCommands...),
				},
			}

			commands, err := devObj.Data.GetCommands(parsercommon.DevfileOptions{})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			commandMap := common.GetCommandsMap(commands)

			err = isEventValid(tt.eventNames, tt.eventType, commandMap)
			if err != nil && !tt.wantErr {
				t.Errorf("TestIsEventValid error: %v", err)
			} else if err != nil && tt.wantErr {
				if !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("TestIsEventValid error mismatch - %s; does not contain: %s", err.Error(), tt.wantErrMsg)
				}
			}
		})
	}

}
