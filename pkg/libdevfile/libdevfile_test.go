package libdevfile

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data"
	devfileFileSystem "github.com/devfile/library/pkg/testingutil/filesystem"
	dfutil "github.com/devfile/library/pkg/util"
	"github.com/golang/mock/gomock"
	"github.com/kylelemons/godebug/pretty"
	"k8s.io/utils/pointer"

	"github.com/redhat-developer/odo/pkg/libdevfile/generator"
	"github.com/redhat-developer/odo/pkg/testingutil"
	"github.com/redhat-developer/odo/pkg/util"
)

var buildGroup = v1alpha2.BuildCommandGroupKind
var runGroup = v1alpha2.RunCommandGroupKind
var debugGroup = v1alpha2.DebugCommandGroupKind

func TestGetDefaultCommand(t *testing.T) {

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
			got, err := GetDefaultCommand(tt.args.devfileObj(), tt.args.kind)
			if err != tt.wantErr {
				t.Errorf("GetDefaultCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDefaultCommand() = %v, want %v", got, tt.want)
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

func TestGetContainerEndpointMapping(t *testing.T) {
	type args struct {
		containers []v1alpha2.Component
	}

	imageComponent := generator.GetImageComponent(generator.ImageComponentParams{
		Name: "image-component",
		Image: v1alpha2.Image{
			ImageName: "an-image-name",
		},
	})

	containerWithNoEndpoints := generator.GetContainerComponent(generator.ContainerComponentParams{
		Name:      "container 1",
		Endpoints: nil,
	})

	containerWithOnePublicEndpoint := generator.GetContainerComponent(generator.ContainerComponentParams{
		Name: "container 2",
		Endpoints: []v1alpha2.Endpoint{
			{
				Name:       "ep1",
				TargetPort: 8080,
				Exposure:   v1alpha2.PublicEndpointExposure,
			},
		},
	})

	containerWithOneInternalEndpoint := generator.GetContainerComponent(generator.ContainerComponentParams{
		Name: "container 3",
		Endpoints: []v1alpha2.Endpoint{
			{
				Name:       "ep2",
				TargetPort: 9090,
				Exposure:   v1alpha2.InternalEndpointExposure,
			},
		},
	})

	tests := []struct {
		name string
		args args
		want map[string][]int
	}{
		{
			name: "invalid input - image components instead of container components",
			args: args{
				containers: []v1alpha2.Component{imageComponent},
			},
			want: map[string][]int{},
		},
		{
			name: "one container with no endpoints exposed",
			args: args{
				containers: []v1alpha2.Component{containerWithNoEndpoints},
			},
			want: map[string][]int{containerWithNoEndpoints.Name: {}},
		},
		{
			name: "multiple containers with varying types of endpoints",
			args: args{
				containers: []v1alpha2.Component{containerWithNoEndpoints, containerWithOnePublicEndpoint, containerWithOneInternalEndpoint},
			},
			want: map[string][]int{containerWithNoEndpoints.Name: {}, containerWithOnePublicEndpoint.Name: {8080}, containerWithOneInternalEndpoint.Name: {9090}},
		},
		{
			name: "invalid input - one image component with rest being containers",
			args: args{
				containers: []v1alpha2.Component{containerWithNoEndpoints, containerWithOnePublicEndpoint, containerWithOneInternalEndpoint, imageComponent},
			},
			want: map[string][]int{containerWithNoEndpoints.Name: {}, containerWithOnePublicEndpoint.Name: {8080}, containerWithOneInternalEndpoint.Name: {9090}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetContainerEndpointMapping(tt.args.containers)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetContainerEndpointMapping() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEndpointsFromDevfile(t *testing.T) {
	type args struct {
		devfileObj      func() parser.DevfileObj
		ignoreExposures []v1alpha2.EndpointExposure
	}
	ep1 := v1alpha2.Endpoint{Name: "ep1", TargetPort: 8080, Exposure: v1alpha2.NoneEndpointExposure}
	ep2 := v1alpha2.Endpoint{Name: "ep2", TargetPort: 9090, Exposure: v1alpha2.InternalEndpointExposure}
	ep3 := v1alpha2.Endpoint{Name: "ep3", TargetPort: 8888, Exposure: v1alpha2.PublicEndpointExposure}

	container := generator.GetContainerComponent(generator.ContainerComponentParams{
		Name:      "container-1",
		Endpoints: []v1alpha2.Endpoint{ep1, ep2, ep3},
	})
	tests := []struct {
		name    string
		args    args
		want    []v1alpha2.Endpoint
		wantErr bool
	}{
		{
			name: "Ignore exposure of type none",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddComponents([]v1alpha2.Component{container})
					return parser.DevfileObj{
						Data: data,
					}
				},
				ignoreExposures: []v1alpha2.EndpointExposure{v1alpha2.NoneEndpointExposure},
			},
			want: []v1alpha2.Endpoint{ep2, ep3},
		},
		{
			name: "Ignore exposure of type public",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddComponents([]v1alpha2.Component{container})
					return parser.DevfileObj{
						Data: data,
					}
				},
				ignoreExposures: []v1alpha2.EndpointExposure{v1alpha2.PublicEndpointExposure},
			},
			want: []v1alpha2.Endpoint{ep1, ep2},
		},
		{
			name: "Ignore exposure of type internal",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddComponents([]v1alpha2.Component{container})
					return parser.DevfileObj{
						Data: data,
					}
				},
				ignoreExposures: []v1alpha2.EndpointExposure{v1alpha2.InternalEndpointExposure},
			},
			want: []v1alpha2.Endpoint{ep1, ep3},
		},
		{
			name: "Ignore none",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddComponents([]v1alpha2.Component{container})
					return parser.DevfileObj{
						Data: data,
					}
				},
				ignoreExposures: []v1alpha2.EndpointExposure{},
			},
			want: []v1alpha2.Endpoint{ep1, ep2, ep3},
		},
		{
			name: "Ignore all exposure types",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddComponents([]v1alpha2.Component{container})
					return parser.DevfileObj{
						Data: data,
					}
				},
				ignoreExposures: []v1alpha2.EndpointExposure{v1alpha2.InternalEndpointExposure, v1alpha2.NoneEndpointExposure, v1alpha2.PublicEndpointExposure},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEndpointsFromDevfile(tt.args.devfileObj(), tt.args.ignoreExposures)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEndpointsFromDevfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEndpointsFromDevfile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetK8sManifestWithVariablesSubstituted(t *testing.T) {
	fakeFs := devfileFileSystem.NewFakeFs()
	cmpName := "my-cmp-1"
	for _, tt := range []struct {
		name           string
		setupFunc      func() error
		devfileObjFunc func() parser.DevfileObj
		wantErr        bool
		want           string
	}{
		{
			name: "Missing Component",
			devfileObjFunc: func() parser.DevfileObj {
				devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion220))
				cmp := generator.GetContainerComponent(generator.ContainerComponentParams{
					Name: "a-different-component",
				})
				s := v1alpha2.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
						Components: []v1alpha2.Component{cmp},
					},
				}
				devfileData.SetDevfileWorkspaceSpec(s)
				return parser.DevfileObj{
					Data: devfileData,
				}
			},
			wantErr: true,
		},
		{
			name: "Multiple Components with the same name",
			devfileObjFunc: func() parser.DevfileObj {
				devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion220))
				cmp1 := generator.GetContainerComponent(generator.ContainerComponentParams{
					Name: cmpName,
				})
				cmp2 := generator.GetImageComponent(generator.ImageComponentParams{
					Name: cmpName,
				})
				s := v1alpha2.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
						Components: []v1alpha2.Component{cmp1, cmp2},
					},
				}
				devfileData.SetDevfileWorkspaceSpec(s)
				return parser.DevfileObj{
					Data: devfileData,
				}
			},
			wantErr: true,
		},
		{
			name: "Container Component",
			devfileObjFunc: func() parser.DevfileObj {
				devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion220))
				cmp := generator.GetContainerComponent(generator.ContainerComponentParams{
					Name: cmpName,
				})
				s := v1alpha2.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
						Components: []v1alpha2.Component{cmp},
					},
				}
				devfileData.SetDevfileWorkspaceSpec(s)
				return parser.DevfileObj{
					Data: devfileData,
				}
			},
			wantErr: true,
		},
		{
			name: "Image Component",
			devfileObjFunc: func() parser.DevfileObj {
				devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion220))
				cmp := generator.GetImageComponent(generator.ImageComponentParams{
					Name: cmpName,
				})
				s := v1alpha2.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
						Components: []v1alpha2.Component{cmp},
					},
				}
				devfileData.SetDevfileWorkspaceSpec(s)
				return parser.DevfileObj{
					Data: devfileData,
				}
			},
			wantErr: true,
		},
		{
			name: "Kubernetes Component - Inlined with no variables",
			devfileObjFunc: func() parser.DevfileObj {
				devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion220))
				cmp := generator.GetKubernetesComponent(generator.KubernetesComponentParams{
					Name: cmpName,
					Kubernetes: &v1alpha2.KubernetesComponent{
						K8sLikeComponent: v1alpha2.K8sLikeComponent{
							K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
								Inlined: "some-text-inlined",
							},
						},
					},
				})
				s := v1alpha2.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
						Components: []v1alpha2.Component{cmp},
					},
				}
				devfileData.SetDevfileWorkspaceSpec(s)
				return parser.DevfileObj{
					Data: devfileData,
				}
			},
			wantErr: false,
			want:    "some-text-inlined",
		},
		{
			name: "Kubernetes Component - Inlined with variables",
			devfileObjFunc: func() parser.DevfileObj {
				devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion220))
				cmp := generator.GetKubernetesComponent(generator.KubernetesComponentParams{
					Name: cmpName,
					Kubernetes: &v1alpha2.KubernetesComponent{
						K8sLikeComponent: v1alpha2.K8sLikeComponent{
							K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
								Inlined: "image: {{MY_CONTAINER_IMAGE}}",
							},
						},
					},
				})
				s := v1alpha2.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
						Variables: map[string]string{
							"MY_CONTAINER_IMAGE": "quay.io/unknown-account/my-image:1.2.3",
						},
						Components: []v1alpha2.Component{cmp},
					},
				}
				devfileData.SetDevfileWorkspaceSpec(s)
				return parser.DevfileObj{
					Data: devfileData,
				}
			},
			wantErr: false,
			want:    "image: quay.io/unknown-account/my-image:1.2.3",
		},
		{
			name: "Kubernetes Component - Inlined with unknown variables",
			devfileObjFunc: func() parser.DevfileObj {
				devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion220))
				cmp := generator.GetKubernetesComponent(generator.KubernetesComponentParams{
					Name: cmpName,
					Kubernetes: &v1alpha2.KubernetesComponent{
						K8sLikeComponent: v1alpha2.K8sLikeComponent{
							K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
								Inlined: "image: {{MY_CONTAINER_IMAGE}}:{{ MY_CONTAINER_IMAGE_VERSION_UNKNOWN }}",
							},
						},
					},
				})
				s := v1alpha2.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
						Variables: map[string]string{
							"MY_CONTAINER_IMAGE": "quay.io/unknown-account/my-image",
						},
						Components: []v1alpha2.Component{cmp},
					},
				}
				devfileData.SetDevfileWorkspaceSpec(s)
				return parser.DevfileObj{
					Data: devfileData,
				}
			},
			wantErr: true,
			want:    "image: quay.io/unknown-account/my-image:{{ MY_CONTAINER_IMAGE_VERSION_UNKNOWN }}",
		},
		{
			name: "Kubernetes Component - non-existing external file",
			devfileObjFunc: func() parser.DevfileObj {
				devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion220))
				cmp := generator.GetKubernetesComponent(generator.KubernetesComponentParams{
					Name: cmpName,
					Kubernetes: &v1alpha2.KubernetesComponent{
						K8sLikeComponent: v1alpha2.K8sLikeComponent{
							K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
								Uri: "kubernetes/my-external-file-with-should-not-exist",
							},
						},
					},
				})
				s := v1alpha2.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
						Components: []v1alpha2.Component{cmp},
					},
				}
				devfileData.SetDevfileWorkspaceSpec(s)
				return parser.DevfileObj{
					Data: devfileData,
				}
			},
			wantErr: true,
		},
		{
			name: "Kubernetes Component - URI with no variables",
			setupFunc: func() error {
				return fakeFs.WriteFile("kubernetes/my-external-file",
					[]byte("some-text-with-no-variables"),
					os.ModePerm)
			},
			devfileObjFunc: func() parser.DevfileObj {
				devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion220))
				cmp := generator.GetKubernetesComponent(generator.KubernetesComponentParams{
					Name: cmpName,
					Kubernetes: &v1alpha2.KubernetesComponent{
						K8sLikeComponent: v1alpha2.K8sLikeComponent{
							K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
								Uri: "kubernetes/my-external-file",
							},
						},
					},
				})
				s := v1alpha2.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
						Components: []v1alpha2.Component{cmp},
					},
				}
				devfileData.SetDevfileWorkspaceSpec(s)
				return parser.DevfileObj{
					Data: devfileData,
				}
			},
			wantErr: false,
			want:    "some-text-with-no-variables",
		},
		{
			name: "Kubernetes Component - URI with variables",
			setupFunc: func() error {
				return fakeFs.WriteFile("kubernetes/my-deployment.yaml",
					[]byte("image: {{ MY_CONTAINER_IMAGE }}:{{MY_CONTAINER_IMAGE_VERSION}}"),
					os.ModePerm)
			},
			devfileObjFunc: func() parser.DevfileObj {
				devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion220))
				cmp := generator.GetKubernetesComponent(generator.KubernetesComponentParams{
					Name: cmpName,
					Kubernetes: &v1alpha2.KubernetesComponent{
						K8sLikeComponent: v1alpha2.K8sLikeComponent{
							K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
								Uri: "kubernetes/my-deployment.yaml",
							},
						},
					},
				})
				s := v1alpha2.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
						Variables: map[string]string{
							"MY_CONTAINER_IMAGE":         "quay.io/unknown-account/my-image",
							"MY_CONTAINER_IMAGE_VERSION": "1.2.3",
						},
						Components: []v1alpha2.Component{cmp},
					},
				}
				devfileData.SetDevfileWorkspaceSpec(s)
				return parser.DevfileObj{
					Data: devfileData,
				}
			},
			wantErr: false,
			want:    "image: quay.io/unknown-account/my-image:1.2.3",
		},
		{
			name: "Kubernetes Component - URI with unknown variables",
			setupFunc: func() error {
				return fakeFs.WriteFile("kubernetes/my-external-file.yaml",
					[]byte("image: {{MY_CONTAINER_IMAGE}}:{{ MY_CONTAINER_IMAGE_VERSION_UNKNOWN }}"),
					os.ModePerm)
			},
			devfileObjFunc: func() parser.DevfileObj {
				devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion220))
				cmp := generator.GetKubernetesComponent(generator.KubernetesComponentParams{
					Name: cmpName,
					Kubernetes: &v1alpha2.KubernetesComponent{
						K8sLikeComponent: v1alpha2.K8sLikeComponent{
							K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
								Uri: "kubernetes/my-external-file.yaml",
							},
						},
					},
				})
				s := v1alpha2.DevWorkspaceTemplateSpec{
					DevWorkspaceTemplateSpecContent: v1alpha2.DevWorkspaceTemplateSpecContent{
						Variables: map[string]string{
							"MY_CONTAINER_IMAGE": "quay.io/unknown-account/my-image",
						},
						Components: []v1alpha2.Component{cmp},
					},
				}
				devfileData.SetDevfileWorkspaceSpec(s)
				return parser.DevfileObj{
					Data: devfileData,
				}
			},
			wantErr: true,
			want:    "image: quay.io/unknown-account/my-image:{{ MY_CONTAINER_IMAGE_VERSION_UNKNOWN }}",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				if err := tt.setupFunc(); err != nil {
					t.Errorf("setup function returned an error: %v", err)
					return
				}
			}
			if tt.devfileObjFunc == nil {
				t.Error("devfileObjFunc function not defined for test case")
				return
			}

			got, err := GetK8sManifestWithVariablesSubstituted(tt.devfileObjFunc(), cmpName, "", fakeFs)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetK8sManifestWithVariablesSubstituted() error = %v, wantErr %v",
					err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetK8sManifestWithVariablesSubstituted() got = %v, want %v",
					got, tt.want)
			}
		})
	}
}

