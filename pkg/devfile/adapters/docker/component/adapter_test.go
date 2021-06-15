package component

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/devfile/library/pkg/devfile/parser/data"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfileParser "github.com/devfile/library/pkg/devfile/parser"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	volumeTypes "github.com/docker/docker/api/types/volume"
	"github.com/golang/mock/gomock"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/lclient"
)

func TestPush(t *testing.T) {

	testComponentName := "test"
	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	command := "ls -la"
	component := "alias1"
	workDir := "/root"

	// create a temp dir for the file indexer
	directory, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("TestPush error: error creating temporary directory for the indexer: %v", err)
	}

	pushParams := adaptersCommon.PushParameters{
		Path:              directory,
		WatchFiles:        []string{},
		WatchDeletedFiles: []string{},
		IgnoredFiles:      []string{},
		ForceBuild:        false,
	}

	execCommands := []devfilev1.Command{
		{
			CommandUnion: devfilev1.CommandUnion{
				Exec: &devfilev1.ExecCommand{
					LabeledCommand: devfilev1.LabeledCommand{
						BaseCommand: devfilev1.BaseCommand{
							Group: &devfilev1.CommandGroup{
								Kind:      devfilev1.RunCommandGroupKind,
								IsDefault: true,
							},
						},
					},
					CommandLine: command,
					Component:   component,
					WorkingDir:  workDir,
				},
			},
		},
	}
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

	tests := []struct {
		name          string
		components    []devfilev1.Component
		componentType devfilev1.ComponentType
		client        *lclient.Client
		wantErr       bool
	}{
		{
			name:          "Case 1: Invalid devfile",
			componentType: "",
			components:    []devfilev1.Component{},
			client:        fakeClient,
			wantErr:       true,
		},
		{
			name:          "Case 2: Valid devfile",
			components:    validComponents,
			componentType: devfilev1.ContainerComponentType,
			client:        fakeClient,
			wantErr:       false,
		},
		{
			name:          "Case 3: Valid devfile, docker client error",
			components:    validComponents,
			componentType: devfilev1.ContainerComponentType,
			client:        fakeErrorClient,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents(tt.components)
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(execCommands)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			componentAdapter := New(adapterCtx, *tt.client)
			// ToDo: Add more meaningful unit tests once Push actually does something with its parameters
			err := componentAdapter.Push(pushParams)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter create unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Remove the temp dir created for the file indexer
	err = os.RemoveAll(directory)
	if err != nil {
		t.Errorf("TestPush error: error deleting the temp dir %s", directory)
	}

}

func TestDockerTest(t *testing.T) {

	testComponentName := "test"
	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	command := "ls -la"
	component := "alias1"
	workDir := "/root"
	id := "testcmd"

	// create a temp dir for the file indexer
	directory, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf("TestPush error: error creating temporary directory for the indexer: %v", err)
	}

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

	tests := []struct {
		name          string
		components    []devfilev1.Component
		componentType devfilev1.ComponentType
		client        *lclient.Client
		execCommands  []devfilev1.Command
		wantErr       bool
	}{
		{
			name:         "Case 1: Invalid devfile",
			components:   validComponents,
			execCommands: []devfilev1.Command{},
			client:       fakeClient,
			wantErr:      true,
		},
		{
			name:       "Case 2: Valid devfile",
			components: validComponents,
			client:     fakeClient,
			execCommands: []devfilev1.Command{
				{
					Id: id,
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{
										Kind: devfilev1.TestCommandGroupKind,
									},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "Case 3: Valid devfile, docker client error",
			components: validComponents,
			client:     fakeErrorClient,
			wantErr:    true,
		},
		{
			name:       "Case 4: No valid containers",
			components: []devfilev1.Component{},
			execCommands: []devfilev1.Command{
				{
					Id: id,
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{
										Kind: devfilev1.TestCommandGroupKind,
									},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			client:  fakeClient,
			wantErr: true,
		},
		{
			name:       "Case 5: Invalid command",
			components: []devfilev1.Component{},
			execCommands: []devfilev1.Command{
				{
					Id: id,
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{
										Kind: devfilev1.TestCommandGroupKind,
									},
								},
							},
							CommandLine: "",
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			client:  fakeClient,
			wantErr: true,
		},
		{
			name:       "Case 6: No valid command group",
			components: []devfilev1.Component{},
			execCommands: []devfilev1.Command{
				{
					Id: id,
					CommandUnion: devfilev1.CommandUnion{
						Exec: &devfilev1.ExecCommand{
							LabeledCommand: devfilev1.LabeledCommand{
								BaseCommand: devfilev1.BaseCommand{
									Group: &devfilev1.CommandGroup{
										Kind: devfilev1.RunCommandGroupKind,
									},
								},
							},
							CommandLine: command,
							Component:   component,
							WorkingDir:  workDir,
						},
					},
				},
			},
			client:  fakeClient,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents(tt.components)
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddCommands(tt.execCommands)
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			componentAdapter := New(adapterCtx, *tt.client)
			err := componentAdapter.Test(id, false)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter create unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Remove the temp dir created for the file indexer
	err = os.RemoveAll(directory)
	if err != nil {
		t.Errorf("TestTest error: error deleting the temp dir %s", directory)
	}

}

func TestAdapterDelete(t *testing.T) {
	type args struct {
		labels map[string]string
	}
	tests := []struct {
		name              string
		args              args
		componentName     string
		componentExists   bool
		skipContainerList bool
		wantErr           bool
	}{
		{
			name: "Case 1: component exists and given labels are valid",
			args: args{labels: map[string]string{
				"component": "component",
			}},
			componentName:   "component",
			componentExists: true,
			wantErr:         false,
		},
		{
			name:              "Case 2: component exists and given labels are not valid",
			args:              args{labels: nil},
			componentName:     "component",
			componentExists:   true,
			wantErr:           true,
			skipContainerList: true,
		},
		{
			name: "Case 3: component doesn't exists",
			args: args{labels: map[string]string{
				"component": "component",
			}},
			componentName:   "component",
			componentExists: false,
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			containerID := "my-id"
			volumeID := "my-volume-name"

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{})
					if err != nil {
						t.Error(err)
					}
					return devfileData
				}(),
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: tt.componentName,
				Devfile:       devObj,
			}

			if !tt.componentExists {
				adapterCtx.ComponentName = "doesNotExists"
			}

			fkclient, mockDockerClient := lclient.FakeNewMockClient(ctrl)

			a := New(adapterCtx, *fkclient)

			labeledContainers := []types.Container{}

			if tt.componentExists {
				labeledContainers = []types.Container{
					{
						ID: containerID,
						Labels: map[string]string{
							"component": tt.componentName,
						},
						Mounts: []types.MountPoint{
							{
								Type: mount.TypeVolume,
								Name: volumeID,
							},
						},
					},
				}

			}

			if !tt.skipContainerList {
				mockDockerClient.EXPECT().ContainerList(gomock.Any(), gomock.Any()).Return(labeledContainers, nil)

				if tt.componentExists {
					mockDockerClient.EXPECT().VolumeList(gomock.Any(), gomock.Any()).Return(volumeTypes.VolumeListOKBody{
						Volumes: []*types.Volume{
							{
								Name: volumeID,
								Labels: map[string]string{
									"component": tt.componentName,
									"type":      "projects",
								},
							},
						},
					}, nil)

					mockDockerClient.EXPECT().ContainerRemove(gomock.Any(), gomock.Eq(containerID), gomock.Any()).Return(nil)

					mockDockerClient.EXPECT().VolumeRemove(gomock.Any(), gomock.Eq(volumeID), gomock.Eq(true)).Return(nil)

				}
			}

			if err := a.Delete(tt.args.labels, false, false); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAdapterDeleteVolumes(t *testing.T) {

	// Convenience func to create a mock ODO-style container with the given volume mounts
	containerWithMount := func(componentName string, mountPoints []types.MountPoint) types.Container {

		return types.Container{
			ID: componentName,
			Labels: map[string]string{
				"component": componentName,
			},
			Mounts: mountPoints,
		}
	}

	componentName := "my-component"
	anotherComponentName := "other-component"

	// The purpose of these tests is to verify the correctness of container deletion, such as:
	// - Only volumes that match the format of an ODO-managed volume (storage or source) are deleted
	// - Ensure that bind mounts are never deleted
	// - Ensure that other component's volumes are never deleted
	// - Ensure that volumes that have only the exact source/storage labels format are deleted

	tests := []struct {
		name           string
		containers     []types.Container
		volumes        []*types.Volume
		expectToDelete []string
	}{
		{
			name: "Case 1: Should delete both storage and source mount",
			containers: []types.Container{
				containerWithMount(componentName,
					[]types.MountPoint{
						{
							Name: "my-src-mount",
							Type: mount.TypeVolume,
						},
						{
							Name: "my-storage-mount",
							Type: mount.TypeVolume,
						},
					}),
			},
			volumes: []*types.Volume{
				{
					Name: "my-src-mount",
					Labels: map[string]string{
						"component": componentName,
						"type":      "projects",
					},
				},
				{
					Name: "my-storage-mount",
					Labels: map[string]string{
						"component":    componentName,
						"storage-name": "anyval",
					},
				},
			},
			expectToDelete: []string{
				"my-src-mount",
				"my-storage-mount",
			},
		},
		{
			name: "Case 2: Should delete storage mount alone",
			containers: []types.Container{
				containerWithMount(componentName,
					[]types.MountPoint{
						{
							Name: "my-storage-mount",
							Type: mount.TypeVolume,
						},
					}),
			},
			volumes: []*types.Volume{
				{
					Name: "my-storage-mount",
					Labels: map[string]string{
						"component":    componentName,
						"storage-name": "anyval",
					},
				},
			},
			expectToDelete: []string{
				"my-storage-mount",
			},
		},
		{
			name: "Case 3: Should not delete a bind mount even if it matches src volume labels",
			containers: []types.Container{
				containerWithMount(componentName,
					[]types.MountPoint{
						{
							Name: "my-src-mount",
							Type: mount.TypeBind,
						},
					}),
			},

			volumes: []*types.Volume{
				{
					Name: "my-src-mount",
					Labels: map[string]string{
						"component": componentName,
						"type":      "projects",
					},
				},
			},
			expectToDelete: []string{},
		},
		{
			name: "Case 4: Should not try to delete other component's volumes",
			containers: []types.Container{
				containerWithMount(componentName,
					[]types.MountPoint{
						{
							Name: "my-src-mount",
							Type: mount.TypeVolume,
						},
						{
							Name: "my-storage-mount",
							Type: mount.TypeVolume,
						},
					}),
				containerWithMount(anotherComponentName,
					[]types.MountPoint{
						{
							Name: "my-src-mount-other-component",
							Type: mount.TypeVolume,
						},
						{
							Name: "my-storage-mount-other-component",
							Type: mount.TypeVolume,
						},
					}),
			},
			volumes: []*types.Volume{
				{
					Name: "my-src-mount",
					Labels: map[string]string{
						"component": componentName,
						"type":      "projects",
					},
				},
				{
					Name: "my-storage-mount",
					Labels: map[string]string{
						"component":    componentName,
						"storage-name": "anyval",
					},
				},
				{
					Name: "my-src-mount-other-component",
					Labels: map[string]string{
						"component": anotherComponentName,
						"type":      "projects",
					},
				},
				{
					Name: "my-storage-mount-other-component",
					Labels: map[string]string{
						"component":    anotherComponentName,
						"storage-name": "anyval",
					},
				},
			},
			expectToDelete: []string{
				"my-src-mount",
				"my-storage-mount",
			},
		},
		{
			name: "Case 5: Should not try to delete a component's non-ODO volumes, even if the format is very close to ODO",
			containers: []types.Container{containerWithMount("my-component",
				[]types.MountPoint{
					{
						Name: "my-src-mount",
						Type: mount.TypeVolume,
					},
					{
						Name: "my-storage-mount",
						Type: mount.TypeVolume,
					},
					{
						Name: "another-volume-1",
						Type: mount.TypeVolume,
					},
					{
						Name: "another-volume-2",
						Type: mount.TypeVolume,
					},
				})},
			volumes: []*types.Volume{
				{
					Name: "my-src-mount",
					Labels: map[string]string{
						"component": componentName,
						"type":      "projects",
					},
				},
				{
					Name: "my-storage-mount",
					Labels: map[string]string{
						"component":    componentName,
						"storage-name": "anyval",
					},
				},
				{
					Name: "another-volume-1",
					Labels: map[string]string{
						"component": componentName,
						"type":      "projects-but-not-really",
					},
				},
				{
					Name: "another-volume-2",
					Labels: map[string]string{
						"component":                   componentName,
						"storage-name-but-not-really": "anyval",
					},
				},
			},
			expectToDelete: []string{
				"my-src-mount",
				"my-storage-mount",
			},
		},
		{
			name: "Case 6: Should not delete a volume that is mounted into another container",
			containers: []types.Container{

				containerWithMount("my-component",
					[]types.MountPoint{
						{
							Name: "my-storage-mount",
							Type: mount.TypeVolume,
						},
					}),

				containerWithMount("a-non-odo-container-for-example",
					[]types.MountPoint{
						{
							Name: "my-storage-mount",
							Type: mount.TypeVolume,
						},
					}),
			},
			volumes: []*types.Volume{
				{
					Name: "my-storage-mount",
					Labels: map[string]string{
						"component":    componentName,
						"storage-name": "anyval",
					},
				},
			},
			expectToDelete: []string{},
		},
		{
			name: "Case 7: Should delete both storage and supervisord mount",
			containers: []types.Container{
				containerWithMount(componentName,
					[]types.MountPoint{
						{
							Name: "my-supervisord-mount",
							Type: mount.TypeVolume,
						},
						{
							Name: "my-storage-mount",
							Type: mount.TypeVolume,
						},
					}),
			},
			volumes: []*types.Volume{
				{
					Name: "my-supervisord-mount",
					Labels: map[string]string{
						"component": componentName,
						"type":      "supervisord",
						"image":     "supervisordimage",
						"version":   "supervisordversion",
					},
				},
				{
					Name: "my-storage-mount",
					Labels: map[string]string{
						"component":    componentName,
						"storage-name": "anyval",
					},
				},
			},
			expectToDelete: []string{
				"my-supervisord-mount",
				"my-storage-mount",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			devObj := devfileParser.DevfileObj{
				Data: func() data.DevfileData {
					devfileData, err := data.NewDevfileData(string(data.APISchemaVersion200))
					if err != nil {
						t.Error(err)
					}
					err = devfileData.AddComponents([]devfilev1.Component{})
					if err != nil {
						t.Error(err)
					}

					return devfileData
				}(),
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: componentName,
				Devfile:       devObj,
			}

			fkclient, mockDockerClient := lclient.FakeNewMockClient(ctrl)

			a := New(adapterCtx, *fkclient)

			arg := map[string]string{
				"component": componentName,
			}

			mockDockerClient.EXPECT().ContainerList(gomock.Any(), gomock.Any()).Return(tt.containers, nil)

			mockDockerClient.EXPECT().ContainerRemove(gomock.Any(), componentName, gomock.Any()).Return(nil)

			mockDockerClient.EXPECT().VolumeList(gomock.Any(), gomock.Any()).Return(volumeTypes.VolumeListOKBody{
				Volumes: tt.volumes,
			}, nil)

			for _, deleteExpected := range tt.expectToDelete {
				mockDockerClient.EXPECT().VolumeRemove(gomock.Any(), deleteExpected, gomock.Any()).Return(nil)
			}

			err := a.Delete(arg, false, false)
			if err != nil {
				t.Errorf("Delete() unexpected error = %v", err)
			}

		})

	}

}
