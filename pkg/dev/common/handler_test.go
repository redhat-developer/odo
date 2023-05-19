package common

import (
	"context"
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/config"
	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"github.com/redhat-developer/odo/pkg/devfile/image"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/platform"
	"github.com/redhat-developer/odo/pkg/podman"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/sethvargo/go-envconfig"
	"k8s.io/utils/pointer"
)

var (
	container1 = v1alpha2.Component{
		Name: "my-container",
		ComponentUnion: v1alpha2.ComponentUnion{
			Container: &v1alpha2.ContainerComponent{
				Container: v1alpha2.Container{
					Image: "my-image",
				},
				Endpoints: []v1alpha2.Endpoint{
					{
						Name:       "http",
						TargetPort: 8080,
					},
					{
						Name:       "debug",
						TargetPort: 5858,
					},
				},
			},
		},
	}
	defaultBuildCommand = v1alpha2.Command{
		Id: "my-build",
		CommandUnion: v1alpha2.CommandUnion{
			Exec: &v1alpha2.ExecCommand{
				CommandLine: "go build main.go",
				Component:   "my-container",
				LabeledCommand: v1alpha2.LabeledCommand{
					BaseCommand: v1alpha2.BaseCommand{
						Group: &v1alpha2.CommandGroup{
							Kind:      v1alpha2.BuildCommandGroupKind,
							IsDefault: pointer.Bool(true),
						},
					},
				},
			},
		},
	}
	kubernetesDeploy = v1alpha2.Component{
		Name: "kubernetes-deploy",
		ComponentUnion: v1alpha2.ComponentUnion{
			Kubernetes: &v1alpha2.KubernetesComponent{
				K8sLikeComponent: v1alpha2.K8sLikeComponent{
					K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
						Inlined: `spec: {}`,
					},
				},
			},
		},
	}

	openshiftDeploy = v1alpha2.Component{
		Name: "openshift-deploy",
		ComponentUnion: v1alpha2.ComponentUnion{
			Openshift: &v1alpha2.OpenshiftComponent{
				K8sLikeComponent: v1alpha2.K8sLikeComponent{
					K8sLikeComponentLocation: v1alpha2.K8sLikeComponentLocation{
						Inlined: `spec: {}`,
					},
				},
			},
		},
	}
	defaultDeployCommandKubernetes = v1alpha2.Command{
		Id: "my-deploy",
		CommandUnion: v1alpha2.CommandUnion{
			Apply: &v1alpha2.ApplyCommand{
				Component: "kubernetes-deploy",
				LabeledCommand: v1alpha2.LabeledCommand{
					BaseCommand: v1alpha2.BaseCommand{
						Group: &v1alpha2.CommandGroup{
							Kind:      v1alpha2.DeployCommandGroupKind,
							IsDefault: pointer.Bool(true),
						},
					},
				},
			},
		},
	}
	defaultDeployCommandOpenshift = v1alpha2.Command{
		Id: "my-deploy",
		CommandUnion: v1alpha2.CommandUnion{
			Apply: &v1alpha2.ApplyCommand{
				Component: "openshift-deploy",
				LabeledCommand: v1alpha2.LabeledCommand{
					BaseCommand: v1alpha2.BaseCommand{
						Group: &v1alpha2.CommandGroup{
							Kind:      v1alpha2.DeployCommandGroupKind,
							IsDefault: pointer.Bool(true),
						},
					},
				},
			},
		},
	}

	imageDeploy = v1alpha2.Component{
		Name: "image-deploy",
		ComponentUnion: v1alpha2.ComponentUnion{
			Image: &v1alpha2.ImageComponent{
				Image: v1alpha2.Image{
					ImageName: "golang",
				},
			},
		},
	}
	defaultDeployCommandImage = v1alpha2.Command{
		Id: "my-deploy",
		CommandUnion: v1alpha2.CommandUnion{
			Apply: &v1alpha2.ApplyCommand{
				Component: "image-deploy",
				LabeledCommand: v1alpha2.LabeledCommand{
					BaseCommand: v1alpha2.BaseCommand{
						Group: &v1alpha2.CommandGroup{
							Kind:      v1alpha2.DeployCommandGroupKind,
							IsDefault: pointer.Bool(true),
						},
					},
				},
			},
		},
	}
)

