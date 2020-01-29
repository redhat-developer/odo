package kclient

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/resource"
)

func TestGenerateContainer(t *testing.T) {

	tests := []struct {
		name         string
		image        string
		isPrivileged bool
		command      []string
		args         []string
		envVars      []corev1.EnvVar
	}{
		{
			name:         "",
			image:        "",
			isPrivileged: false,
			command:      []string{},
			args:         []string{},
			envVars:      []corev1.EnvVar{},
		},
		{
			name:         "container1",
			image:        "quay.io/eclipse/che-java8-maven:nightly",
			isPrivileged: true,
			command:      []string{"tail"},
			args:         []string{"-f", "/dev/null"},
			envVars: []corev1.EnvVar{
				{
					Name:  "test",
					Value: "123",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			container := GenerateContainer(tt.name, tt.image, tt.isPrivileged, tt.command, tt.args, tt.envVars)

			if container.Name != tt.name {
				t.Errorf("expected %s, actual %s", tt.name, container.Name)
			}

			if container.Image != tt.image {
				t.Errorf("expected %s, actual %s", tt.image, container.Image)
			}

			if tt.isPrivileged {
				if *container.SecurityContext.Privileged != tt.isPrivileged {
					t.Errorf("expected %t, actual %t", tt.isPrivileged, *container.SecurityContext.Privileged)
				}
			} else if tt.isPrivileged == false && container.SecurityContext != nil {
				t.Errorf("expected security context to be nil but it was defined")
			}

			if len(container.Command) != len(tt.command) {
				t.Errorf("expected %d, actual %d", len(tt.command), len(container.Command))
			} else {
				for i := range container.Command {
					if container.Command[i] != tt.command[i] {
						t.Errorf("expected %s, actual %s", tt.command[i], container.Command[i])
					}
				}
			}

			if len(container.Args) != len(tt.args) {
				t.Errorf("expected %d, actual %d", len(tt.args), len(container.Args))
			} else {
				for i := range container.Args {
					if container.Args[i] != tt.args[i] {
						t.Errorf("expected %s, actual %s", tt.args[i], container.Args[i])
					}
				}
			}

			if len(container.Env) != len(tt.envVars) {
				t.Errorf("expected %d, actual %d", len(tt.envVars), len(container.Env))
			} else {
				for i := range container.Env {
					if container.Env[i].Name != tt.envVars[i].Name {
						t.Errorf("expected name %s, actual name %s", tt.envVars[i].Name, container.Env[i].Name)
					}
					if container.Env[i].Value != tt.envVars[i].Value {
						t.Errorf("expected value %s, actual value %s", tt.envVars[i].Value, container.Env[i].Value)
					}
				}
			}

		})
	}
}

func TestGeneratePodSpec(t *testing.T) {

	container := &corev1.Container{
		Name:            "container1",
		Image:           "image1",
		ImagePullPolicy: corev1.PullAlways,

		Command: []string{"tail"},
		Args:    []string{"-f", "/dev/null"},
		Env:     []corev1.EnvVar{},
	}

	tests := []struct {
		podName        string
		namespace      string
		serviceAccount string
		labels         map[string]string
	}{
		{
			podName:        "podSpecTest",
			namespace:      "default",
			serviceAccount: "default",
			labels: map[string]string{
				"app":       "app",
				"component": "frontend",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.podName, func(t *testing.T) {

			podSpec := GeneratePodTemplateSpec(tt.podName, tt.namespace, tt.serviceAccount, tt.labels, []corev1.Container{*container})

			if podSpec.Name != tt.podName {
				t.Errorf("expected %s, actual %s", tt.podName, podSpec.Name)
			}

			if podSpec.Namespace != tt.namespace {
				t.Errorf("expected %s, actual %s", tt.namespace, podSpec.Namespace)
			}

			if len(podSpec.Labels) != len(tt.labels) {
				t.Errorf("expected %d, actual %d", len(tt.labels), len(podSpec.Labels))
			} else {
				for i := range podSpec.Labels {
					if podSpec.Labels[i] != tt.labels[i] {
						t.Errorf("expected %s, actual %s", tt.labels[i], podSpec.Labels[i])
					}
				}
			}

		})
	}
}

func TestGeneratePVCSpec(t *testing.T) {

	tests := []struct {
		name    string
		size    string
		wantErr bool
	}{
		{
			name:    "1",
			size:    "1Gi",
			wantErr: false,
		},
		{
			name:    "2",
			size:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			quantity, err := resource.ParseQuantity(tt.size)
			// Checks for unexpected error cases
			if !tt.wantErr == (err != nil) {
				t.Errorf("resource.ParseQuantity unexpected error %v, wantErr %v", err, tt.wantErr)
			}

			pvcSpec := GeneratePVCSpec(quantity)
			if pvcSpec.AccessModes[0] != corev1.ReadWriteMany {
				t.Errorf("AccessMode Error: expected %s, actual %s", corev1.ReadWriteMany, pvcSpec.AccessModes[0])
			}

			pvcSpecQuantity := pvcSpec.Resources.Requests["storage"]
			if pvcSpecQuantity.String() != quantity.String() {
				t.Errorf("pvcSpec.Resources.Requests Error: expected %v, actual %v", pvcSpecQuantity.String(), quantity.String())
			}
		})
	}
}
