package kclient

import (
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// constants for deployments
const (
	DeploymentKind       = "Deployment"
	DeploymentAPIVersion = "apps/v1"
)

// CreateDeployment creates a deployment based on the given deployment spec
func (c *Client) CreateDeployment(deploymentSpec appsv1.DeploymentSpec) (*appsv1.Deployment, error) {
	// inherit ObjectMeta from deployment spec so that namespace, labels, owner references etc will be the same
	objectMeta := deploymentSpec.Template.ObjectMeta

	deployment := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       DeploymentKind,
			APIVersion: DeploymentAPIVersion,
		},
		ObjectMeta: objectMeta,
		Spec:       deploymentSpec,
	}

	deploy, err := c.KubeClient.AppsV1().Deployments(c.Namespace).Create(&deployment)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to create Deployment %s", objectMeta.Name)
	}
	return deploy, nil
}
