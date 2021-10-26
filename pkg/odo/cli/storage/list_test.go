package storage

import (
	"testing"

	"github.com/openshift/odo/v2/pkg/localConfigProvider"
	"github.com/openshift/odo/v2/pkg/storage"
)

func Test_isContainerDisplay(t *testing.T) {
	generateStorage := func(storage storage.Storage, status storage.StorageStatus, containerName string) storage.Storage {
		storage.Status = status
		storage.Spec.ContainerName = containerName
		return storage
	}

	type args struct {
		storageList storage.StorageList
		obj         []localConfigProvider.LocalContainer
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "case 1: storage is mounted on all the containers on the same path",
			args: args{
				storageList: storage.StorageList{
					Items: []storage.Storage{
						generateStorage(storage.NewStorage("pvc-1", "1Gi", "/data"), storage.StateTypePushed, "container-0"),
						generateStorage(storage.NewStorage("pvc-1", "1Gi", "/data"), storage.StateTypePushed, "container-1"),
					},
				},
				obj: []localConfigProvider.LocalContainer{
					{
						Name: "container-0",
					},
					{
						Name: "container-1",
					},
				},
			},
			want: false,
		},
		{
			name: "case 2: storage is mounted on different paths",
			args: args{
				storageList: storage.StorageList{
					Items: []storage.Storage{
						generateStorage(storage.NewStorage("pvc-1", "1Gi", "/data"), storage.StateTypePushed, "container-0"),
						generateStorage(storage.NewStorage("pvc-1", "1Gi", "/path"), storage.StateTypePushed, "container-1"),
					},
				},
				obj: []localConfigProvider.LocalContainer{
					{
						Name: "container-0",
					},
					{
						Name: "container-1",
					},
				},
			},
			want: true,
		},
		{
			name: "case 3: storage is mounted to the same path on all the containers but states are different",
			args: args{
				storageList: storage.StorageList{
					Items: []storage.Storage{
						generateStorage(storage.NewStorage("pvc-1", "1Gi", "/data"), storage.StateTypePushed, "container-0"),
						generateStorage(storage.NewStorage("pvc-1", "1Gi", "/data"), storage.StateTypeNotPushed, "container-1"),
					},
				},
				obj: []localConfigProvider.LocalContainer{
					{
						Name: "container-0",
					},
					{
						Name: "container-1",
					},
				},
			},
			want: true,
		},
		{
			name: "case 4: storage is not mounted on all the containers",
			args: args{
				storageList: storage.StorageList{
					Items: []storage.Storage{
						generateStorage(storage.NewStorage("pvc-1", "1Gi", "/data"), storage.StateTypePushed, "container-0"),
					},
				},
				obj: []localConfigProvider.LocalContainer{
					{
						Name: "container-0",
					},
					{
						Name: "container-1",
					},
				},
			},
			want: true,
		},
		{
			name: "case 5: storage is mounted on a container deleted locally from the devfile",
			args: args{
				storageList: storage.StorageList{
					Items: []storage.Storage{
						generateStorage(storage.NewStorage("pvc-1", "1Gi", "/data"), storage.StateTypePushed, "container-0"),
						generateStorage(storage.NewStorage("pvc-1", "1Gi", "/data"), storage.StateTypePushed, "container-1"),
					},
				},
				obj: []localConfigProvider.LocalContainer{
					{
						Name: "container-0",
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isContainerDisplay(tt.args.storageList, tt.args.obj); got != tt.want {
				t.Errorf("isContainerDisplay() = %v, want %v", got, tt.want)
			}
		})
	}
}
