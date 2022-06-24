package validate

import (
	"testing"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"

	"github.com/redhat-developer/odo/pkg/util"
)

var buildGroup = devfilev1.BuildCommandGroupKind

func TestValidateCommand(t *testing.T) {

	tests := []struct {
		name    string
		command devfilev1.Command
		wantErr bool
	}{
		{
			name: "valid Exec Command",
			command: devfilev1.Command{
				Id: "somecommand",
				CommandUnion: devfilev1.CommandUnion{
					Exec: &devfilev1.ExecCommand{},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid odo Command",
			command: devfilev1.Command{
				Id: "invalid-odo-command",
				CommandUnion: devfilev1.CommandUnion{
					Custom: &devfilev1.CustomCommand{
						CommandClass: "cmd-class",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid Apply Command",
			command: devfilev1.Command{
				Id: "my-apply-command",
				CommandUnion: devfilev1.CommandUnion{
					Apply: &devfilev1.ApplyCommand{},
				},
			},
			wantErr: false,
		},
		{
			name: "valid Composite Command",
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
