package component

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
