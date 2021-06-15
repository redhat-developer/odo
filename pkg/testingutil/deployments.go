package testingutil

import (
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// CreateFakeDeployment creates a fake deployment with the given pod name and labels
func CreateFakeDeployment(podName string) *appsv1.Deployment {
	fakeUID := types.UID("12345")

	deployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: podName,
			UID:  fakeUID,
			Labels: map[string]string{
				applabels.App:                  "app",
				applabels.ApplicationLabel:     "app",
				componentlabels.ComponentLabel: podName,
				applabels.ManagedBy:            "odo",
			},
		},
	}
	return &deployment
}
