package component

import (
	"testing"

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

			// Verify that a comopnent with the specified name exists
			componentExists := componentAdapter.DoesComponentExist(tt.getComponentName)
			if componentExists != tt.want {
				t.Errorf("expected %v, actual %v", tt.want, componentExists)
			}

		})
	}

}
