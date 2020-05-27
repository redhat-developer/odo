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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestTriggerTemplateBuilder(t *testing.T) {
	defaultValue1 := "value1"
	defaultValue2 := "value2"
	tests := []struct {
		name    string
		normal  *v1alpha1.TriggerTemplate
		builder *v1alpha1.TriggerTemplate
	}{
		{
			name:    "Empty",
			normal:  &v1alpha1.TriggerTemplate{},
			builder: TriggerTemplate("", ""),
		},
		{
			name: "Name and Namespace",
			normal: &v1alpha1.TriggerTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
			},
			builder: TriggerTemplate("name", "namespace"),
		},
		{
			name: "One Param",
			normal: &v1alpha1.TriggerTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: v1alpha1.TriggerTemplateSpec{
					Params: []v1alpha1.ParamSpec{
						{
							Name:        "param1",
							Description: "description",
							Default:     &defaultValue1,
						},
					},
				},
			},
			builder: TriggerTemplate("name", "namespace",
				TriggerTemplateSpec(
					TriggerTemplateParam("param1", "description", "value1"),
				),
			),
		},
		{
			name: "Two Param",
			normal: &v1alpha1.TriggerTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: v1alpha1.TriggerTemplateSpec{
					Params: []v1alpha1.ParamSpec{
						{
							Name:        "param1",
							Description: "description",
							Default:     &defaultValue1,
						},
						{
							Name:        "param2",
							Description: "description",
							Default:     &defaultValue2,
						},
					},
				},
			},
			builder: TriggerTemplate("name", "namespace",
				TriggerTemplateSpec(
					TriggerTemplateParam("param1", "description", "value1"),
					TriggerTemplateParam("param2", "description", "value2"),
				),
			),
		},
		{
			name: "One Resource Template",
			normal: &v1alpha1.TriggerTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: v1alpha1.TriggerTemplateSpec{
					ResourceTemplates: []v1alpha1.TriggerResourceTemplate{
						{
							RawExtension: runtime.RawExtension{Raw: []byte(`{"rt1": "value"}`)},
						},
					},
				},
			},
			builder: TriggerTemplate("name", "namespace",
				TriggerTemplateSpec(
					TriggerResourceTemplate(runtime.RawExtension{Raw: []byte(`{"rt1": "value"}`)}),
				),
			),
		},
		{
			name: "Two Resource Template",
			normal: &v1alpha1.TriggerTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: v1alpha1.TriggerTemplateSpec{
					ResourceTemplates: []v1alpha1.TriggerResourceTemplate{
						{
							RawExtension: runtime.RawExtension{Raw: []byte(`{"rt1": "value"}`)},
						},
						{
							RawExtension: runtime.RawExtension{Raw: []byte(`{"rt2": "value"}`)},
						},
					},
				},
			},
			builder: TriggerTemplate("name", "namespace",
				TriggerTemplateSpec(
					TriggerResourceTemplate(runtime.RawExtension{Raw: []byte(`{"rt1": "value"}`)}),
					TriggerResourceTemplate(runtime.RawExtension{Raw: []byte(`{"rt2": "value"}`)}),
				),
			),
		},
		{
			name: "Resource Templates, Params and extra Meta",
			normal: &v1alpha1.TriggerTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
					Labels: map[string]string{
						"key": "value",
					},
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "TriggerTemplate",
					APIVersion: "v1alpha1",
				},
				Spec: v1alpha1.TriggerTemplateSpec{
					Params: []v1alpha1.ParamSpec{
						{
							Name:        "param1",
							Description: "description",
							Default:     &defaultValue1,
						},
						{
							Name:        "param2",
							Description: "description",
							Default:     &defaultValue2,
						},
					},
					ResourceTemplates: []v1alpha1.TriggerResourceTemplate{
						{
							RawExtension: runtime.RawExtension{Raw: []byte(`{"rt1": "value"}`)},
						},
						{
							RawExtension: runtime.RawExtension{Raw: []byte(`{"rt2": "value"}`)},
						},
					},
				},
			},
			builder: TriggerTemplate("name", "namespace",
				TriggerTemplateMeta(
					TypeMeta("TriggerTemplate", "v1alpha1"),
					Label("key", "value"),
				),
				TriggerTemplateSpec(
					TriggerTemplateParam("param1", "description", "value1"),
					TriggerTemplateParam("param2", "description", "value2"),
					TriggerResourceTemplate(runtime.RawExtension{Raw: []byte(`{"rt1": "value"}`)}),
					TriggerResourceTemplate(runtime.RawExtension{Raw: []byte(`{"rt2": "value"}`)}),
				),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if diff := cmp.Diff(tt.normal, tt.builder); diff != "" {
				t.Errorf("TriggerBinding(): -want +got: %s", diff)
			}
		})
	}
}
