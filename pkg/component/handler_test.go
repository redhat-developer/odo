package component

import (
	"context"
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/config"
	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"github.com/redhat-developer/odo/pkg/configAutomount"
	"github.com/redhat-developer/odo/pkg/devfile/image"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/platform"
	"github.com/redhat-developer/odo/pkg/podman"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/sethvargo/go-envconfig"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/utils/pointer"
)

var (
	// Components

	containerComponent = v1alpha2.Component{
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

	kubernetesComponent = v1alpha2.Component{
		Name: "my-kubernetes",
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

	openshiftComponent = v1alpha2.Component{
		Name: "my-openshift",
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

	imageComponent = v1alpha2.Component{
		Name: "my-image",
		ComponentUnion: v1alpha2.ComponentUnion{
			Image: &v1alpha2.ImageComponent{
				Image: v1alpha2.Image{
					ImageName: "golang",
				},
			},
		},
	}

	// Commands

	execOnContainer = v1alpha2.Command{
		Id: "my-exec-on-container",
		CommandUnion: v1alpha2.CommandUnion{
			Exec: &v1alpha2.ExecCommand{
				CommandLine: "go build main.go",
				Component:   "my-container",
			},
		},
	}

	applyKubernetes = v1alpha2.Command{
		Id: "my-apply-kubernetes",
		CommandUnion: v1alpha2.CommandUnion{
			Apply: &v1alpha2.ApplyCommand{
				Component: "my-kubernetes",
			},
		},
	}
	applyOpenshift = v1alpha2.Command{
		Id: "my-apply-openshift",
		CommandUnion: v1alpha2.CommandUnion{
			Apply: &v1alpha2.ApplyCommand{
				Component: "my-openshift",
			},
		},
	}

	applyImage = v1alpha2.Command{
		Id: "my-apply-image",
		CommandUnion: v1alpha2.CommandUnion{
			Apply: &v1alpha2.ApplyCommand{
				Component: "my-image",
			},
		},
	}
)

func CommandWithKind(command v1alpha2.Command, kind v1alpha2.CommandGroupKind, isDefault *bool) v1alpha2.Command {
	group := &v1alpha2.CommandGroup{
		Kind:      kind,
		IsDefault: isDefault,
	}

	if command.Exec != nil {
		command.Exec.Group = group
	}

	if command.Apply != nil {
		command.Apply.Group = group
	}

	if command.Composite != nil {
		command.Composite.Group = group
	}

	return command
}

func TestDeploy(t *testing.T) {

	appName := "app"
	componentName := "componentName"

	tests := []struct {
		name                  string
		devfileObj            func() parser.DevfileObj
		podName               string
		msg                   string
		show                  bool
		componentExists       bool
		platformClient        func(ctrl *gomock.Controller) platform.Client
		execClient            func(ctrl *gomock.Controller) exec.Client
		configAutomountClient func(ctrl *gomock.Controller) configAutomount.Client
		imageBackend          func(ctrl *gomock.Controller) image.Backend
		env                   map[string]string
		wantErr               bool
	}{
		{
			name: "Devfile with Apply Kubernetes command with Deploy kind on cluster",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				_ = devfileData.AddComponents([]v1alpha2.Component{
					kubernetesComponent,
				})
				_ = devfileData.AddCommands([]v1alpha2.Command{
					CommandWithKind(applyKubernetes, v1alpha2.DeployCommandGroupKind, pointer.Bool(true)),
				})

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
			configAutomountClient: func(ctrl *gomock.Controller) configAutomount.Client {
				return nil
			},
			imageBackend: func(ctrl *gomock.Controller) image.Backend {
				return nil
			},
		},
		{
			name: "Devfile with Apply Kubernetes command with Deploy kind on podman",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				_ = devfileData.AddComponents([]v1alpha2.Component{
					kubernetesComponent,
				})
				_ = devfileData.AddCommands([]v1alpha2.Command{
					CommandWithKind(applyKubernetes, v1alpha2.DeployCommandGroupKind, pointer.Bool(true)),
				})

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
			configAutomountClient: func(ctrl *gomock.Controller) configAutomount.Client {
				return nil
			},
			imageBackend: func(ctrl *gomock.Controller) image.Backend {
				return nil

			},
			wantErr: false,
		},

		{
			name: "Devfile with Apply Openshift command with Deploy kind on cluster",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				_ = devfileData.AddComponents([]v1alpha2.Component{
					openshiftComponent,
				})
				_ = devfileData.AddCommands([]v1alpha2.Command{
					CommandWithKind(applyOpenshift, v1alpha2.DeployCommandGroupKind, pointer.Bool(true)),
				})

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
			configAutomountClient: func(ctrl *gomock.Controller) configAutomount.Client {
				return nil
			},
			imageBackend: func(ctrl *gomock.Controller) image.Backend {
				return nil

			},
		},
		{
			name: "Devfile with Apply Openshift command with Deploy kind on podman",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				_ = devfileData.AddComponents([]v1alpha2.Component{
					openshiftComponent,
				})
				_ = devfileData.AddCommands([]v1alpha2.Command{
					CommandWithKind(applyOpenshift, v1alpha2.DeployCommandGroupKind, pointer.Bool(true)),
				})

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
			configAutomountClient: func(ctrl *gomock.Controller) configAutomount.Client {
				return nil
			},
			imageBackend: func(ctrl *gomock.Controller) image.Backend {
				return nil

			},
			wantErr: false,
		},

		{
			name: "Devfile with Apply Image command with Deploy kind on cluster",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				_ = devfileData.AddComponents([]v1alpha2.Component{
					imageComponent,
				})
				_ = devfileData.AddCommands([]v1alpha2.Command{
					CommandWithKind(applyImage, v1alpha2.DeployCommandGroupKind, pointer.Bool(true)),
				})

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
			configAutomountClient: func(ctrl *gomock.Controller) configAutomount.Client {
				return nil
			},
			imageBackend: func(ctrl *gomock.Controller) image.Backend {
				client := image.NewMockBackend(ctrl)
				client.EXPECT().Build(gomock.Any(), gomock.Any(), gomock.Any())
				client.EXPECT().Push("golang")
				return client

			},
		},
		{
			name: "Devfile with Apply Image command with Deploy kind on podman",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				_ = devfileData.AddComponents([]v1alpha2.Component{
					imageComponent,
				})
				_ = devfileData.AddCommands([]v1alpha2.Command{
					CommandWithKind(applyImage, v1alpha2.DeployCommandGroupKind, pointer.Bool(true)),
				})

				devfileObj := parser.DevfileObj{
					Data: devfileData,
				}
				return devfileObj
			},
			platformClient: func(ctrl *gomock.Controller) platform.Client {
				client := podman.NewMockClient(ctrl)
				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				return client
			},
			configAutomountClient: func(ctrl *gomock.Controller) configAutomount.Client {
				return nil
			},
			imageBackend: func(ctrl *gomock.Controller) image.Backend {
				client := image.NewMockBackend(ctrl)
				client.EXPECT().Build(gomock.Any(), gomock.Any(), gomock.Any())
				client.EXPECT().Push("golang")
				return client

			},
			wantErr: false,
		},

		{
			name: "Devfile with Apply Image command with Deploy kind on cluster and push disabled",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				_ = devfileData.AddComponents([]v1alpha2.Component{
					imageComponent,
				})
				_ = devfileData.AddCommands([]v1alpha2.Command{
					CommandWithKind(applyImage, v1alpha2.DeployCommandGroupKind, pointer.Bool(true)),
				})

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
			configAutomountClient: func(ctrl *gomock.Controller) configAutomount.Client {
				return nil
			},
			imageBackend: func(ctrl *gomock.Controller) image.Backend {
				client := image.NewMockBackend(ctrl)
				client.EXPECT().Build(gomock.Any(), gomock.Any(), gomock.Any())
				return client

			},
			env: map[string]string{
				"ODO_PUSH_IMAGES": "false",
			},
		},
		{
			name: "Devfile with Apply Image command with Deploy kind on podman and push disabled",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				_ = devfileData.AddComponents([]v1alpha2.Component{
					imageComponent,
				})
				_ = devfileData.AddCommands([]v1alpha2.Command{
					CommandWithKind(applyImage, v1alpha2.DeployCommandGroupKind, pointer.Bool(true)),
				})

				devfileObj := parser.DevfileObj{
					Data: devfileData,
				}
				return devfileObj
			},
			platformClient: func(ctrl *gomock.Controller) platform.Client {
				client := podman.NewMockClient(ctrl)
				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				return client
			},
			configAutomountClient: func(ctrl *gomock.Controller) configAutomount.Client {
				return nil
			},
			imageBackend: func(ctrl *gomock.Controller) image.Backend {
				client := image.NewMockBackend(ctrl)
				client.EXPECT().Build(gomock.Any(), gomock.Any(), gomock.Any())
				return client

			},
			env: map[string]string{
				"ODO_PUSH_IMAGES": "false",
			},
		},

		{
			name: "Devfile with Exec on Container command with Deploy kind on cluster",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				_ = devfileData.AddComponents([]v1alpha2.Component{
					containerComponent,
				})
				_ = devfileData.AddCommands([]v1alpha2.Command{
					CommandWithKind(execOnContainer, v1alpha2.DeployCommandGroupKind, pointer.Bool(true)),
				})

				devfileObj := parser.DevfileObj{
					Data: devfileData,
				}
				return devfileObj
			},
			platformClient: func(ctrl *gomock.Controller) platform.Client {
				client := kclient.NewMockClientInterface(ctrl)
				client.EXPECT().GetCurrentNamespacePolicy()
				client.EXPECT().ListJobs(gomock.Any()).Return(&batchv1.JobList{}, nil)
				createdJob := batchv1.Job{}
				createdJob.SetName("job")
				client.EXPECT().CreateJob(gomock.Any(), gomock.Any()).Return(&createdJob, nil)
				client.EXPECT().WaitForJobToComplete(gomock.Any())
				client.EXPECT().DeleteJob("job")
				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				return client
			},
			configAutomountClient: func(ctrl *gomock.Controller) configAutomount.Client {
				client := configAutomount.NewMockClient(ctrl)
				client.EXPECT().GetAutomountingVolumes()
				return client
			},
			imageBackend: func(ctrl *gomock.Controller) image.Backend {
				return nil

			},
		},
		{
			name: "Devfile with Exec on Container command with Deploy kind on podman",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				_ = devfileData.AddComponents([]v1alpha2.Component{
					containerComponent,
				})
				_ = devfileData.AddCommands([]v1alpha2.Command{
					CommandWithKind(execOnContainer, v1alpha2.DeployCommandGroupKind, pointer.Bool(true)),
				})

				devfileObj := parser.DevfileObj{
					Data: devfileData,
				}
				return devfileObj
			},
			platformClient: func(ctrl *gomock.Controller) platform.Client {
				client := podman.NewMockClient(ctrl)
				// Not implemented on Podman
				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				return client
			},
			configAutomountClient: func(ctrl *gomock.Controller) configAutomount.Client {
				client := configAutomount.NewMockClient(ctrl)
				return client
			},
			imageBackend: func(ctrl *gomock.Controller) image.Backend {
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx := context.Background()
			ctx = odocontext.WithDevfilePath(ctx, "/devfile.yaml")
			ctx = odocontext.WithApplication(ctx, appName)
			ctx = odocontext.WithComponentName(ctx, componentName)
			envConfig, err := config.GetConfigurationWith(envconfig.MapLookuper(tt.env))
			if err != nil {
				t.Error("error reading config")
			}
			ctx = envcontext.WithEnvConfig(ctx, *envConfig)
			cmdHandler := &runHandler{
				ctx:                   ctx,
				fs:                    filesystem.NewFakeFs(),
				execClient:            tt.execClient(ctrl),
				platformClient:        tt.platformClient(ctrl),
				configAutomountClient: tt.configAutomountClient(ctrl),
				imageBackend:          tt.imageBackend(ctrl),
				devfile:               tt.devfileObj(),
			}
			err = libdevfile.Deploy(ctx, tt.devfileObj(), cmdHandler)
			if (err != nil) != tt.wantErr {
				t.Errorf("Err expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

/*
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
				_ = devfileData.AddComponents([]v1alpha2.Component{imageComponent})
				_ = devfileData.AddCommands([]v1alpha2.Command{applyImage})

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
				_ = devfileData.AddComponents([]v1alpha2.Component{imageComponent})
				_ = devfileData.AddCommands([]v1alpha2.Command{applyImage})

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
			ctx = odocontext.WithApplication(ctx, tt.appName)
			ctx = odocontext.WithComponentName(ctx, tt.componentName)
			ctrl := gomock.NewController(t)
			cmdHandler := &runHandler{
				ctx:            ctx,
				fs:             filesystem.NewFakeFs(),
				execClient:     tt.execClient(ctrl),
				platformClient: tt.platformClient(ctrl),
				imageBackend:   tt.imageBackend(ctrl),
				devfile:        tt.devfileObj(),
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
				_ = devfileData.AddComponents([]v1alpha2.Component{containerComponent})
				_ = devfileData.AddCommands([]v1alpha2.Command{execOnContainer})

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
			ctx := context.Background()
			ctx = odocontext.WithApplication(ctx, "app")
			ctx = odocontext.WithComponentName(ctx, "componentName")
			cmdHandler := &runHandler{
				ctx:            ctx,
				fs:             filesystem.NewFakeFs(),
				execClient:     tt.execClient(ctrl),
				platformClient: tt.platformClient(ctrl),
				devfile:        tt.devfileObj(),
				podName:        tt.podName,
			}
			err := libdevfile.Build(ctx, tt.devfileObj(), "", cmdHandler)
			if (err != nil) != tt.wantErr {
				t.Errorf("Err expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}
*/
