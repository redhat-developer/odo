package component

import (
	"strings"
	"testing"

	"github.com/devfile/library/pkg/devfile/parser/data"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/lclient"
)

func TestExecDevfile(t *testing.T) {

	testComponentName := "test"
	command := "ls -la"
	workDir := "/tmp"
	component := "alias1"

	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name                string
		client              *lclient.Client
		pushDevfileCommands adaptersCommon.PushCommandsMap
		componentExists     bool
		wantErr             bool
	}{
		{
			name:   "Case 1: Successful devfile command exec of devbuild and devrun",
			client: fakeClient,
			pushDevfileCommands: adaptersCommon.PushCommandsMap{
				devfilev1.RunCommandGroupKind: devfilev1.Command{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: devfilev1.RunCommandGroupKind},
								},
							},
							CommandLine: command,
							WorkingDir:  workDir,
							Component:   component,
						},
					},
				},
				devfilev1.BuildCommandGroupKind: devfilev1.Command{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: devfilev1.BuildCommandGroupKind},
								},
							},
							CommandLine: command,
							WorkingDir:  workDir,
							Component:   component,
						},
					},
				},
			},
			componentExists: false,
			wantErr:         false,
		},
		{
			name:   "Case 2: Successful devfile command exec of devrun",
			client: fakeClient,
			pushDevfileCommands: adaptersCommon.PushCommandsMap{
				devfilev1.RunCommandGroupKind: devfilev1.Command{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: devfilev1.RunCommandGroupKind},
								},
							},
							CommandLine: command,
							WorkingDir:  workDir,
							Component:   component,
						},
					},
				},
			},
			componentExists: true,
			wantErr:         false,
		},
		{
			name:                "Case 3: No devfile push commands should result in an err",
			client:              fakeClient,
			pushDevfileCommands: adaptersCommon.PushCommandsMap{},
			componentExists:     false,
			wantErr:             true,
		},
		{
			name:   "Case 4: Unsuccessful devfile command exec of devrun",
			client: fakeErrorClient,
			pushDevfileCommands: adaptersCommon.PushCommandsMap{
				devfilev1.RunCommandGroupKind: devfilev1.Command{
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{Kind: devfilev1.RunCommandGroupKind},
								},
							},
							CommandLine: command,
							WorkingDir:  workDir,
							Component:   component,
						},
					},
				},
			},
			componentExists: true,
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error()
					}
					err = devfileData.AddComponents([]devfilev1.Component{})
					if err != nil {
						t.Error()
					}
					return devfileData
				}(),
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			componentAdapter := New(adapterCtx, *tt.client)
			err := componentAdapter.ExecDevfile(tt.pushDevfileCommands, tt.componentExists, adaptersCommon.PushParameters{Show: false})
			if !tt.wantErr && err != nil {
				t.Errorf("TestExecDevfile error: unexpected error during executing devfile commands: %v", err)
			}
		})
	}
}

