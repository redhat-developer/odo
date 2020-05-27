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

// TriggerBindingOp is an operation which modifies the TriggerBinding.
type TriggerBindingOp func(*v1alpha1.TriggerBinding)

// TriggerBindingSpecOp is an operation which modifies the TriggerBindingSpec.
type TriggerBindingSpecOp func(*v1alpha1.TriggerBindingSpec)

// TriggerBinding creates a TriggerBinding with default values.
// Any number of TriggerBinding modifiers can be passed.
func TriggerBinding(name, namespace string, ops ...TriggerBindingOp) *v1alpha1.TriggerBinding {
	t := &v1alpha1.TriggerBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	for _, op := range ops {
		op(t)
	}
	return t
}

// TriggerBindingMeta sets the Meta structs of the TriggerBinding.
// Any number of MetaOp modifiers can be passed.
func TriggerBindingMeta(ops ...MetaOp) TriggerBindingOp {
	return func(t *v1alpha1.TriggerBinding) {
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

// TriggerBindingSpec sets the specified spec of the TriggerBinding.
// Any number of TriggerBindingSpecOp modifiers can be passed.
func TriggerBindingSpec(ops ...TriggerBindingSpecOp) TriggerBindingOp {
	return func(t *v1alpha1.TriggerBinding) {
		for _, op := range ops {
			op(&t.Spec)
		}
	}
}

// TriggerBindingParam adds a param to the TriggerBindingSpec.
func TriggerBindingParam(name, value string) TriggerBindingSpecOp {
	return func(spec *v1alpha1.TriggerBindingSpec) {
		spec.Params = append(spec.Params, v1alpha1.Param{
			Name:  name,
			Value: value,
		})
	}
}
