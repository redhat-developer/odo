package service

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfile "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser/data/v2/common"

	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	"github.com/devfile/library/pkg/devfile/parser/data"
	devfileFileSystem "github.com/devfile/library/pkg/testingutil/filesystem"
	"reflect"
	"testing"
)

func TestAddKubernetesComponentToDevfile(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()

	type args struct {
		crd        string
		name       string
		devfileObj parser.DevfileObj
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    []v1alpha2.Component
	}{
		{
			name: "Case 1: Add service CRD to devfile.yaml",
			args: args{
				crd:  "test CRD",
				name: "testName",
				devfileObj: parser.DevfileObj{
					Data: func() data.DevfileData {
						devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
						if err != nil {
							t.Error(err)
						}
						return devfileData
					}(),
					Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				},
			},
			wantErr: false,
			want: []v1alpha2.Component{{
				Name: "testName",
				ComponentUnion: devfile.ComponentUnion{
					Kubernetes: &devfile.KubernetesComponent{
						K8sLikeComponent: devfile.K8sLikeComponent{
							BaseComponent: devfile.BaseComponent{},
							K8sLikeComponentLocation: devfile.K8sLikeComponentLocation{
								Inlined: "test CRD",
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
			if err := AddKubernetesComponentToDevfile(tt.args.crd, tt.args.name, tt.args.devfileObj); (err != nil) != tt.wantErr {
				t.Errorf("AddKubernetesComponentToDevfile() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := tt.args.devfileObj.Data.GetComponents(common.DevfileOptions{})
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteKubernetesComponentFromDevfile(t *testing.T) {
	fs := devfileFileSystem.NewFakeFs()

	type args struct {
		name       string
		devfileObj parser.DevfileObj
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    []v1alpha2.Component
	}{
		{
			name: "Case 1: Remove a CRD from devfile.yaml",
			args: args{
				name: "testName",
				devfileObj: parser.DevfileObj{
					Data: func() data.DevfileData {
						devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
						if err != nil {
							t.Error(err)
						}
						err = devfileData.AddComponents([]v1alpha2.Component{{
							Name: "testName",
							ComponentUnion: devfile.ComponentUnion{
								Kubernetes: &devfile.KubernetesComponent{
									K8sLikeComponent: devfile.K8sLikeComponent{
										BaseComponent: devfile.BaseComponent{},
										K8sLikeComponentLocation: devfile.K8sLikeComponentLocation{
											Inlined: "test CRD",
										},
									},
								},
							},
						},
						})
						if err != nil {
							t.Error(err)
						}
						return devfileData
					}(),
					Ctx: devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
				},
			},
			wantErr: false,
			want:    []v1alpha2.Component{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteKubernetesComponentFromDevfile(tt.args.name, tt.args.devfileObj); (err != nil) != tt.wantErr {
				t.Errorf("DeleteKubernetesComponentFromDevfile() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := tt.args.devfileObj.Data.GetComponents(common.DevfileOptions{})
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}
