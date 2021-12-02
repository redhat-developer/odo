package storage

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/redhat-developer/odo/pkg/localConfigProvider"
	storageLabels "github.com/redhat-developer/odo/pkg/storage/labels"
)

func getStorageLabels(storageName, componentName, applicationName string) map[string]string {
	return storageLabels.GetLabels(storageName, componentName, applicationName, true)
}

func TestPush(t *testing.T) {
	componentName := "nodejs"

	localStorage0 := localConfigProvider.LocalStorage{
		Name:      "storage-0",
		Size:      "1Gi",
		Path:      "/data",
		Container: "runtime-0",
	}
	localStorage1 := localConfigProvider.LocalStorage{
		Name:      "storage-1",
		Size:      "5Gi",
		Path:      "/path",
		Container: "runtime-1",
	}

	clusterStorage0 := NewStorageWithContainer("storage-0", "1Gi", "/data", "runtime-0")
	clusterStorage1 := NewStorageWithContainer("storage-1", "5Gi", "/path", "runtime-1")

	tests := []struct {
		name                string
		returnedFromLocal   []localConfigProvider.LocalStorage
		returnedFromCluster StorageList
		createdItems        []localConfigProvider.LocalStorage
		deletedItems        []string
		wantErr             bool
	}{
		{
			name:                "case 1: no storage in both local and cluster",
			returnedFromLocal:   []localConfigProvider.LocalStorage{},
			returnedFromCluster: StorageList{},
		},
		{
			name:                "case 2: two storage in local and no on cluster",
			returnedFromLocal:   []localConfigProvider.LocalStorage{localStorage0, localStorage1},
			returnedFromCluster: StorageList{},
			createdItems: []localConfigProvider.LocalStorage{
				{
					Name:      "storage-0",
					Size:      "1Gi",
					Path:      "/data",
					Container: "runtime-0",
				},
				{
					Name:      "storage-1",
					Size:      "5Gi",
					Path:      "/path",
					Container: "runtime-1",
				},
			},
		},
		{
			name:              "case 3: 0 storage in local and two on cluster",
			returnedFromLocal: []localConfigProvider.LocalStorage{},
			returnedFromCluster: StorageList{
				Items: []Storage{clusterStorage0, clusterStorage1},
			},
			createdItems: []localConfigProvider.LocalStorage{},
			deletedItems: []string{"storage-0", "storage-1"},
		},
		{
			name:              "case 4: same two storage in local and cluster",
			returnedFromLocal: []localConfigProvider.LocalStorage{localStorage0, localStorage1},
			returnedFromCluster: StorageList{
				Items: []Storage{clusterStorage0, clusterStorage1},
			},
			createdItems: []localConfigProvider.LocalStorage{},
			deletedItems: []string{},
		},
		{
			name: "case 5: two storage in both local and cluster but two of them are different and the other two are same",
			returnedFromLocal: []localConfigProvider.LocalStorage{localStorage0,
				{
					Name:      "storage-1-1",
					Size:      "5Gi",
					Path:      "/path",
					Container: "runtime-1",
				},
			},
			returnedFromCluster: StorageList{
				Items: []Storage{
					clusterStorage0,
					clusterStorage1,
				},
			},
			createdItems: []localConfigProvider.LocalStorage{
				{
					Name:      "storage-1-1",
					Size:      "5Gi",
					Path:      "/path",
					Container: "runtime-1",
				},
			},
			deletedItems: []string{clusterStorage1.Name},
		},
		{
			name: "case 6: spec mismatch",
			returnedFromLocal: []localConfigProvider.LocalStorage{
				{
					Name:      "storage-1",
					Size:      "3Gi",
					Path:      "/path",
					Container: "runtime-1",
				},
			},
			returnedFromCluster: StorageList{
				Items: []Storage{
					clusterStorage1,
				},
			},
			createdItems: []localConfigProvider.LocalStorage{},
			deletedItems: []string{},
			wantErr:      true,
		},
		{
			name: "case 7: only one PVC created for two storage with same name but on different containers",
			returnedFromLocal: []localConfigProvider.LocalStorage{
				{
					Name:      "storage-0",
					Size:      "1Gi",
					Path:      "/data",
					Container: "runtime-0",
				},
				{
					Name:      "storage-0",
					Size:      "1Gi",
					Path:      "/path",
					Container: "runtime-1",
				},
			},
			returnedFromCluster: StorageList{},
			createdItems: []localConfigProvider.LocalStorage{
				{
					Name:      "storage-0",
					Size:      "1Gi",
					Path:      "/path",
					Container: "runtime-1",
				},
			},
		},
		{
			name: "case 8: only path spec mismatch",
			returnedFromLocal: []localConfigProvider.LocalStorage{
				{
					Name:      "storage-1",
					Size:      "5Gi",
					Path:      "/data",
					Container: "runtime-1",
				},
			},
			returnedFromCluster: StorageList{
				Items: []Storage{
					clusterStorage1,
				},
			},
		},
		{
			name:              "case 9: only one PVC deleted for two storage with same name but on different containers",
			returnedFromLocal: []localConfigProvider.LocalStorage{},
			returnedFromCluster: StorageList{
				Items: []Storage{
					NewStorageWithContainer("storage-0", "1Gi", "/data", "runtime-0"),
					NewStorageWithContainer("storage-0", "1Gi", "/data", "runtime-1"),
				},
			},
			deletedItems: []string{"storage-0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			fakeStorageClient := NewMockClient(ctrl)
			fakeLocalConfig := localConfigProvider.NewMockLocalConfigProvider(ctrl)

			fakeLocalConfig.EXPECT().GetName().Return(componentName).AnyTimes()

			fakeStorageClient.EXPECT().ListFromCluster().Return(tt.returnedFromCluster, nil).AnyTimes()
			fakeLocalConfig.EXPECT().ListStorage().Return(tt.returnedFromLocal, nil).AnyTimes()

			convert := ConvertListLocalToMachine(tt.createdItems)
			for i := range convert.Items {
				fakeStorageClient.EXPECT().Create(convert.Items[i]).Return(nil).Times(1)
			}

			for i := range tt.deletedItems {
				fakeStorageClient.EXPECT().Delete(tt.deletedItems[i]).Return(nil).Times(1)
			}

			if err := Push(fakeStorageClient, fakeLocalConfig); (err != nil) != tt.wantErr {
				t.Errorf("Push() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
