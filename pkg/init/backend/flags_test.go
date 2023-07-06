package backend

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/utils/pointer"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	parsercontext "github.com/devfile/library/v2/pkg/devfile/parser/context"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	dffilesystem "github.com/devfile/library/v2/pkg/testingutil/filesystem"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/registry"
	"github.com/redhat-developer/odo/pkg/testingutil"
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

func TestFlagsBackend_HandleApplicationPorts(t *testing.T) {
	type devfileProvider func(fs dffilesystem.Filesystem) (parser.DevfileObj, error)

	zeroDevfileProvider := func(fs dffilesystem.Filesystem) (parser.DevfileObj, error) {
		return parser.DevfileObj{}, nil
	}
	fakeDevfileProvider := func(fs dffilesystem.Filesystem) (parser.DevfileObj, error) {
		devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion220))
		obj := parser.DevfileObj{
			Ctx:  parsercontext.FakeContext(fs, "/tmp/devfile.yaml"),
			Data: devfileData,
		}
		return obj, nil
	}
	type fields struct {
		registryClient registry.Client
	}
	type args struct {
		devfileObjProvider devfileProvider
		flags              map[string]string
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantProvider devfileProvider
		wantErr      bool
	}{
		{
			name: "no run-port flag",
			args: args{
				devfileObjProvider: fakeDevfileProvider,
				flags: map[string]string{
					"opt1":    "val1",
					FLAG_NAME: "my-name",
				},
			},
			wantProvider: fakeDevfileProvider,
		},
		{
			name: "flag string value not enclosed within []",
			args: args{
				devfileObjProvider: fakeDevfileProvider,
				flags: map[string]string{
					FLAG_NAME:     "my-name",
					FLAG_RUN_PORT: "aaa,bbb",
				},
			},
			wantProvider: fakeDevfileProvider,
		},
		{
			name: "invalid port type",
			args: args{
				devfileObjProvider: fakeDevfileProvider,
				flags: map[string]string{
					FLAG_NAME:     "my-name",
					FLAG_RUN_PORT: "[8080,abcde]",
				},
			},
			wantErr:      true,
			wantProvider: zeroDevfileProvider,
		},
		{
			name: "devfile with no command, but --run-port set",
			args: args{
				devfileObjProvider: func(fs dffilesystem.Filesystem) (parser.DevfileObj, error) {
					devfileObj, err := fakeDevfileProvider(fs)
					if err != nil {
						return parser.DevfileObj{}, err
					}
					err = devfileObj.Data.AddComponents([]v1alpha2.Component{
						testingutil.GetFakeContainerComponent("my-cont1", 1234, 2345),
						testingutil.GetFakeContainerComponent("my-cont2", 4321, 5432),
					})
					if err != nil {
						return parser.DevfileObj{}, err
					}
					return devfileObj, nil
				},
				flags: map[string]string{
					FLAG_NAME:     "my-name",
					FLAG_RUN_PORT: "[8080,8081]",
				},
			},
			wantProvider: func(fs dffilesystem.Filesystem) (parser.DevfileObj, error) {
				devfileObj, err := fakeDevfileProvider(fs)
				if err != nil {
					return parser.DevfileObj{}, err
				}
				err = devfileObj.Data.AddComponents([]v1alpha2.Component{
					testingutil.GetFakeContainerComponent("my-cont1", 1234, 2345),
					testingutil.GetFakeContainerComponent("my-cont2", 4321, 5432),
				})
				if err != nil {
					return parser.DevfileObj{}, err
				}
				return devfileObj, nil
			},
		},
		{
			name: "devfile with more than one default run commands, --run-port set",
			args: args{
				devfileObjProvider: func(fs dffilesystem.Filesystem) (parser.DevfileObj, error) {
					devfileObj, err := fakeDevfileProvider(fs)
					if err != nil {
						return parser.DevfileObj{}, err
					}
					err = devfileObj.Data.AddComponents([]v1alpha2.Component{
						testingutil.GetFakeContainerComponent("my-cont1", 1234, 2345),
						testingutil.GetFakeContainerComponent("my-cont2", 4321, 5432),
					})
					if err != nil {
						return parser.DevfileObj{}, err
					}
					err = devfileObj.Data.AddCommands([]v1alpha2.Command{
						{
							Id: "devrun1",
							CommandUnion: v1alpha2.CommandUnion{
								Exec: &v1alpha2.ExecCommand{
									LabeledCommand: v1alpha2.LabeledCommand{
										BaseCommand: v1alpha2.BaseCommand{
											Group: &v1alpha2.CommandGroup{
												Kind:      v1alpha2.RunCommandGroupKind,
												IsDefault: pointer.Bool(true),
											},
										},
									},
								},
							},
						},
						{
							Id: "devrun2",
							CommandUnion: v1alpha2.CommandUnion{
								Exec: &v1alpha2.ExecCommand{
									LabeledCommand: v1alpha2.LabeledCommand{
										BaseCommand: v1alpha2.BaseCommand{
											Group: &v1alpha2.CommandGroup{
												Kind:      v1alpha2.RunCommandGroupKind,
												IsDefault: pointer.Bool(true),
											},
										},
									},
								},
							},
						},
					})
					if err != nil {
						return parser.DevfileObj{}, err
					}
					return devfileObj, nil
				},
				flags: map[string]string{
					FLAG_NAME:     "my-name",
					FLAG_RUN_PORT: "[8080,8081]",
				},
			},
			wantErr:      true,
			wantProvider: zeroDevfileProvider,
		},
		{
			name: "devfile with more than one non-default run commands, --run-port set",
			args: args{
				devfileObjProvider: func(fs dffilesystem.Filesystem) (parser.DevfileObj, error) {
					devfileObj, err := fakeDevfileProvider(fs)
					if err != nil {
						return parser.DevfileObj{}, err
					}
					err = devfileObj.Data.AddComponents([]v1alpha2.Component{
						testingutil.GetFakeContainerComponent("my-cont1", 1234, 2345),
						testingutil.GetFakeContainerComponent("my-cont2", 4321, 5432),
					})
					if err != nil {
						return parser.DevfileObj{}, err
					}
					err = devfileObj.Data.AddCommands([]v1alpha2.Command{
						{
							Id: "devrun1",
							CommandUnion: v1alpha2.CommandUnion{
								Exec: &v1alpha2.ExecCommand{
									LabeledCommand: v1alpha2.LabeledCommand{
										BaseCommand: v1alpha2.BaseCommand{
											Group: &v1alpha2.CommandGroup{
												Kind:      v1alpha2.RunCommandGroupKind,
												IsDefault: pointer.Bool(false),
											},
										},
									},
								},
							},
						},
						{
							Id: "devrun2",
							CommandUnion: v1alpha2.CommandUnion{
								Exec: &v1alpha2.ExecCommand{
									LabeledCommand: v1alpha2.LabeledCommand{
										BaseCommand: v1alpha2.BaseCommand{
											Group: &v1alpha2.CommandGroup{
												Kind:      v1alpha2.RunCommandGroupKind,
												IsDefault: pointer.Bool(false),
											},
										},
									},
								},
							},
						},
					})
					if err != nil {
						return parser.DevfileObj{}, err
					}
					return devfileObj, nil
				},
				flags: map[string]string{
					FLAG_NAME:     "my-name",
					FLAG_RUN_PORT: "[8080,8081]",
				},
			},
			wantProvider: func(fs dffilesystem.Filesystem) (parser.DevfileObj, error) {
				devfileObj, err := fakeDevfileProvider(fs)
				if err != nil {
					return parser.DevfileObj{}, err
				}
				err = devfileObj.Data.AddComponents([]v1alpha2.Component{
					testingutil.GetFakeContainerComponent("my-cont1", 1234, 2345),
					testingutil.GetFakeContainerComponent("my-cont2", 4321, 5432),
				})
				if err != nil {
					return parser.DevfileObj{}, err
				}
				err = devfileObj.Data.AddCommands([]v1alpha2.Command{
					{
						Id: "devrun1",
						CommandUnion: v1alpha2.CommandUnion{
							Exec: &v1alpha2.ExecCommand{
								LabeledCommand: v1alpha2.LabeledCommand{
									BaseCommand: v1alpha2.BaseCommand{
										Group: &v1alpha2.CommandGroup{
											Kind:      v1alpha2.RunCommandGroupKind,
											IsDefault: pointer.Bool(false),
										},
									},
								},
							},
						},
					},
					{
						Id: "devrun2",
						CommandUnion: v1alpha2.CommandUnion{
							Exec: &v1alpha2.ExecCommand{
								LabeledCommand: v1alpha2.LabeledCommand{
									BaseCommand: v1alpha2.BaseCommand{
										Group: &v1alpha2.CommandGroup{
											Kind:      v1alpha2.RunCommandGroupKind,
											IsDefault: pointer.Bool(false),
										},
									},
								},
							},
						},
					},
				})
				if err != nil {
					return parser.DevfileObj{}, err
				}
				return devfileObj, nil
			},
		},
		{
			name: "devfile with no run command, --run-port set",
			args: args{
				devfileObjProvider: func(fs dffilesystem.Filesystem) (parser.DevfileObj, error) {
					devfileObj, err := fakeDevfileProvider(fs)
					if err != nil {
						return parser.DevfileObj{}, err
					}
					err = devfileObj.Data.AddComponents([]v1alpha2.Component{
						testingutil.GetFakeContainerComponent("my-cont1", 1234, 2345),
						testingutil.GetFakeContainerComponent("my-cont2", 4321, 5432),
					})
					if err != nil {
						return parser.DevfileObj{}, err
					}
					err = devfileObj.Data.AddCommands([]v1alpha2.Command{
						{
							Id: "devdebug",
							CommandUnion: v1alpha2.CommandUnion{
								Exec: &v1alpha2.ExecCommand{
									LabeledCommand: v1alpha2.LabeledCommand{
										BaseCommand: v1alpha2.BaseCommand{
											Group: &v1alpha2.CommandGroup{Kind: v1alpha2.DebugCommandGroupKind},
										},
									},
								},
							},
						},
					})
					if err != nil {
						return parser.DevfileObj{}, err
					}
					return devfileObj, nil
				},
				flags: map[string]string{
					FLAG_NAME:     "my-name",
					FLAG_RUN_PORT: "[8080,8081]",
				},
			},
			wantProvider: func(fs dffilesystem.Filesystem) (parser.DevfileObj, error) {
				devfileObj, err := fakeDevfileProvider(fs)
				if err != nil {
					return parser.DevfileObj{}, err
				}
				err = devfileObj.Data.AddComponents([]v1alpha2.Component{
					testingutil.GetFakeContainerComponent("my-cont1", 1234, 2345),
					testingutil.GetFakeContainerComponent("my-cont2", 4321, 5432),
				})
				if err != nil {
					return parser.DevfileObj{}, err
				}
				err = devfileObj.Data.AddCommands([]v1alpha2.Command{
					{
						Id: "devdebug",
						CommandUnion: v1alpha2.CommandUnion{
							Exec: &v1alpha2.ExecCommand{
								LabeledCommand: v1alpha2.LabeledCommand{
									BaseCommand: v1alpha2.BaseCommand{
										Group: &v1alpha2.CommandGroup{Kind: v1alpha2.DebugCommandGroupKind},
									},
								},
							},
						},
					},
				})
				if err != nil {
					return parser.DevfileObj{}, err
				}
				return devfileObj, nil
			},
		},
		{
			name: "devfile with a default non-exec (apply) run command, --run-port set",
			args: args{
				devfileObjProvider: func(fs dffilesystem.Filesystem) (parser.DevfileObj, error) {
					devfileObj, err := fakeDevfileProvider(fs)
					if err != nil {
						return parser.DevfileObj{}, err
					}
					err = devfileObj.Data.AddComponents([]v1alpha2.Component{
						testingutil.GetFakeContainerComponent("my-cont1", 1234, 2345),
						testingutil.GetFakeContainerComponent("my-cont2", 4321, 5432),
					})
					if err != nil {
						return parser.DevfileObj{}, err
					}
					err = devfileObj.Data.AddCommands([]v1alpha2.Command{
						{
							Id: "devrun1",
							CommandUnion: v1alpha2.CommandUnion{
								Apply: &v1alpha2.ApplyCommand{
									LabeledCommand: v1alpha2.LabeledCommand{
										BaseCommand: v1alpha2.BaseCommand{
											Group: &v1alpha2.CommandGroup{
												Kind:      v1alpha2.RunCommandGroupKind,
												IsDefault: pointer.Bool(true),
											},
										},
									},
								},
							},
						},
					})
					if err != nil {
						return parser.DevfileObj{}, err
					}
					return devfileObj, nil
				},
				flags: map[string]string{
					FLAG_NAME:     "my-name",
					FLAG_RUN_PORT: "[8080,8081]",
				},
			},
			wantErr:      true,
			wantProvider: zeroDevfileProvider,
		},
		{
			name: "devfile with a default non-exec (composite) run command, --run-port set",
			args: args{
				devfileObjProvider: func(fs dffilesystem.Filesystem) (parser.DevfileObj, error) {
					devfileObj, err := fakeDevfileProvider(fs)
					if err != nil {
						return parser.DevfileObj{}, err
					}
					err = devfileObj.Data.AddComponents([]v1alpha2.Component{
						testingutil.GetFakeContainerComponent("my-cont1", 1234, 2345),
						testingutil.GetFakeContainerComponent("my-cont2", 4321, 5432),
					})
					if err != nil {
						return parser.DevfileObj{}, err
					}
					err = devfileObj.Data.AddCommands([]v1alpha2.Command{
						{
							Id: "devrun1",
							CommandUnion: v1alpha2.CommandUnion{
								Composite: &v1alpha2.CompositeCommand{
									LabeledCommand: v1alpha2.LabeledCommand{
										BaseCommand: v1alpha2.BaseCommand{
											Group: &v1alpha2.CommandGroup{
												Kind:      v1alpha2.RunCommandGroupKind,
												IsDefault: pointer.Bool(true),
											},
										},
									},
								},
							},
						},
					})
					if err != nil {
						return parser.DevfileObj{}, err
					}
					return devfileObj, nil
				},
				flags: map[string]string{
					FLAG_NAME:     "my-name",
					FLAG_RUN_PORT: "[8080,8081]",
				},
			},
			wantErr:      true,
			wantProvider: zeroDevfileProvider,
		},

		{
			name: "devfile with an exec run command with non-container component, --run-port set",
			args: args{
				devfileObjProvider: func(fs dffilesystem.Filesystem) (parser.DevfileObj, error) {
					devfileObj, err := fakeDevfileProvider(fs)
					if err != nil {
						return parser.DevfileObj{}, err
					}
					err = devfileObj.Data.AddCommands([]v1alpha2.Command{
						{
							Id: "devrun1",
							CommandUnion: v1alpha2.CommandUnion{
								Exec: &v1alpha2.ExecCommand{
									LabeledCommand: v1alpha2.LabeledCommand{
										BaseCommand: v1alpha2.BaseCommand{
											Group: &v1alpha2.CommandGroup{
												Kind:      v1alpha2.RunCommandGroupKind,
												IsDefault: pointer.Bool(true),
											},
										},
									},
									Component: "some-random-name",
								},
							},
						},
					})
					if err != nil {
						return parser.DevfileObj{}, err
					}
					return devfileObj, nil
				},
				flags: map[string]string{
					FLAG_NAME:     "my-name",
					FLAG_RUN_PORT: "[8080,8081]",
				},
			},
			wantErr:      true,
			wantProvider: zeroDevfileProvider,
		},
		{
			name: "devfile with an exec run command with non-container component, --run-port set",
			args: args{
				devfileObjProvider: func(fs dffilesystem.Filesystem) (parser.DevfileObj, error) {
					devfileObj, err := fakeDevfileProvider(fs)
					if err != nil {
						return parser.DevfileObj{}, err
					}
					err = devfileObj.Data.AddComponents([]v1alpha2.Component{
						testingutil.GetFakeContainerComponent("my-cont1", 1234, 2345),
						testingutil.GetFakeContainerComponent("my-cont2", 4321, 5432),
						{
							Name: "k8s-comp1",
							ComponentUnion: v1alpha2.ComponentUnion{
								Kubernetes: &v1alpha2.KubernetesComponent{
									K8sLikeComponent: v1alpha2.K8sLikeComponent{
										K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
											Inlined: "some-k8s-def",
										},
									},
								},
							},
						},
					})
					if err != nil {
						return parser.DevfileObj{}, err
					}
					err = devfileObj.Data.AddCommands([]v1alpha2.Command{
						{
							Id: "devrun1",
							CommandUnion: v1alpha2.CommandUnion{
								Exec: &v1alpha2.ExecCommand{
									LabeledCommand: v1alpha2.LabeledCommand{
										BaseCommand: v1alpha2.BaseCommand{
											Group: &v1alpha2.CommandGroup{
												Kind:      v1alpha2.RunCommandGroupKind,
												IsDefault: pointer.Bool(true),
											},
										},
									},
									Component: "k8s-comp1",
								},
							},
						},
					})
					if err != nil {
						return parser.DevfileObj{}, err
					}
					return devfileObj, nil
				},
				flags: map[string]string{
					FLAG_NAME:     "my-name",
					FLAG_RUN_PORT: "[8080,8081]",
				},
			},
			wantErr:      true,
			wantProvider: zeroDevfileProvider,
		},
		{
			name: "devfile with default exec run command with container component, --run-port set",
			args: args{
				devfileObjProvider: func(fs dffilesystem.Filesystem) (parser.DevfileObj, error) {
					devfileObj, err := fakeDevfileProvider(fs)
					if err != nil {
						return parser.DevfileObj{}, err
					}
					err = devfileObj.Data.AddComponents([]v1alpha2.Component{
						testingutil.GetFakeContainerComponent("my-cont1", 1234, 2345),
						testingutil.GetFakeContainerComponent("my-cont2", 4321, 5432),
					})
					if err != nil {
						return parser.DevfileObj{}, err
					}
					err = devfileObj.Data.AddCommands([]v1alpha2.Command{
						{
							Id: "devrun1",
							CommandUnion: v1alpha2.CommandUnion{
								Exec: &v1alpha2.ExecCommand{
									LabeledCommand: v1alpha2.LabeledCommand{
										BaseCommand: v1alpha2.BaseCommand{
											Group: &v1alpha2.CommandGroup{
												Kind:      v1alpha2.RunCommandGroupKind,
												IsDefault: pointer.Bool(true),
											},
										},
									},
									Component: "my-cont1",
								},
							},
						},
						{
							Id: "devdebug1",
							CommandUnion: v1alpha2.CommandUnion{
								Exec: &v1alpha2.ExecCommand{
									LabeledCommand: v1alpha2.LabeledCommand{
										BaseCommand: v1alpha2.BaseCommand{
											Group: &v1alpha2.CommandGroup{
												Kind:      v1alpha2.DebugCommandGroupKind,
												IsDefault: pointer.Bool(true),
											},
										},
									},
									Component: "my-cont2",
								},
							},
						},
					})
					if err != nil {
						return parser.DevfileObj{}, err
					}
					return devfileObj, nil
				},
				flags: map[string]string{
					FLAG_NAME:     "my-name",
					FLAG_RUN_PORT: "[8080,8081]",
				},
			},
			wantErr: false,
			wantProvider: func(fs dffilesystem.Filesystem) (parser.DevfileObj, error) {
				devfileObj, err := fakeDevfileProvider(fs)
				if err != nil {
					return parser.DevfileObj{}, err
				}
				// only my-cont1 (referenced by the default run command) should change
				cont1 := testingutil.GetFakeContainerComponent("my-cont1")
				cont1.Container.Endpoints = append(cont1.Container.Endpoints,
					v1alpha2.Endpoint{
						Name:       "port-8080-tcp",
						TargetPort: 8080,
						Protocol:   "tcp",
					},
					v1alpha2.Endpoint{
						Name:       "port-8081-tcp",
						TargetPort: 8081,
						Protocol:   "tcp",
					},
				)
				err = devfileObj.Data.AddComponents([]v1alpha2.Component{
					cont1,
					testingutil.GetFakeContainerComponent("my-cont2", 4321, 5432),
				})
				if err != nil {
					return parser.DevfileObj{}, err
				}
				err = devfileObj.Data.AddCommands([]v1alpha2.Command{
					{
						Id: "devrun1",
						CommandUnion: v1alpha2.CommandUnion{
							Exec: &v1alpha2.ExecCommand{
								LabeledCommand: v1alpha2.LabeledCommand{
									BaseCommand: v1alpha2.BaseCommand{
										Group: &v1alpha2.CommandGroup{
											Kind:      v1alpha2.RunCommandGroupKind,
											IsDefault: pointer.Bool(true),
										},
									},
								},
								Component: "my-cont1",
							},
						},
					},
					{
						Id: "devdebug1",
						CommandUnion: v1alpha2.CommandUnion{
							Exec: &v1alpha2.ExecCommand{
								LabeledCommand: v1alpha2.LabeledCommand{
									BaseCommand: v1alpha2.BaseCommand{
										Group: &v1alpha2.CommandGroup{
											Kind:      v1alpha2.DebugCommandGroupKind,
											IsDefault: pointer.Bool(true),
										},
									},
								},
								Component: "my-cont2",
							},
						},
					},
				})
				if err != nil {
					return parser.DevfileObj{}, err
				}
				return devfileObj, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &FlagsBackend{
				registryClient: tt.fields.registryClient,
			}
			fs := dffilesystem.NewFakeFs()
			devfileObj, err := tt.args.devfileObjProvider(fs)
			if err != nil {
				t.Errorf("error building input DevfileObj: %v", err)
				return
			}
			want, err := tt.wantProvider(fs)
			if err != nil {
				t.Errorf("error building expected DevfileObj: %v", err)
				return
			}

			got, err := o.HandleApplicationPorts(devfileObj, nil, tt.args.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleApplicationPorts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(parsercontext.DevfileCtx{})); diff != "" {
				t.Errorf("HandleApplicationPorts() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
