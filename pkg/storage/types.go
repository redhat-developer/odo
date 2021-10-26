package storage

import (
	"github.com/openshift/odo/v2/pkg/machineoutput"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const StorageKind = "Storage"

// Storage holds the information about storage attached to the component
type Storage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              StorageSpec   `json:"spec,omitempty"`
	Status            StorageStatus `json:"status,omitempty"`
}

// StorageState
type StorageStatus string

const (
	// StateTypePushed means that Storage is present both locally and on cluster
	StateTypePushed StorageStatus = "Pushed"
	// StateTypeNotPushed means that Storage is only in local config, but not on the cluster
	StateTypeNotPushed = "Not Pushed"
	// StateTypeLocallyDeleted means that Storage was deleted from the local config, but it is still present on the cluster
	StateTypeLocallyDeleted = "Locally Deleted"
)

// StorageSpec indicates size and path of storage
type StorageSpec struct {
	Size string `json:"size,omitempty"`
	// if path is empty, it indicates that the storage is not mounted in any component
	Path string `json:"path,omitempty"`

	ContainerName string `json:"containerName,omitempty"`
}

// StorageList is a list of storages
type StorageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Storage `json:"items"`
}

// NewStorageList returns an instance of a list containning the `items` storages
func NewStorageList(items []Storage) StorageList {
	return StorageList{
		TypeMeta: metav1.TypeMeta{
			Kind:       machineoutput.ListKind,
			APIVersion: machineoutput.APIVersion,
		},
		ListMeta: metav1.ListMeta{},
		Items:    items,
	}
}

// NewStorage returns an instance of Storage
// storagePath indicates the path to which the storage is mounted to, "" if not mounted
func NewStorage(storageName, storageSize, storagePath string) Storage {
	return Storage{
		TypeMeta: metav1.TypeMeta{
			Kind:       StorageKind,
			APIVersion: machineoutput.APIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{Name: storageName},
		Spec: StorageSpec{
			Size: storageSize,
			Path: storagePath,
		},
	}
}

// NewStorageWithContainer returns an instance of Storage with container specified
// storagePath indicates the path to which the storage is mounted to, "" if not mounted
func NewStorageWithContainer(storageName, storageSize, storagePath string, container string) Storage {
	storage := NewStorage(storageName, storageSize, storagePath)
	storage.Spec.ContainerName = container
	return storage
}
