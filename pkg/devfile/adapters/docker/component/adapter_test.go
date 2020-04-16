package component

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/mount"
	volumeTypes "github.com/docker/docker/api/types/volume"
	"github.com/golang/mock/gomock"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/lclient"
	"github.com/openshift/odo/pkg/testingutil"
)

func TestPush(t *testing.T) {

	testComponentName := "test"
	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name          string
		componentType versionsCommon.DevfileComponentType
		client        *lclient.Client
		wantErr       bool
	}{
		{
			name:          "Case 1: Invalid devfile",
			componentType: "",
			client:        fakeClient,
			wantErr:       true,
		},
		{
			name:          "Case 2: Valid devfile",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			client:        fakeClient,
			wantErr:       false,
		},
		{
			name:          "Case 3: Valid devfile, docker client error",
			componentType: versionsCommon.DevfileComponentTypeDockerimage,
			client:        fakeErrorClient,
			wantErr:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: tt.componentType,
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			componentAdapter := New(adapterCtx, *tt.client)
			// ToDo: Add more meaningful unit tests once Push actually does something with its parameters
			err := componentAdapter.Push(adaptersCommon.PushParameters{})

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter create unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestDoesComponentExist(t *testing.T) {
	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	tests := []struct {
		name             string
		client           *lclient.Client
		componentType    versionsCommon.DevfileComponentType
		componentName    string
		getComponentName string
		want             bool
	}{
		{
			name:             "Case 1: Valid component name",
			client:           fakeClient,
			componentType:    versionsCommon.DevfileComponentTypeDockerimage,
			componentName:    "golang",
			getComponentName: "golang",
			want:             true,
		},
		{
			name:             "Case 2: Non-existent component name",
			componentType:    versionsCommon.DevfileComponentTypeDockerimage,
			client:           fakeClient,
			componentName:    "test-name",
			getComponentName: "fake-component",
			want:             false,
		},
		{
			name:             "Case 3: Docker client error",
			componentType:    versionsCommon.DevfileComponentTypeDockerimage,
			client:           fakeErrorClient,
			componentName:    "test-name",
			getComponentName: "fake-component",
			want:             false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: tt.componentType,
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: tt.componentName,
				Devfile:       devObj,
			}

			componentAdapter := New(adapterCtx, *tt.client)

			// Verify that a component with the specified name exists
			componentExists := componentAdapter.DoesComponentExist(tt.getComponentName)
			if componentExists != tt.want {
				t.Errorf("expected %v, actual %v", tt.want, componentExists)
			}

		})
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
			name: "case 1: component exists and given labels are valid",
			args: args{labels: map[string]string{
				"component": "component",
			}},
			componentName:   "component",
			componentExists: true,
			wantErr:         false,
		},
		{
			name:              "case 2: component exists and given labels are not valid",
			args:              args{labels: nil},
			componentName:     "component",
			componentExists:   true,
			wantErr:           true,
			skipContainerList: true,
		},
		{
			name: "case 3: component doesn't exists",
			args: args{labels: map[string]string{
				"component": "component",
			}},
			componentName:   "component",
			componentExists: false,
			wantErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			containerID := "my-id"
			volumeID := "my-volume-name"

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			devObj := devfileParser.DevfileObj{
				Data: testingutil.TestDevfileData{
					ComponentType: "nodejs",
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: tt.componentName,
				Devfile:       devObj,
			}

			if !tt.componentExists {
				adapterCtx.ComponentName = "doesNotExists"
			}

			fkclient, mockDockerClient := lclient.FakeNewMockClient(ctrl)

			a := Adapter{
				Client:         *fkclient,
				AdapterContext: adapterCtx,
			}

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

			if err := a.Delete(tt.args.labels); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
