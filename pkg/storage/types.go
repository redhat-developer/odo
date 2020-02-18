package storage

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Storage
type Storage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              StorageSpec   `json:"spec,omitempty"`
	Status            StorageStatus `json:"status,omitempty"`
	State             StorageState  `json:"state,omitempty"`
}

// StorageState
type StorageState string

const (
	// StateTypePushed means that Storage is present both locally and on cluster
	StateTypePushed StorageState = "Pushed"
	// StateTypeNotPushed means that Storage is only in local config, but not on the cluster
	StateTypeNotPushed = "Not Pushed"
	// StateTypeLocallyDeleted means that Storage was deleted from the local config, but it is still present on the cluster
	StateTypeLocallyDeleted = "Locally Deleted"
)

// StorageSpec indicates size and path of storage
type StorageSpec struct {
	Size string `json:"size,omitempty"`
}

// StorageList is a list of storages
type StorageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Storage `json:"items"`
}

// StorageStatus is status of storage
type StorageStatus struct {
	// if path is empty, it indicates that the storage is not mounted in any component
	Path string `json:"path,omitempty"`
}
