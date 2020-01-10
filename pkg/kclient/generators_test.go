package kclient

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestGenerateContainerSpec(t *testing.T) {

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

			containerSpec := GenerateContainerSpec(tt.name, tt.image, tt.isPrivileged, tt.command, tt.args, tt.envVars)

			if containerSpec.Name != tt.name {
				t.Errorf("expected %s, actual %s", tt.name, containerSpec.Name)
			}

			if containerSpec.Image != tt.image {
				t.Errorf("expected %s, actual %s", tt.image, containerSpec.Image)
			}

			if tt.isPrivileged {
				if *containerSpec.SecurityContext.Privileged != tt.isPrivileged {
					t.Errorf("expected %t, actual %t", tt.isPrivileged, *containerSpec.SecurityContext.Privileged)
				}
			} else if tt.isPrivileged == false && containerSpec.SecurityContext != nil {
				t.Errorf("expected security context to be nil but it was defined")
			}

			if len(containerSpec.Command) != len(tt.command) {
				t.Errorf("expected %d, actual %d", len(tt.command), len(containerSpec.Command))
			} else {
				for i := range containerSpec.Command {
					if containerSpec.Command[i] != tt.command[i] {
						t.Errorf("expected %s, actual %s", tt.command[i], containerSpec.Command[i])
					}
				}
			}

			if len(containerSpec.Args) != len(tt.args) {
				t.Errorf("expected %d, actual %d", len(tt.args), len(containerSpec.Args))
			} else {
				for i := range containerSpec.Args {
					if containerSpec.Args[i] != tt.args[i] {
						t.Errorf("expected %s, actual %s", tt.args[i], containerSpec.Args[i])
					}
				}
			}

			if len(containerSpec.Env) != len(tt.envVars) {
				t.Errorf("expected %d, actual %d", len(tt.envVars), len(containerSpec.Env))
			} else {
				for i := range containerSpec.Env {
					if containerSpec.Env[i].Name != tt.envVars[i].Name {
						t.Errorf("expected name %s, actual name %s", tt.envVars[i].Name, containerSpec.Env[i].Name)
					}
					if containerSpec.Env[i].Value != tt.envVars[i].Value {
						t.Errorf("expected value %s, actual value %s", tt.envVars[i].Value, containerSpec.Env[i].Value)
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

			podSpec := GeneratePodSpec(tt.podName, tt.namespace, tt.serviceAccount, tt.labels, []corev1.Container{*container})

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
