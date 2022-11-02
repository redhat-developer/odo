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
	"k8s.io/utils/pointer"

	"github.com/redhat-developer/odo/pkg/libdevfile/generator"
	"github.com/redhat-developer/odo/pkg/testingutil"
	"github.com/redhat-developer/odo/pkg/util"
)

var buildGroup = v1alpha2.BuildCommandGroupKind
var runGroup = v1alpha2.RunCommandGroupKind

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
		wantPresent    bool
	}{
		{
			name: "Case 1: Valid devfile",
			execCommands: []v1alpha2.Command{
				getExecCommand("build", buildGroup),
				getExecCommand("run", runGroup),
			},
			requestedType: []v1alpha2.CommandGroupKind{buildGroup, runGroup},
			wantErr:       false,
			wantPresent:   true,
		},
		{
			name: "Case 2: Valid devfile with devrun and devbuild",
			execCommands: []v1alpha2.Command{
				getExecCommand("build", buildGroup),
				getExecCommand("run", runGroup),
			},
			requestedType: []v1alpha2.CommandGroupKind{buildGroup, runGroup},
			wantErr:       false,
			wantPresent:   true,
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
			wantPresent:   true,
		},
		{
			name: "Case 4.1: Mismatched command type",
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
			name: "Case 4.2: Matching command by name and type",
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
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			reqCommandName: "build command",
			requestedType:  []v1alpha2.CommandGroupKind{buildGroup},
			wantErr:        false,
			wantPresent:    true,
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
			wantPresent:    true,
		},
		{
			name: "Case 5.1: if only one command is present, it is returned and assumed as default",
			execCommands: []v1alpha2.Command{
				{
					Id: "defaultRunCommand",
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
			wantPresent:    true,
		},
		{
			name: "Case 5.2: if multiple default commands are present, error is returned",
			execCommands: []v1alpha2.Command{
				{
					Id: "runCommand1",
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
				}, {
					Id: "runCommand2",
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

			requestedType: []v1alpha2.CommandGroupKind{runGroup},
			wantErr:       true,
		},
		{
			name: "Case 5.2: if multiple default commands are present, error is returned",
			execCommands: []v1alpha2.Command{
				{
					Id: "runCommand1",
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
				}, {
					Id: "runCommand2",
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

			requestedType: []v1alpha2.CommandGroupKind{runGroup},
			wantErr:       true,
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
			wantPresent:    true,
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
				cmd, ok, err := GetCommand(devObj, tt.reqCommandName, gtype)
				if tt.wantErr != (err != nil) {
					t.Errorf("TestGetCommand unexpected error for command: %v wantErr: %v err: %v", gtype, tt.wantErr, err)
					return
				} else if tt.wantErr {
					return
				}
				if tt.wantPresent != ok {
					t.Errorf("TestGetCommand unexpected presence for command: %v wantPresent: %v ok: %v", gtype, tt.wantPresent, ok)
					return
				}

				if len(tt.retCommandName) > 0 && cmd.Id != tt.retCommandName {
					t.Errorf("TestGetCommand error: command names do not match expected: %v actual: %v", tt.retCommandName, cmd.Id)
				}
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
					dData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = dData.AddCommands([]v1alpha2.Command{deployDefault1, applyImageCommand, applyDeploymentCommand, applyServiceCommand})
					_ = dData.AddComponents([]v1alpha2.Component{imageComponent, deploymentComponent, serviceComponent})
					return parser.DevfileObj{
						Data: dData,
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
		{
			name: "deploy with multiple deploy and no default",
			args: args{
				devfileObj: func() parser.DevfileObj {
					dData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = dData.AddCommands([]v1alpha2.Command{applyServiceCommand, applyDeploymentCommand})
					_ = dData.AddComponents([]v1alpha2.Component{deploymentComponent, serviceComponent})
					return parser.DevfileObj{
						Data: dData,
					}
				},
				handler: func(ctrl *gomock.Controller) Handler {
					return NewMockHandler(ctrl)
				},
			},
			wantErr: true,
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

func TestBuild(t *testing.T) {
	containerComp := v1alpha2.Component{
		Name: "my-container",
		ComponentUnion: v1alpha2.ComponentUnion{
			Container: &v1alpha2.ContainerComponent{
				Container: v1alpha2.Container{
					Image: "my-image",
				},
			},
		},
	}
	defaultBuildCommand := generator.GetExecCommand(generator.ExecCommandParams{
		Kind:        v1alpha2.BuildCommandGroupKind,
		Id:          "my-default-build-command",
		IsDefault:   pointer.BoolPtr(true),
		CommandLine: "build my-app",
		Component:   containerComp.Name,
	})
	nonDefaultBuildCommandExplicit := generator.GetExecCommand(generator.ExecCommandParams{
		Kind:        v1alpha2.BuildCommandGroupKind,
		Id:          "my-explicit-non-default-build-command",
		IsDefault:   pointer.BoolPtr(false),
		CommandLine: "build my-app",
		Component:   containerComp.Name,
	})
	nonDefaultBuildCommandImplicit := generator.GetExecCommand(generator.ExecCommandParams{
		Kind:        v1alpha2.BuildCommandGroupKind,
		Id:          "my-implicit-non-default-build-command",
		CommandLine: "build my-app",
		Component:   containerComp.Name,
	})
	type args struct {
		devfileObj func() parser.DevfileObj
		handler    func(ctrl *gomock.Controller) Handler
		cmdName    string
	}
	for _, tt := range []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "missing default command",
			args: args{
				devfileObj: func() parser.DevfileObj {
					dData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = dData.AddCommands([]v1alpha2.Command{nonDefaultBuildCommandExplicit, nonDefaultBuildCommandImplicit})
					_ = dData.AddComponents([]v1alpha2.Component{containerComp})
					return parser.DevfileObj{
						Data: dData,
					}
				},
				handler: func(ctrl *gomock.Controller) Handler {
					h := NewMockHandler(ctrl)
					h.EXPECT().Execute(gomock.Eq(defaultBuildCommand)).Times(0)
					h.EXPECT().Execute(gomock.Eq(nonDefaultBuildCommandExplicit)).Times(0)
					h.EXPECT().Execute(gomock.Eq(nonDefaultBuildCommandImplicit)).Times(0)
					return h
				},
			},
		},
		{
			name: "with default command",
			args: args{
				devfileObj: func() parser.DevfileObj {
					dData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = dData.AddCommands([]v1alpha2.Command{defaultBuildCommand, nonDefaultBuildCommandExplicit, nonDefaultBuildCommandImplicit})
					_ = dData.AddComponents([]v1alpha2.Component{containerComp})
					return parser.DevfileObj{
						Data: dData,
					}
				},
				handler: func(ctrl *gomock.Controller) Handler {
					h := NewMockHandler(ctrl)
					h.EXPECT().Execute(gomock.Eq(defaultBuildCommand)).Times(1)
					h.EXPECT().Execute(gomock.Eq(nonDefaultBuildCommandExplicit)).Times(0)
					h.EXPECT().Execute(gomock.Eq(nonDefaultBuildCommandImplicit)).Times(0)
					return h
				},
			},
		},
		{
			name: "missing custom command",
			args: args{
				devfileObj: func() parser.DevfileObj {
					dData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = dData.AddCommands([]v1alpha2.Command{defaultBuildCommand})
					_ = dData.AddComponents([]v1alpha2.Component{containerComp})
					return parser.DevfileObj{
						Data: dData,
					}
				},
				handler: func(ctrl *gomock.Controller) Handler {
					h := NewMockHandler(ctrl)
					h.EXPECT().Execute(gomock.Eq(defaultBuildCommand)).Times(0)
					h.EXPECT().Execute(gomock.Eq(nonDefaultBuildCommandExplicit)).Times(0)
					h.EXPECT().Execute(gomock.Eq(nonDefaultBuildCommandImplicit)).Times(0)
					return h
				},
				cmdName: "my-explicit-non-default-build-command",
			},
			wantErr: true,
		},
		{
			name: "with custom command",
			args: args{
				devfileObj: func() parser.DevfileObj {
					dData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = dData.AddCommands([]v1alpha2.Command{nonDefaultBuildCommandExplicit, nonDefaultBuildCommandImplicit})
					_ = dData.AddComponents([]v1alpha2.Component{containerComp})
					return parser.DevfileObj{
						Data: dData,
					}
				},
				handler: func(ctrl *gomock.Controller) Handler {
					h := NewMockHandler(ctrl)
					h.EXPECT().Execute(gomock.Eq(defaultBuildCommand)).Times(0)
					h.EXPECT().Execute(gomock.Eq(nonDefaultBuildCommandExplicit)).Times(1)
					h.EXPECT().Execute(gomock.Eq(nonDefaultBuildCommandImplicit)).Times(0)
					return h
				},
				cmdName: "my-explicit-non-default-build-command",
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := Build(tt.args.devfileObj(), tt.args.cmdName, tt.args.handler(gomock.NewController(t)))
			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
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

	containerWithOneNoneInternalEndpoint := generator.GetContainerComponent(generator.ContainerComponentParams{
		Name: "container-none-endpoint",
		Endpoints: []v1alpha2.Endpoint{
			{
				Name:       "debug",
				TargetPort: 9099,
				Exposure:   v1alpha2.NoneEndpointExposure,
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
				containers: []v1alpha2.Component{
					containerWithNoEndpoints,
					containerWithOnePublicEndpoint,
					containerWithOneInternalEndpoint,
					containerWithOneNoneInternalEndpoint,
				},
			},
			want: map[string][]int{
				containerWithNoEndpoints.Name:             {},
				containerWithOnePublicEndpoint.Name:       {8080},
				containerWithOneInternalEndpoint.Name:     {9090},
				containerWithOneNoneInternalEndpoint.Name: {9099},
			},
		},
		{
			name: "invalid input - one image component with rest being containers",
			args: args{
				containers: []v1alpha2.Component{
					containerWithNoEndpoints,
					containerWithOnePublicEndpoint,
					containerWithOneInternalEndpoint,
					containerWithOneNoneInternalEndpoint,
					imageComponent,
				},
			},
			want: map[string][]int{
				containerWithNoEndpoints.Name:             {},
				containerWithOnePublicEndpoint.Name:       {8080},
				containerWithOneInternalEndpoint.Name:     {9090},
				containerWithOneNoneInternalEndpoint.Name: {9099},
			},
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

func TestValidateAndGetCommand(t *testing.T) {

	commands := [...]string{"ls -la", "pwd"}
	components := [...]string{"alias1", "alias2"}

	tests := []struct {
		name           string
		requestedType  []v1alpha2.CommandGroupKind
		execCommands   []v1alpha2.Command
		compCommands   []v1alpha2.Command
		reqCommandName string
		retCommandName []string
		wantErr        bool
	}{
		{
			name: "Case 1: Valid devfile, default command returned even if it is not marked as IsDefault",
			execCommands: []v1alpha2.Command{
				getExecCommand("build", buildGroup),
				getExecCommand("run", runGroup),
			},
			requestedType:  []v1alpha2.CommandGroupKind{buildGroup, runGroup},
			wantErr:        false,
			retCommandName: []string{"build", "run"},
		},
		{
			name: "Case 2: Valid devfile, but error returned because multiple build commands without default",
			execCommands: []v1alpha2.Command{
				getExecCommand("build", buildGroup),
				getExecCommand("build2", buildGroup),
				getExecCommand("run", runGroup),
			},
			requestedType: []v1alpha2.CommandGroupKind{buildGroup, runGroup},
			wantErr:       true,
		},
		{
			name: "Case 3: Valid devfile with empty workdir",
			execCommands: []v1alpha2.Command{
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
			requestedType:  []v1alpha2.CommandGroupKind{runGroup},
			wantErr:        false,
			retCommandName: []string{"run"},
		},
		{
			name: "Case 4.1: Mismatched command type",
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
			name: "Case 4.2: Matching command by name and type",
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
							CommandLine: commands[0],
							Component:   components[0],
						},
					},
				},
			},
			reqCommandName: "build command",
			requestedType:  []v1alpha2.CommandGroupKind{buildGroup},
			retCommandName: []string{"build command"},
			wantErr:        false,
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
			retCommandName: []string{"defaultRunCommand"},
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
			retCommandName: []string{"myComposite"},
			requestedType:  []v1alpha2.CommandGroupKind{buildGroup},
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if !tt.wantErr && (len(tt.requestedType) != len(tt.retCommandName)) {
				t.Errorf("Invalid test definition %q requestedType length must match retCommandName length.", tt.name)
			}
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

			for i, gtype := range tt.requestedType {
				cmd, err := ValidateAndGetCommand(devObj, tt.reqCommandName, gtype)
				if tt.wantErr != (err != nil) {
					t.Errorf("TestGetCommand unexpected error for command: %v wantErr: %v err: %v", gtype, tt.wantErr, err)
					return
				} else if tt.wantErr {
					return
				}

				if cmd.Id != tt.retCommandName[i] {
					t.Errorf("TestGetCommand error: command names do not match expected: %v actual: %v", tt.retCommandName[i], cmd.Id)
				}
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

	defaultBuildCommand := v1alpha2.Command{
		Id: "default build command",
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
			name:         "Case 2: Default Build Command, and Provided Run Command",
			buildCommand: emptyString,
			runCommand:   "customcommand",
			execCommands: execCommands,
			//only the specified run command is returned, because the build command is not marked as default
			numberOfCommands: 2,
			wantErr:          false,
		},
		{
			name:             "Case 2.2: Default Build Command, and Default Run Command",
			buildCommand:     emptyString,
			runCommand:       emptyString,
			execCommands:     append(execCommands, defaultBuildCommand),
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

			pushCommands, err := ValidateAndGetPushCommands(devObj, tt.buildCommand, tt.runCommand)
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
