package common

import (
	"reflect"
	"testing"
)

func TestGetID(t *testing.T) {

	tests := []struct {
		name    string
		command DevfileCommand
		want    string
	}{
		{
			name: "Case 1: Exec command ID",
			command: DevfileCommand{
				Id: "exec1",
				Exec: &Exec{
					Component: "nodejs",
				},
			},
			want: "exec1",
		},
		{
			name: "Case 2: Composite command ID",
			command: DevfileCommand{
				Id: "composite1",
				Composite: &Composite{
					Parallel: false,
				},
			},
			want: "composite1",
		},
		{
			name:    "Case 3: Empty command",
			command: DevfileCommand{},
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commandID := tt.command.GetID()
			if commandID != tt.want {
				t.Errorf("expected %v, actual %v", tt.want, commandID)
			}
		})
	}

}

func TestGetGroup(t *testing.T) {

	tests := []struct {
		name    string
		command DevfileCommand
		want    *Group
	}{
		{
			name: "Case 1: Exec command group",
			command: DevfileCommand{
				Id: "exec1",
				Exec: &Exec{
					Group: &Group{
						IsDefault: true,
						Kind:      RunCommandGroupType,
					},
				},
			},
			want: &Group{
				IsDefault: true,
				Kind:      RunCommandGroupType,
			},
		},
		{
			name: "Case 2: Composite command group",
			command: DevfileCommand{
				Id: "composite1",
				Composite: &Composite{
					Group: &Group{
						IsDefault: true,
						Kind:      BuildCommandGroupType,
					},
				},
			},
			want: &Group{
				IsDefault: true,
				Kind:      BuildCommandGroupType,
			},
		},
		{
			name:    "Case 3: Empty command",
			command: DevfileCommand{},
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commandGroup := tt.command.GetGroup()
			if !reflect.DeepEqual(commandGroup, tt.want) {
				t.Errorf("expected %v, actual %v", tt.want, commandGroup)
			}
		})
	}

}

func TestGetExecComponent(t *testing.T) {

	tests := []struct {
		name    string
		command DevfileCommand
		want    string
	}{
		{
			name: "Case 1: Exec component present",
			command: DevfileCommand{
				Id: "exec1",
				Exec: &Exec{
					Component: "component1",
				},
			},
			want: "component1",
		},
		{
			name: "Case 2: Exec component absent",
			command: DevfileCommand{
				Id:   "exec1",
				Exec: &Exec{},
			},
			want: "",
		},
		{
			name:    "Case 3: Empty command",
			command: DevfileCommand{},
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := tt.command.GetExecComponent()
			if component != tt.want {
				t.Errorf("expected %v, actual %v", tt.want, component)
			}
		})
	}

}

func TestGetExecCommandLine(t *testing.T) {

	tests := []struct {
		name    string
		command DevfileCommand
		want    string
	}{
		{
			name: "Case 1: Exec command line present",
			command: DevfileCommand{
				Id: "exec1",
				Exec: &Exec{
					CommandLine: "commandline1",
				},
			},
			want: "commandline1",
		},
		{
			name: "Case 2: Exec command line absent",
			command: DevfileCommand{
				Id:   "exec1",
				Exec: &Exec{},
			},
			want: "",
		},
		{
			name:    "Case 3: Empty command",
			command: DevfileCommand{},
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commandLine := tt.command.GetExecCommandLine()
			if commandLine != tt.want {
				t.Errorf("expected %v, actual %v", tt.want, commandLine)
			}
		})
	}

}

func TestGetExecWorkingDir(t *testing.T) {

	tests := []struct {
		name    string
		command DevfileCommand
		want    string
	}{
		{
			name: "Case 1: Exec working dir present",
			command: DevfileCommand{
				Id: "exec1",
				Exec: &Exec{
					WorkingDir: "workingdir1",
				},
			},
			want: "workingdir1",
		},
		{
			name: "Case 2: Exec working dir absent",
			command: DevfileCommand{
				Id:   "exec1",
				Exec: &Exec{},
			},
			want: "",
		},
		{
			name:    "Case 3: Empty command",
			command: DevfileCommand{},
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workingDir := tt.command.GetExecWorkingDir()
			if workingDir != tt.want {
				t.Errorf("expected %v, actual %v", tt.want, workingDir)
			}
		})
	}

}

func TestIsComposite(t *testing.T) {

	tests := []struct {
		name    string
		command DevfileCommand
		want    bool
	}{
		{
			name: "Case 1: Exec command",
			command: DevfileCommand{
				Id:   "exec1",
				Exec: &Exec{},
			},
			want: false,
		},
		{
			name: "Case 2: composite command",
			command: DevfileCommand{
				Id:        "comp1",
				Composite: &Composite{},
			},
			want: true,
		},
		{
			name:    "Case 3: Empty command",
			command: DevfileCommand{},
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isCompositeCmd := tt.command.IsComposite()
			if isCompositeCmd != tt.want {
				t.Errorf("expected %v, actual %v", tt.want, isCompositeCmd)
			}
		})
	}

}
