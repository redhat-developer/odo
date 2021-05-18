package testingutil

import (
	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"
	devfileCtx "github.com/devfile/library/pkg/devfile/parser/context"
	"github.com/devfile/library/pkg/devfile/parser/data"
	devfilefs "github.com/devfile/library/pkg/testingutil/filesystem"
)

// GetFakeContainerComponent returns a fake container component for testing
func GetFakeContainerComponent(name string) v1.Component {
	image := "docker.io/maven:latest"
	memoryLimit := "128Mi"
	volumeName := "myvolume1"
	volumePath := "/my/volume/mount/path1"
	mountSources := true

	return v1.Component{
		Name: name,
		ComponentUnion: v1.ComponentUnion{
			Container: &v1.ContainerComponent{
				Container: v1.Container{
					Image:       image,
					Env:         []v1.EnvVar{},
					MemoryLimit: memoryLimit,
					VolumeMounts: []v1.VolumeMount{{
						Name: volumeName,
						Path: volumePath,
					}},
					MountSources: &mountSources,
				},
			}}}

}

// GetFakeVolumeComponent returns a fake volume component for testing
func GetFakeVolumeComponent(name, size string) v1.Component {
	return v1.Component{
		Name: name,
		ComponentUnion: v1.ComponentUnion{
			Volume: &v1.VolumeComponent{
				Volume: v1.Volume{
					Size: size,
				}}}}

}

// GetFakeExecRunCommands returns fake commands for testing
func GetFakeExecRunCommands() []v1.Command {
	return []v1.Command{
		{
			CommandUnion: v1.CommandUnion{
				Exec: &v1.ExecCommand{
					LabeledCommand: v1.LabeledCommand{
						BaseCommand: v1.BaseCommand{
							Group: &v1.CommandGroup{
								Kind:      v1.RunCommandGroupKind,
								IsDefault: true,
							},
						},
					},
					CommandLine: "ls -a",
					Component:   "alias1",
					WorkingDir:  "/root",
				},
			},
		},
	}
}

// GetFakeExecRunCommands returns a fake env for testing
func GetFakeEnv(name, value string) v1.EnvVar {
	return v1.EnvVar{
		Name:  name,
		Value: value,
	}
}

// GetFakeVolumeMount returns a fake volume mount for testing
func GetFakeVolumeMount(name, path string) v1.VolumeMount {
	return v1.VolumeMount{
		Name: name,
		Path: path,
	}
}

// GetTestDevfileObj returns a devfile object for testing
func GetTestDevfileObj(fs devfilefs.Filesystem) parser.DevfileObj {
	devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
	_ = devfileData.AddCommands([]v1.Command{
		{
			Id: "devbuild",
			CommandUnion: v1.CommandUnion{
				Exec: &v1.ExecCommand{
					WorkingDir: "/projects/nodejs-starter",
				},
			},
		},
	})
	_ = devfileData.AddComponents([]v1.Component{
		{
			Name: "runtime",
			ComponentUnion: v1.ComponentUnion{
				Container: &v1.ContainerComponent{
					Container: v1.Container{
						Image: "quay.io/nodejs-12",
					},
					Endpoints: []v1.Endpoint{
						{
							Name:       "port-3030",
							TargetPort: 3000,
						},
					},
				},
			},
		},
		{
			Name: "loadbalancer",
			ComponentUnion: v1.ComponentUnion{
				Container: &v1.ContainerComponent{
					Container: v1.Container{
						Image: "quay.io/nginx",
					},
				},
			},
		},
	})

	return parser.DevfileObj{
		Ctx:  devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
		Data: devfileData,
	}
}

// GetTestDevfileObjWithMultipleEndpoints returns a devfile object with multiple endpoints for testing
func GetTestDevfileObjWithMultipleEndpoints(fs devfilefs.Filesystem) parser.DevfileObj {
	devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))
	_ = devfileData.AddComponents([]v1.Component{
		{
			Name: "runtime",
			ComponentUnion: v1.ComponentUnion{
				Container: &v1.ContainerComponent{
					Endpoints: []v1.Endpoint{
						{
							Name:       "port-3030",
							TargetPort: 3030,
						},
						{
							Name:       "port-3000",
							TargetPort: 3000,
						},
					},
				},
			},
		},
		{
			Name: "runtime-debug",
			ComponentUnion: v1.ComponentUnion{
				Container: &v1.ContainerComponent{
					Endpoints: []v1.Endpoint{
						{
							Name:       "port-8080",
							TargetPort: 8080,
						},
					},
				},
			},
		},
	})
	return parser.DevfileObj{
		Ctx:  devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
		Data: devfileData,
	}
}

