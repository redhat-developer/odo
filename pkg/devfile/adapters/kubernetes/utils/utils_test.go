package utils

import (
	"errors"
	"reflect"
	"testing"

	devfilepkg "github.com/devfile/api/v2/pkg/devfile"
	"github.com/devfile/library/pkg/devfile/parser/data"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/util"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	appsv1 "k8s.io/api/apps/v1"

	adaptersCommon "github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/kclient"
	odoTestingUtil "github.com/redhat-developer/odo/pkg/testingutil"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktesting "k8s.io/client-go/testing"
)

func TestComponentExists(t *testing.T) {

	tests := []struct {
		name             string
		componentType    devfilev1.ComponentType
		componentName    string
		appName          string
		getComponentName string
		want             bool
		wantErr          bool
	}{
		{
			name:             "Case 1: Valid component name",
			componentName:    "test-name",
			appName:          "app",
			getComponentName: "test-name",
			want:             true,
			wantErr:          false,
		},
		{
			name:             "Case 2: Non-existent component name",
			componentName:    "test-name",
			appName:          "",
			getComponentName: "fake-component",
			want:             false,
			wantErr:          false,
		},
		{
			name:             "Case 3: Error condition",
			componentName:    "test-name",
			appName:          "app",
			getComponentName: "test-name",
			want:             false,
			wantErr:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fkclient, fkclientset := kclient.FakeNew()
			fkclientset.Kubernetes.PrependReactor("list", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				emptyDeployment := odoTestingUtil.CreateFakeDeployment("")
				deployment := odoTestingUtil.CreateFakeDeployment(tt.getComponentName)

				if tt.wantErr {
					return true, &appsv1.DeploymentList{
						Items: []appsv1.Deployment{
							*emptyDeployment,
						},
					}, errors.New("deployment get error")
				} else if tt.getComponentName == tt.componentName {
					return true, &appsv1.DeploymentList{
						Items: []appsv1.Deployment{
							*deployment,
						},
					}, nil
				}

				return true, &appsv1.DeploymentList{
					Items: []appsv1.Deployment{},
				}, nil
			})

			// Verify that a component with the specified name exists
			componentExists, err := component.ComponentExists(fkclient, tt.getComponentName, tt.appName)
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			} else if !tt.wantErr && componentExists != tt.want {
				t.Errorf("expected %v, actual %v", tt.want, componentExists)
			}

		})
	}

}

func TestAddOdoProjectVolume(t *testing.T) {

	tests := []struct {
		name                         string
		containers                   []corev1.Container
		containerWithProjectVolMount []string
		volMount                     map[string]string
	}{
		{
			name:       "Case: nil passed as containers slice",
			containers: nil,
		},
		{
			name: "Case: Various containers with and without PROJECTS_ROOT",
			containers: []corev1.Container{
				{
					Name: "container1",
					Env: []corev1.EnvVar{
						{
							Name:  adaptersCommon.EnvProjectsRoot,
							Value: "/path1",
						},
					},
				},
				{
					Name: "container2",
					Env: []corev1.EnvVar{
						{
							Name:  adaptersCommon.EnvProjectsRoot,
							Value: "/path2",
						},
					},
				},
				{
					Name: "container3",
					Env: []corev1.EnvVar{
						{
							Name:  "RANDOM",
							Value: "/path3",
						},
					},
				},
			},
			containerWithProjectVolMount: []string{"container1", "container2"},
			volMount: map[string]string{
				"container1": "/path1",
				"container2": "/path2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.containers == nil {
				AddOdoProjectVolume(nil)
			} else {
				AddOdoProjectVolume(&tt.containers)
			}

			for wantContainerName, wantMountPath := range tt.volMount {
				matched := false
				for _, container := range tt.containers {
					if container.Name == wantContainerName {
						for _, volMount := range container.VolumeMounts {
							if volMount.Name == storage.OdoSourceVolume && volMount.MountPath == wantMountPath {
								matched = true
							}
						}
					}
				}

				if !matched {
					t.Error("TestAddOdoProjectVolume error: did not match the volume mount for odo-projects")
				}
			}
		})
	}
}

