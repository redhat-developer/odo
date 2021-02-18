package testingutil

import corev1 "k8s.io/api/core/v1"

// CreateFakeContainer creates a container with the given containerName
func CreateFakeContainer(containerName string) corev1.Container {
	return corev1.Container{
		Name: containerName,
	}
}
