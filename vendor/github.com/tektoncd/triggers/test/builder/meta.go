package builder

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MetaOp is an interface that is used in other builders.
// Other builders should have a Meta function that accepts ...MetaOp where ObjectMetaOp/TypeMetaOp are the underlying type.
type MetaOp interface{}

// ObjectMetaOp is an operation which modifies the ObjectMeta.
type ObjectMetaOp func(m *metav1.ObjectMeta)

// TypeMetaOp is an operation which modifies the TypeMeta.
type TypeMetaOp func(m *metav1.TypeMeta)

// Label adds a single label to the ObjectMeta.
func Label(key, value string) ObjectMetaOp {
	return func(m *metav1.ObjectMeta) {
		if m.Labels == nil {
			m.Labels = make(map[string]string)
		}
		m.Labels[key] = value
	}
}

// TypeMeta sets the TypeMeta struct with default values.
func TypeMeta(kind, apiVersion string) TypeMetaOp {
	return func(m *metav1.TypeMeta) {
		m.Kind = kind
		m.APIVersion = apiVersion
	}
}
