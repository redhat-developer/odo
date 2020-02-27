package common

import (
	corev1 "k8s.io/api/core/v1"
)

// ComponentAdapter defines the component functions that platform-specific adapters must implement
type ComponentAdapter interface {
	Initialize() (*corev1.PodTemplateSpec, error)
	Start(*corev1.PodTemplateSpec) error
}

// StorageAdapter defines the storage functions that platform-specific adapters must implement
type StorageAdapter interface {
	Start(*corev1.PodTemplateSpec) error
}
