package envinfo

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"
	devfileFileSystem "github.com/devfile/library/pkg/testingutil/filesystem"
	"reflect"
	"testing"
)

func TestEnvSpecificInfo_AddServiceToDevfile(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()
	type fields struct {
		devfilePath       string
		Filename          string
		EnvInfo           EnvInfo
		envinfoFileExists bool
	}
	type args struct {
		crd  string
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    []v1alpha2.Component
	}{
		{
			name: "Case 1: Add service CRD to devfile.yaml",
			fields: fields{
				devfilePath: "",
				Filename:    "",
				EnvInfo: EnvInfo{
					devfileObj: parser.DevfileObj{
						Data: func() data.DevfileData {
							devfileData, _ := data.NewDevfileData(string(data.APIVersion200))
							return devfileData
						}(),
						Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
					},
				},
				envinfoFileExists: false,
			},
			args: args{
				crd:  "this is a test CRD",
				name: "testName",
			},
			wantErr: false,
			want: []v1alpha2.Component{
				{
					Name: "testName",
					ComponentUnion: devfile.ComponentUnion{
						Kubernetes: &devfile.KubernetesComponent{
							K8sLikeComponent: devfile.K8sLikeComponent{
								BaseComponent: devfile.BaseComponent{},
								K8sLikeComponentLocation: devfile.K8sLikeComponentLocation{
									Inlined: "this is a test CRD",
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			esi := &EnvSpecificInfo{
				devfilePath:       tt.fields.devfilePath,
				Filename:          tt.fields.Filename,
				EnvInfo:           tt.fields.EnvInfo,
				envinfoFileExists: tt.fields.envinfoFileExists,
			}
			if err := esi.AddServiceToDevfile(tt.args.crd, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("AddServiceToDevfile() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, _ := esi.devfileObj.Data.GetComponents(common.DevfileOptions{})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvSpecificInfo_DeleteServiceFromDevfile(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()
	type fields struct {
		devfilePath       string
		Filename          string
		EnvInfo           EnvInfo
		envinfoFileExists bool
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    []v1alpha2.Component
	}{
		{
			name: "Case 1: Remove a CRD from devfile.yaml",
			fields: fields{
				devfilePath: "",
				Filename:    "",
				EnvInfo: EnvInfo{
					devfileObj: parser.DevfileObj{
						Data: func() data.DevfileData {
							devfileData, _ := data.NewDevfileData(string(data.APIVersion200))
							_ = devfileData.AddComponents([]v1alpha2.Component{
								{
									Name: "testName",
									ComponentUnion: devfile.ComponentUnion{
										Kubernetes: &devfile.KubernetesComponent{
											K8sLikeComponent: devfile.K8sLikeComponent{
												BaseComponent: devfile.BaseComponent{},
												K8sLikeComponentLocation: devfile.K8sLikeComponentLocation{
													Inlined: "this is a test CRD",
												},
											},
										},
									},
								},
							})
							return devfileData
						}(),
						Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
					},
				},
			},
			args:    args{name: "testName"},
			wantErr: false,
			want:    []v1alpha2.Component{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			esi := &EnvSpecificInfo{
				devfilePath:       tt.fields.devfilePath,
				Filename:          tt.fields.Filename,
				EnvInfo:           tt.fields.EnvInfo,
				envinfoFileExists: tt.fields.envinfoFileExists,
			}
			if err := esi.DeleteServiceFromDevfile(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("DeleteServiceFromDevfile() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, _ := esi.devfileObj.Data.GetComponents(common.DevfileOptions{})
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}
