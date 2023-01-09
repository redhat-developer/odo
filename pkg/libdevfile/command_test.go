package libdevfile

import (
	"fmt"
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/redhat-developer/odo/pkg/libdevfile/generator"
	"k8s.io/utils/pointer"
)

func Test_newCommand(t *testing.T) {

	execCommand := generator.GetExecCommand(generator.ExecCommandParams{
		Kind:      v1alpha2.RunCommandGroupKind,
		Id:        "exec-command",
		IsDefault: pointer.BoolPtr(true),
	})
	compositeCommand := generator.GetCompositeCommand(generator.CompositeCommandParams{
		Kind:      v1alpha2.DeployCommandGroupKind,
		Id:        "composite-command",
		IsDefault: pointer.BoolPtr(true),
	})
	applyCommand := generator.GetApplyCommand(generator.ApplyCommandParams{
		Kind:      v1alpha2.DeployCommandGroupKind,
		Id:        "apply-command",
		IsDefault: pointer.BoolPtr(false),
	})

	data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
	_ = data.AddCommands([]v1alpha2.Command{execCommand, compositeCommand, applyCommand})
	devfileObj := parser.DevfileObj{
		Data: data,
	}

	type args struct {
		devfileObj parser.DevfileObj
		devfileCmd v1alpha2.Command
	}
	tests := []struct {
		name     string
		args     args
		wantType string
		wantErr  bool
	}{
		{
			name: "exec command",
			args: args{
				devfileObj: devfileObj,
				devfileCmd: execCommand,
			},
			wantType: "*libdevfile.execCommand",
		},
		{
			name: "composite command",
			args: args{
				devfileObj: devfileObj,
				devfileCmd: compositeCommand,
			},
			wantType: "*libdevfile.compositeCommand",
		},
		{
			name: "apply command",
			args: args{
				devfileObj: devfileObj,
				devfileCmd: applyCommand,
			},
			wantType: "*libdevfile.applyCommand",
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newCommand(tt.args.devfileObj, tt.args.devfileCmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("newCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			gotType := fmt.Sprintf("%T", got)
			if gotType != tt.wantType {
				t.Errorf("newCommand() type = %v, want %v", got, tt.wantType)
			}
		})
	}
}
