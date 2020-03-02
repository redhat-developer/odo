package kclient

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// CreateService generates and creates the service
// commonObjectMeta is the ObjectMeta for the service
func (c *Client) CreateService(commonObjectMeta metav1.ObjectMeta, containerPorts []corev1.ContainerPort) (*corev1.Service, error) {
	// generate Service Spec
	var svcPorts []corev1.ServicePort
	for _, containerPort := range containerPorts {
		svcPort := corev1.ServicePort{

			Name:       containerPort.Name,
			Port:       containerPort.ContainerPort,
			TargetPort: intstr.FromInt(int(containerPort.ContainerPort)),
		}
		svcPorts = append(svcPorts, svcPort)
	}
	svc := corev1.Service{
		ObjectMeta: commonObjectMeta,
		Spec: corev1.ServiceSpec{
			Ports: svcPorts,
			Selector: map[string]string{
				"deploymentconfig": commonObjectMeta.Name,
			},
		},
	}

	service, err := c.KubeClient.CoreV1().Services(c.Namespace).Create(&svc)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create Service for %s", commonObjectMeta.Name)
	}
	return service, err
}

// UpdateService updates a service based on the given servcie spec
func (c *Client) UpdateService(commonObjectMeta metav1.ObjectMeta, containerPorts []corev1.ContainerPort) (*corev1.Service, error) {
	// generate Service Spec
	var svcPorts []corev1.ServicePort
	for _, containerPort := range containerPorts {
		svcPort := corev1.ServicePort{

			Name:       containerPort.Name,
			Port:       containerPort.ContainerPort,
			TargetPort: intstr.FromInt(int(containerPort.ContainerPort)),
		}
		svcPorts = append(svcPorts, svcPort)
	}
	svc := corev1.Service{
		ObjectMeta: commonObjectMeta,
		Spec: corev1.ServiceSpec{
			Ports: svcPorts,
			Selector: map[string]string{
				"deploymentconfig": commonObjectMeta.Name,
			},
		},
	}

	service, err := c.KubeClient.CoreV1().Services(c.Namespace).Update(&svc)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to update Service for %s", commonObjectMeta.Name)
	}
	return service, err
}
