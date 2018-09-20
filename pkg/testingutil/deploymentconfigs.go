package testingutil

import (
	"fmt"

	v1 "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getContainerPort(containerPort int32, containerProtocol corev1.Protocol) (container corev1.ContainerPort) {
	return corev1.ContainerPort{
		Name:          fmt.Sprintf("%v/%v", containerPort, containerProtocol),
		ContainerPort: containerPort,
		Protocol:      containerProtocol,
	}
}

func getContainer(componentName string, applicationName string, ports []corev1.ContainerPort) corev1.Container {
	return corev1.Container{
		Name:  fmt.Sprintf("%v-%v", componentName, applicationName),
		Image: fmt.Sprintf("%v-%v", componentName, applicationName),
		Ports: ports,
	}
}

func getDeploymentConfig(namespace string, componentName string, componentType string, applicationName string, containers []corev1.Container) v1.DeploymentConfig {
	return v1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DeploymentConfig",
			APIVersion: "apps.openshift.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%v-%v", componentName, applicationName),
			Namespace: namespace,
			Labels: map[string]string{
				"app":                              "app",
				"app.kubernetes.io/component-name": componentName,
				"app.kubernetes.io/component-type": componentType,
				"app.kubernetes.io/name":           applicationName,
			},
			Annotations: map[string]string{ // Convert into separate function when other source types required in tests
				"app.kubernetes.io/component-source-type": "git",
				"app.kubernetes.io/url":                   fmt.Sprintf("https://github.com/%s/%s.git", componentName, applicationName),
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
					Containers: containers,
				},
			},
		},
	}
}

func FakeDeploymentConfigs() *v1.DeploymentConfigList {

	var componentName string
	var applicationName string
	var componentType string

	// DC1 with multiple containers each with multiple ports
	componentType = "python"
	componentName = "python"
	applicationName = "app"
	c1 := getContainer(componentName, applicationName, []corev1.ContainerPort{
		{
			Name:          fmt.Sprintf("%v-%v-p1", componentName, applicationName),
			ContainerPort: 8080,
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          fmt.Sprintf("%v-%v-p2", componentName, applicationName),
			ContainerPort: 9090,
			Protocol:      corev1.ProtocolUDP,
		},
	})
	c2 := getContainer(componentName, applicationName, []corev1.ContainerPort{
		{
			Name:          fmt.Sprintf("%v-%v-p1", componentName, applicationName),
			ContainerPort: 10080,
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          fmt.Sprintf("%v-%v-p2", componentName, applicationName),
			ContainerPort: 10090,
			Protocol:      corev1.ProtocolUDP,
		},
	})
	dc1 := getDeploymentConfig("myproject", componentName, componentType, applicationName, []corev1.Container{c1, c2})

	// DC2 with single container and single port
	componentType = "nodejs"
	componentName = "nodejs"
	applicationName = "app"
	c3 := getContainer(componentName, applicationName, []corev1.ContainerPort{
		{
			Name:          fmt.Sprintf("%v-%v-p1", componentName, applicationName),
			ContainerPort: 8080,
			Protocol:      corev1.ProtocolTCP,
		},
	})
	dc2 := getDeploymentConfig("myproject", componentName, componentType, applicationName, []corev1.Container{c3})

	// DC3 with single container and multiple ports
	componentType = "wildfly"
	componentName = "wildfly"
	applicationName = "app"
	c4 := getContainer(componentName, applicationName, []corev1.ContainerPort{
		{
			Name:          fmt.Sprintf("%v-%v-p1", componentName, applicationName),
			ContainerPort: 8080,
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          fmt.Sprintf("%v-%v-p1", componentName, applicationName),
			ContainerPort: 8090,
			Protocol:      corev1.ProtocolTCP,
		},
	})
	dc3 := getDeploymentConfig("myproject", componentName, componentType, applicationName, []corev1.Container{c4})

	return &v1.DeploymentConfigList{
		Items: []v1.DeploymentConfig{
			dc1,
			dc2,
			dc3,
		},
	}
}
