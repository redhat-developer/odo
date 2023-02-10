package podmandev

import (
	"testing"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/v2/pkg/devfile/parser"
	"github.com/devfile/library/v2/pkg/devfile/parser/data"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/libdevfile/generator"
	"github.com/redhat-developer/odo/pkg/version"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

var (
	devfileName = "mycmp"
	appName     = "app"

	command = generator.GetExecCommand(generator.ExecCommandParams{
		Id:          "run",
		Component:   "mycomponent",
		CommandLine: "./run",
		IsDefault:   pointer.Bool(true),
		Kind:        v1alpha2.RunCommandGroupKind,
	})

	baseComponent = generator.GetContainerComponent(generator.ContainerComponentParams{
		Name: "mycomponent",
		Container: v1alpha2.Container{
			Image: "myimage",
		},
	})

	volume = generator.GetVolumeComponent(generator.VolumeComponentParams{
		Name: "myvolume",
	})

	basePod = &corev1.Pod{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "mycmp-app",
			Labels: map[string]string{
				"app":                                  appName,
				"app.kubernetes.io/instance":           devfileName,
				"app.kubernetes.io/managed-by":         "odo",
				"app.kubernetes.io/managed-by-version": version.VERSION,
				"app.kubernetes.io/part-of":            appName,
				"component":                            devfileName,
				"odo.dev/mode":                         labels.ComponentDevMode,
				"odo.dev/project-type":                 "Not available",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Args:    []string{"-f", "/dev/null"},
					Command: []string{"tail"},
					Env: []corev1.EnvVar{
						{
							Name:  "PROJECTS_ROOT",
							Value: "/projects",
						},
						{
							Name:  "PROJECT_SOURCE",
							Value: "/projects",
						},
					},
					Image:           "myimage",
					ImagePullPolicy: "Always",
					Name:            "mycomponent",
					VolumeMounts: []corev1.VolumeMount{
						{
							MountPath: "/projects",
							Name:      "odo-projects",
						},
						{
							MountPath: "/opt/odo/",
							Name:      "odo-shared-data",
						},
					},
				},
				{
					Args:    []string{"-f", "/dev/null"},
					Command: []string{"tail"},
					Image:   portForwardingHelperImage,
					Name:    portForwardingHelperContainerName,
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "odo-projects",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "odo-projects-mycmp-app",
						},
					},
				},
				{
					Name: "odo-shared-data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "odo-shared-data-mycmp-app",
						},
					},
				},
			},
		},
	}
)

