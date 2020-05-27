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
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterTriggerBindingOp is an operation which modifies the ClusterTriggerBinding.
type ClusterTriggerBindingOp func(*v1alpha1.ClusterTriggerBinding)

// ClusterTriggerBinding creates a ClusterTriggerBinding with default values.
// Any number of ClusterTriggerBinding modifiers can be passed.
func ClusterTriggerBinding(name string, ops ...ClusterTriggerBindingOp) *v1alpha1.ClusterTriggerBinding {
	t := &v1alpha1.ClusterTriggerBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	for _, op := range ops {
		op(t)
	}
	return t
}

// ClusterTriggerBindingSpec sets the specified spec of the ClusterTriggerBinding.
// Any number of ClusterTriggerBindingSpecOp modifiers can be passed.
func ClusterTriggerBindingSpec(ops ...TriggerBindingSpecOp) ClusterTriggerBindingOp {
	return func(t *v1alpha1.ClusterTriggerBinding) {
		for _, op := range ops {
			op(&t.Spec)
		}
	}
}

// ClusterTriggerBindingSpec sets the specified spec of the ClusterTriggerBinding.
// Any number of ClusterTriggerBindingSpecOp modifiers can be passed.
func ClusterTriggerBindingMeta(ops ...MetaOp) ClusterTriggerBindingOp {
	return func(t *v1alpha1.ClusterTriggerBinding) {
		for _, op := range ops {
			switch o := op.(type) {
			case ObjectMetaOp:
				o(&t.ObjectMeta)
			case TypeMetaOp:
				o(&t.TypeMeta)
			}
		}
	}
}