func TestGetCommand(t *testing.T) {

	commands := [...]string{"ls -la", "pwd"}
	components := [...]string{"alias1", "alias2"}

	tests := []struct {
		name           string
		requestedType  []v1alpha2.CommandGroupKind
		execCommands   []v1alpha2.Command
		compCommands   []v1alpha2.Command
		reqCommandName string
		retCommandName string
		wantErr        bool
	}{
		{
			name: "Case 1: Valid devfile",
			execCommands: []v1alpha2.Command{
				getExecCommand("build", buildGroup),
				getExecCommand("run", runGroup),
			},
			requestedType: []v1alpha2.CommandGroupKind{buildGroup, runGroup},
			wantErr:       false,
		},
		{
			name: "Case 2: Valid devfile with devrun and devbuild",
			execCommands: []v1alpha2.Command{
				getExecCommand("build", buildGroup),
				getExecCommand("run", runGroup),
			},
			requestedType: []v1alpha2.CommandGroupKind{buildGroup, runGroup},
			wantErr:       false,
		},
		{
			name: "Case 3: Valid devfile with empty workdir",
			execCommands: []v1alpha2.Command{
				{
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			requestedType: []v1alpha2.CommandGroupKind{runGroup},
			wantErr:       false,
		},
		{
			name: "Case 4: Mismatched command type",
			execCommands: []v1alpha2.Command{
				{
					Id: "build command",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			reqCommandName: "build command",
			requestedType:  []v1alpha2.CommandGroupKind{buildGroup},
			wantErr:        true,
		},
		{
			name: "Case 5: Default command is returned",
			execCommands: []v1alpha2.Command{
				{
					Id: "defaultRunCommand",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
				{
					Id: "runCommand",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			retCommandName: "defaultRunCommand",
			requestedType:  []v1alpha2.CommandGroupKind{runGroup},
			wantErr:        false,
		},
		{
			name: "Case 6: Composite command is returned",
			execCommands: []v1alpha2.Command{
				{
					Id: "build",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(false)},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
				{
					Id: "run",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			compCommands: []v1alpha2.Command{
				{
					Id: "myComposite",
					CommandUnion: v1alpha2.CommandUnion{
						Composite: &v1alpha2.CompositeCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							Commands: []string{"build", "run"},
						},
					},
				},
			},
			retCommandName: "myComposite",
			requestedType:  []v1alpha2.CommandGroupKind{buildGroup},
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := []v1alpha2.Component{testingutil.GetFakeContainerComponent(tt.execCommands[0].Exec.Component)}
			devObj := parser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.execCommands)
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.compCommands)
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents(components)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			for _, gtype := range tt.requestedType {
				cmd, err := getCommand(devObj.Data, tt.reqCommandName, gtype)
				if !tt.wantErr == (err != nil) {
					t.Errorf("TestGetCommand unexpected error for command: %v wantErr: %v err: %v", gtype, tt.wantErr, err)
					return
				} else if tt.wantErr {
					return
				}

				if len(tt.retCommandName) > 0 && cmd.Id != tt.retCommandName {
					t.Errorf("TestGetCommand error: command names do not match expected: %v actual: %v", tt.retCommandName, cmd.Id)
				}
			}
		})
	}

}

func Test_getCommandAssociatedToGroup(t *testing.T) {

	commands := [...]string{"ls -la", "pwd"}
	components := [...]string{"alias1", "alias2"}

	tests := []struct {
		name           string
		requestedType  []v1alpha2.CommandGroupKind
		execCommands   []v1alpha2.Command
		compCommands   []v1alpha2.Command
		retCommandName string
		wantErr        bool
	}{
		{
			name: "Case 1: Valid devfile",
			execCommands: []v1alpha2.Command{
				getExecCommand("", buildGroup),
				getExecCommand("", runGroup),
			},
			requestedType: []v1alpha2.CommandGroupKind{buildGroup, runGroup},
			wantErr:       false,
		},
		{
			name: "Case 2: Valid devfile with devrun and devbuild",
			execCommands: []v1alpha2.Command{
				getExecCommand("", buildGroup),
				getExecCommand("", runGroup),
			},
			requestedType: []v1alpha2.CommandGroupKind{buildGroup, runGroup},
			wantErr:       false,
		},
		{
			name: "Case 3: Valid devfile with empty workdir",
			execCommands: []v1alpha2.Command{
				{
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			requestedType: []v1alpha2.CommandGroupKind{runGroup},
			wantErr:       false,
		},
		{
			name: "Case 4: Default command is returned",
			execCommands: []v1alpha2.Command{
				{
					Id: "defaultruncommand",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
				{
					Id: "runcommand",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			retCommandName: "defaultruncommand",
			requestedType:  []v1alpha2.CommandGroupKind{runGroup},
			wantErr:        false,
		},
		{
			name: "Case 5: Valid devfile, has composite command",
			execCommands: []v1alpha2.Command{
				{
					Id: "build1",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: buildGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
				{
					Id: "build2",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: buildGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
				{
					Id: "run",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			compCommands: []v1alpha2.Command{
				{
					Id: "mycomp",
					CommandUnion: v1alpha2.CommandUnion{
						Composite: &v1alpha2.CompositeCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							Commands: []string{"build1", "run"},
						},
					},
				},
			},
			retCommandName: "mycomp",
			requestedType:  []v1alpha2.CommandGroupKind{buildGroup},
			wantErr:        false,
		},
		{
			name: "Case 6: Default composite command",
			execCommands: []v1alpha2.Command{
				{
					Id: "build",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(false)},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
				{
					Id: "run",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			compCommands: []v1alpha2.Command{
				{
					Id: "mycomp",
					CommandUnion: v1alpha2.CommandUnion{
						Composite: &v1alpha2.CompositeCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							Commands: []string{"build", "run"},
						},
					},
				},
				{
					Id: "mycomp2",
					CommandUnion: v1alpha2.CommandUnion{
						Composite: &v1alpha2.CompositeCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(false)},
								},
							},
							Commands: []string{"build", "run"},
						},
					},
				},
			},
			retCommandName: "mycomp",
			requestedType:  []v1alpha2.CommandGroupKind{buildGroup},
			wantErr:        false,
		},
		{
			name: "Case 7: no build and debug commands",
			execCommands: []v1alpha2.Command{
				getExecCommand("", runGroup),
			},
			requestedType: []v1alpha2.CommandGroupKind{buildGroup, debugGroup},
			wantErr:       false,
		},
		{
			name: "Case 8: no default build and debug commands",
			execCommands: []v1alpha2.Command{
				getExecCommand("build-0", buildGroup),
				getExecCommand("build-1", buildGroup),
				getExecCommand("debug-0", debugGroup),
				getExecCommand("debug-1", debugGroup),
				getExecCommand("", runGroup),
			},
			requestedType: []v1alpha2.CommandGroupKind{buildGroup, debugGroup},
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := []v1alpha2.Component{testingutil.GetFakeContainerComponent(tt.execCommands[0].Exec.Component)}
			devObj := parser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.execCommands)
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.compCommands)
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents(components)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			for _, gtype := range tt.requestedType {
				cmd, err := getCommandAssociatedToGroup(devObj.Data, gtype)
				if !tt.wantErr == (err != nil) {
					t.Errorf("TestGetCommandFromDevfile unexpected error for command: %v wantErr: %v err: %v", gtype, tt.wantErr, err)
					return
				} else if tt.wantErr {
					return
				}

				if len(tt.retCommandName) > 0 && cmd.Id != tt.retCommandName {
					t.Errorf("TestGetCommandFromDevfile error: command names do not match expected: %v actual: %v", tt.retCommandName, cmd.Id)
				}
			}
		})
	}

}

func Test_getCommandByName(t *testing.T) {

	commands := [...]string{"ls -la", "pwd"}
	components := [...]string{"alias1", "alias2"}
	invalidComponent := "garbagealias"

	tests := []struct {
		name           string
		requestedType  v1alpha2.CommandGroupKind
		execCommands   []v1alpha2.Command
		compCommands   []v1alpha2.Command
		reqCommandName string
		retCommandName string
		wantErr        bool
	}{
		{
			name: "Case 1: Valid devfile",
			execCommands: []v1alpha2.Command{
				getExecCommand("a", buildGroup),
				getExecCommand("b", runGroup),
			},
			reqCommandName: "b",
			retCommandName: "b",
			requestedType:  runGroup,
			wantErr:        false,
		},
		{
			name: "Case 2: Valid devfile with empty workdir",
			execCommands: []v1alpha2.Command{
				{
					Id: "build command",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			reqCommandName: "build command",
			retCommandName: "build command",
			requestedType:  runGroup,
			wantErr:        false,
		},
		{
			name: "Case 3: Invalid command",
			execCommands: []v1alpha2.Command{
				{
					Id: "build command",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   invalidComponent,
						},
					},
				},
			},
			reqCommandName: "build command wrong",
			requestedType:  runGroup,
			wantErr:        true,
		},
		{
			name: "Case 4: Mismatched command type",
			execCommands: []v1alpha2.Command{
				{
					Id: "build command",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			reqCommandName: "build command",
			requestedType:  buildGroup,
			wantErr:        true,
		},
		{
			name: "Case 5: Multiple default commands but should be with the flag",
			execCommands: []v1alpha2.Command{
				{
					Id: "defaultruncommand",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
				{
					Id: "runcommand",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			reqCommandName: "defaultruncommand",
			retCommandName: "defaultruncommand",
			requestedType:  runGroup,
			wantErr:        false,
		},
		{
			name: "Case 6: No default command but should be with the flag",
			execCommands: []v1alpha2.Command{
				{
					Id: "defaultruncommand",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
				{
					Id: "runcommand",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			reqCommandName: "defaultruncommand",
			retCommandName: "defaultruncommand",
			requestedType:  runGroup,
			wantErr:        false,
		},
		{
			name: "Case 7: No Command Group",
			execCommands: []v1alpha2.Command{
				{
					Id: "defaultruncommand",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			reqCommandName: "defaultruncommand",
			retCommandName: "defaultruncommand",
			requestedType:  runGroup,
			wantErr:        false,
		},
		{
			name: "Case 8: Valid devfile with composite commands",
			execCommands: []v1alpha2.Command{
				{
					Id: "build",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(false)},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
				{
					Id: "run",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			compCommands: []v1alpha2.Command{
				{
					Id: "mycomp",
					CommandUnion: v1alpha2.CommandUnion{
						Composite: &v1alpha2.CompositeCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							Commands: []string{"build", "run"},
						},
					},
				},
				{
					Id: "mycomp2",
					CommandUnion: v1alpha2.CommandUnion{
						Composite: &v1alpha2.CompositeCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(false)},
								},
							},
							Commands: []string{"build", "run"},
						},
					},
				},
			},
			reqCommandName: "mycomp",
			retCommandName: "mycomp",
			requestedType:  buildGroup,
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := []v1alpha2.Component{testingutil.GetFakeContainerComponent(tt.execCommands[0].Exec.Component)}
			if tt.execCommands[0].Exec.Component == invalidComponent {
				components = []v1alpha2.Component{testingutil.GetFakeContainerComponent("randomComponent")}
			}
			devObj := parser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.execCommands)
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.compCommands)
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents(components)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			cmd, err := getCommandByName(devObj.Data, tt.requestedType, tt.reqCommandName)
			if !tt.wantErr == (err != nil) {
				t.Errorf("TestGetCommand unexpected error for command: %v wantErr: %v err: %v", tt.requestedType, tt.wantErr, err)
				return
			} else if tt.wantErr {
				return
			}

			if cmd.Exec != nil {
				if cmd.Id != tt.retCommandName {
					t.Errorf("TestGetCommand error: command names do not match expected: %v actual: %v", tt.retCommandName, cmd.Id)
				}
			}
		})
	}

}

func TestGetBuildCommand(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""

	tests := []struct {
		name         string
		commandName  string
		execCommands []v1alpha2.Command
		wantCommand  v1alpha2.Command
		wantErr      bool
	}{
		{
			name:        "Case 1: Default Build Command",
			commandName: emptyString,
			execCommands: []v1alpha2.Command{
				{
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantCommand: v1alpha2.Command{
				CommandUnion: v1alpha2.CommandUnion{
					Exec: &v1alpha2.ExecCommand{
						LabeledCommand: v1alpha2.LabeledCommand{
							BaseCommand: v1alpha2.BaseCommand{
								Group: &v1alpha2.CommandGroup{Kind: buildGroup, IsDefault: util.GetBoolPtr(true)},
							},
						},
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case 2: Build Command passed through the odo flag",
			commandName: "flagcommand",
			execCommands: []v1alpha2.Command{
				{
					Id: "flagcommand",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: buildGroup},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
				{
					Id: "build command",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: buildGroup},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantCommand: v1alpha2.Command{
				Id: "flagcommand",
				CommandUnion: v1alpha2.CommandUnion{
					Exec: &v1alpha2.ExecCommand{
						LabeledCommand: v1alpha2.LabeledCommand{
							BaseCommand: v1alpha2.BaseCommand{
								Group: &v1alpha2.CommandGroup{Kind: buildGroup},
							},
						},
						CommandLine: command,
						Component:   component,
						WorkingDir:  workDir,
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case 3: Build Command not found",
			commandName: "customcommand123",
			execCommands: []v1alpha2.Command{
				{
					Id: "build command",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: buildGroup},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := parser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.execCommands)
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]v1alpha2.Component{testingutil.GetFakeContainerComponent(component)})
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			command, err := GetBuildCommand(devObj.Data, tt.commandName)

			if !tt.wantErr == (err != nil) {
				t.Errorf("TestGetBuildCommand: unexpected error for command \"%v\" expected: %v actual: %v", tt.commandName, tt.wantErr, err)
			} else if !tt.wantErr && !reflect.DeepEqual(tt.wantCommand, command) {
				t.Errorf("TestGetBuildCommand: unexpected command returned: %v", pretty.Compare(tt.wantCommand, command))
			}

		})
	}

}

func TestGetDebugCommand(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""

	var emptyCommand v1alpha2.Command

	tests := []struct {
		name         string
		commandName  string
		execCommands []v1alpha2.Command
		wantErr      bool
	}{
		{
			name:        "Case: Default Debug Command",
			commandName: emptyString,
			execCommands: []v1alpha2.Command{
				{
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{
										IsDefault: util.GetBoolPtr(true),
										Kind:      v1alpha2.DebugCommandGroupKind,
									},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Custom Debug Command",
			commandName: "customdebugcommand",
			execCommands: []v1alpha2.Command{
				{
					Id: "customdebugcommand",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{
										IsDefault: util.GetBoolPtr(false),
										Kind:      v1alpha2.DebugCommandGroupKind,
									},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Missing Debug Command",
			commandName: "customcommand123",
			execCommands: []v1alpha2.Command{
				{
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{
										IsDefault: util.GetBoolPtr(true),
										Kind:      v1alpha2.BuildCommandGroupKind,
									},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := parser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]v1alpha2.Component{testingutil.GetFakeContainerComponent(component)})
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.execCommands)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			command, err := GetDebugCommand(devObj.Data, tt.commandName)

			if tt.wantErr && err == nil {
				t.Errorf("Error was expected but got no error")
			} else if !tt.wantErr {
				if err != nil {
					t.Errorf("TestGetDebugCommand: unexpected error for command \"%v\" expected: %v actual: %v", tt.commandName, tt.wantErr, err)
				} else if reflect.DeepEqual(emptyCommand, command) {
					t.Errorf("TestGetDebugCommand: unexpected empty command returned for command: %v", tt.commandName)
				}
			}
		})
	}
}

func TestGetTestCommand(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""

	var emptyCommand v1alpha2.Command

	tests := []struct {
		name         string
		commandName  string
		execCommands []v1alpha2.Command
		wantErr      bool
	}{
		{
			name:        "Case: Default Test Command",
			commandName: emptyString,
			execCommands: []v1alpha2.Command{
				{
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{
										IsDefault: util.GetBoolPtr(true),
										Kind:      v1alpha2.TestCommandGroupKind,
									},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Custom Test Command",
			commandName: "customtestcommand",
			execCommands: []v1alpha2.Command{
				{
					Id: "customtestcommand",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{
										IsDefault: util.GetBoolPtr(false),
										Kind:      v1alpha2.TestCommandGroupKind,
									},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case: Missing Test Command",
			commandName: "customcommand123",
			execCommands: []v1alpha2.Command{
				{
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{
										IsDefault: util.GetBoolPtr(true),
										Kind:      v1alpha2.BuildCommandGroupKind,
									},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := parser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]v1alpha2.Component{testingutil.GetFakeContainerComponent(component)})
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.execCommands)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			command, err := GetTestCommand(devObj.Data, tt.commandName)

			if tt.wantErr && err == nil {
				t.Errorf("Error was expected but got no error")
			} else if !tt.wantErr {
				if err != nil {
					t.Errorf("TestGetTestCommand: unexpected error for command \"%v\" expected: %v actual: %v", tt.commandName, tt.wantErr, err)
				} else if reflect.DeepEqual(emptyCommand, command) {
					t.Errorf("TestGetTestCommand: unexpected empty command returned for command: %v", tt.commandName)
				}
			}
		})
	}
}

func TestGetRunCommand(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""

	var emptyCommand v1alpha2.Command

	tests := []struct {
		name         string
		commandName  string
		execCommands []v1alpha2.Command
		wantErr      bool
	}{
		{
			name:        "Case 1: Default Run Command",
			commandName: emptyString,
			execCommands: []v1alpha2.Command{
				{
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup, IsDefault: util.GetBoolPtr(true)},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case 2: Run Command passed through odo flag",
			commandName: "flagcommand",
			execCommands: []v1alpha2.Command{
				{
					Id: "flagcommand",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
				{
					Id: "run command",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "Case 3: Missing Run Command",
			commandName: "",
			execCommands: []v1alpha2.Command{
				{
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: buildGroup},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := parser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]v1alpha2.Component{testingutil.GetFakeContainerComponent(component)})
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.execCommands)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			command, err := GetRunCommand(devObj.Data, tt.commandName)

			if !tt.wantErr == (err != nil) {
				t.Errorf("TestGetRunCommand: unexpected error for command \"%v\" expected: %v actual: %v", tt.commandName, tt.wantErr, err)
			} else if !tt.wantErr && reflect.DeepEqual(emptyCommand, command) {
				t.Errorf("TestGetRunCommand: unexpected empty command returned for command: %v", tt.commandName)
			}
		})
	}

}

func TestValidateAndGetDebugCommands(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""

	execCommands := []v1alpha2.Command{
		{
			CommandUnion: v1alpha2.CommandUnion{
				Exec: &v1alpha2.ExecCommand{
					LabeledCommand: v1alpha2.LabeledCommand{
						BaseCommand: v1alpha2.BaseCommand{
							Group: &v1alpha2.CommandGroup{
								IsDefault: util.GetBoolPtr(true),
								Kind:      v1alpha2.DebugCommandGroupKind,
							},
						},
					},
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
				},
			},
		},
		{
			Id: "customdebugcommand",
			CommandUnion: v1alpha2.CommandUnion{
				Exec: &v1alpha2.ExecCommand{
					LabeledCommand: v1alpha2.LabeledCommand{
						BaseCommand: v1alpha2.BaseCommand{
							Group: &v1alpha2.CommandGroup{
								IsDefault: util.GetBoolPtr(false),
								Kind:      v1alpha2.DebugCommandGroupKind,
							},
						},
					},
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
				},
			},
		},
	}

	tests := []struct {
		name          string
		debugCommand  string
		componentType v1alpha2.ComponentType
		wantErr       bool
	}{
		{
			name:          "Case: Default Devfile Commands",
			debugCommand:  emptyString,
			componentType: v1alpha2.ContainerComponentType,
			wantErr:       false,
		},
		{
			name:          "Case: provided debug Command",
			debugCommand:  "customdebugcommand",
			componentType: v1alpha2.ContainerComponentType,
			wantErr:       false,
		},
		{
			name:          "Case: invalid debug Command",
			debugCommand:  "invaliddebugcommand",
			componentType: v1alpha2.ContainerComponentType,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := parser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]v1alpha2.Component{testingutil.GetFakeContainerComponent(component)})
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(execCommands)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			debugCommand, err := ValidateAndGetDebugCommands(devObj.Data, tt.debugCommand)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Error was expected but got no error")
				} else {
					return
				}
			} else {
				if err != nil {
					t.Errorf("TestValidateAndGetDebugDevfileCommands: unexpected error %v", err)
				}
			}

			if !reflect.DeepEqual(nil, debugCommand) && debugCommand.Id != tt.debugCommand {
				t.Errorf("TestValidateAndGetDebugDevfileCommands name of debug command is wrong want: %v got: %v", tt.debugCommand, debugCommand.Id)
			}
		})
	}
}

func TestValidateAndGetPushCommands(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""

	execCommands := []v1alpha2.Command{
		{
			Id: "run command",
			CommandUnion: v1alpha2.CommandUnion{
				Exec: &v1alpha2.ExecCommand{
					LabeledCommand: v1alpha2.LabeledCommand{
						BaseCommand: v1alpha2.BaseCommand{
							Group: &v1alpha2.CommandGroup{
								Kind:      runGroup,
								IsDefault: util.GetBoolPtr(true),
							},
						},
					},
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
				},
			},
		},

		{
			Id: "build command",
			CommandUnion: v1alpha2.CommandUnion{
				Exec: &v1alpha2.ExecCommand{
					LabeledCommand: v1alpha2.LabeledCommand{
						BaseCommand: v1alpha2.BaseCommand{
							Group: &v1alpha2.CommandGroup{Kind: buildGroup},
						},
					},
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
				},
			},
		},

		{
			Id: "customcommand",
			CommandUnion: v1alpha2.CommandUnion{
				Exec: &v1alpha2.ExecCommand{
					LabeledCommand: v1alpha2.LabeledCommand{
						BaseCommand: v1alpha2.BaseCommand{
							Group: &v1alpha2.CommandGroup{Kind: runGroup},
						},
					},
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
				},
			},
		},
	}

	wrongCompTypeCmd := v1alpha2.Command{

		Id: "wrong",
		CommandUnion: v1alpha2.CommandUnion{
			Exec: &v1alpha2.ExecCommand{
				LabeledCommand: v1alpha2.LabeledCommand{
					BaseCommand: v1alpha2.BaseCommand{
						Group: &v1alpha2.CommandGroup{Kind: runGroup},
					},
				},
				CommandLine: command,
				Component:   "",
				WorkingDir:  workDir,
			},
		},
	}

	tests := []struct {
		name                string
		buildCommand        string
		runCommand          string
		execCommands        []v1alpha2.Command
		numberOfCommands    int
		missingBuildCommand bool
		wantErr             bool
	}{
		{
			name:             "Case 1: Default Devfile Commands",
			buildCommand:     emptyString,
			runCommand:       emptyString,
			execCommands:     execCommands,
			numberOfCommands: 2,
			wantErr:          false,
		},
		{
			name:             "Case 2: Default Build Command, and Provided Run Command",
			buildCommand:     emptyString,
			runCommand:       "customcommand",
			execCommands:     execCommands,
			numberOfCommands: 2,
			wantErr:          false,
		},
		{
			name:             "Case 3: Empty Component",
			buildCommand:     "customcommand",
			runCommand:       "customcommand",
			execCommands:     append(execCommands, wrongCompTypeCmd),
			numberOfCommands: 0,
			wantErr:          true,
		},
		{
			name:             "Case 4: Provided Wrong Build Command and Provided Run Command",
			buildCommand:     "customcommand123",
			runCommand:       "customcommand",
			execCommands:     execCommands,
			numberOfCommands: 1,
			wantErr:          true,
		},
		{
			name:         "Case 5: Missing Build Command, and Provided Run Command",
			buildCommand: emptyString,
			runCommand:   "customcommand",
			execCommands: []v1alpha2.Command{
				{
					Id: "customcommand",
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{
							LabeledCommand: v1alpha2.LabeledCommand{
								BaseCommand: v1alpha2.BaseCommand{
									Group: &v1alpha2.CommandGroup{Kind: runGroup},
								},
							},
							Component:   component,
							CommandLine: command,
						},
					},
				},
			},
			numberOfCommands: 1,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := parser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.execCommands)
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents(([]v1alpha2.Component{testingutil.GetFakeContainerComponent(component)}))
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			pushCommands, err := ValidateAndGetPushCommands(devObj.Data, tt.buildCommand, tt.runCommand)
			if !tt.wantErr == (err != nil) {
				t.Errorf("TestValidateAndGetPushDevfileCommands unexpected error when validating commands wantErr: %v err: %v", tt.wantErr, err)
			} else if tt.wantErr && err != nil {
				return
			}

			if len(pushCommands) != tt.numberOfCommands {
				t.Errorf("TestValidateAndGetPushDevfileCommands error: wrong number of validated commands expected: %v actual :%v", tt.numberOfCommands, len(pushCommands))
			}
		})
	}

}

func TestValidateAndGetTestCommands(t *testing.T) {

	command := "ls -la"
	component := "alias1"
	workDir := "/"
	emptyString := ""

	execCommands := []v1alpha2.Command{
		{
			CommandUnion: v1alpha2.CommandUnion{
				Exec: &v1alpha2.ExecCommand{
					LabeledCommand: v1alpha2.LabeledCommand{
						BaseCommand: v1alpha2.BaseCommand{
							Group: &v1alpha2.CommandGroup{
								IsDefault: util.GetBoolPtr(true),
								Kind:      v1alpha2.TestCommandGroupKind,
							},
						},
					},
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
				},
			},
		},
		{
			Id: "customtestcommand",
			CommandUnion: v1alpha2.CommandUnion{
				Exec: &v1alpha2.ExecCommand{
					LabeledCommand: v1alpha2.LabeledCommand{
						BaseCommand: v1alpha2.BaseCommand{
							Group: &v1alpha2.CommandGroup{
								IsDefault: util.GetBoolPtr(false),
								Kind:      v1alpha2.TestCommandGroupKind,
							},
						},
					},
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
				},
			},
		},
	}

	tests := []struct {
		name          string
		testCommand   string
		componentType v1alpha2.ComponentType
		wantErr       bool
	}{
		{
			name:          "Case: Default Devfile Commands",
			testCommand:   emptyString,
			componentType: v1alpha2.ContainerComponentType,
			wantErr:       false,
		},
		{
			name:          "Case: provided test Command",
			testCommand:   "customtestcommand",
			componentType: v1alpha2.ContainerComponentType,
			wantErr:       false,
		},
		{
			name:          "Case: invalid test Command",
			testCommand:   "invalidtestcommand",
			componentType: v1alpha2.ContainerComponentType,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := parser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]v1alpha2.Component{testingutil.GetFakeContainerComponent(component)})
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(execCommands)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			testCommand, err := ValidateAndGetTestCommands(devObj.Data, tt.testCommand)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Error was expected but got no error")
				} else {
					return
				}
			} else {
				if err != nil {
					t.Errorf("TestValidateAndGetTestDevfileCommands: unexpected error %v", err)
				}
			}

			if !reflect.DeepEqual(nil, testCommand) && testCommand.Id != tt.testCommand {
				t.Errorf("TestValidateAndGetTestDevfileCommands name of test command is wrong want: %v got: %v", tt.testCommand, testCommand.Id)
			}
		})
	}
}

func TestGetContainerComponentsForCommand(t *testing.T) {
	devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion220))
	devfileData.SetMetadata(devfilepkg.DevfileMetadata{Name: "my-app"})
	_ = devfileData.AddComponents([]v1alpha2.Component{
		{
			Name: "my-container1",
			ComponentUnion: v1alpha2.ComponentUnion{
				Container: &v1alpha2.ContainerComponent{
					Container: v1alpha2.Container{Image: "my-image"},
				},
			},
		},
		{
			Name: "my-container2",
			ComponentUnion: v1alpha2.ComponentUnion{
				Container: &v1alpha2.ContainerComponent{
					Container: v1alpha2.Container{Image: "my-image"},
				},
			},
		},
		{
			Name: "my-container3",
			ComponentUnion: v1alpha2.ComponentUnion{
				Container: &v1alpha2.ContainerComponent{
					Container: v1alpha2.Container{Image: "my-image"},
				},
			},
		},
		{
			Name: "my-image1",
			ComponentUnion: v1alpha2.ComponentUnion{
				Image: &v1alpha2.ImageComponent{
					Image: v1alpha2.Image{
						ImageName: "my-image",
					},
				},
			},
		},
		{
			Name: "my-k8s",
			ComponentUnion: v1alpha2.ComponentUnion{
				Kubernetes: &v1alpha2.KubernetesComponent{
					K8sLikeComponent: v1alpha2.K8sLikeComponent{
						K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
							Inlined: "---",
						},
					},
				},
			},
		},
	})

	customCmd := v1alpha2.Command{
		Id: "custom",
		CommandUnion: v1alpha2.CommandUnion{
			Custom: &v1alpha2.CustomCommand{},
		},
	}
	execCmdCont1 := v1alpha2.Command{
		Id: "execCmdCont1",
		CommandUnion: v1alpha2.CommandUnion{
			Exec: &v1alpha2.ExecCommand{Component: "my-container1"},
		},
	}
	execCmdCont2 := v1alpha2.Command{
		Id: "execCmdCont2",
		CommandUnion: v1alpha2.CommandUnion{
			Exec: &v1alpha2.ExecCommand{Component: "my-container2"},
		},
	}
	execCmdCont3 := v1alpha2.Command{
		Id: "execCmdCont3",
		CommandUnion: v1alpha2.CommandUnion{
			Exec: &v1alpha2.ExecCommand{Component: "my-container3"},
		},
	}
	applyCmdCont1 := v1alpha2.Command{
		Id: "applyCmdCont1",
		CommandUnion: v1alpha2.CommandUnion{
			Apply: &v1alpha2.ApplyCommand{Component: "my-container1"},
		},
	}
	applyCmdImg1 := v1alpha2.Command{
		Id: "applyCmdImg1",
		CommandUnion: v1alpha2.CommandUnion{
			Apply: &v1alpha2.ApplyCommand{Component: "my-image1"},
		},
	}
	applyK8s1 := v1alpha2.Command{
		Id: "applyK8s1",
		CommandUnion: v1alpha2.CommandUnion{
			Apply: &v1alpha2.ApplyCommand{Component: "my-k8s"},
		},
	}
	childCompositeCmd := v1alpha2.Command{
		Id: "child-composite",
		CommandUnion: v1alpha2.CommandUnion{
			Composite: &v1alpha2.CompositeCommand{
				Commands: []string{execCmdCont3.Id, applyK8s1.Id},
			},
		},
	}

	_ = devfileData.AddCommands([]v1alpha2.Command{
		customCmd, execCmdCont1, execCmdCont2, execCmdCont3, applyCmdCont1, applyCmdImg1, applyK8s1, childCompositeCmd})

	devfileObj := parser.DevfileObj{Data: devfileData}

	type args struct {
		cmd v1alpha2.Command
	}
	for _, tt := range []struct {
		name    string
		args    args
		wantErr bool
		want    []string
	}{
		{
			name: "zero -value command",
			args: args{cmd: v1alpha2.Command{}},
		},
		{
			name:    "GetCommandType returning an error",
			args:    args{cmd: v1alpha2.Command{Id: "unknown"}},
			wantErr: true,
		},
		{
			name:    "non-supported command type",
			args:    args{cmd: customCmd},
			wantErr: true,
		},
		{
			name: "exec command matching existing container component",
			args: args{cmd: execCmdCont1},
			want: []string{"my-container1"},
		},
		{
			name: "exec command not matching existing container component",
			args: args{
				cmd: v1alpha2.Command{
					CommandUnion: v1alpha2.CommandUnion{
						Exec: &v1alpha2.ExecCommand{Component: "my-k8s"},
					},
				},
			},
		},
		{
			name: "apply command matching existing container component",
			args: args{cmd: applyCmdCont1},
			want: []string{"my-container1"},
		},
		{
			name: "apply command not matching existing container component",
			args: args{cmd: applyCmdImg1},
		},
		{
			name: "composite command with one command missing not declared in Devfile commands",
			args: args{
				cmd: v1alpha2.Command{
					CommandUnion: v1alpha2.CommandUnion{
						Composite: &v1alpha2.CompositeCommand{
							Commands: []string{execCmdCont1.Id, "a-command-not-found-in-devfile"},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "composite command with at least one unsupported component",
			args: args{
				cmd: v1alpha2.Command{
					CommandUnion: v1alpha2.CommandUnion{
						Composite: &v1alpha2.CompositeCommand{
							Commands: []string{childCompositeCmd.Id, customCmd.Id, applyCmdImg1.Id},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "composite command with no non-unsupported component",
			args: args{
				cmd: v1alpha2.Command{
					CommandUnion: v1alpha2.CommandUnion{
						Composite: &v1alpha2.CompositeCommand{
							Commands: []string{childCompositeCmd.Id, applyCmdImg1.Id, execCmdCont1.Id},
						},
					},
				},
			},
			want: []string{"my-container3", "my-container1"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetContainerComponentsForCommand(devfileObj, tt.args.cmd)

			if tt.wantErr != (err != nil) {
				t.Errorf("unexpected error, wantErr: %v, err: %v", tt.wantErr, err)
			}
			if !reflect.DeepEqual(tt.want, got) {
				t.Errorf("want: %v, got %v", tt.want, got)
			}
		})
	}
}

func getExecCommand(id string, group v1alpha2.CommandGroupKind) v1alpha2.Command {
	if len(id) == 0 {
		id = fmt.Sprintf("%s-%s", "cmd", dfutil.GenerateRandomString(10))
	}
	commands := [...]string{"ls -la", "pwd"}
	components := [...]string{"alias1", "alias2"}
	workDir := [...]string{"/", "/root"}

	return v1alpha2.Command{
		Id: id,
		CommandUnion: v1alpha2.CommandUnion{
			Exec: &v1alpha2.ExecCommand{
				LabeledCommand: v1alpha2.LabeledCommand{
					BaseCommand: v1alpha2.BaseCommand{
						Group: &v1alpha2.CommandGroup{Kind: group},
					},
				},
				CommandLine: commands[0],
				Component:   components[0],
				WorkingDir:  workDir[0],
			},
		},
	}

}
