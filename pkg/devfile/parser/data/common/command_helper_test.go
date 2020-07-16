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
				Exec: &Exec{
					Id: "exec1",
				},
			},
			want: "exec1",
		},
		{
			name: "Case 2: Composite command ID",
			command: DevfileCommand{
				Composite: &Composite{
					Id: "composite1",
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
				Exec: &Exec{
					Id: "exec1",
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
				Composite: &Composite{
					Id: "composite1",
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
