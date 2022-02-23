package libdevfile

import (
	"reflect"
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/libdevfile/generator"
	"k8s.io/utils/pointer"
)

func Test_getDefaultCommand(t *testing.T) {

	runDefault1 := generator.GetExecCommand(generator.ExecCommandParams{
		Kind:      v1alpha2.RunCommandGroupKind,
		Id:        "run-default-1",
		IsDefault: pointer.BoolPtr(true),
	})
	deployDefault1 := generator.GetCompositeCommand(generator.CompositeCommandParams{
		Kind:      v1alpha2.DeployCommandGroupKind,
		Id:        "deploy-default-1",
		IsDefault: pointer.BoolPtr(true),
	})
	deployDefault2 := generator.GetExecCommand(generator.ExecCommandParams{
		Kind:      v1alpha2.DeployCommandGroupKind,
		Id:        "deploy-default-2",
		IsDefault: pointer.BoolPtr(true),
	})
	deployNoDefault1 := generator.GetApplyCommand(generator.ApplyCommandParams{
		Kind:      v1alpha2.DeployCommandGroupKind,
		Id:        "deploy-no-default-1",
		IsDefault: pointer.BoolPtr(false),
	})
	deployUnspecDefault1 := generator.GetCompositeCommand(generator.CompositeCommandParams{
		Kind:      v1alpha2.DeployCommandGroupKind,
		Id:        "deploy-unspec-default-1",
		IsDefault: nil,
	})

	type args struct {
		devfileObj func() parser.DevfileObj
		kind       v1alpha2.CommandGroupKind
	}
	tests := []struct {
		name    string
		args    args
		want    v1alpha2.Command
		wantErr error
	}{
		{
			name: "a single deploy command, default",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddCommands([]v1alpha2.Command{runDefault1, deployDefault1})
					return parser.DevfileObj{
						Data: data,
					}
				},
				kind: v1alpha2.DeployCommandGroupKind,
			},
			wantErr: nil,
			want:    deployDefault1,
		},
		{
			name: "a single deploy command, not default",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddCommands([]v1alpha2.Command{runDefault1, deployNoDefault1})
					return parser.DevfileObj{
						Data: data,
					}
				},
				kind: v1alpha2.DeployCommandGroupKind,
			},
			wantErr: nil,
			want:    deployNoDefault1,
		},
		{
			name: "a single deploy command, unspecified default",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddCommands([]v1alpha2.Command{runDefault1, deployUnspecDefault1})
					return parser.DevfileObj{
						Data: data,
					}
				},
				kind: v1alpha2.DeployCommandGroupKind,
			},
			wantErr: nil,
			want:    deployUnspecDefault1,
		},
		{
			name: "several deploy commands, only one is default",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddCommands([]v1alpha2.Command{runDefault1, deployDefault1, deployNoDefault1, deployUnspecDefault1})
					return parser.DevfileObj{
						Data: data,
					}
				},
				kind: v1alpha2.DeployCommandGroupKind,
			},
			wantErr: nil,
			want:    deployDefault1,
		},
		{
			name: "no deploy command",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddCommands([]v1alpha2.Command{runDefault1})
					return parser.DevfileObj{
						Data: data,
					}
				},
				kind: v1alpha2.DeployCommandGroupKind,
			},
			wantErr: NewNoCommandFoundError(v1alpha2.DeployCommandGroupKind),
		},
		{
			name: "two deploy default commands",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddCommands([]v1alpha2.Command{runDefault1, deployDefault1, deployDefault2})
					return parser.DevfileObj{
						Data: data,
					}
				},
				kind: v1alpha2.DeployCommandGroupKind,
			},
			wantErr: NewMoreThanOneDefaultCommandFoundError(v1alpha2.DeployCommandGroupKind),
		},
		{
			name: "two deploy commands, no one is default",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddCommands([]v1alpha2.Command{runDefault1, deployNoDefault1, deployUnspecDefault1})
					return parser.DevfileObj{
						Data: data,
					}
				},
				kind: v1alpha2.DeployCommandGroupKind,
			},
			wantErr: NewNoDefaultCommandFoundError(v1alpha2.DeployCommandGroupKind),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDefaultCommand(tt.args.devfileObj(), tt.args.kind)
			if err != tt.wantErr {
				t.Errorf("getDefaultCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDefaultCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeploy(t *testing.T) {
	deployDefault1 := generator.GetCompositeCommand(generator.CompositeCommandParams{
		Kind:      v1alpha2.DeployCommandGroupKind,
		Id:        "deploy-default-1",
		IsDefault: pointer.BoolPtr(true),
		Commands:  []string{"image-command", "deployment-command", "service-command"},
	})
	applyImageCommand := generator.GetApplyCommand(generator.ApplyCommandParams{
		Kind:      v1alpha2.DeployCommandGroupKind,
		Id:        "image-command",
		IsDefault: pointer.BoolPtr(false),
		Component: "image-component",
	})
	applyDeploymentCommand := generator.GetApplyCommand(generator.ApplyCommandParams{
		Kind:      v1alpha2.DeployCommandGroupKind,
		Id:        "deployment-command",
		IsDefault: pointer.BoolPtr(false),
		Component: "deployment-component",
	})
	applyServiceCommand := generator.GetApplyCommand(generator.ApplyCommandParams{
		Kind:      v1alpha2.DeployCommandGroupKind,
		Id:        "service-command",
		IsDefault: pointer.BoolPtr(false),
		Component: "service-component",
	})

	imageComponent := generator.GetImageComponent(generator.ImageComponentParams{
		Name: "image-component",
		Image: v1alpha2.Image{
			ImageName: "an-image-name",
		},
	})
	deploymentComponent := generator.GetKubernetesComponent(generator.KubernetesComponentParams{
		Name:       "deployment-component",
		Kubernetes: &v1alpha2.KubernetesComponent{},
	})
	serviceComponent := generator.GetKubernetesComponent(generator.KubernetesComponentParams{
		Name:       "service-component",
		Kubernetes: &v1alpha2.KubernetesComponent{},
	})

	type args struct {
		devfileObj func() parser.DevfileObj
		handler    func(ctrl *gomock.Controller) Handler
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "deploy an image and two kubernetes components",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddCommands([]v1alpha2.Command{deployDefault1, applyImageCommand, applyDeploymentCommand, applyServiceCommand})
					_ = data.AddComponents([]v1alpha2.Component{imageComponent, deploymentComponent, serviceComponent})
					return parser.DevfileObj{
						Data: data,
					}
				},
				handler: func(ctrl *gomock.Controller) Handler {
					h := NewMockHandler(ctrl)
					h.EXPECT().ApplyImage(imageComponent)
					h.EXPECT().ApplyKubernetes(deploymentComponent)
					h.EXPECT().ApplyKubernetes(serviceComponent)
					return h
				},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			if err := Deploy(tt.args.devfileObj(), tt.args.handler(ctrl)); (err != nil) != tt.wantErr {
				t.Errorf("Deploy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
