package validate

import (
	"strings"
	"testing"

	"github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/testingutil"
)

func TestValidateEvents(t *testing.T) {

	containers := []string{"container1", "container2"}
	dummyComponents := []common.DevfileComponent{
		{
			Name: containers[0],
			Container: &common.Container{
				Image: "",
			},
		},
		{
			Name: containers[1],
			Container: &common.Container{
				Image: "",
			},
		},
	}

	tests := []struct {
		name         string
		events       common.DevfileEvents
		components   []common.DevfileComponent
		execCommands []common.DevfileCommand
		compCommands []common.DevfileCommand
		wantErr      bool
	}{
		{
			name:       "Case 1: Valid events",
			components: dummyComponents,
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
			events: common.DevfileEvents{
				PostStart: []string{
					"command1",
				},
				PreStop: []string{
					"composite1",
				},
			},
			wantErr: false,
		},
		{
			name:       "Case 2: Invalid events with wrong mapping to devfile command",
			components: dummyComponents,
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
			events: common.DevfileEvents{
				PostStart: []string{
					"command1iswrong",
				},
				PreStop: []string{
					"composite1",
				},
			},
			wantErr: true,
		},
		{
			name:       "Case 3: Invalid event command with mapping to wrong devfile container component",
			components: dummyComponents,
			execCommands: []common.DevfileCommand{
				{
					Id: "command1",
					Exec: &common.Exec{
						CommandLine: "/some/command1",
						Component:   "wrongcomponent",
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
			events: common.DevfileEvents{
				PostStart: []string{
					"command1",
				},
				PreStop: []string{
					"composite1",
				},
			},
			wantErr: true,
		},
		{
			name:       "Case 4: Invalid events with wrong child command in composite command",
			components: dummyComponents,
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
						Commands: []string{"command1iswrong", "command2"},
					},
				},
			},
			events: common.DevfileEvents{
				PostStart: []string{
					"command1",
				},
				PreStop: []string{
					"composite1",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := parser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: tt.components,
					Commands:   append(tt.execCommands, tt.compCommands...),
					Events:     tt.events,
				},
			}
			err := validateEvents(devObj.Data)
			if err != nil && !tt.wantErr {
				t.Errorf("TestValidateEvents error - %v", err)
			}
		})
	}

}

func TestIsEventValid(t *testing.T) {

	containers := []string{"container1", "container2"}
	dummyComponents := []common.DevfileComponent{
		{
			Name: containers[0],
			Container: &common.Container{
				Image: "image",
			},
		},
		{
			Name: containers[1],
			Container: &common.Container{
				Image: "image",
			},
		},
	}

	tests := []struct {
		name         string
		eventType    string
		components   []common.DevfileComponent
		execCommands []common.DevfileCommand
		compCommands []common.DevfileCommand
		eventNames   []string
		wantErr      bool
		wantErrMsg   string
	}{
		{
			name:       "Case 1: Valid events",
			eventType:  "preStart",
			components: dummyComponents,
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
			name:       "Case 2: Invalid events with wrong mapping to devfile command",
			eventType:  "preStart",
			components: dummyComponents,
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
		{
			name:       "Case 3: Invalid event command with mapping to wrong devfile container component",
			eventType:  "preStart",
			components: dummyComponents,
			execCommands: []common.DevfileCommand{
				{
					Id: "command1",
					Exec: &common.Exec{
						CommandLine: "/some/command1",
						Component:   "wrongcomponent",
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
			wantErr:    true,
			wantErrMsg: "does not map to a supported component",
		},
		{
			name:       "Case 4: Invalid events with wrong child command in composite command",
			eventType:  "preStart",
			components: dummyComponents,
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
						Commands: []string{"command1iswrong", "command2"},
					},
				},
			},
			eventNames: []string{
				"command1",
				"composite1",
			},
			wantErr:    true,
			wantErrMsg: "does not exist in the devfile",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := parser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: tt.components,
					Commands:   append(tt.execCommands, tt.compCommands...),
				},
			}

			err := isEventValid(devObj.Data, tt.eventNames, tt.eventType)
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
