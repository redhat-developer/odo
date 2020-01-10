package kclient

import (

	// api resource types
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GenerateContainerSpec creates a container spec that can be used when creating a pod
func GenerateContainerSpec(name, image string, isPrivileged bool, command, args []string, envVars []corev1.EnvVar) corev1.Container {
	container := &corev1.Container{
		Name:            name,
		Image:           image,
		ImagePullPolicy: corev1.PullAlways,

		Command: command,
		Args:    args,
		Env:     envVars,
	}

	if isPrivileged {
		container.SecurityContext = &corev1.SecurityContext{
			Privileged: &isPrivileged,
		}
	}

	return *container
}

// GeneratePodSpec creates a pod spec
func GeneratePodSpec(podName, namespace, serviceAccountName string, labels map[string]string, containers []corev1.Container) *corev1.Pod {
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: serviceAccountName,
			Containers:         containers,
		},
	}

	return pod
}
