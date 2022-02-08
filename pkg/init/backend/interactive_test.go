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

	"github.com/redhat-developer/odo/pkg/catalog"
	"github.com/redhat-developer/odo/pkg/init/asker"
)

func TestInteractiveBackend_SelectDevfile(t *testing.T) {
	type fields struct {
		buildAsker         func(ctrl *gomock.Controller) asker.Asker
		buildCatalogClient func(ctrl *gomock.Controller) catalog.Client
	}
	tests := []struct {
		name    string
		fields  fields
		want    *DevfileLocation
		wantErr bool
	}{
		{
			name: "direct selection",
			fields: fields{
				buildAsker: func(ctrl *gomock.Controller) asker.Asker {
					client := asker.NewMockAsker(ctrl)
					client.EXPECT().AskLanguage(gomock.Any()).Return("java", nil)
					client.EXPECT().AskType(gomock.Any()).Return(false, catalog.DevfileComponentType{
						Name: "a-devfile-name",
						Registry: catalog.Registry{
							Name: "MyRegistry1",
						},
					}, nil)
					return client
				},
				buildCatalogClient: func(ctrl *gomock.Controller) catalog.Client {
					client := catalog.NewMockClient(ctrl)
					client.EXPECT().ListDevfileComponents(gomock.Any())
					return client
				},
			},
			want: &DevfileLocation{
				Devfile:         "a-devfile-name",
				DevfileRegistry: "MyRegistry1",
			},
		},
		{
			name: "selection with back",
			fields: fields{
				buildAsker: func(ctrl *gomock.Controller) asker.Asker {
					client := asker.NewMockAsker(ctrl)
					client.EXPECT().AskLanguage(gomock.Any()).Return("java", nil)
					client.EXPECT().AskType(gomock.Any()).Return(true, catalog.DevfileComponentType{}, nil)
					client.EXPECT().AskLanguage(gomock.Any()).Return("go", nil)
					client.EXPECT().AskType(gomock.Any()).Return(false, catalog.DevfileComponentType{
						Name: "a-devfile-name",
						Registry: catalog.Registry{
							Name: "MyRegistry1",
						},
					}, nil)
					return client
				},
				buildCatalogClient: func(ctrl *gomock.Controller) catalog.Client {
					client := catalog.NewMockClient(ctrl)
					client.EXPECT().ListDevfileComponents(gomock.Any())
					return client
				},
			},
			want: &DevfileLocation{
				Devfile:         "a-devfile-name",
				DevfileRegistry: "MyRegistry1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			o := &InteractiveBackend{
				asker:         tt.fields.buildAsker(ctrl),
				catalogClient: tt.fields.buildCatalogClient(ctrl),
			}
			_, got, err := o.SelectDevfile(map[string]string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("InteractiveBuilder.ParamsBuild() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InteractiveBuilder.ParamsBuild() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInteractiveBackend_SelectStarterProject(t *testing.T) {
	type fields struct {
		asker         func(ctrl *gomock.Controller) asker.Asker
		catalogClient catalog.Client
	}
	type args struct {
		devfile func() parser.DevfileObj
		flags   map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantOK  bool
		want    *v1alpha2.StarterProject
		wantErr bool
	}{
		{
			name: "some flags passed",
			args: args{
				devfile: func() parser.DevfileObj {
					devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					return parser.DevfileObj{
						Data: devfileData,
					}
				},
				flags: map[string]string{
					"devfile": "adevfile",
				},
			},
			wantOK:  false,
			want:    nil,
			wantErr: false,
		},
		{
			name: "no flags, no starter selected",
			fields: fields{
				asker: func(ctrl *gomock.Controller) asker.Asker {
					client := asker.NewMockAsker(ctrl)
					client.EXPECT().AskStarterProject(gomock.Any()).Return(false, 0, nil)
					return client
				},
			},
			args: args{
				devfile: func() parser.DevfileObj {
					devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					return parser.DevfileObj{
						Data: devfileData,
					}
				},
				flags: map[string]string{},
			},
			wantOK:  true,
			want:    nil,
			wantErr: false,
		},
		{
			name: "no flags, starter selected",
			fields: fields{
				asker: func(ctrl *gomock.Controller) asker.Asker {
					client := asker.NewMockAsker(ctrl)
					client.EXPECT().AskStarterProject(gomock.Any()).Return(true, 1, nil)
					return client
				},
			},
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
				flags: map[string]string{},
			},
			wantOK: true,
			want: &v1alpha2.StarterProject{
				Name: "starter2",
			},
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			var askerClient asker.Asker
			if tt.fields.asker != nil {
				askerClient = tt.fields.asker(ctrl)
			}
			o := &InteractiveBackend{
				asker:         askerClient,
				catalogClient: tt.fields.catalogClient,
			}
			got, got1, err := o.SelectStarterProject(tt.args.devfile(), tt.args.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("InteractiveBackend.SelectStarterProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantOK {
				t.Errorf("InteractiveBackend.SelectStarterProject() got = %v, wantOK %v", got, tt.wantOK)
			}
			if !reflect.DeepEqual(got1, tt.want) {
				t.Errorf("InteractiveBackend.SelectStarterProject() got1 = %v, want %v", got1, tt.want)
			}
		})
	}
}

func TestInteractiveBackend_PersonalizeName(t *testing.T) {
	type fields struct {
		asker         func(ctrl *gomock.Controller) asker.Asker
		catalogClient catalog.Client
	}
	type args struct {
		devfile func(fs filesystem.Filesystem) parser.DevfileObj
		flags   map[string]string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		want        bool
		wantErr     bool
		checkResult func(devfile parser.DevfileObj, args args) bool
	}{
		{
			name: "no flag",
			fields: fields{
				asker: func(ctrl *gomock.Controller) asker.Asker {
					client := asker.NewMockAsker(ctrl)
					client.EXPECT().AskName(gomock.Any()).Return("aname", nil)
					return client
				},
			},
			args: args{
				devfile: func(fs filesystem.Filesystem) parser.DevfileObj {
					devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					obj := parser.DevfileObj{
						Ctx:  parsercontext.FakeContext(fs, "/tmp/devfile.yaml"),
						Data: devfileData,
					}
					return obj
				},
				flags: map[string]string{},
			},
			want:    true,
			wantErr: false,
			checkResult: func(devfile parser.DevfileObj, args args) bool {
				return devfile.GetMetadataName() == "aname"
			},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			var askerClient asker.Asker
			if tt.fields.asker != nil {
				askerClient = tt.fields.asker(ctrl)
			}
			o := &InteractiveBackend{
				asker:         askerClient,
				catalogClient: tt.fields.catalogClient,
			}
			fs := filesystem.NewFakeFs()
			devfile := tt.args.devfile(fs)
			got, err := o.PersonalizeName(devfile, tt.args.flags)
			if (err != nil) != tt.wantErr {
				t.Errorf("InteractiveBackend.PersonalizeName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("InteractiveBackend.PersonalizeName() = %v, want %v", got, tt.want)
			}
			if tt.checkResult != nil && !tt.checkResult(devfile, tt.args) {
				t.Errorf("InteractiveBackend.PersonalizeName(), checking result failed")
			}
		})
	}
}
