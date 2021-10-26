package validate

import (
	"testing"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/openshift/odo/v2/pkg/util"
)

var buildGroup = devfilev1.BuildCommandGroupKind
var runGroup = devfilev1.RunCommandGroupKind

func TestValidateCommand(t *testing.T) {

	tests := []struct {
		name    string
		command devfilev1.Command
		wantErr bool
	}{
		{
			name: "Case 1: Valid Exec Command",
			command: devfilev1.Command{
				Id: "somecommand",
				CommandUnion: devfilev1.CommandUnion{
					Exec: &devfilev1.ExecCommand{},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 2: Valid Composite Command",
			command: devfilev1.Command{
				Id: "composite1",
				CommandUnion: devfilev1.CommandUnion{
					Composite: &devfilev1.CompositeCommand{
						LabeledCommand: devfilev1.LabeledCommand{
							BaseCommand: devfilev1.BaseCommand{
								Group: &devfilev1.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(true)},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 3: Invalid Composite Command with Run Kind",
			command: devfilev1.Command{
				Id: "composite1",
				CommandUnion: devfilev1.CommandUnion{
					Composite: &devfilev1.CompositeCommand{
						LabeledCommand: devfilev1.LabeledCommand{
							BaseCommand: devfilev1.BaseCommand{
								Group: &devfilev1.CommandGroup{Kind: runGroup, IsDefault: util.GetBoolPtr(true)},
							},
						},
					},
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
		command devfilev1.Command
		wantErr bool
	}{
		{
			name: "Case 1: Valid Composite Command",
			command: devfilev1.Command{

				Id: "command1",
				CommandUnion: devfilev1.CommandUnion{
					Composite: &devfilev1.CompositeCommand{
						LabeledCommand: devfilev1.LabeledCommand{
							BaseCommand: devfilev1.BaseCommand{
								Group: &devfilev1.CommandGroup{Kind: buildGroup},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 2: Invalid Composite Run Kind Command",
			command: devfilev1.Command{
				Id: "command1",
				CommandUnion: devfilev1.CommandUnion{
					Composite: &devfilev1.CompositeCommand{
						LabeledCommand: devfilev1.LabeledCommand{
							BaseCommand: devfilev1.BaseCommand{
								Group: &devfilev1.CommandGroup{Kind: runGroup},
							},
						},
					},
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
