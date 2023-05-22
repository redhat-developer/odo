package common

/*
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
		platformClient  func(ctrl *gomock.Controller) platform.Client
		execClient      func(ctrl *gomock.Controller) exec.Client
		appName         string
		componentName   string
		podName         string
		msg             string
		show            bool
		componentExists bool
		devfileObj      func() parser.DevfileObj
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
		},
		{
			name: "Devfile with Exec deploy command",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				_ = devfileData.AddComponents([]v1alpha2.Component{kubernetesDeploy})
				_ = devfileData.AddCommands([]v1alpha2.Command{defaultDeployCommandKubernetes})

				devfileObj := parser.DevfileObj{
					Data: devfileData,
				}
				return devfileObj
			},
			platformClient: func(ctrl *gomock.Controller) platform.Client {
				client := platform.NewMockClient(ctrl)
				// Nothing happens as Apply Kubernetes component is not implemented by handler
				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				// Nothing happens as Apply Kubernetes component is not implemented by handler
				return client
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			cmdHandler := NewExecHandler(tt.platformClient(ctrl), tt.execClient(ctrl), tt.appName, tt.componentName, tt.podName, tt.msg, tt.show, tt.componentExists)
			ctx := context.Background()
			_ = libdevfile.Deploy(ctx, tt.devfileObj(), cmdHandler)
		})
	}
}

func TestApplyOpenshift(t *testing.T) {

	tests := []struct {
		name            string
		platformClient  func(ctrl *gomock.Controller) platform.Client
		execClient      func(ctrl *gomock.Controller) exec.Client
		appName         string
		componentName   string
		podName         string
		msg             string
		show            bool
		componentExists bool
		devfileObj      func() parser.DevfileObj
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
		},
		{
			name: "Devfile with Exec deploy command",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				_ = devfileData.AddComponents([]v1alpha2.Component{openshiftDeploy})
				_ = devfileData.AddCommands([]v1alpha2.Command{defaultDeployCommandOpenshift})

				devfileObj := parser.DevfileObj{
					Data: devfileData,
				}
				return devfileObj
			},
			platformClient: func(ctrl *gomock.Controller) platform.Client {
				client := platform.NewMockClient(ctrl)
				// Nothing happens as Apply Openshift component is not implemented by handler
				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				// Nothing happens as Apply Openshift component is not implemented by handler
				return client
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			cmdHandler := NewExecHandler(tt.platformClient(ctrl), tt.execClient(ctrl), tt.appName, tt.componentName, tt.podName, tt.msg, tt.show, tt.componentExists)
			ctx := context.Background()
			_ = libdevfile.Deploy(ctx, tt.devfileObj(), cmdHandler)
		})
	}
}

func TestApplyImage(t *testing.T) {

	tests := []struct {
		name            string
		platformClient  func(ctrl *gomock.Controller) platform.Client
		execClient      func(ctrl *gomock.Controller) exec.Client
		appName         string
		componentName   string
		podName         string
		msg             string
		show            bool
		componentExists bool
		devfileObj      func() parser.DevfileObj
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
		},
		{
			name: "Devfile with Exec deploy command",
			devfileObj: func() parser.DevfileObj {
				devfileData, err := data.NewDevfileData("2.1.0")
				if err != nil {
					t.Error(err)
				}
				devfileData.SetSchemaVersion("2.1.0")
				_ = devfileData.AddComponents([]v1alpha2.Component{imageDeploy})
				_ = devfileData.AddCommands([]v1alpha2.Command{defaultDeployCommandImage})

				devfileObj := parser.DevfileObj{
					Data: devfileData,
				}
				return devfileObj
			},
			platformClient: func(ctrl *gomock.Controller) platform.Client {
				client := platform.NewMockClient(ctrl)
				// Nothing happens as Apply Image component is not implemented by handler
				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				// Nothing happens as Apply Image component is not implemented by handler
				return client
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			cmdHandler := NewExecHandler(tt.platformClient(ctrl), tt.execClient(ctrl), tt.appName, tt.componentName, tt.podName, tt.msg, tt.show, tt.componentExists)
			ctx := context.Background()
			_ = libdevfile.Deploy(ctx, tt.devfileObj(), cmdHandler)
		})
	}
}

func TestExecute(t *testing.T) {

	tests := []struct {
		name            string
		platformClient  func(ctrl *gomock.Controller) platform.Client
		execClient      func(ctrl *gomock.Controller) exec.Client
		appName         string
		componentName   string
		podName         string
		msg             string
		show            bool
		componentExists bool
		devfileObj      func() parser.DevfileObj
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
				// Nothing happens as there is no default Build command on the Devfile
				return client
			},
			execClient: func(ctrl *gomock.Controller) exec.Client {
				client := exec.NewMockClient(ctrl)
				// Nothing happens as there is no default Build command on the Devfile
				return client
			},
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
				_ = devfileData.AddComponents([]v1alpha2.Component{container1})
				_ = devfileData.AddCommands([]v1alpha2.Command{defaultBuildCommand})

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
				client.EXPECT().ExecuteCommand(gomock.Any(), gomock.Any(), "a-pod-name", "my-container", true, gomock.Any(), gomock.Any())
				return client
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			cmdHandler := NewExecHandler(tt.platformClient(ctrl), tt.execClient(ctrl), tt.appName, tt.componentName, tt.podName, tt.msg, tt.show, tt.componentExists)
			ctx := context.Background()
			_ = libdevfile.Build(ctx, tt.devfileObj(), "", cmdHandler)
		})
	}
}
*/
