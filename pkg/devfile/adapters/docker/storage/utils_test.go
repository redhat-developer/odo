package storage

import (
	"strings"
	"testing"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/lclient"
)

func TestCreateComponentStorage(t *testing.T) {
	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	testComponentName := "test"
	volNames := [...]string{"vol1", "vol2"}
	volSize := "5Gi"

	tests := []struct {
		name     string
		storages []common.Storage
		client   *lclient.Client
		wantErr  bool
	}{
		{
			name: "Case 1: Multiple volumes defined, no Docker client error",
			storages: []common.Storage{
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
			name: "Case 1: Multiple volumes defined, Docker client error",
			storages: []common.Storage{
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
			err := CreateComponentStorage(tt.client, tt.storages, testComponentName)
			if !tt.wantErr == (err != nil) {
				t.Errorf("Storage adapter create unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestStorageCreate(t *testing.T) {
	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	testComponentName := "test"
	volNames := [...]string{"vol1", "vol2"}
	volSize := "5Gi"

	tests := []struct {
		name    string
		storage common.Storage
		client  *lclient.Client
		wantErr bool
	}{
		{
			name: "Case 1: Valid volume, no Docker client error",
			storage: common.Storage{
				Name: "vol1",
				Volume: common.DevfileVolume{
					Name: volNames[0],
					Size: volSize,
				},
			},
			client:  fakeClient,
			wantErr: false,
		},
		{
			name: "Case 2: Docker client error",
			storage: common.Storage{
				Name: "vol-name",
				Volume: common.DevfileVolume{
					Name: volNames[0],
					Size: volSize,
				},
			},
			client:  fakeErrorClient,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create one of the test volumes
			_, err := Create(tt.client, tt.storage.Volume.Name, testComponentName, tt.storage.Name)
			if !tt.wantErr == (err != nil) {
				t.Errorf("Docker volume create unexpected error %v, wantErr %v", err, tt.wantErr)
			}

		})
	}

}

func TestProcessVolumes(t *testing.T) {
	fakeClient := lclient.FakeNew()
	fakeErrorClient := lclient.FakeErrorNew()

	testComponentName := "test"
	volumeNames := []string{"vol1", "vol2", "vol3"}
	volumePaths := []string{"/path1", "/path2", "/path3"}
	volumeSizes := []string{"1Gi", "2Gi", "3Gi"}
	tests := []struct {
		name               string
		client             *lclient.Client
		aliasVolumeMapping map[string][]common.DevfileVolume
		wantErr            bool
		wantStorage        []common.Storage
	}{
		{
			name:               "Case 1: No volumes defined",
			aliasVolumeMapping: nil,
			client:             fakeClient,
			wantStorage:        nil,
			wantErr:            false,
		},
		{
			name: "Case 2: One volume defined, one component",
			aliasVolumeMapping: map[string][]common.DevfileVolume{
				"some-component": []common.DevfileVolume{
					{
						Name:          volumeNames[0],
						ContainerPath: volumePaths[0],
						Size:          volumeSizes[0],
					},
				},
			},
			client: fakeClient,
			wantStorage: []common.Storage{
				{
					Volume: common.DevfileVolume{
						Name:          volumeNames[0],
						ContainerPath: volumePaths[0],
						Size:          volumeSizes[0],
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 3: Multiple volumes defined, one component",
			aliasVolumeMapping: map[string][]common.DevfileVolume{
				"some-component": []common.DevfileVolume{
					{
						Name:          volumeNames[0],
						ContainerPath: volumePaths[0],
						Size:          volumeSizes[0],
					},
					{
						Name:          volumeNames[1],
						ContainerPath: volumePaths[1],
						Size:          volumeSizes[1],
					},
					{
						Name:          volumeNames[2],
						ContainerPath: volumePaths[2],
						Size:          volumeSizes[2],
					},
				},
			},
			client: fakeClient,
			wantStorage: []common.Storage{
				{
					Volume: common.DevfileVolume{
						Name:          volumeNames[0],
						ContainerPath: volumePaths[0],
						Size:          volumeSizes[0],
					},
				},
				{
					Volume: common.DevfileVolume{
						Name:          volumeNames[1],
						ContainerPath: volumePaths[1],
						Size:          volumeSizes[1],
					},
				},
				{
					Volume: common.DevfileVolume{
						Name:          volumeNames[2],
						ContainerPath: volumePaths[2],
						Size:          volumeSizes[2],
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 4: Multiple volumes defined, multiple components",
			aliasVolumeMapping: map[string][]common.DevfileVolume{
				"some-component": []common.DevfileVolume{
					{
						Name:          volumeNames[0],
						ContainerPath: volumePaths[0],
						Size:          volumeSizes[0],
					},
					{
						Name:          volumeNames[1],
						ContainerPath: volumePaths[1],
						Size:          volumeSizes[1],
					},
				},
				"second-component": []common.DevfileVolume{
					{
						Name:          volumeNames[0],
						ContainerPath: volumePaths[0],
						Size:          volumeSizes[0],
					},
				},
				"third-component": []common.DevfileVolume{
					{
						Name:          volumeNames[1],
						ContainerPath: volumePaths[1],
						Size:          volumeSizes[1],
					},
					{
						Name:          volumeNames[2],
						ContainerPath: volumePaths[2],
						Size:          volumeSizes[2],
					},
				},
			},
			client: fakeClient,
			wantStorage: []common.Storage{
				{
					Volume: common.DevfileVolume{
						Name:          volumeNames[0],
						ContainerPath: volumePaths[0],
						Size:          volumeSizes[0],
					},
				},
				{
					Volume: common.DevfileVolume{
						Name:          volumeNames[1],
						ContainerPath: volumePaths[1],
						Size:          volumeSizes[1],
					},
				},
				{
					Volume: common.DevfileVolume{
						Name:          volumeNames[2],
						ContainerPath: volumePaths[2],
						Size:          volumeSizes[2],
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Case 5: Docker client error",
			aliasVolumeMapping: map[string][]common.DevfileVolume{
				"some-component": []common.DevfileVolume{
					{
						Name:          volumeNames[0],
						ContainerPath: volumePaths[0],
						Size:          volumeSizes[0],
					},
				},
			},
			client:      fakeErrorClient,
			wantStorage: nil,
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create one of the test volumes
			uniqueStorage, _, err := ProcessVolumes(tt.client, testComponentName, tt.aliasVolumeMapping)
			if !tt.wantErr == (err != nil) {
				t.Errorf("Docker volume create unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			storageLength := len(uniqueStorage)
			wantStorageLength := len(tt.wantStorage)
			if storageLength != wantStorageLength {
				t.Errorf("expected %v, wanted %v", storageLength, wantStorageLength)
			}

			if storageLength > 0 {
				for i := range uniqueStorage {
					var volExists bool
					for j := range tt.wantStorage {
						if uniqueStorage[i].Volume.Name == tt.wantStorage[j].Volume.Name && uniqueStorage[i].Volume.ContainerPath == tt.wantStorage[j].Volume.ContainerPath {
							volExists = true
						}
					}

					if !volExists {
						t.Errorf("expected %v, wanted %v", uniqueStorage[i].Volume, tt.wantStorage[i].Volume)
					}
				}
			}

		})
	}

}

func TestGenerateVolName(t *testing.T) {

	tests := []struct {
		name        string
		volName     string
		cmpName     string
		wantVolName string
		wantErr     bool
	}{
		{
			name:        "Case 1: Valid volume and component name",
			volName:     "myVol",
			cmpName:     "myCmp",
			wantVolName: "myVol-myCmp",
			wantErr:     false,
		},
		{
			name:        "Case 2: Valid volume name, empty component name",
			volName:     "myVol",
			cmpName:     "",
			wantVolName: "myVol-",
			wantErr:     false,
		},
		{
			name:        "Case 3: Long Valid volume and component name",
			volName:     "myVolmyVolmyVolmyVolmyVolmyVolmyVolmyVolmyVol",
			cmpName:     "myCmpmyCmpmyCmpmyCmpmyCmpmyCmpmyCmpmyCmpmyCmp",
			wantVolName: "myVolmyVolmyVolmyVolmyVolmyVolmyVolmyVolmyVol-",
			wantErr:     false,
		},
		{
			name:        "Case 4: Empty volume name",
			volName:     "",
			cmpName:     "myCmp",
			wantVolName: "",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generatedVolName, err := GenerateVolName(tt.volName, tt.cmpName)
			if !tt.wantErr && err != nil {
				t.Errorf("TestGenerateVolName error: unexpected error when generating volume name: %v", err)
			} else if !tt.wantErr && !strings.Contains(generatedVolName, tt.wantVolName) {
				t.Errorf("TestGenerateVolName error: generating volume name does not semi match wanted volume name, wanted: %s got: %s", tt.wantVolName, generatedVolName)
			}
		})
	}

}
