package init

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	_init "github.com/redhat-developer/odo/pkg/init"
	"github.com/redhat-developer/odo/pkg/odo/cli/init/params"
	"github.com/redhat-developer/odo/pkg/odo/cmdline"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func TestInitOptions_Complete(t *testing.T) {
	type fields struct {
		backends func(*gomock.Controller) []params.ParamsBuilder
	}
	tests := []struct {
		name           string
		fields         fields
		cmdlineExpects func(*cmdline.MockCmdline)
		fsysPopulate   func(fsys filesystem.Filesystem)
		wantErr        bool
	}{
		{
			name: "directory not empty",
			fsysPopulate: func(fsys filesystem.Filesystem) {
				_ = fsys.WriteFile(".emptyfile", []byte(""), 0644)
			},
			wantErr: true,
		},
		{
			name: "second backend used",
			fields: fields{
				backends: func(ctrl *gomock.Controller) []params.ParamsBuilder {
					b1 := params.NewMockParamsBuilder(ctrl)
					b2 := params.NewMockParamsBuilder(ctrl)
					b1.EXPECT().IsAdequate(gomock.Any()).Return(false)
					b2.EXPECT().IsAdequate(gomock.Any()).Return(true)
					b2.EXPECT().ParamsBuild().Times(1)
					return []params.ParamsBuilder{b1, b2}
				},
			},
			cmdlineExpects: func(mock *cmdline.MockCmdline) {
				mock.EXPECT().GetFlags()
				mock.EXPECT().Context().Return(context.Background())
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := filesystem.NewFakeFs()
			if tt.fsysPopulate != nil {
				tt.fsysPopulate(fsys)
			}
			ctrl := gomock.NewController(t)
			var backends []params.ParamsBuilder
			if tt.fields.backends != nil {
				backends = tt.fields.backends(ctrl)
			}
			prefClient := preference.NewMockClient(ctrl)
			initClient := _init.NewMockClient(ctrl)
			o := NewInitOptions(backends, fsys, initClient, prefClient)

			cmdline := cmdline.NewMockCmdline(ctrl)
			if tt.cmdlineExpects != nil {
				tt.cmdlineExpects(cmdline)
			}
			if err := o.Complete(cmdline, []string{}); (err != nil) != tt.wantErr {
				t.Errorf("InitOptions.Complete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