func TestAddOdoMandatoryVolume(t *testing.T) {
	findContainerByName := func(containers []corev1.Container, name string) (corev1.Container, bool) {
		for _, c := range containers {
			if c.Name == name {
				return c, true
			}
		}
		return corev1.Container{}, false
	}

	hasVolumeMount := func(volumeMounts []corev1.VolumeMount, mountPath string, volName string) bool {
		for _, v := range volumeMounts {
			if v.Name == volName && v.MountPath == mountPath {
				return true
			}
		}
		return false
	}

	for _, tt := range []struct {
		name             string
		containers       []corev1.Container
		wantVolumeMounts map[string]map[string]string
	}{
		{
			name:       "nil as containers slice",
			containers: nil,
		},
		{
			name: "containers with no existing volume mounts",
			containers: []corev1.Container{
				{
					Name: "container1",
				},
				{
					Name: "container2",
				},
			},
			wantVolumeMounts: map[string]map[string]string{
				"container1": {storage.SharedDataMountPath: storage.SharedDataVolumeName},
				"container2": {storage.SharedDataMountPath: storage.SharedDataVolumeName},
			},
		},
		{
			name: "containers with existing volume mounts",
			containers: []corev1.Container{
				{
					Name: "container1",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "vol1",
							MountPath: "/container1/vol1",
						},
						{
							Name:      "vol2",
							MountPath: "/container1/vol2",
						},
					},
				},
				{
					Name: "container2",
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "vol2",
							MountPath: "/container2/vol2",
						},
					},
				},
				{
					Name: "container3",
				},
			},
			wantVolumeMounts: map[string]map[string]string{
				"container1": {
					"/container1/vol1":          "vol1",
					"/container1/vol2":          "vol2",
					storage.SharedDataMountPath: storage.SharedDataVolumeName,
				},
				"container2": {
					"/container2/vol2":          "vol2",
					storage.SharedDataMountPath: storage.SharedDataVolumeName,
				},
				"container3": {storage.SharedDataMountPath: storage.SharedDataVolumeName},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.containers == nil {
				AddOdoMandatoryVolume(nil)
			} else {
				AddOdoMandatoryVolume(&tt.containers)
			}

			for containerName, volMounts := range tt.wantVolumeMounts {
				c, ok := findContainerByName(tt.containers, containerName)
				if !ok {
					t.Errorf("container %s defined in expected volume mounts, but not in container list for test",
						containerName)
				}
				for mountPath, vol := range volMounts {
					if !hasVolumeMount(c.VolumeMounts, mountPath, vol) {
						t.Errorf("expected %s to be mounted under %s in container %s", vol, mountPath, c.Name)
					}
				}
			}
		})
	}
}

