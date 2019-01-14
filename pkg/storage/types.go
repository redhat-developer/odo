package storage

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Storage
type Storage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              StorageSpec `json:"spec,omitempty"`
}

// StorageSpec
type StorageSpec struct {
	Size string `json:"size,omitempty"`
	// if path is empty, it indicates that the storage is not mounted in any component
	Path string `json:"path,omitempty"`
}

// AppList is a list of applications
type StorageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Storage `json:"items"`
}
