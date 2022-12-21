package backend

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/redhat-developer/alizer/go/pkg/apis/model"
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
		wantLocation *api.DetectionResult
		wantErr      bool
	}{
		{
			name: "error while trying to detect devfile",
			fields: fields{
				askerClient: func(ctrl *gomock.Controller) asker.Asker {
					askerClient := asker.NewMockAsker(ctrl)
					askerClient.EXPECT().AskCorrect().Return(true, nil).Times(0)
					return askerClient
				},
				alizerClient: func(ctrl *gomock.Controller) alizer.Client {
					alizerClient := alizer.NewMockClient(ctrl)
					alizerClient.EXPECT().DetectFramework(gomock.Any(), gomock.Any()).Return(model.DevFileType{
						Name: "a-devfile-name",
					}, "1.0.0", api.Registry{
						Name: "a-registry",
					}, errors.New("unable to detect framework"))
					return alizerClient
				},
			},
			args: args{
				fs:  filesystem.DefaultFs{},
				dir: GetTestProjectPath("nodejs"),
			},
			wantErr: true,
		},
		{
			name: "error while trying to detect ports",
			fields: fields{
				askerClient: func(ctrl *gomock.Controller) asker.Asker {
					askerClient := asker.NewMockAsker(ctrl)
					askerClient.EXPECT().AskCorrect().Return(true, nil).Times(0)
					return askerClient
				},
				alizerClient: func(ctrl *gomock.Controller) alizer.Client {
					alizerClient := alizer.NewMockClient(ctrl)
					alizerClient.EXPECT().DetectFramework(gomock.Any(), gomock.Any()).Return(model.DevFileType{
						Name: "a-devfile-name",
					}, "1.0.0", api.Registry{
						Name: "a-registry",
					}, nil)
					alizerClient.EXPECT().DetectPorts(gomock.Any()).Return(nil, errors.New("unable to detect ports"))
					return alizerClient
				},
			},
			args: args{
				fs:  filesystem.DefaultFs{},
				dir: GetTestProjectPath("nodejs"),
			},
			wantErr: true,
		},
		{
			name: "error while asking consent to user",
			fields: fields{
				askerClient: func(ctrl *gomock.Controller) asker.Asker {
					askerClient := asker.NewMockAsker(ctrl)
					askerClient.EXPECT().AskCorrect().Return(false, errors.New("error while prompting user"))
					return askerClient
				},
				alizerClient: func(ctrl *gomock.Controller) alizer.Client {
					alizerClient := alizer.NewMockClient(ctrl)
					alizerClient.EXPECT().DetectFramework(gomock.Any(), gomock.Any()).Return(model.DevFileType{
						Name: "a-devfile-name",
					}, "1.0.0", api.Registry{
						Name: "a-registry",
					}, nil)
					alizerClient.EXPECT().DetectPorts(gomock.Any()).Return(nil, nil)
					return alizerClient
				},
			},
			args: args{
				fs:  filesystem.DefaultFs{},
				dir: GetTestProjectPath("nodejs"),
			},
			wantErr: true,
		},
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
					alizerClient.EXPECT().DetectFramework(gomock.Any(), gomock.Any()).Return(model.DevFileType{
						Name: "a-devfile-name",
					}, "1.0.0", api.Registry{
						Name: "a-registry",
					}, nil)
					alizerClient.EXPECT().DetectPorts(gomock.Any()).Return(nil, nil)
					return alizerClient
				},
			},
			args: args{
				fs:  filesystem.DefaultFs{},
				dir: GetTestProjectPath("nodejs"),
			},
			wantLocation: &api.DetectionResult{
				Devfile:         "a-devfile-name",
				DevfileRegistry: "a-registry",
				DevfileVersion:  "1.0.0",
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
					alizerClient.EXPECT().DetectFramework(gomock.Any(), gomock.Any()).Return(model.DevFileType{}, "", api.Registry{}, nil)
					alizerClient.EXPECT().DetectPorts(gomock.Any()).Return(nil, nil)
					return alizerClient
				},
			},
			args: args{
				fs:  filesystem.DefaultFs{},
				dir: GetTestProjectPath("nodejs"),
			},
			wantLocation: nil,
		},
		{
			name: "devfile and ports detected and accepted",
			fields: fields{
				askerClient: func(ctrl *gomock.Controller) asker.Asker {
					askerClient := asker.NewMockAsker(ctrl)
					askerClient.EXPECT().AskCorrect().Return(true, nil)
					return askerClient
				},
				alizerClient: func(ctrl *gomock.Controller) alizer.Client {
					alizerClient := alizer.NewMockClient(ctrl)
					alizerClient.EXPECT().DetectFramework(gomock.Any(), gomock.Any()).Return(model.DevFileType{
						Name: "a-devfile-name",
					}, "1.0.0", api.Registry{
						Name: "a-registry",
					}, nil)
					alizerClient.EXPECT().DetectPorts(gomock.Any()).Return([]int{1234, 5678}, nil)
					return alizerClient
				},
			},
			args: args{
				fs:  filesystem.DefaultFs{},
				dir: GetTestProjectPath("nodejs"),
			},
			wantLocation: &api.DetectionResult{
				Devfile:          "a-devfile-name",
				DevfileRegistry:  "a-registry",
				ApplicationPorts: []int{1234, 5678},
				DevfileVersion:   "1.0.0",
			},
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
