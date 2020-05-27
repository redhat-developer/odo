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
	"k8s.io/apimachinery/pkg/runtime"
)

// TriggerTemplateOp is an operation which modifies an TriggerTemplate struct.
type TriggerTemplateOp func(*v1alpha1.TriggerTemplate)

// TriggerTemplateSpecOp is an operation which modifies a TriggerTemplateSpec struct.
type TriggerTemplateSpecOp func(*v1alpha1.TriggerTemplateSpec)

// TriggerTemplate creates a TriggerTemplate with default values.
// Any number of TriggerTemplate modifiers can be passed.
func TriggerTemplate(name, namespace string, ops ...TriggerTemplateOp) *v1alpha1.TriggerTemplate {
	t := &v1alpha1.TriggerTemplate{
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

// TriggerTemplateMeta sets the Meta structs of the TriggerTemplate.
// Any number of MetaOp modifiers can be passed.
func TriggerTemplateMeta(ops ...MetaOp) TriggerTemplateOp {
	return func(t *v1alpha1.TriggerTemplate) {
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

// TriggerTemplateSpec sets the TriggerTemplateSpec.
// Any number of TriggerTemplate modifiers can be passed.
func TriggerTemplateSpec(ops ...TriggerTemplateSpecOp) TriggerTemplateOp {
	return func(t *v1alpha1.TriggerTemplate) {
		spec := &t.Spec
		for _, op := range ops {
			op(spec)
		}
		t.Spec = *spec
	}
}

// TriggerResourceTemplate adds a ResourceTemplate to the TriggerTemplateSpec.
func TriggerResourceTemplate(resourceTemplate runtime.RawExtension) TriggerTemplateSpecOp {
	return func(spec *v1alpha1.TriggerTemplateSpec) {
		spec.ResourceTemplates = append(spec.ResourceTemplates,
			v1alpha1.TriggerResourceTemplate{
				RawExtension: resourceTemplate,
			})
	}
}

// TriggerTemplateParam adds a ParamSpec to the TriggerTemplateSpec.
func TriggerTemplateParam(name, description, defaultValue string) TriggerTemplateSpecOp {
	return func(spec *v1alpha1.TriggerTemplateSpec) {
		spec.Params = append(spec.Params,
			v1alpha1.ParamSpec{
				Name:        name,
				Description: description,
				Default:     &defaultValue,
			})
	}
}
