package testingutil

import (
	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateFakePod creates a fake pod with the given pod name and component name
func CreateFakePod(componentName, podName string) *corev1.Pod {
	fakePod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   podName,
			Labels: componentlabels.GetLabels(componentName, "app", false),
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
	return fakePod
}

// CreateFakePodWithContainers creates a fake pod with the given pod name, container name and containers
func CreateFakePodWithContainers(componentName, podName string, containers []corev1.Container) *corev1.Pod {
	fakePod := CreateFakePod(componentName, podName)
	fakePod.Spec.Containers = containers
	return fakePod
}
