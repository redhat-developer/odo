package generic

import (
	"strings"
	"testing"

	"github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/testingutil"
)

func TestIsEventValid(t *testing.T) {

	containers := []string{"container1", "container2"}

	tests := []struct {
		name         string
		eventType    string
		execCommands []common.DevfileCommand
		compCommands []common.DevfileCommand
		eventNames   []string
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name:      "Case 1: Valid events",
			eventType: "preStart",
			execCommands: []common.DevfileCommand{
				{
					Id: "command1",
					Exec: &common.Exec{
						CommandLine: "/some/command1",
						Component:   containers[0],
						WorkingDir:  "workDir",
					},
				},
				{
					Id: "command2",
					Exec: &common.Exec{
						CommandLine: "/some/command2",
						Component:   containers[1],
						WorkingDir:  "workDir",
					},
				},
			},
			compCommands: []common.DevfileCommand{
				{
					Id: "composite1",
					Composite: &common.Composite{
						Commands: []string{"command1", "command2"},
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
			execCommands: []common.DevfileCommand{
				{
					Id: "command1",
					Exec: &common.Exec{
						CommandLine: "/some/command1",
						Component:   containers[0],
						WorkingDir:  "workDir",
					},
				},
				{
					Id: "command2",
					Exec: &common.Exec{
						CommandLine: "/some/command2",
						Component:   containers[1],
						WorkingDir:  "workDir",
					},
				},
			},
			compCommands: []common.DevfileCommand{
				{
					Id: "composite1",
					Composite: &common.Composite{
						Commands: []string{"command1", "command2"},
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

			commands := devObj.Data.GetCommands()

			err := IsEventValid(tt.eventNames, tt.eventType, commands)
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