func Test_createPodFromComponent(t *testing.T) {

	type args struct {
		devfileObj    func() parser.DevfileObj
		componentName string
		appName       string
		debug         bool
		buildCommand  string
		runCommand    string
		debugCommand  string
	}
	tests := []struct {
		name        string
		args        args
		wantPod     func() *corev1.Pod
		wantFwPorts []api.ForwardedPort
		wantErr     bool
	}{
		{
			name: "basic component without command",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddCommands([]v1alpha2.Command{command})
					_ = data.AddComponents([]v1alpha2.Component{baseComponent})
					return parser.DevfileObj{
						Data: data,
					}
				},
				componentName: devfileName,
				appName:       appName,
			},
			wantPod: func() *corev1.Pod {
				pod := basePod.DeepCopy()
				return pod
			},
		},
		{
			name: "basic component with command",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddCommands([]v1alpha2.Command{command})
					cmp := baseComponent.DeepCopy()
					cmp.Container.Command = []string{"./cmd"}
					cmp.Container.Args = []string{"arg1", "arg2"}
					_ = data.AddComponents([]v1alpha2.Component{*cmp})
					return parser.DevfileObj{
						Data: data,
					}
				},
				componentName: devfileName,
				appName:       appName,
			},
			wantPod: func() *corev1.Pod {
				pod := basePod.DeepCopy()
				pod.Spec.Containers[0].Command = []string{"./cmd"}
				pod.Spec.Containers[0].Args = []string{"arg1", "arg2"}
				return pod
			},
		},
		{
			name: "basic component + memory limit",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddCommands([]v1alpha2.Command{command})
					cmp := baseComponent.DeepCopy()
					cmp.Container.MemoryLimit = "1Gi"
					_ = data.AddComponents([]v1alpha2.Component{*cmp})
					return parser.DevfileObj{
						Data: data,
					}
				},
				componentName: devfileName,
				appName:       appName,
			},
			wantPod: func() *corev1.Pod {
				pod := basePod.DeepCopy()
				pod.Spec.Containers[0].Resources.Limits = corev1.ResourceList{
					"memory": resource.MustParse("1Gi"),
				}
				return pod
			},
		},
		{
			name: "basic component + application endpoint",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddCommands([]v1alpha2.Command{command})
					cmp := baseComponent.DeepCopy()
					cmp.Container.Endpoints = append(cmp.Container.Endpoints, v1alpha2.Endpoint{
						Name:       "http",
						TargetPort: 8080,
					})
					_ = data.AddComponents([]v1alpha2.Component{*cmp})
					return parser.DevfileObj{
						Data: data,
					}
				},
				componentName: devfileName,
				appName:       appName,
			},
			wantPod: func() *corev1.Pod {
				pod := basePod.DeepCopy()
				pod.Spec.Containers[1].Ports = append(pod.Spec.Containers[1].Ports, corev1.ContainerPort{
					Name:          "http",
					ContainerPort: 20001,
					HostPort:      20001,
				})
				return pod
			},
			wantFwPorts: []api.ForwardedPort{
				{
					Platform:      "podman",
					ContainerName: "mycomponent",
					PortName:      "http",
					LocalAddress:  "127.0.0.1",
					LocalPort:     20001,
					ContainerPort: 8080,
					IsDebug:       false,
				},
			},
		},
		{
			name: "basic component + application endpoint + debug endpoint - without debug",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddCommands([]v1alpha2.Command{command})
					cmp := baseComponent.DeepCopy()
					cmp.Container.Endpoints = append(cmp.Container.Endpoints, v1alpha2.Endpoint{
						Name:       "http",
						TargetPort: 8080,
					})
					cmp.Container.Endpoints = append(cmp.Container.Endpoints, v1alpha2.Endpoint{
						Name:       "debug",
						TargetPort: 5858,
					})
					_ = data.AddComponents([]v1alpha2.Component{*cmp})
					return parser.DevfileObj{
						Data: data,
					}
				},
				componentName: devfileName,
				appName:       appName,
			},
			wantPod: func() *corev1.Pod {
				pod := basePod.DeepCopy()
				pod.Spec.Containers[1].Ports = append(pod.Spec.Containers[1].Ports, corev1.ContainerPort{
					Name:          "http",
					ContainerPort: 20001,
					HostPort:      20001,
				})
				return pod
			},
			wantFwPorts: []api.ForwardedPort{
				{
					Platform:      "podman",
					ContainerName: "mycomponent",
					PortName:      "http",
					LocalAddress:  "127.0.0.1",
					LocalPort:     20001,
					ContainerPort: 8080,
					IsDebug:       false,
				},
			},
		},
		{
			name: "basic component + application endpoint + debug endpoint - with debug",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddCommands([]v1alpha2.Command{command})
					cmp := baseComponent.DeepCopy()
					cmp.Container.Endpoints = append(cmp.Container.Endpoints, v1alpha2.Endpoint{
						Name:       "http",
						TargetPort: 8080,
					})
					cmp.Container.Endpoints = append(cmp.Container.Endpoints, v1alpha2.Endpoint{
						Name:       "debug",
						TargetPort: 5858,
					})
					_ = data.AddComponents([]v1alpha2.Component{*cmp})
					return parser.DevfileObj{
						Data: data,
					}
				},
				componentName: devfileName,
				appName:       appName,
				debug:         true,
			},
			wantPod: func() *corev1.Pod {
				pod := basePod.DeepCopy()
				pod.Spec.Containers[1].Ports = append(pod.Spec.Containers[1].Ports, corev1.ContainerPort{
					Name:          "http",
					ContainerPort: 20001,
					HostPort:      20001,
				})
				pod.Spec.Containers[1].Ports = append(pod.Spec.Containers[1].Ports, corev1.ContainerPort{
					Name:          "debug",
					ContainerPort: 20002,
					HostPort:      20002,
				})
				return pod
			},
			wantFwPorts: []api.ForwardedPort{
				{
					Platform:      "podman",
					ContainerName: "mycomponent",
					PortName:      "http",
					LocalAddress:  "127.0.0.1",
					LocalPort:     20001,
					ContainerPort: 8080,
					IsDebug:       false,
				},
				{
					Platform:      "podman",
					ContainerName: "mycomponent",
					PortName:      "debug",
					LocalAddress:  "127.0.0.1",
					LocalPort:     20002,
					ContainerPort: 5858,
					IsDebug:       true,
				},
			},
		},
		{
			name: "basic component with volume mount",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddCommands([]v1alpha2.Command{command})
					_ = data.AddComponents([]v1alpha2.Component{baseComponent, volume})
					_ = data.AddVolumeMounts(baseComponent.Name, []v1alpha2.VolumeMount{
						{
							Name: volume.Name,
							Path: "/path/to/mount",
						},
					})

					return parser.DevfileObj{
						Data: data,
					}
				},
				componentName: devfileName,
				appName:       appName,
			},
			wantPod: func() *corev1.Pod {
				pod := basePod.DeepCopy()
				pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{
					Name: volume.Name,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: volume.Name + "-" + devfileName + "-" + appName,
						},
					},
				})
				pod.Spec.Containers[0].VolumeMounts = append(pod.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
					Name:      volume.Name,
					MountPath: "/path/to/mount",
				})
				return pod
			},
		},
		{
			name: "basic component + application endpoint + debug endpoint + container ports known - with debug",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddCommands([]v1alpha2.Command{command})
					cmp := baseComponent.DeepCopy()
					cmp.Container.Endpoints = append(cmp.Container.Endpoints, v1alpha2.Endpoint{
						Name:       "http",
						TargetPort: 20001,
					})
					cmp.Container.Endpoints = append(cmp.Container.Endpoints, v1alpha2.Endpoint{
						Name:       "debug",
						TargetPort: 20002,
					})
					cmp.Container.Endpoints = append(cmp.Container.Endpoints, v1alpha2.Endpoint{
						Name:       "debug-1",
						TargetPort: 5858,
					})
					_ = data.AddComponents([]v1alpha2.Component{*cmp})
					return parser.DevfileObj{
						Data: data,
					}
				},
				componentName: devfileName,
				appName:       appName,
				debug:         true,
			},
			wantPod: func() *corev1.Pod {
				pod := basePod.DeepCopy()
				pod.Spec.Containers[1].Ports = append(pod.Spec.Containers[1].Ports, corev1.ContainerPort{
					Name:          "http",
					ContainerPort: 20003,
					HostPort:      20003,
				})
				pod.Spec.Containers[1].Ports = append(pod.Spec.Containers[1].Ports, corev1.ContainerPort{
					Name:          "debug",
					ContainerPort: 20004,
					HostPort:      20004,
				})
				pod.Spec.Containers[1].Ports = append(pod.Spec.Containers[1].Ports, corev1.ContainerPort{
					Name:          "debug-1",
					ContainerPort: 20005,
					HostPort:      20005,
				})
				return pod
			},
			wantFwPorts: []api.ForwardedPort{
				{
					Platform:      "podman",
					ContainerName: "mycomponent",
					PortName:      "http",
					LocalAddress:  "127.0.0.1",
					LocalPort:     20003,
					ContainerPort: 20001,
					IsDebug:       false,
				},
				{
					Platform:      "podman",
					ContainerName: "mycomponent",
					PortName:      "debug",
					LocalAddress:  "127.0.0.1",
					LocalPort:     20004,
					ContainerPort: 20002,
					IsDebug:       true,
				},
				{
					Platform:      "podman",
					ContainerName: "mycomponent",
					PortName:      "debug-1",
					LocalAddress:  "127.0.0.1",
					LocalPort:     20005,
					ContainerPort: 5858,
					IsDebug:       true,
				},
			},
		},
		{
			name: "basic component + application endpoint + debug endpoint + container ports known - without debug",
			args: args{
				devfileObj: func() parser.DevfileObj {
					data, _ := data.NewDevfileData(string(data.APISchemaVersion200))
					_ = data.AddCommands([]v1alpha2.Command{command})
					cmp := baseComponent.DeepCopy()
					cmp.Container.Endpoints = append(cmp.Container.Endpoints, v1alpha2.Endpoint{
						Name:       "http",
						TargetPort: 20001,
					})
					cmp.Container.Endpoints = append(cmp.Container.Endpoints, v1alpha2.Endpoint{
						Name:       "debug",
						TargetPort: 20002,
					})
					cmp.Container.Endpoints = append(cmp.Container.Endpoints, v1alpha2.Endpoint{
						Name:       "debug-1",
						TargetPort: 5858,
					})
					cmp.Container.Endpoints = append(cmp.Container.Endpoints, v1alpha2.Endpoint{
						Name:       "http-1",
						TargetPort: 8080,
					})
					_ = data.AddComponents([]v1alpha2.Component{*cmp})
					return parser.DevfileObj{
						Data: data,
					}
				},
				componentName: devfileName,
				appName:       appName,
				debug:         false,
			},
			wantPod: func() *corev1.Pod {
				pod := basePod.DeepCopy()
				pod.Spec.Containers[1].Ports = append(pod.Spec.Containers[1].Ports, corev1.ContainerPort{
					Name:          "http",
					ContainerPort: 20002,
					HostPort:      20002,
				})
				pod.Spec.Containers[1].Ports = append(pod.Spec.Containers[1].Ports, corev1.ContainerPort{
					Name:          "http-1",
					ContainerPort: 20003,
					HostPort:      20003,
				})
				return pod
			},
			wantFwPorts: []api.ForwardedPort{
				{
					Platform:      "podman",
					ContainerName: "mycomponent",
					PortName:      "http",
					LocalAddress:  "127.0.0.1",
					LocalPort:     20002,
					ContainerPort: 20001,
					IsDebug:       false,
				},
				{
					Platform:      "podman",
					ContainerName: "mycomponent",
					PortName:      "http-1",
					LocalAddress:  "127.0.0.1",
					LocalPort:     20003,
					ContainerPort: 8080,
					IsDebug:       false,
				},
			},
		},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotFwPorts, err := createPodFromComponent(
				tt.args.devfileObj(),
				tt.args.componentName,
				tt.args.appName,
				tt.args.debug,
				tt.args.buildCommand,
				tt.args.runCommand,
				tt.args.debugCommand,
				false,
				[]int{20001, 20002, 20003, 20004, 20005},
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("createPodFromComponent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(tt.wantPod(), got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("createPodFromComponent() pod mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantFwPorts, gotFwPorts, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("createPodFromComponent() fwPorts mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
