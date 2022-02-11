package backend

import (
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	parsercontext "github.com/devfile/library/pkg/devfile/parser/context"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/devfile/library/pkg/testingutil/filesystem"

	"github.com/redhat-developer/odo/pkg/preference"
)

func TestFlagsBackend_SelectDevfile(t *testing.T) {
	type fields struct {
		flags map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *DevfileLocation
		wantErr bool
	}{
		{
			name: "all fields defined",
			fields: fields{
				flags: map[string]string{
					FLAG_DEVFILE:          "adevfile",
					FLAG_DEVFILE_PATH:     "apath",
					FLAG_DEVFILE_REGISTRY: "aregistry",
				},
			},
			wantErr: false,
			want: &DevfileLocation{
				Devfile:         "adevfile",
				DevfilePath:     "apath",
				DevfileRegistry: "aregistry",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &FlagsBackend{}
			got, err := o.SelectDevfile(tt.fields.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("FlagsBackend.SelectDevfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FlagsBackend.SelectDevfile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlagsBackend_Validate(t *testing.T) {
	type fields struct {
	}
	type args struct {
		flags map[string]string
	}
	tests := []struct {
		name               string
		fields             fields
		args               args
		registryNameExists bool
		registryList       []preference.Registry
		wantErr            bool
	}{
		{
			name: "no name passed",
			args: args{
				flags: map[string]string{
					"name": "",
				},
			},
			wantErr: true,
		},
		{
			name: "no devfile info passed",
			args: args{
				flags: map[string]string{
					"name": "aname",
				},
			},
			wantErr: true,
		},
		{
			name: "devfile passed with a single registry",
			args: args{
				flags: map[string]string{
					"name":    "aname",
					"devfile": "adevfile",
				},
			},
			registryList: []preference.Registry{
				{
					Name: "aregistry",
				},
			},
			registryNameExists: true,
			wantErr:            false,
		},
		{
			name: "devfile and devfile-path passed",
			args: args{
				flags: map[string]string{
					"name":         "aname",
					"devfile":      "adevfile",
					"devfile-path": "apath",
				},
			},
			wantErr: true,
		},
		{
			name: "devfile and devfile-registry passed",
			args: args{
				flags: map[string]string{
					"name":             "aname",
					"devfile":          "adevfile",
					"devfile-registry": "aregistry",
				},
			},
			registryNameExists: true,
			wantErr:            false,
		},
		{
			name: "devfile and devfile-registry passed with non existing registry",
			args: args{
				flags: map[string]string{
					"name":             "aname",
					"devfile":          "adevfile",
					"devfile-registry": "aregistry",
				},
			},
			registryNameExists: false,
			wantErr:            true,
		},
		{
			name: "devfile-path and devfile-registry passed",
			args: args{
				flags: map[string]string{
					"name":             "aname",
					"devfile-path":     "apath",
					"devfile-registry": "aregistry",
				},
			},
			registryNameExists: true,
			wantErr:            true,
		},
		{
			name: "numeric name",
			args: args{
				flags: map[string]string{
					"name":    "1234",
					"devfile": "adevfile",
				},
			},
			wantErr: true,
		},
		{
			name: "non DNS name",
			args: args{
				flags: map[string]string{
					"name":    "WrongName",
					"devfile": "adevfile",
				},
			},
			wantErr: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			prefClient := preference.NewMockClient(ctrl)
			prefClient.EXPECT().RegistryNameExists(gomock.Any()).Return(tt.registryNameExists).AnyTimes()
			prefClient.EXPECT().RegistryList().Return(&tt.registryList).AnyTimes()

			o := &FlagsBackend{
				preferenceClient: prefClient,
			}
			if err := o.Validate(tt.args.flags); (err != nil) != tt.wantErr {
				t.Errorf("FlagsBackend.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFlagsBackend_SelectStarterProject(t *testing.T) {
	type fields struct {
		preferenceClient preference.Client
	}
	type args struct {
		devfile func() parser.DevfileObj
		flags   map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *v1alpha2.StarterProject
		wantErr bool
	}{
		{
			name: "some flags, but not starter",
			args: args{
				devfile: func() parser.DevfileObj {
					return parser.DevfileObj{}
				},
				flags: map[string]string{
					"devfile": "adevfile",
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "starter flag defined and starter exists",
			args: args{
				devfile: func() parser.DevfileObj {
					devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = devfileData.AddStarterProjects([]v1alpha2.StarterProject{
						{
							Name: "starter1",
						},
						{
							Name: "starter2",
						},
						{
							Name: "starter3",
						},
					})
					return parser.DevfileObj{
						Data: devfileData,
					}
				},
				flags: map[string]string{
					"devfile": "adevfile",
					"starter": "starter2",
				},
			},
			want: &v1alpha2.StarterProject{
				Name: "starter2",
			},
			wantErr: false,
		},
		{
			name: "starter flag defined and starter does not exist",
			args: args{
				devfile: func() parser.DevfileObj {
					devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = devfileData.AddStarterProjects([]v1alpha2.StarterProject{
						{
							Name: "starter1",
						},
						{
							Name: "starter2",
						},
						{
							Name: "starter3",
						},
					})
					return parser.DevfileObj{
						Data: devfileData,
					}
				},
				flags: map[string]string{
					"devfile": "adevfile",
					"starter": "starter4",
				},
			},
			want:    nil,
			wantErr: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &FlagsBackend{
				preferenceClient: tt.fields.preferenceClient,
			}
			got1, err := o.SelectStarterProject(tt.args.devfile(), tt.args.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("FlagsBackend.SelectStarterProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got1, tt.want) {
				t.Errorf("FlagsBackend.SelectStarterProject() got1 = %v, want %v", got1, tt.want)
			}
		})
	}
}

func TestFlagsBackend_PersonalizeName(t *testing.T) {
	type fields struct {
		preferenceClient preference.Client
	}
	type args struct {
		devfile func(fs filesystem.Filesystem) parser.DevfileObj
		flags   map[string]string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		checkResult func(devfile parser.DevfileObj, args args) bool
	}{
		{
			name: "name flag",
			args: args{
				devfile: func(fs filesystem.Filesystem) parser.DevfileObj {
					devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					obj := parser.DevfileObj{
						Ctx:  parsercontext.FakeContext(fs, "/tmp/devfile.yaml"),
						Data: devfileData,
					}
					return obj
				},
				flags: map[string]string{
					"devfile": "adevfile",
					"name":    "a-name",
				},
			},
			wantErr: false,
			checkResult: func(devfile parser.DevfileObj, args args) bool {
				return devfile.GetMetadataName() == args.flags["name"]
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &FlagsBackend{
				preferenceClient: tt.fields.preferenceClient,
			}
			fs := filesystem.NewFakeFs()
			devfile := tt.args.devfile(fs)
			err := o.PersonalizeName(devfile, tt.args.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("FlagsBackend.PersonalizeName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkResult != nil && !tt.checkResult(devfile, tt.args) {
				t.Errorf("FlagsBackend.PersonalizeName(), checking result failed")
			}
		})
	}
}
