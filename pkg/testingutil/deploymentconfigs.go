package testingutil

import (
	"fmt"

	v1 "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// getDeploymentConfig("myproject", "python", "python", "app", 8080)
func getDeploymentConfig(namespace string, componentName string, componentType string, applicationName string, containerPort int32) v1.DeploymentConfig {
	return v1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%v-%v", componentName, applicationName),
			Namespace: namespace,
			Labels: map[string]string{
				"app": "app",
				"app.kubernetes.io/component-name": componentName,
				"app.kubernetes.io/component-type": componentType,
				"app.kubernetes.io/name":           applicationName,
			},
			Annotations: map[string]string{
				"app.kubernetes.io/component-source-type": "git",
				"app.kubernetes.io/url":                   "https://github.com/openshift/django-ex.git",
			},
		},
		Spec: v1.DeploymentConfigSpec{
			Replicas: 1,
			Selector: map[string]string{
				"deploymentconfig": fmt.Sprintf("%v-%v", componentName, applicationName),
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"deploymentconfig": fmt.Sprintf("%v-%v", componentName, applicationName),
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  fmt.Sprintf("%v-%v", componentName, applicationName),
							Image: fmt.Sprintf("%v-%v", componentName, applicationName),
							Ports: []corev1.ContainerPort{
								{
									Name:          fmt.Sprintf("%v-%v", componentName, applicationName),
									ContainerPort: containerPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
						},
					},
				},
			},
		},
	}
}

func FakeDeploymentConfigs(namespace string, componentName string, componentType string, applicationName string, containerPort int32) *v1.DeploymentConfigList {
	dc := getDeploymentConfig("myproject", "python", "python", "app", 8080)
	return &v1.DeploymentConfigList{
		Items: []v1.DeploymentConfig{
			dc,
		},
	}
}
