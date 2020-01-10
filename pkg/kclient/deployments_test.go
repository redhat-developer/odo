package kclient

import (
	"testing"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ktesting "k8s.io/client-go/testing"
)

func TestCreateDeployment(t *testing.T) {

	container := &corev1.Container{
		Name:            "container1",
		Image:           "image1",
		ImagePullPolicy: corev1.PullAlways,

		Command: []string{"tail"},
		Args:    []string{"-f", "/dev/null"},
		Env:     []corev1.EnvVar{},
	}

	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "",
			Namespace: "default",
			Labels: map[string]string{
				"app":       "app",
				"component": "frontend",
			},
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: "default",
			Containers:         []corev1.Container{*container},
		},
	}

	tests := []struct {
		name           string
		deploymentName string
		wantErr        bool
	}{
		{
			name:           "Case: Valid deployment name",
			deploymentName: "pod",
			wantErr:        false,
		},
		{
			name:           "Case: Invalid deployment name",
			deploymentName: "",
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// initialising the fakeclient
			fkclient, fkclientset := FakeNew()
			fkclient.Namespace = "default"

			fkclientset.Kubernetes.PrependReactor("create", "deployments", func(action ktesting.Action) (bool, runtime.Object, error) {
				if tt.deploymentName == "" {
					return true, nil, errors.Errorf("deployment name is empty")
				}
				deployment := appsv1.Deployment{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Deployment",
						APIVersion: "apps/v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: tt.deploymentName,
					},
				}
				return true, &deployment, nil
			})

			pod.ObjectMeta.Name = tt.deploymentName
			createdDeployment, err := fkclient.CreateDeployment(pod)

			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("fkclient.CreateDeployment(pod) unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {

				if len(fkclientset.Kubernetes.Actions()) != 1 {
					t.Errorf("expected 1 action in StartDeployment got: %v", fkclientset.Kubernetes.Actions())
				} else {
					if createdDeployment.Name != tt.deploymentName {
						t.Errorf("deployment name does not match the expected name, expected: %s, got %s", tt.deploymentName, createdDeployment.Name)
					}
				}

			}

		})
	}
}
