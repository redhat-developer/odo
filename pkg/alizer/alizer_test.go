package alizer

import (
	"context"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/registry"
)

// Below functions are from:
// https://github.com/redhat-developer/alizer/blob/main/go/test/apis/language_recognizer_test.go
func GetTestProjectPath(folder string) string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	return filepath.Join(basepath, "..", "..", "tests/examples/source/", folder)
}

var types = []api.DevfileStack{
	{
		Name:        "java-maven",
		Language:    "java",
		ProjectType: "maven",
		Tags:        []string{"Java", "Maven"},
		Registry: api.Registry{
			Name: "registry1",
		},
	},
	{
		Name:        "java-quarkus",
		Language:    "java",
		ProjectType: "quarkus",
		Tags:        []string{"Java", "Quarkus"},
		Registry: api.Registry{
			Name: "registry1",
		},
	},
	{
		Name:        "java-wildfly",
		Language:    "java",
		ProjectType: "wildfly",
		Tags:        []string{"Java", "WildFly"},
		Registry: api.Registry{
			Name: "registry2",
		},
	},
	{
		Name:        "nodejs",
		Language:    "JavaScript",
		ProjectType: "nodejs",
		Tags:        []string{"NodeJS", "Express", "ubi8"},
		Registry: api.Registry{
			Name: "registry2",
		},
	},
	{
		Name:        "python",
		Language:    "python",
		ProjectType: "python",
		Tags:        []string{"Python", "pip"},
		Registry: api.Registry{
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
			registryClient := registry.NewMockClient(ctrl)
			ctx := context.Background()
			registryClient.EXPECT().ListDevfileStacks(ctx, "", "", "", false, false).Return(list, nil)
			alizerClient := NewAlizerClient(registryClient)
			// Run function DetectFramework

			detected, err := alizerClient.DetectFramework(ctx, tt.args.path)
			if !tt.wantErr == (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
				return
			}

			if detected.Type.Name != tt.wantedDevfile {
				t.Errorf("unexpected devfile %v, wantedDevfile %v", detected, tt.wantedDevfile)
			}
			if detected.Registry.Name != tt.wantedRegistry {
				t.Errorf("unexpected registry %v, wantedRegistry %v", detected.Registry, tt.wantedRegistry)
			}
		})
	}
}

func TestDetectName(t *testing.T) {

	type args struct {
		path string
	}
	tests := []struct {
		name       string
		args       args
		wantedName string
		wantErr    bool
	}{
		{
			name: "Case 1: Detect Node.JS name through package.json",
			args: args{
				path: GetTestProjectPath("nodejs"),
			},
			wantedName: "node-echo",
			wantErr:    false,
		},
		{
			// NOTE
			// Alizer does NOT support Python yet, so this test is expected to fail once Python support
			// is implemented
			name: "Case 2: Detect Python name through DIRECTORY name",
			args: args{
				path: GetTestProjectPath("python"),
			},
			// Directory name is 'python' so expect that name to be returned
			wantedName: "python",
			wantErr:    false,
		},
		{

			// NOTE
			// Returns "insultapp" instead of "InsultApp" as it does DNS1123 sanitization
			// See DetectName function
			name: "Case 3: Detect Java name through pom.xml",
			args: args{
				path: GetTestProjectPath("wildfly"),
			},
			wantedName: "insultapp",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			registryClient := registry.NewMockClient(ctrl)
			alizerClient := NewAlizerClient(registryClient)

			name, err := alizerClient.DetectName(tt.args.path)

			if !tt.wantErr == (err != nil) {
				t.Errorf("unexpected error %v, wantErr %v", err, tt.wantErr)
				return
			}

			if name != tt.wantedName {
				t.Errorf("unexpected name %q, wanted: %q", name, tt.wantedName)
			}
		})
	}
}

func TestAlizer_DetectPorts(t *testing.T) {
	type fields struct {
		registryClient registry.Client
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []int
		wantErr bool
	}{
		{
			name: "Detect Node.JS example",
			args: args{
				path: GetTestProjectPath("nodejs"),
			},
			want:    []int{8080},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &Alizer{
				registryClient: tt.fields.registryClient,
			}
			got, err := o.DetectPorts(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Alizer.DetectPorts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Alizer.DetectPorts() = %v, want %v", got, tt.want)
			}
		})
	}
}
