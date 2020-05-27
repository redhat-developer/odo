/*
Copyright 2019 The Tekton Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
