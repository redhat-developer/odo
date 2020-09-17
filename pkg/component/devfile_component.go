package component

import (
	applabels "github.com/openshift/odo/pkg/application/labels"
	componentlabels "github.com/openshift/odo/pkg/component/labels"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewDevfileComponent(componentName string) DevfileComponent {
	return DevfileComponent{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DevfileComponent",
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: componentName,
		},
	}
}

func DevfileComponentsFromDeployments(deployList *appsv1.DeploymentList) []DevfileComponent {
	compList := []DevfileComponent{}
	for _, deployment := range deployList.Items {
		app := deployment.Labels[applabels.ApplicationLabel]
		cmpType := deployment.Labels[componentlabels.ComponentTypeLabel]

		comp := NewDevfileComponent(deployment.Name)
		comp.Status.State = StateTypePushed
		comp.Spec.Namespace = deployment.Namespace
		comp.Spec.Application = app
		comp.Spec.ComponentType = cmpType
		compList = append(compList, comp)
	}
	return compList
}
