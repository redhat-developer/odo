package component

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type OdoComponent interface {
	GetName() string
	GetApplication() string
}

type KubernetesComponent interface {
	OdoComponent
	GetOwnerReference() metav1.OwnerReference
}