func TestExecTestCmd(t *testing.T) {

	testComponentName := "test"
	command := "ls -la"
	workDir := "/tmp"
	component := "alias1"

	validComponents := []devfilev1.Component{
		{
			Name: component,
			ComponentUnion: devfilev1.ComponentUnion{
				Container: &devfilev1.ContainerComponent{
					Container: devfilev1.Container{
						Image: "image",
					},
				},
			},
		},
	}

	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name               string
		client             *lclient.Client
		testDevfileCommand devfilev1.Command
		wantErr            bool
	}{
		{
			name:   "Case 1: Successful execute test command",
			client: fakeClient,
			testDevfileCommand: devfilev1.Command{
				CommandUnion: devfilev1.CommandUnion{
					Exec: &devfilev1.ExecCommand{
						LabeledCommand: devfilev1.LabeledCommand{
							BaseCommand: devfilev1.BaseCommand{
								Group: &devfilev1.CommandGroup{Kind: devfilev1.TestCommandGroupKind},
							},
						},
						CommandLine: command,
						WorkingDir:  workDir,
						Component:   component,
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "Case 2: No devfile test commands should result in an err",
			client: fakeClient,
			testDevfileCommand: devfilev1.Command{
				CommandUnion: devfilev1.CommandUnion{
					Exec: &devfilev1.ExecCommand{
						LabeledCommand: devfilev1.LabeledCommand{
							BaseCommand: devfilev1.BaseCommand{
								Group: &devfilev1.CommandGroup{Kind: devfilev1.BuildCommandGroupKind},
							},
						},
						CommandLine: command,
						WorkingDir:  workDir,
						Component:   component,
					},
				},
			},
			wantErr: true,
		},
		{
			name:   "Case 3: Unsuccessful exec test command",
			client: fakeErrorClient,
			testDevfileCommand: devfilev1.Command{
				CommandUnion: devfilev1.CommandUnion{
					Exec: &devfilev1.ExecCommand{
						LabeledCommand: devfilev1.LabeledCommand{
							BaseCommand: devfilev1.BaseCommand{
								Group: &devfilev1.CommandGroup{Kind: devfilev1.TestCommandGroupKind},
							},
						},
						CommandLine: command,
						WorkingDir:  workDir,
						Component:   component,
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error()
					}
					err = devfileData.AddComponents(validComponents)
					if err != nil {
						t.Error()
					}
					return devfileData
				}(),
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			componentAdapter := New(adapterCtx, *tt.client)
			err := componentAdapter.ExecuteDevfileCommand(tt.testDevfileCommand, false)
			if !tt.wantErr && err != nil {
				t.Errorf("TestExecTestCmd error: unexpected error during executing devfile commands: %v", err)
			}
		})
	}
}

func TestCreateProjectVolumeIfReqd(t *testing.T) {
	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name           string
		componentName  string
		client         *lclient.Client
		wantVolumeName string
		wantErr        bool
	}{
		{
			name:           "Case 1: Volume does not exist",
			componentName:  "somecomponent",
			client:         fakeClient,
			wantVolumeName: lclient.ProjectSourceVolumeName + "-somecomponent",
			wantErr:        false,
		},
		{
			name:           "Case 2: Volume exist",
			componentName:  "test",
			client:         fakeClient,
			wantVolumeName: lclient.ProjectSourceVolumeName + "-test",
			wantErr:        false,
		},
		{
			name:           "Case 3: More than one project volume exist",
			componentName:  "duplicate",
			client:         fakeClient,
			wantVolumeName: "",
			wantErr:        true,
		},
		{
			name:           "Case 4: Client error",
			componentName:  "random",
			client:         fakeErrorClient,
			wantVolumeName: "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error()
					}
					err = devfileData.AddComponents([]devfilev1.Component{})
					if err != nil {
						t.Error()
					}
					return devfileData
				}(),
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: tt.componentName,
				Devfile:       devObj,
			}

			componentAdapter := New(adapterCtx, *tt.client)
			volumeName, err := componentAdapter.createProjectVolumeIfReqd()
			if !tt.wantErr && err != nil {
				t.Errorf("TestCreateAndGetProjectVolume error: Unexpected error: %v", err)
			} else if !tt.wantErr && !strings.Contains(volumeName, tt.wantVolumeName) {
				t.Errorf("TestCreateAndGetProjectVolume error: project volume name did not match, expected: %v got: %v", tt.wantVolumeName, volumeName)
			}
		})
	}
}

func TestStartBootstrapSupervisordInitContainer(t *testing.T) {

	supervisordVolumeName := "supervisord"
	componentName := "myComponent"

	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name    string
		client  *lclient.Client
		wantErr bool
	}{
		{
			name:    "Case 1: Successfully create a bootstrap container",
			client:  fakeClient,
			wantErr: false,
		},
		{
			name:    "Case 2: Failed to create a bootstrap container ",
			client:  fakeErrorClient,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error()
					}
					err = devfileData.AddComponents([]devfilev1.Component{})
					if err != nil {
						t.Error()
					}
					return devfileData
				}(),
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: componentName,
				Devfile:       devObj,
			}

			componentAdapter := New(adapterCtx, *tt.client)
			err := componentAdapter.startBootstrapSupervisordInitContainer(supervisordVolumeName)
			if !tt.wantErr && err != nil {
				t.Errorf("TestStartBootstrapSupervisordInitContainer: unexpected error got: %v wanted: %v", err, tt.wantErr)
			}
		})
	}

}

func TestCreateAndInitSupervisordVolumeIfReqd(t *testing.T) {

	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	componentName := "myComponent"

	tests := []struct {
		name    string
		client  *lclient.Client
		wantErr bool
	}{
		{
			name:    "Case 1: Successfully create a bootstrap vol and container",
			client:  fakeClient,
			wantErr: false,
		},
		{
			name:    "Case 2: Failed to create a bootstrap vol and container ",
			client:  fakeErrorClient,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error()
					}
					err = devfileData.AddComponents([]devfilev1.Component{})
					if err != nil {
						t.Error()
					}
					return devfileData
				}(),
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: componentName,
				Devfile:       devObj,
			}

			componentAdapter := New(adapterCtx, *tt.client)
			volName, err := componentAdapter.createAndInitSupervisordVolumeIfReqd(false)
			if !tt.wantErr && err != nil {
				t.Errorf("TestCreateAndInitSupervisordVolume: unexpected error %v, wanted %v", err, tt.wantErr)
			} else if !tt.wantErr && !strings.Contains(volName, adaptersCommon.SupervisordVolumeName+"-"+componentName) {
				t.Errorf("TestCreateAndInitSupervisordVolume: unexpected supervisord vol name, expected: %v got: %v", adaptersCommon.SupervisordVolumeName, volName)
			}
		})
	}

}
