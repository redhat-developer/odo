package sub

import (
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"

	"github.com/google/go-cmp/cmp"
)

func newCommand(group string, id string) v1alpha2.Command {
	return v1alpha2.Command{
		Id: id,
		CommandUnion: v1alpha2.CommandUnion{
			Exec: &v1alpha2.ExecCommand{
				LabeledCommand: v1alpha2.LabeledCommand{
					BaseCommand: v1alpha2.BaseCommand{
						Group: &v1alpha2.CommandGroup{
							Kind: v1alpha2.CommandGroupKind(group),
						},
					},
				},
			},
		},
	}
}

func Test_moveCommandSub(t *testing.T) {
	type args struct {
		commands      []v1alpha2.Command
		previousGroup string
		newGroup      string
		previousIndex int
		newIndex      int
	}
	tests := []struct {
		name    string
		args    args
		want    map[string][]v1alpha2.Command
		wantErr bool
	}{
		{
			name: "Move from run to test",
			args: args{
				commands: []v1alpha2.Command{
					newCommand("build", "build1"),
					newCommand("run", "runToTest"),
					newCommand("", "other1"),
				},
				previousGroup: "run",
				previousIndex: 0,
				newGroup:      "test",
				newIndex:      0,
			},
			want: map[string][]v1alpha2.Command{
				"build": {
					newCommand("build", "build1"),
				},
				"run": {},
				"test": {
					newCommand("test", "runToTest"),
				},
				"": {
					newCommand("", "other1"),
				},
			},
		},
		{
			name: "Move from other to build",
			args: args{
				commands: []v1alpha2.Command{
					newCommand("build", "build1"),
					newCommand("run", "run"),
					newCommand("other", "otherToBuild"),
				},
				previousGroup: "other",
				previousIndex: 0,
				newGroup:      "build",
				newIndex:      1,
			},
			want: map[string][]v1alpha2.Command{
				"build": {
					newCommand("build", "build1"),
					newCommand("build", "otherToBuild"),
				},
				"run": {
					newCommand("run", "run"),
				},
				"other": {},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MoveCommand(tt.args.commands, tt.args.previousGroup, tt.args.newGroup, tt.args.previousIndex, tt.args.newIndex)
			if (err != nil) != tt.wantErr {
				t.Errorf("moveCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("moveCommand() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
