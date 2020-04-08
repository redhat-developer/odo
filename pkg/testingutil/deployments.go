package testingutil

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// createFakeDeployment creates a fake deployment with the given pod name and labels
func CreateFakeDeployment(podName string) *appsv1.Deployment {
	fakeUID := types.UID("12345")

	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
			UID:  fakeUID,
		},
	}
	return &deployment
}
