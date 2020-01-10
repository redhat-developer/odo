package kclient

import (
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateDeployment creates a deployment based on the given pod
func (c *Client) CreateDeployment(pod *corev1.Pod) (*appsv1.Deployment, error) {

	replicas := int32(1)
	deployment := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: pod.ObjectMeta,
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: pod.Labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: pod.ObjectMeta,
				Spec:       pod.Spec,
			},
		},
	}

	deploy, err := c.KubeClient.AppsV1().Deployments(c.Namespace).Create(&deployment)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create Deployment for %s", pod.ObjectMeta.Name)
	}
	return deploy, nil
}
