package libdevfile

import (
	"os"
	"reflect"
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/devfile/library/pkg/devfile/parser/data"
	devfileFileSystem "github.com/devfile/library/pkg/testingutil/filesystem"
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

func TestGetComponentResourceManifestContentWithVariablesResolved(t *testing.T) {
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

			got, err := GetComponentResourceManifestContentWithVariablesResolved(tt.devfileObjFunc(), cmpName, "", fakeFs)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetComponentResourceManifestContentWithVariablesResolved() error = %v, wantErr %v",
					err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetComponentResourceManifestContentWithVariablesResolved() got = %v, want %v",
					got, tt.want)
			}
		})
	}
}
