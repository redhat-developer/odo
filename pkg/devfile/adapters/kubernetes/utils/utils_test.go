package utils

import (
	"testing"

	"github.com/devfile/library/pkg/devfile/parser/data"
	"github.com/google/go-cmp/cmp"

	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/util"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	corev1 "k8s.io/api/core/v1"
)

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
							Name:  _envProjectsRoot,
							Value: "/path1",
						},
					},
				},
				{
					Name: "container2",
					Env: []corev1.EnvVar{
						{
							Name:  _envProjectsRoot,
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

func TestUpdateContainersEntrypointsIfNeeded(t *testing.T) {
	const (
		buildCommand            = "my-build"
		buildCmdLine            = "echo my-build-command-line"
		buildContainerComponent = "build-container-component"
	)
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

	execBuildGroup := devfilev1.CommandGroup{
		IsDefault: util.GetBoolPtr(true),
		Kind:      devfilev1.BuildCommandGroupKind,
	}
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
		buildCommand          string
		runCommand            string
		debugCommand          string
		buildContainerCommand []string
		runContainerCommand   []string
		debugContainerCommand []string
		buildContainerArgs    []string
		runContainerArgs      []string
		debugContainerArgs    []string
		wantErr               bool
		// key is the container name
		expectedContainerCommand map[string][]string
		// key is the container name
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
			name:         "missing build command specified by name",
			buildCommand: buildCommand + "-not-found",
			runCommand:   runCommand,
			debugCommand: debugCommand,
			wantErr:      true,
		},
		{
			name:         "containers without any command or args => must be overridden with 'tail -f /dev/null'",
			buildCommand: buildCommand,
			runCommand:   runCommand,
			debugCommand: debugCommand,
			wantErr:      false,
			expectedContainerCommand: map[string][]string{
				buildContainerComponent: {"tail"},
				runContainerComponent:   {"tail"},
				debugContainerComponent: {"tail"},
			},
			expectedContainerArgs: map[string][]string{
				buildContainerComponent: {"-f", "/dev/null"},
				runContainerComponent:   {"-f", "/dev/null"},
				debugContainerComponent: {"-f", "/dev/null"},
			},
		},
		{
			name:                  "containers with one without any command or args => must be overridden with 'tail -f /dev/null'",
			buildCommand:          buildCommand,
			runCommand:            runCommand,
			debugCommand:          debugCommand,
			wantErr:               false,
			buildContainerCommand: []string{"npm"},
			buildContainerArgs:    []string{"install"},
			runContainerCommand:   []string{"printenv"},
			runContainerArgs:      []string{"HOSTNAME"},
			expectedContainerCommand: map[string][]string{
				buildContainerComponent: {"npm"},
				runContainerComponent:   {"printenv"},
				debugContainerComponent: {"tail"},
			},
			expectedContainerArgs: map[string][]string{
				buildContainerComponent: {"install"},
				runContainerComponent:   {"HOSTNAME"},
				debugContainerComponent: {"-f", "/dev/null"},
			},
		},
		{
			name:                "default build command, containers with one without any command or args => must be overridden with 'tail -f /dev/null'",
			runCommand:          runCommand,
			debugCommand:        debugCommand,
			wantErr:             false,
			runContainerCommand: []string{"printenv"},
			runContainerArgs:    []string{"HOSTNAME"},
			expectedContainerCommand: map[string][]string{
				buildContainerComponent: {"tail"},
				runContainerComponent:   {"printenv"},
				debugContainerComponent: {"tail"},
			},
			expectedContainerArgs: map[string][]string{
				buildContainerComponent: {"-f", "/dev/null"},
				runContainerComponent:   {"HOSTNAME"},
				debugContainerComponent: {"-f", "/dev/null"},
			},
		},
		{
			name:                  "containers with explicit command or args",
			buildCommand:          buildCommand,
			runCommand:            runCommand,
			debugCommand:          debugCommand,
			wantErr:               false,
			buildContainerCommand: []string{"npm"},
			buildContainerArgs:    []string{"install"},
			runContainerCommand:   []string{"printenv"},
			runContainerArgs:      []string{"HOSTNAME"},
			debugContainerCommand: []string{"tail"},
			debugContainerArgs:    []string{"-f", "/path/to/my/custom/log/file"},
			expectedContainerCommand: map[string][]string{
				buildContainerComponent: {"npm"},
				runContainerComponent:   {"printenv"},
				debugContainerComponent: {"tail"},
			},
			expectedContainerArgs: map[string][]string{
				buildContainerComponent: {"install"},
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
					Name: buildContainerComponent,
					ComponentUnion: devfilev1.ComponentUnion{
						Container: &devfilev1.ContainerComponent{
							Container: devfilev1.Container{},
						},
					},
				},
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
					Id: buildCommand,
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							CommandLine: buildCmdLine,
							Component:   buildContainerComponent,
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &execBuildGroup,
								},
							},
						},
					},
				},
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
					Name:    buildContainerComponent,
					Command: tt.buildContainerCommand,
					Args:    tt.buildContainerArgs,
				},
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

			containers, err := UpdateContainersEntrypointsIfNeeded(devfileObj, containerForComponents, tt.buildCommand, tt.runCommand, tt.debugCommand)
			if tt.wantErr != (err != nil) {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
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

				if diff := cmp.Diff(tt.expectedContainerCommand[c.Name], c.Command); diff != "" {
					t.Errorf("UpdateContainersEntrypointsIfNeeded() expectedContainerCommand[%s] mismatch (-want +got):\n%s", c.Name, diff)
				}
				if diff := cmp.Diff(tt.expectedContainerArgs[c.Name], c.Args); diff != "" {
					t.Errorf("UpdateContainersEntrypointsIfNeeded() expectedContainerArgs[%s] mismatch (-want +got):\n%s", c.Name, diff)
				}
			}

		})
	}
}
