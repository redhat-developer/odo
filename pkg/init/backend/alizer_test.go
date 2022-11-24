package backend

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/redhat-developer/alizer/go/pkg/apis/recognizer"
	"github.com/redhat-developer/odo/pkg/alizer"
	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/init/asker"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

// Below functions are from:
// https://github.com/redhat-developer/alizer/blob/main/go/test/apis/language_recognizer_test.go
func GetTestProjectPath(folder string) string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	return filepath.Join(basepath, "..", "..", "..", "tests/examples/source/", folder)
}

func TestAlizerBackend_SelectDevfile(t *testing.T) {
	type fields struct {
		askerClient  func(ctrl *gomock.Controller) asker.Asker
		alizerClient func(ctrl *gomock.Controller) alizer.Client
	}
	type args struct {
		flags map[string]string
		fs    filesystem.Filesystem
		dir   string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantLocation *api.DevfileLocation
		wantErr      bool
	}{
		{
			name: "devfile found and accepted",
			fields: fields{
				askerClient: func(ctrl *gomock.Controller) asker.Asker {
					askerClient := asker.NewMockAsker(ctrl)
					askerClient.EXPECT().AskCorrect().Return(true, nil)
					return askerClient
				},
				alizerClient: func(ctrl *gomock.Controller) alizer.Client {
					alizerClient := alizer.NewMockClient(ctrl)
					alizerClient.EXPECT().DetectFramework(gomock.Any(), gomock.Any()).Return(recognizer.DevFileType{
						Name: "a-devfile-name",
					}, api.Registry{
						Name: "a-registry",
					}, nil)
					return alizerClient
				},
			},
			args: args{
				fs:  filesystem.DefaultFs{},
				dir: GetTestProjectPath("nodejs"),
			},
			wantLocation: &api.DevfileLocation{
				Devfile:         "a-devfile-name",
				DevfileRegistry: "a-registry",
			},
		},
		{
			name: "devfile found but not accepted",
			fields: fields{
				askerClient: func(ctrl *gomock.Controller) asker.Asker {
					askerClient := asker.NewMockAsker(ctrl)
					askerClient.EXPECT().AskCorrect().Return(false, nil)
					return askerClient
				},
				alizerClient: func(ctrl *gomock.Controller) alizer.Client {
					alizerClient := alizer.NewMockClient(ctrl)
					alizerClient.EXPECT().DetectFramework(gomock.Any(), gomock.Any()).Return(recognizer.DevFileType{}, api.Registry{}, nil)
					return alizerClient
				},
			},
			args: args{
				fs:  filesystem.DefaultFs{},
				dir: GetTestProjectPath("nodejs"),
			},
			wantLocation: nil,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			o := &AlizerBackend{
				askerClient:  tt.fields.askerClient(ctrl),
				alizerClient: tt.fields.alizerClient(ctrl),
			}
			ctx := context.Background()
			gotLocation, err := o.SelectDevfile(ctx, tt.args.flags, tt.args.fs, tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("AlizerBackend.SelectDevfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.wantLocation, gotLocation); diff != "" {
				t.Errorf("AlizerBackend.SelectDevfile() wantLocation mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
