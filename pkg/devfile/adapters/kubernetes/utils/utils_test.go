package utils

import (
	"reflect"
	"testing"

	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	"github.com/openshift/odo/pkg/devfile/parser/data/common"
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/testingutil"

	corev1 "k8s.io/api/core/v1"
)

func TestUpdateContainersWithSupervisord(t *testing.T) {

	command := "ls -la"
	component := "alias1"

	debugCommand := "nodemon --inspect={DEBUG_PORT}"
	debugComponent := "alias2"

	image := "image1"
	workDir := "/root"
	emptyString := ""
	defaultCommand := []string{"tail"}
	execRunGroup := versionsCommon.Group{
		IsDefault: true,
		Kind:      versionsCommon.RunCommandGroupType,
	}
	execDebugGroup := versionsCommon.Group{
		IsDefault: true,
		Kind:      versionsCommon.DebugCommandGroupType,
	}
	defaultArgs := []string{"-f", "/dev/null"}
	supervisordCommand := []string{adaptersCommon.SupervisordBinaryPath}
	supervisordArgs := []string{"-c", adaptersCommon.SupervisordConfFile}

	tests := []struct {
		name                    string
		runCommand              string
		debugCommand            string
		containers              []corev1.Container
		execCommands            []common.Exec
		componentType           common.DevfileComponentType
		isSupervisordEntrypoint bool
		wantErr                 bool
	}{
		{
			name:       "Case: Container With Command and Args",
			runCommand: emptyString,
			containers: []corev1.Container{
				{
					Name:            component,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Command:         defaultCommand,
					Args:            defaultArgs,
					Env:             []corev1.EnvVar{},
				},
			},
			execCommands: []versionsCommon.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &execRunGroup,
				},
			},
			componentType:           common.ContainerComponentType,
			isSupervisordEntrypoint: false,
			wantErr:                 false,
		},
		{
			name:       "Case: Container With Command and Args but Missing Work Dir",
			runCommand: emptyString,
			containers: []corev1.Container{
				{
					Name:            component,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Command:         defaultCommand,
					Args:            defaultArgs,
					Env:             []corev1.EnvVar{},
				},
			},
			execCommands: []versionsCommon.Exec{
				{
					CommandLine: command,
					Component:   component,
					Group:       &execRunGroup,
				},
			},
			componentType:           common.ContainerComponentType,
			isSupervisordEntrypoint: false,
			wantErr:                 false,
		},
		{
			name:       "Case: Container With No Command and Args ",
			runCommand: emptyString,
			containers: []corev1.Container{
				{
					Name:            component,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Env:             []corev1.EnvVar{},
				},
			},
			execCommands: []versionsCommon.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &execRunGroup,
				},
			},
			componentType:           common.ContainerComponentType,
			isSupervisordEntrypoint: true,
			wantErr:                 false,
		},
		{
			name:       "Case: Custom Command Container With No Command and Args ",
			runCommand: "customcommand",
			containers: []corev1.Container{
				{
					Name:            component,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Env:             []corev1.EnvVar{},
				},
			},
			execCommands: []versionsCommon.Exec{
				{
					Id:          "customcommand",
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &execRunGroup,
				},
			},
			componentType:           common.ContainerComponentType,
			isSupervisordEntrypoint: true,
			wantErr:                 false,
		},
		{
			name:       "Case: Wrong Custom Command Container",
			runCommand: "customcommand123",
			containers: []corev1.Container{
				{
					Name:            component,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Env:             []corev1.EnvVar{},
				},
			},
			execCommands: []versionsCommon.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &execRunGroup,
				},
			},
			componentType:           common.ContainerComponentType,
			isSupervisordEntrypoint: true,
			wantErr:                 true,
		},

		{
			name:         "Case: empty debug command",
			runCommand:   emptyString,
			debugCommand: emptyString,
			containers: []corev1.Container{
				{
					Name:            component,
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
			execCommands: []versionsCommon.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &execRunGroup,
				},
				{
					CommandLine: debugCommand,
					Component:   debugComponent,
					WorkingDir:  workDir,
					Group:       &execDebugGroup,
				},
			},
			componentType:           common.ContainerComponentType,
			isSupervisordEntrypoint: true,
			wantErr:                 false,
		},
		{
			name:         "Case: custom debug command",
			runCommand:   "",
			debugCommand: "customdebugcommand",
			containers: []corev1.Container{
				{
					Name:            component,
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Env:             []corev1.EnvVar{},
				},
			},
			execCommands: []versionsCommon.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &execRunGroup,
				},
				{
					Id:          "customdebugcommand",
					CommandLine: debugCommand,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &execDebugGroup,
				},
			},
			componentType:           common.ContainerComponentType,
			isSupervisordEntrypoint: true,
			wantErr:                 false,
		},
		{
			name:         "Case: wrong custom debug command",
			runCommand:   "",
			debugCommand: "customdebugcommand123",
			containers: []corev1.Container{
				{
					Name:            component,
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
			execCommands: []versionsCommon.Exec{
				{
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
					Group:       &execRunGroup,
				},
				{
					CommandLine: debugCommand,
					Component:   debugComponent,
					WorkingDir:  workDir,
					Group: &versionsCommon.Group{
						IsDefault: true,
						Kind:      versionsCommon.BuildCommandGroupType,
					},
				},
			},
			componentType:           common.ContainerComponentType,
			isSupervisordEntrypoint: true,
			wantErr:                 true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					Components: []versionsCommon.DevfileComponent{
						{
							Container: &versionsCommon.Container{
								Name: component,
							},
						},
						{
							Container: &versionsCommon.Container{
								Name: debugComponent,
							},
						},
					},
					ExecCommands: tt.execCommands,
				},
			}

			containers, err := UpdateContainersWithSupervisord(devObj, tt.containers, tt.runCommand, tt.debugCommand, 5858)

			if !tt.wantErr && err != nil {
				t.Errorf("TestUpdateContainersWithSupervisord unxpected error: %v", err)
			} else if tt.wantErr && err != nil {
				// return since we dont want to test anything further
				return
			} else if tt.wantErr && err == nil {
				t.Error("wanted error but got not error")
			}

			// Check if the supervisord volume has been mounted
			supervisordVolumeMountMatched := false
			envRunMatched := false
			envWorkDirMatched := false
			envDebugMatched := false
			envDebugWorkDirMatched := false

			if tt.execCommands[0].WorkingDir == "" {
				// if workdir is not present, dont test for matching the env
				envWorkDirMatched = true
			}

			if len(tt.execCommands) >= 2 && tt.execCommands[1].WorkingDir == "" {
				// if workdir is not present, dont test for matching the env
				envDebugWorkDirMatched = true
			}

			for _, container := range containers {
				for _, testContainer := range tt.containers {
					if container.Name == testContainer.Name {
						for _, volumeMount := range container.VolumeMounts {
							if volumeMount.Name == adaptersCommon.SupervisordVolumeName && volumeMount.MountPath == adaptersCommon.SupervisordMountPath {
								supervisordVolumeMountMatched = true
							}
						}

						for _, envVar := range container.Env {
							if envVar.Name == adaptersCommon.EnvOdoCommandRun && envVar.Value == tt.execCommands[0].CommandLine {
								envRunMatched = true
							}
							if tt.execCommands[0].WorkingDir != "" && envVar.Name == adaptersCommon.EnvOdoCommandRunWorkingDir && envVar.Value == tt.execCommands[0].WorkingDir {
								envWorkDirMatched = true
							}
							if len(tt.execCommands) >= 2 && envVar.Name == adaptersCommon.EnvOdoCommandDebug && envVar.Value == tt.execCommands[1].CommandLine {
								envDebugMatched = true
							}
							if len(tt.execCommands) >= 2 && tt.execCommands[1].WorkingDir != "" && envVar.Name == adaptersCommon.EnvOdoCommandDebugWorkingDir && envVar.Value == tt.execCommands[1].WorkingDir {
								envDebugWorkDirMatched = true
							}
						}

						if tt.isSupervisordEntrypoint && (!reflect.DeepEqual(container.Command, supervisordCommand) || !reflect.DeepEqual(container.Args, supervisordArgs)) {
							t.Errorf("TestUpdateContainersWithSupervisord error: commands and args mismatched for container %v, expected command: %v actual command: %v, expected args: %v actual args: %v", component, supervisordCommand, container.Command, supervisordArgs, container.Args)
						} else if !tt.isSupervisordEntrypoint && (!reflect.DeepEqual(container.Command, defaultCommand) || !reflect.DeepEqual(container.Args, defaultArgs)) {
							t.Errorf("TestUpdateContainersWithSupervisord error: commands and args mismatched for container %v, expected command: %v actual command: %v, expected args: %v actual args: %v", component, defaultCommand, container.Command, defaultArgs, container.Args)
						}
					}
				}
			}
			if !supervisordVolumeMountMatched {
				t.Errorf("TestUpdateContainersWithSupervisord error: could not find supervisord volume mounts for container %v", component)
			}
			if !envRunMatched || !envWorkDirMatched {
				t.Errorf("TestUpdateContainersWithSupervisord error: could not find env vars for supervisord in container %v, found command env: %v, found work dir env: %v", component, envRunMatched, envWorkDirMatched)
			}

			if len(tt.execCommands) >= 2 && (!envDebugMatched || !envDebugWorkDirMatched) {
				t.Errorf("TestUpdateContainersWithSupervisord error: could not find env vars for supervisord in container %v, found debug env: %v, found work dir env: %v", component, envDebugMatched, envDebugWorkDirMatched)
			}
		})
	}

}
