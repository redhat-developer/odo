package testingutil

import corev1 "k8s.io/api/core/v1"

// CreateFakeContainer creates a container with the given containerName
func CreateFakeContainer(containerName string) corev1.Container {
	return corev1.Container{
		Name: containerName,
	}
}

// CreateFakeContainerWithVolumeMounts creates a container with the given containerName and volumeMounts
func CreateFakeContainerWithVolumeMounts(containerName string, volumeMounts []corev1.VolumeMount) corev1.Container {
	container := CreateFakeContainer(containerName)
	container.VolumeMounts = volumeMounts
	return container
}
