package backend

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	parsercontext "github.com/devfile/library/v2/pkg/devfile/parser/context"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	dffilesystem "github.com/devfile/library/v2/pkg/testingutil/filesystem"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/registry"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func TestFlagsBackend_SelectDevfile(t *testing.T) {
	type fields struct {
		flags map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		want    *api.DetectionResult
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
			want: &api.DetectionResult{
				Devfile:         "adevfile",
				DevfilePath:     "apath",
				DevfileRegistry: "aregistry",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &FlagsBackend{}
			ctx := context.Background()
			got, err := o.SelectDevfile(ctx, tt.fields.flags, nil, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("FlagsBackend.SelectDevfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("FlagsBackend.SelectDevfile() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFlagsBackend_Validate(t *testing.T) {
	type fields struct {
	}
	type args struct {
		flags map[string]string
		fsys  func() filesystem.Filesystem
		dir   string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		registryList []api.Registry
		wantErr      bool
	}{
		{
			name: "no name passed",
			args: args{
				flags: map[string]string{
					"name": "",
				},
				fsys: func() filesystem.Filesystem {
					fs := filesystem.NewFakeFs()
					_ = fs.MkdirAll("/tmp", 0644)
					return fs
				},
				dir: "/tmp",
			},
			wantErr: true,
		},
		{
			name: "no devfile info passed",
			args: args{
				flags: map[string]string{
					"name": "aname",
				},
				fsys: func() filesystem.Filesystem {
					fs := filesystem.NewFakeFs()
					_ = fs.MkdirAll("/tmp", 0644)
					return fs
				},
				dir: "/tmp",
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
				fsys: func() filesystem.Filesystem {
					fs := filesystem.NewFakeFs()
					_ = fs.MkdirAll("/tmp", 0644)
					return fs
				},
				dir: "/tmp",
			},
			registryList: []api.Registry{
				{
					Name: "aregistry",
				},
			},
			wantErr: false,
		},
		{
			name: "devfile and devfile-path passed",
			args: args{
				flags: map[string]string{
					"name":         "aname",
					"devfile":      "adevfile",
					"devfile-path": "apath",
				},
				fsys: func() filesystem.Filesystem {
					fs := filesystem.NewFakeFs()
					_ = fs.MkdirAll("/tmp", 0644)
					return fs
				},
				dir: "/tmp",
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
				fsys: func() filesystem.Filesystem {
					fs := filesystem.NewFakeFs()
					_ = fs.MkdirAll("/tmp", 0644)
					return fs
				},
				dir: "/tmp",
			},
			registryList: []api.Registry{
				{
					Name: "aregistry",
				},
			},
			wantErr: false,
		},
		{
			name: "devfile and devfile-registry passed with non existing registry",
			args: args{
				flags: map[string]string{
					"name":             "aname",
					"devfile":          "adevfile",
					"devfile-registry": "aregistry",
				},
				fsys: func() filesystem.Filesystem {
					fs := filesystem.NewFakeFs()
					_ = fs.MkdirAll("/tmp", 0644)
					return fs
				},
				dir: "/tmp",
			},
			wantErr: true,
		},
		{
			name: "devfile-path and devfile-registry passed",
			args: args{
				flags: map[string]string{
					"name":             "aname",
					"devfile-path":     "apath",
					"devfile-registry": "aregistry",
				},
				fsys: func() filesystem.Filesystem {
					fs := filesystem.NewFakeFs()
					_ = fs.MkdirAll("/tmp", 0644)
					return fs
				},
				dir: "/tmp",
			},
			registryList: []api.Registry{
				{
					Name: "aregistry",
				},
			},
			wantErr: true,
		},
		{
			name: "numeric name",
			args: args{
				flags: map[string]string{
					"name":    "1234",
					"devfile": "adevfile",
				},
				fsys: func() filesystem.Filesystem {
					fs := filesystem.NewFakeFs()
					_ = fs.MkdirAll("/tmp", 0644)
					return fs
				},
				dir: "/tmp",
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
				fsys: func() filesystem.Filesystem {
					fs := filesystem.NewFakeFs()
					_ = fs.MkdirAll("/tmp", 0644)
					return fs
				},
				dir: "/tmp",
			},
			wantErr: true,
		},
		{
			name: "starter flag with an empty directory",
			args: args{
				flags: map[string]string{
					"name":    "aname",
					"devfile": "adevfile",
					"starter": "astarter",
				},
				fsys: func() filesystem.Filesystem {
					fs := filesystem.NewFakeFs()
					_ = fs.MkdirAll("/tmp", 0644)
					return fs
				},
				dir: "/tmp",
			},
			wantErr: false,
		},
		{
			name: "starter flag with a non empty directory",
			args: args{
				flags: map[string]string{
					"name":    "aname",
					"devfile": "adevfile",
					"starter": "astarter",
				},
				fsys: func() filesystem.Filesystem {
					fs := filesystem.NewFakeFs()
					_ = fs.MkdirAll("/tmp", 0644)
					_ = fs.WriteFile("/tmp/main.go", []byte("package main"), 0644)
					return fs
				},
				dir: "/tmp",
			},
			wantErr: true,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			registryClient := registry.NewMockClient(ctrl)
			registryClient.EXPECT().GetDevfileRegistries(gomock.Eq(tt.args.flags[FLAG_DEVFILE_REGISTRY])).Return(tt.registryList, nil).AnyTimes()

			o := &FlagsBackend{
				registryClient: registryClient,
			}
			if err := o.Validate(tt.args.flags, tt.args.fsys(), tt.args.dir); (err != nil) != tt.wantErr {
				t.Errorf("FlagsBackend.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFlagsBackend_SelectStarterProject(t *testing.T) {
	type fields struct {
		registryClient registry.Client
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
				registryClient: tt.fields.registryClient,
			}
			got1, err := o.SelectStarterProject(tt.args.devfile(), tt.args.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("FlagsBackend.SelectStarterProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got1); diff != "" {
				t.Errorf("FlagsBackend.SelectStarterProject() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFlagsBackend_PersonalizeName(t *testing.T) {
	type fields struct {
		registryClient registry.Client
	}
	type args struct {
		devfile func(fs dffilesystem.Filesystem) parser.DevfileObj
		flags   map[string]string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		checkResult func(newName string, args args) bool
	}{
		{
			name: "name flag",
			args: args{
				devfile: func(fs dffilesystem.Filesystem) parser.DevfileObj {
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
			checkResult: func(newName string, args args) bool {
				return newName == args.flags["name"]
			},
		},
		{
			name: "invalid name flag",
			args: args{
				devfile: func(fs dffilesystem.Filesystem) parser.DevfileObj {
					devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					obj := parser.DevfileObj{
						Ctx:  parsercontext.FakeContext(fs, "/tmp/devfile.yaml"),
						Data: devfileData,
					}
					return obj
				},
				flags: map[string]string{
					"devfile": "adevfile",
					"name":    "1234",
				},
			},
			wantErr: true,
			checkResult: func(newName string, args args) bool {
				return newName == ""
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &FlagsBackend{
				registryClient: tt.fields.registryClient,
			}
			fs := dffilesystem.NewFakeFs()
			newName, err := o.PersonalizeName(tt.args.devfile(fs), tt.args.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("FlagsBackend.PersonalizeName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkResult != nil && !tt.checkResult(newName, tt.args) {
				t.Errorf("FlagsBackend.PersonalizeName(), checking result failed")
			}
		})
	}
}
