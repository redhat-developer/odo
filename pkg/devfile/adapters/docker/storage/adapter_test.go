package storage

import (
	"testing"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	adaptersCommon "github.com/openshift/odo/pkg/devfile/adapters/common"
	devfileParser "github.com/openshift/odo/pkg/devfile/parser"
	versionsCommon "github.com/openshift/odo/pkg/devfile/parser/data/common"
	"github.com/openshift/odo/pkg/lclient"
	"github.com/openshift/odo/pkg/testingutil"
)

func TestCreate(t *testing.T) {

	testComponentName := "test"
	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	volNames := [...]string{"vol1", "vol2"}
	volSize := "5Gi"

	tests := []struct {
		name    string
		storage []common.Storage
		client  *lclient.Client
		wantErr bool
	}{
		{
			name:    "Case 1: No volumes",
			storage: nil,
			client:  fakeClient,
			wantErr: false,
		},
		{
			name: "Case 2: Multiple volumes",
			storage: []common.Storage{
				{
					Name: "vol1",
					Volume: common.DevfileVolume{
						Name: volNames[0],
						Size: volSize,
					},
				},
				{
					Name: "vol2",
					Volume: common.DevfileVolume{
						Name: volNames[1],
						Size: volSize,
					},
				},
			},
			client:  fakeClient,
			wantErr: false,
		},
		{
			name: "Case 3: Docker client error",
			storage: []common.Storage{
				{
					Name: "vol1",
					Volume: common.DevfileVolume{
						Name: volNames[0],
						Size: volSize,
					},
				},
				{
					Name: "vol2",
					Volume: common.DevfileVolume{
						Name: volNames[1],
						Size: volSize,
					},
				},
			},
			client:  fakeErrorClient,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devObj := devfileParser.DevfileObj{
				Data: &testingutil.TestDevfileData{
					Components: []versionsCommon.DevfileComponent{},
				},
			}

			adapterCtx := adaptersCommon.AdapterContext{
				ComponentName: testComponentName,
				Devfile:       devObj,
			}

			storageAdapter := New(adapterCtx, *tt.client)
			// ToDo: Add more meaningful unit tests once Push actually does something with its parameters
			err := storageAdapter.Create(tt.storage)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("component adapter create unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}
