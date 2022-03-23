package backend

import (
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/init/asker"
	"github.com/redhat-developer/odo/pkg/registry"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

// Below functions are from:
// https://github.com/redhat-developer/alizer/blob/main/go/test/apis/language_recognizer_test.go
func GetTestProjectPath(folder string) string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	return filepath.Join(basepath, "..", "..", "..", "tests/examples/source/", folder)
}

var types = []registry.DevfileStack{
	{
		Name:        "java-maven",
		Language:    "java",
		ProjectType: "maven",
		Tags:        []string{"Java", "Maven"},
		Registry: registry.Registry{
			Name: "registry1",
		},
	},
	{
		Name:        "java-quarkus",
		Language:    "java",
		ProjectType: "quarkus",
		Tags:        []string{"Java", "Quarkus"},
		Registry: registry.Registry{
			Name: "registry1",
		},
	},
	{
		Name:        "java-wildfly",
		Language:    "java",
		ProjectType: "wildfly",
		Tags:        []string{"Java", "WildFly"},
		Registry: registry.Registry{
			Name: "registry2",
		},
	},
	{
		Name:        "nodejs",
		Language:    "javascript",
		ProjectType: "nodejs",
		Tags:        []string{"NodeJS", "Express", "ubi8"},
		Registry: registry.Registry{
			Name: "registry2",
		},
	},
	{
		Name:        "python",
		Language:    "python",
		ProjectType: "python",
		Tags:        []string{"Python", "pip"},
		Registry: registry.Registry{
			Name: "registry3",
		},
	},
}
var list = registry.DevfileStackList{
	Items: types,
}

func TestDetectFramework(t *testing.T) {

	type args struct {
		path string
	}
	tests := []struct {
		name           string
		args           args
		wantedDevfile  string
		wantedRegistry string
		wantErr        bool
	}{
		{
			name: "Detect Node.JS example",
			args: args{
				path: GetTestProjectPath("nodejs"),
			},
			wantedDevfile:  "nodejs",
			wantedRegistry: "registry2",
			wantErr:        false,
		},
		{
			name: "Detect java openjdk example",
			args: args{
				path: GetTestProjectPath("openjdk"),
			},
			wantedDevfile:  "java-maven",
			wantedRegistry: "registry1",
			wantErr:        false,
		},
		{
			name: "Detect python example",
			args: args{
				path: GetTestProjectPath("python"),
			},
			wantedDevfile:  "python",
			wantedRegistry: "registry3",
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			askerClient := asker.NewMockAsker(ctrl)
			registryClient := registry.NewMockClient(ctrl)
			registryClient.EXPECT().ListDevfileStacks("").Return(list, nil)
			alizerClient := NewAlizerBackend(askerClient, registryClient)
			// Run function DetectFramework
			detected, registry, err := alizerClient.detectFramework(tt.args.path)

			if !tt.wantErr == (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
				return
			}

			if detected.Name != tt.wantedDevfile {
				t.Errorf("unexpected devfile %v, wantedDevfile %v", detected, tt.wantedDevfile)
			}
			if registry.Name != tt.wantedRegistry {
				t.Errorf("unexpected registry %v, wantedRegistry %v", registry, tt.wantedRegistry)
			}
		})
	}
}

func TestAlizerBackend_SelectDevfile(t *testing.T) {
	type fields struct {
		askerClient    func(ctrl *gomock.Controller) asker.Asker
		registryClient func(ctrl *gomock.Controller) registry.Client
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
		wantLocation *DevfileLocation
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
				registryClient: func(ctrl *gomock.Controller) registry.Client {
					registryClient := registry.NewMockClient(ctrl)
					registryClient.EXPECT().ListDevfileStacks("").Return(list, nil)
					return registryClient
				},
			},
			args: args{
				fs:  filesystem.DefaultFs{},
				dir: GetTestProjectPath("nodejs"),
			},
			wantLocation: &DevfileLocation{
				Devfile:         "nodejs",
				DevfileRegistry: "registry2",
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
				registryClient: func(ctrl *gomock.Controller) registry.Client {
					registryClient := registry.NewMockClient(ctrl)
					registryClient.EXPECT().ListDevfileStacks("").Return(list, nil)
					return registryClient
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
				askerClient:    tt.fields.askerClient(ctrl),
				registryClient: tt.fields.registryClient(ctrl),
			}
			gotLocation, err := o.SelectDevfile(tt.args.flags, tt.args.fs, tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("AlizerBackend.SelectDevfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotLocation, tt.wantLocation) {
				t.Errorf("AlizerBackend.SelectDevfile() = %v, want %v", gotLocation, tt.wantLocation)
			}
		})
	}
}