func TestUpdateContainerEnvVars(t *testing.T) {
	cmd := "ls -la"
	cmp := "alias1"

	debugCommand := "nodemon --inspect={DEBUG_PORT}"
	debugComponent := "alias2"

	image := "image1"
	workDir := "/root"
	defaultCommand := []string{"tail"}
	execRunGroup := devfilev1.CommandGroup{
		IsDefault: util.GetBoolPtr(true),
		Kind:      devfilev1.RunCommandGroupKind,
	}
	execDebugGroup := devfilev1.CommandGroup{
		IsDefault: util.GetBoolPtr(true),
		Kind:      devfilev1.DebugCommandGroupKind,
	}
	defaultArgs := []string{"-f", "/dev/null"}

	devfileData, _ := data.NewDevfileData(string(data.APISchemaVersion220))
	devfileData.SetMetadata(devfilepkg.DevfileMetadata{Name: "my-app"})
	_ = devfileData.AddCommands([]devfilev1.Command{
		{
			Id: "debug-cmd",
			CommandUnion: devfilev1.CommandUnion{
				Exec: &devfilev1.ExecCommand{
					Component: cmp,
					LabeledCommand: devfilev1.LabeledCommand{
						BaseCommand: devfilev1.BaseCommand{Group: &execDebugGroup},
					},
				},
			},
		},
	})
	_ = devfileData.AddComponents([]devfilev1.Component{
		{
			Name: cmp,
			ComponentUnion: devfilev1.ComponentUnion{
				Container: &devfilev1.ContainerComponent{
					Container: devfilev1.Container{
						Image: "my-image",
					},
				},
			},
		},
	})
	devfileObj := devfileParser.DevfileObj{
		Data: devfileData,
	}

	tests := []struct {
		name         string
		debugCommand string
		debugPort    int
		containers   []corev1.Container
		execCommands []devfilev1.Command
		wantErr      bool
	}{
		{
			name: "Case: Container With Command and Args",
			containers: []corev1.Container{
				{
					Name:            cmp,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Command:         defaultCommand,
					Args:            defaultArgs,
					Env:             []corev1.EnvVar{},
				},
			},
			execCommands: []devfilev1.Command{
				{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execRunGroup,
								},
							},
							CommandLine: cmd,
							Component:   cmp,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case: Container With Command and Args but Missing Work Dir",
			containers: []corev1.Container{
				{
					Name:            cmp,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Command:         defaultCommand,
					Args:            defaultArgs,
					Env:             []corev1.EnvVar{},
				},
			},
			execCommands: []devfilev1.Command{
				{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execRunGroup,
								},
							},
							CommandLine: cmd,
							Component:   cmp,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case: Container With No Command and Args ",
			containers: []corev1.Container{
				{
					Name:            cmp,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Env:             []corev1.EnvVar{},
				},
			},
			execCommands: []devfilev1.Command{
				{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execRunGroup,
								},
							},
							CommandLine: cmd,
							Component:   cmp,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case: Custom Command Container With No Command and Args ",
			containers: []corev1.Container{
				{
					Name:            cmp,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Env:             []corev1.EnvVar{},
				},
			},
			execCommands: []devfilev1.Command{
				{
					Id: "customcommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execRunGroup,
								},
							},
							CommandLine: cmd,
							Component:   cmp,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:      "Case: empty debug command",
			debugPort: 5858,
			containers: []corev1.Container{
				{
					Name:            cmp,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Env:             []corev1.EnvVar{},
				},
				{
					Name:            debugComponent,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Env:             []corev1.EnvVar{},
				},
			},
			execCommands: []devfilev1.Command{
				{
					Id: "customruncommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execRunGroup,
								},
							},
							CommandLine: cmd,
							Component:   cmp,
							WorkingDir:  workDir,
						},
					},
				},
				{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execDebugGroup,
								},
							},
							CommandLine: debugCommand,
							Component:   debugComponent,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "Case: custom debug command",
			debugCommand: "customdebugcommand",
			debugPort:    3000,
			containers: []corev1.Container{
				{
					Name:            cmp,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Env:             []corev1.EnvVar{},
				},
			},
			execCommands: []devfilev1.Command{
				{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execRunGroup,
								},
							},
							CommandLine: cmd,
							Component:   cmp,
							WorkingDir:  workDir,
						},
					},
				},
				{
					Id: "customdebugcommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execDebugGroup,
								},
							},
							CommandLine: debugCommand,
							Component:   cmp,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "Case: custom debug command with DEBUG_PORT env already set",
			debugCommand: "customdebugcommand",
			debugPort:    3000,
			containers: []corev1.Container{
				{
					Name:            cmp,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Env: []corev1.EnvVar{
						{
							Name:  "DEBUG_PORT",
							Value: "5858",
						},
					},
				},
			},
			execCommands: []devfilev1.Command{
				{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execRunGroup,
								},
							},
							CommandLine: cmd,
							Component:   cmp,
							WorkingDir:  workDir,
						},
					},
				},
				{
					Id: "customdebugcommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execDebugGroup,
								},
							},
							CommandLine: debugCommand,
							Component:   cmp,
							WorkingDir:  workDir,
							Env:         []devfilev1.EnvVar{},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "Case: wrong custom debug command",
			debugCommand: "customdebugcommand123",
			debugPort:    9090,
			containers: []corev1.Container{
				{
					Name:            cmp,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Env:             []corev1.EnvVar{},
				},
				{
					Name:            debugComponent,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Env:             []corev1.EnvVar{},
				},
			},
			execCommands: []devfilev1.Command{
				{
					Id: "run",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execRunGroup,
								},
							},
							CommandLine: cmd,
							Component:   cmp,
							WorkingDir:  workDir,
						},
					},
				},
				{
					Id: "debug",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{
										IsDefault: util.GetBoolPtr(true),
										Kind:      devfilev1.BuildCommandGroupKind,
									},
								},
							},
							CommandLine: debugCommand,
							Component:   debugComponent,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case: custom run command with single environment variable",
			containers: []corev1.Container{
				{
					Name:            cmp,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Env:             []corev1.EnvVar{},
				},
			},
			execCommands: []devfilev1.Command{
				{
					Id: "customruncommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execRunGroup,
								},
							},
							CommandLine: cmd,
							Component:   cmp,
							WorkingDir:  workDir,
							Env: []devfilev1.EnvVar{
								{
									Name:  "env1",
									Value: "value1",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case: custom run command with multiple environment variable",
			containers: []corev1.Container{
				{
					Name:            cmp,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Env:             []corev1.EnvVar{},
				},
			},
			execCommands: []devfilev1.Command{
				{
					Id: "customruncommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execRunGroup,
								},
							},
							CommandLine: cmd,
							Component:   cmp,
							WorkingDir:  workDir,
							Env: []devfilev1.EnvVar{
								{
									Name:  "env1",
									Value: "value1",
								},
								{
									Name:  "env2",
									Value: "value2 with space",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "Case: custom debug command with single environment variable",
			debugCommand: "customdebugcommand",
			debugPort:    3000,
			containers: []corev1.Container{
				{
					Name:            cmp,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Env:             []corev1.EnvVar{},
				},
			},
			execCommands: []devfilev1.Command{
				{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execRunGroup,
								},
							},
							CommandLine: cmd,
							Component:   cmp,
							WorkingDir:  workDir,
						},
					},
				},
				{
					Id: "customdebugcommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execDebugGroup,
								},
							},
							CommandLine: debugCommand,
							Component:   cmp,
							WorkingDir:  workDir,
							Env: []devfilev1.EnvVar{
								{
									Name:  "env1",
									Value: "value1",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "Case: custom debug command with multiple environment variables",
			debugCommand: "customdebugcommand",
			debugPort:    3000,
			containers: []corev1.Container{
				{
					Name:            cmp,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Env:             []corev1.EnvVar{},
				},
			},
			execCommands: []devfilev1.Command{
				{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execRunGroup,
								},
							},
							CommandLine: cmd,
							Component:   cmp,
							WorkingDir:  workDir,
						},
					},
				},
				{
					Id: "customdebugcommand",
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execDebugGroup,
								},
							},
							CommandLine: debugCommand,
							Component:   cmp,
							WorkingDir:  workDir,
							Env: []devfilev1.EnvVar{
								{
									Name:  "env1",
									Value: "value1",
								},
								{
									Name:  "env2",
									Value: "value2 with space",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			containers, err := UpdateContainerEnvVars(devfileObj, tt.containers, "debug-cmd", tt.debugPort)

			envDebugPortMatched := false

			for _, container := range containers {
				for _, testContainer := range tt.containers {
					if container.Name == testContainer.Name {
						for _, envVar := range container.Env {
							// if the debug command is also present
							if len(tt.execCommands) >= 2 {
								if envVar.Name == adaptersCommon.EnvDebugPort {
									// check if the debug command's debugPort env was set properly
									envDebugPortMatched = true
								}
							}
						}
					}
				}
			}

			if tt.wantErr != (err != nil) {
				t.Errorf("unexpected error, wantErr: %v, err: %v", tt.wantErr, err)
			}

			if len(tt.execCommands) >= 2 && !envDebugPortMatched {
				t.Errorf("TestUpdateContainerEnvVars error: missing env var %s in container %q",
					adaptersCommon.EnvDebugPort, cmp)
			}
		})
	}
}