// DevfileObjWithInternalNoneEndpoints returns a devfile object with internal endpoints for testing
func DevfileObjWithInternalNoneEndpoints(fs devfilefs.Filesystem) parser.DevfileObj {
	devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))

	_ = devfileData.AddComponents([]v1.Component{
		{
			Name: "runtime",
			ComponentUnion: v1.ComponentUnion{
				Container: &v1.ContainerComponent{
					Endpoints: []v1.Endpoint{
						{
							Name:       "port-3030",
							TargetPort: 3030,
							Exposure:   v1.NoneEndpointExposure,
						},
						{
							Name:       "port-3000",
							TargetPort: 3000,
						},
					},
				},
			},
		},
		{
			Name: "runtime-debug",
			ComponentUnion: v1.ComponentUnion{
				Container: &v1.ContainerComponent{
					Endpoints: []v1.Endpoint{
						{
							Name:       "port-8080",
							TargetPort: 8080,
							Exposure:   v1.InternalEndpointExposure,
						},
					},
				},
			},
		},
	})

	return parser.DevfileObj{
		Ctx:  devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
		Data: devfileData,
	}
}

// DevfileObjWithSecureEndpoints returns a devfile object with internal endpoints for testing
func DevfileObjWithSecureEndpoints(fs devfilefs.Filesystem) parser.DevfileObj {
	devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))

	_ = devfileData.AddComponents([]v1.Component{
		{
			Name: "runtime",
			ComponentUnion: v1.ComponentUnion{
				Container: &v1.ContainerComponent{
					Endpoints: []v1.Endpoint{
						{
							Name:       "port-3030",
							TargetPort: 3030,
							Protocol:   v1.WSSEndpointProtocol,
						},
						{
							Name:       "port-3000",
							TargetPort: 3000,
							Protocol:   v1.HTTPSEndpointProtocol,
						},
					},
				},
			},
		},
		{
			Name: "runtime-debug",
			ComponentUnion: v1.ComponentUnion{
				Container: &v1.ContainerComponent{
					Endpoints: []v1.Endpoint{
						{
							Name:       "port-8080",
							TargetPort: 8080,
							Secure:     true,
						},
					},
				},
			},
		},
	})
	return parser.DevfileObj{
		Ctx:  devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
		Data: devfileData,
	}
}

// GetTestDevfileObjWithPath returns a devfile object for testing
func GetTestDevfileObjWithPath(fs devfilefs.Filesystem) parser.DevfileObj {
	devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion200))

	_ = devfileData.AddCommands([]v1.Command{
		{
			Id: "devbuild",
			CommandUnion: v1.CommandUnion{
				Exec: &v1.ExecCommand{
					WorkingDir: "/projects/nodejs-starter",
				},
			},
		},
	})
	_ = devfileData.AddComponents([]v1.Component{
		{
			Name: "runtime",
			ComponentUnion: v1.ComponentUnion{
				Container: &v1.ContainerComponent{
					Container: v1.Container{
						Image: "quay.io/nodejs-12",
					},
					Endpoints: []v1.Endpoint{
						{
							Name:       "port-3030",
							TargetPort: 3000,
							Path:       "/test",
						},
					},
				},
			},
		},
		{
			Name: "loadbalancer",
			ComponentUnion: v1.ComponentUnion{
				Container: &v1.ContainerComponent{
					Container: v1.Container{
						Image: "quay.io/nginx",
					},
				},
			},
		},
	})
	return parser.DevfileObj{
		Ctx:  devfileCtx.FakeContext(fs, parser.OutputDevfileYamlPath),
		Data: devfileData,
	}
}
