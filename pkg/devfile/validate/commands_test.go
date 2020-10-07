package validate

import (
	"testing"

	"github.com/openshift/odo/pkg/devfile/parser/data/common"
)

var buildGroup = common.BuildCommandGroupType
var runGroup = common.RunCommandGroupType

func TestValidateCommand(t *testing.T) {

	tests := []struct {
		name    string
		command common.DevfileCommand
		wantErr bool
	}{
		{
			name: "Case 1: Valid Exec Command",
			command: common.DevfileCommand{
				Id:   "somecommand",
				Exec: &common.Exec{},
			},
			wantErr: false,
		},
		{
			name: "Case 2: Valid Composite Command",
			command: common.DevfileCommand{
				Id: "composite1",
				Composite: &common.Composite{
					Group: &common.Group{Kind: buildGroup, IsDefault: true},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 3: Invalid Composite Command with Run Kind",
			command: common.DevfileCommand{
				Id: "composite1",
				Composite: &common.Composite{
					Group: &common.Group{Kind: runGroup, IsDefault: true},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCommand(tt.command)
			if !tt.wantErr == (err != nil) {
				t.Errorf("TestValidateCommand unexpected error: %v", err)
				return
			}
		})
	}

}

func TestValidateCompositeCommand(t *testing.T) {

	tests := []struct {
		name    string
		command common.DevfileCommand
		wantErr bool
	}{
		{
			name: "Case 1: Valid Composite Command",
			command: common.DevfileCommand{

				Id: "command1",
				Composite: &common.Composite{
					Group: &common.Group{Kind: buildGroup},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 2: Invalid Composite Run Kind Command",
			command: common.DevfileCommand{
				Id: "command1",
				Composite: &common.Composite{
					Group: &common.Group{Kind: runGroup},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCompositeCommand(tt.command)
			if !tt.wantErr == (err != nil) {
				t.Errorf("TestValidateCompositeCommand unexpected error: %v", err)
			}
		})
	}
}