func TestUpdateContainersEntrypointsIfNeeded(t *testing.T) {
	const (
		runCommand            = "my-run"
		runCmdLine            = "echo my-run-command-line"
		runContainerComponent = "run-container-component"
	)
	const (
		debugCommand            = "my-debug"
		debugCmdLine            = "echo my-run-command-line"
		debugContainerComponent = "debug-container-component"
	)

	execRunGroup := devfilev1.CommandGroup{
		IsDefault: util.GetBoolPtr(true),
		Kind:      devfilev1.RunCommandGroupKind,
	}
	execDebugGroup := devfilev1.CommandGroup{
		IsDefault: util.GetBoolPtr(true),
		Kind:      devfilev1.DebugCommandGroupKind,
	}

	for _, tt := range []struct {
		name                  string
		commands              []devfilev1.Command
		components            []devfilev1.Component
		runCommand            string
		debugCommand          string
		runContainerCommand   []string
		debugContainerCommand []string
		runContainerArgs      []string
		debugContainerArgs    []string
		wantErr               bool
		//key is the container name
		expectedContainerCommand map[string][]string
		//key is the container name
		expectedContainerArgs map[string][]string
	}{
		{
			name:       "invalid run command",
			runCommand: "a-non-existing-run-name",
			wantErr:    true,
		},
		{
			name:         "invalid debug command",
			runCommand:   runCommand,
			debugCommand: "a-non-existing-debug-name",
			wantErr:      true,
		},
		{
			name:         "containers without any command or args => must be overridden with 'tail -f /dev/null'",
			runCommand:   runCommand,
			debugCommand: debugCommand,
			wantErr:      false,
			expectedContainerCommand: map[string][]string{
				runContainerComponent:   {"tail"},
				debugContainerComponent: {"tail"},
			},
			expectedContainerArgs: map[string][]string{
				runContainerComponent:   {"-f", "/dev/null"},
				debugContainerComponent: {"-f", "/dev/null"},
			},
		},
		{
			name:                "containers with one without any command or args => must be overridden with 'tail -f /dev/null'",
			runCommand:          runCommand,
			debugCommand:        debugCommand,
			wantErr:             false,
			runContainerCommand: []string{"printenv"},
			runContainerArgs:    []string{"HOSTNAME"},
			expectedContainerCommand: map[string][]string{
				runContainerComponent:   {"printenv"},
				debugContainerComponent: {"tail"},
			},
			expectedContainerArgs: map[string][]string{
				runContainerComponent:   {"HOSTNAME"},
				debugContainerComponent: {"-f", "/dev/null"},
			},
		},
		{
			name:                  "containers with explicit command or args",
			runCommand:            runCommand,
			debugCommand:          debugCommand,
			wantErr:               false,
			runContainerCommand:   []string{"printenv"},
			runContainerArgs:      []string{"HOSTNAME"},
			debugContainerCommand: []string{"tail"},
			debugContainerArgs:    []string{"-f", "/path/to/my/custom/log/file"},
			expectedContainerCommand: map[string][]string{
				runContainerComponent:   {"printenv"},
				debugContainerComponent: {"tail"},
			},
			expectedContainerArgs: map[string][]string{
				runContainerComponent:   {"HOSTNAME"},
				debugContainerComponent: {"-f", "/path/to/my/custom/log/file"},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			devfileData, err := data.NewDevfileData(string(data.APISchemaVersion220))
			if err != nil {
				t.Error(err)
			}
			err = devfileData.AddComponents([]devfilev1.Component{
				{
					Name: runContainerComponent,
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{},
						},
					},
				},
				{
					Name: debugContainerComponent,
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{},
						},
					},
				},
			})
			if err != nil {
				t.Error(err)
			}
			err = devfileData.AddCommands([]devfilev1.Command{
				{
					Id: runCommand,
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							CommandLine: runCmdLine,
							Component:   runContainerComponent,
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execRunGroup,
								},
							},
						},
					},
				},
				{
					Id: debugCommand,
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							CommandLine: debugCmdLine,
							Component:   debugContainerComponent,
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execDebugGroup,
								},
							},
						},
					},
				},
			})
			if err != nil {
				t.Error(err)
			}
			devfileObj := devfileParser.DevfileObj{
				Data: devfileData,
			}

			containerForComponents := []corev1.Container{
				{
					Name:    runContainerComponent,
					Command: tt.runContainerCommand,
					Args:    tt.runContainerArgs,
				},
				{
					Name:    debugContainerComponent,
					Command: tt.debugContainerCommand,
					Args:    tt.debugContainerArgs,
				},
			}

			containers, err := UpdateContainersEntrypointsIfNeeded(devfileObj, containerForComponents, tt.runCommand, tt.debugCommand)
			if tt.wantErr != (err != nil) {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(containers) != len(tt.expectedContainerCommand) {
				t.Errorf("length of expectedContainerCommand must match the one of containers, please fix test %q", tt.name)
			}
			if len(containers) != len(tt.expectedContainerArgs) {
				t.Errorf("length of expectedContainerArgs must match the one of containers, please fix test %q", tt.name)
			}
			for _, c := range containers {
				if len(c.Command) == 0 {
					t.Errorf("empty command for container %q", c.Name)
				}
				if len(c.Args) == 0 {
					t.Errorf("empty command for container %q", c.Args)
				}

				if !reflect.DeepEqual(tt.expectedContainerCommand[c.Name], c.Command) {
					t.Errorf("unexpected command for container %q, expected=%v, got %v",
						c.Name, tt.expectedContainerCommand[c.Name], c.Command)
				}
				if !reflect.DeepEqual(tt.expectedContainerArgs[c.Name], c.Args) {
					t.Errorf("unexpected args for container %q, expected=%v, got %v",
						c.Name, tt.expectedContainerArgs[c.Name], c.Args)
				}
			}

		})
	}
}
