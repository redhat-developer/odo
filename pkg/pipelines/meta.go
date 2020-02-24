package pipelines

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apitypes "k8s.io/apimachinery/pkg/types"
)

func typeMeta(kind, apiVersion string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       kind,
		APIVersion: apiVersion,
	}
}

func objectMeta(n apitypes.NamespacedName) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      n.Name,
		Namespace: n.Namespace,
	}
}

func namespacedName(ns, name string) apitypes.NamespacedName {
	return apitypes.NamespacedName{
		Namespace: ns,
		Name:      name,
	}
}