func TestApplyKubernetes(t *testing.T) {

	tests := []struct {
		name            string
		devfileObj      func() parser.DevfileObj
		appName         string
		componentName   string
		podName         string
		msg             string
		show            bool
		componentExists bool
		platformClient  func(ctrl *gomock.Controller) platform.Client
		execClient      func(ctrl *gomock.Controller) exec.Client
		wantErr         bool
	}{
		{
			name: "empty Devfile",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				devfileObj := parser.DevfileObj{
					Data: devfileData,
				}
				return devfileObj
			},
			platformClient: func(ctrl *gomock.Controller) platform.Client {
				client := kclient.NewMockClientInterface(ctrl)
				// Nothing happens as there is no Deploy commands on the Devfile
				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				// Nothing happens as there is no Deploy commands on the Devfile
				return client
			},
			wantErr: true,
		},
		{
			name: "Devfile with Apply Kubernetes command",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				devfileData.AddComponents([]v1alpha2.Component{kubernetesDeploy})
				devfileData.AddCommands([]v1alpha2.Command{defaultDeployCommandKubernetes})

				devfileObj := parser.DevfileObj{
					Data: devfileData,
				}
				return devfileObj
			},
			platformClient: func(ctrl *gomock.Controller) platform.Client {
				client := kclient.NewMockClientInterface(ctrl)

				// Expects the resource is applied to the cluster
				client.EXPECT().GetRestMappingFromUnstructured(gomock.Any())
				client.EXPECT().IsServiceBindingSupported()
				client.EXPECT().PatchDynamicResource(gomock.Any())

				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				return client
			},
		},
		{
			name: "Devfile with Apply Kubernetes command on podman",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				devfileData.AddComponents([]v1alpha2.Component{kubernetesDeploy})
				devfileData.AddCommands([]v1alpha2.Command{defaultDeployCommandKubernetes})

				devfileObj := parser.DevfileObj{
					Data: devfileData,
				}
				return devfileObj
			},
			platformClient: func(ctrl *gomock.Controller) platform.Client {
				client := podman.NewMockClient(ctrl)
				// Nothing, as this is not implemented on podman
				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				return client
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			cmdHandler := &RunHandler{
				FS:             filesystem.NewFakeFs(),
				ExecClient:     tt.execClient(ctrl),
				PlatformClient: tt.platformClient(ctrl),
				Devfile:        tt.devfileObj(),
			}
			ctx := context.Background()
			err := libdevfile.Deploy(ctx, tt.devfileObj(), cmdHandler)
			if (err != nil) != tt.wantErr {
				t.Errorf("Err expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestApplyOpenshift(t *testing.T) {

	tests := []struct {
		name            string
		devfileObj      func() parser.DevfileObj
		appName         string
		componentName   string
		podName         string
		msg             string
		show            bool
		componentExists bool
		platformClient  func(ctrl *gomock.Controller) platform.Client
		execClient      func(ctrl *gomock.Controller) exec.Client
		wantErr         bool
	}{
		{
			name: "empty Devfile",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				devfileObj := parser.DevfileObj{
					Data: devfileData,
				}
				return devfileObj
			},
			platformClient: func(ctrl *gomock.Controller) platform.Client {
				client := platform.NewMockClient(ctrl)
				// Nothing happens as there is no Deploy commands on the Devfile
				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				// Nothing happens as there is no Deploy commands on the Devfile
				return client
			},
			wantErr: true,
		},
		{
			name: "Devfile with Deploy OpenShift command",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				devfileData.AddComponents([]v1alpha2.Component{openshiftDeploy})
				devfileData.AddCommands([]v1alpha2.Command{defaultDeployCommandOpenshift})

				devfileObj := parser.DevfileObj{
					Data: devfileData,
				}
				return devfileObj
			},
			platformClient: func(ctrl *gomock.Controller) platform.Client {
				client := kclient.NewMockClientInterface(ctrl)

				// Expects the resource is applied to the cluster
				client.EXPECT().GetRestMappingFromUnstructured(gomock.Any())
				client.EXPECT().IsServiceBindingSupported()
				client.EXPECT().PatchDynamicResource(gomock.Any())

				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				return client
			},
		},
		{
			name: "Devfile with Deploy OpenShift command on Podman",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				devfileData.AddComponents([]v1alpha2.Component{openshiftDeploy})
				devfileData.AddCommands([]v1alpha2.Command{defaultDeployCommandOpenshift})

				devfileObj := parser.DevfileObj{
					Data: devfileData,
				}
				return devfileObj
			},
			platformClient: func(ctrl *gomock.Controller) platform.Client {
				client := podman.NewMockClient(ctrl)
				// Nothing, as this is not implemented on Podman
				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				return client
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			cmdHandler := &RunHandler{
				FS:             filesystem.NewFakeFs(),
				ExecClient:     tt.execClient(ctrl),
				PlatformClient: tt.platformClient(ctrl),
				Devfile:        tt.devfileObj(),
			}
			ctx := context.Background()
			err := libdevfile.Deploy(ctx, tt.devfileObj(), cmdHandler)
			if (err != nil) != tt.wantErr {
				t.Errorf("Err expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestApplyImage(t *testing.T) {

	tests := []struct {
		name            string
		devfileObj      func() parser.DevfileObj
		appName         string
		componentName   string
		podName         string
		msg             string
		show            bool
		componentExists bool
		platformClient  func(ctrl *gomock.Controller) platform.Client
		execClient      func(ctrl *gomock.Controller) exec.Client
		imageBackend    func(ctrl *gomock.Controller) image.Backend
		env             map[string]string
		wantErr         bool
	}{
		{
			name: "empty Devfile",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				devfileObj := parser.DevfileObj{
					Data: devfileData,
				}
				return devfileObj
			},
			platformClient: func(ctrl *gomock.Controller) platform.Client {
				client := platform.NewMockClient(ctrl)
				// Nothing happens as there is no Deploy commands on the Devfile
				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				// Nothing happens as there is no Deploy commands on the Devfile
				return client
			},
			imageBackend: func(ctrl *gomock.Controller) image.Backend {
				return nil
			},
			wantErr: true,
		},
		{
			name: "Devfile with Apply Image command",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				devfileData.AddComponents([]v1alpha2.Component{imageDeploy})
				devfileData.AddCommands([]v1alpha2.Command{defaultDeployCommandImage})

				devfileObj := parser.DevfileObj{
					Data: devfileData,
				}
				return devfileObj
			},
			platformClient: func(ctrl *gomock.Controller) platform.Client {
				client := kclient.NewMockClientInterface(ctrl)
				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				return client
			},
			imageBackend: func(ctrl *gomock.Controller) image.Backend {
				client := image.NewMockBackend(ctrl)
				client.EXPECT().Build(gomock.Any(), gomock.Any(), gomock.Any())
				client.EXPECT().Push("golang")
				return client

			},
		},
		{
			name: "Devfile with Apply Image command and push disabled",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				devfileData.AddComponents([]v1alpha2.Component{imageDeploy})
				devfileData.AddCommands([]v1alpha2.Command{defaultDeployCommandImage})

				devfileObj := parser.DevfileObj{
					Data: devfileData,
				}
				return devfileObj
			},
			env: map[string]string{
				"ODO_PUSH_IMAGES": "false",
			},
			platformClient: func(ctrl *gomock.Controller) platform.Client {
				client := kclient.NewMockClientInterface(ctrl)
				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				return client
			},
			imageBackend: func(ctrl *gomock.Controller) image.Backend {
				client := image.NewMockBackend(ctrl)
				client.EXPECT().Build(gomock.Any(), gomock.Any(), gomock.Any())
				return client

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			envConfig, err := config.GetConfigurationWith(envconfig.MapLookuper(tt.env))
			if err != nil {
				t.Error("error reading config")
			}
			ctx = envcontext.WithEnvConfig(ctx, *envConfig)
			ctx = odocontext.WithDevfilePath(ctx, "/devfile.yaml")
			ctrl := gomock.NewController(t)
			cmdHandler := &RunHandler{
				Ctx:            ctx,
				FS:             filesystem.NewFakeFs(),
				ExecClient:     tt.execClient(ctrl),
				PlatformClient: tt.platformClient(ctrl),
				ImageBackend:   tt.imageBackend(ctrl),
				Devfile:        tt.devfileObj(),
			}
			err = libdevfile.Deploy(ctx, tt.devfileObj(), cmdHandler)
			if (err != nil) != tt.wantErr {
				t.Errorf("Err expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestExecute(t *testing.T) {

	tests := []struct {
		name            string
		devfileObj      func() parser.DevfileObj
		appName         string
		componentName   string
		podName         string
		msg             string
		show            bool
		componentExists bool
		platformClient  func(ctrl *gomock.Controller) platform.Client
		execClient      func(ctrl *gomock.Controller) exec.Client
		wantErr         bool
	}{
		{
			name: "empty Devfile",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				devfileObj := parser.DevfileObj{
					Data: devfileData,
				}
				return devfileObj
			},
			platformClient: func(ctrl *gomock.Controller) platform.Client {
				client := kclient.NewMockClientInterface(ctrl)
				// Nothing happens as there is no default Build command on the Devfile
				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				// Nothing happens as there is no default Build command on the Devfile
				return client
			},
			wantErr: false,
		},
		{
			name:    "Devfile with exec Build command",
			podName: "a-pod-name",
			show:    true,

			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				devfileData.AddComponents([]v1alpha2.Component{container1})
				devfileData.AddCommands([]v1alpha2.Command{defaultBuildCommand})

				devfileObj := parser.DevfileObj{
					Data: devfileData,
				}
				return devfileObj
			},
			platformClient: func(ctrl *gomock.Controller) platform.Client {
				client := platform.NewMockClient(ctrl)
				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Any(), "a-pod-name", "my-container", false, gomock.Any(), gomock.Any()).AnyTimes()
				return client
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			cmdHandler := &RunHandler{
				FS:             filesystem.NewFakeFs(),
				ExecClient:     tt.execClient(ctrl),
				PlatformClient: tt.platformClient(ctrl),
				Devfile:        tt.devfileObj(),
				PodName:        tt.podName,
				AppName:        tt.appName,
				ComponentName:  tt.componentName,
			}
			ctx := context.Background()
			err := libdevfile.Build(ctx, tt.devfileObj(), "", cmdHandler)
			if (err != nil) != tt.wantErr {
				t.Errorf("Err expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}
