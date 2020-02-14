package builder

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTriggerTemplateBuilder(t *testing.T) {
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
					Params: []pipelinev1.ParamSpec{
						{
							Name:        "param1",
							Description: "description",
							Default: &pipelinev1.ArrayOrString{
								StringVal: "value1",
								Type:      pipelinev1.ParamTypeString,
							},
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
					Params: []pipelinev1.ParamSpec{
						{
							Name:        "param1",
							Description: "description",
							Default: &pipelinev1.ArrayOrString{
								StringVal: "value1",
								Type:      pipelinev1.ParamTypeString,
							},
						},
						{
							Name:        "param2",
							Description: "description",
							Default: &pipelinev1.ArrayOrString{
								StringVal: "value2",
								Type:      pipelinev1.ParamTypeString,
							},
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
							RawMessage: json.RawMessage(`{"rt1": "value"}`),
						},
					},
				},
			},
			builder: TriggerTemplate("name", "namespace",
				TriggerTemplateSpec(
					TriggerResourceTemplate(json.RawMessage(`{"rt1": "value"}`)),
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
							RawMessage: json.RawMessage(`{"rt1": "value"}`),
						},
						{
							RawMessage: json.RawMessage(`{"rt2": "value"}`),
						},
					},
				},
			},
			builder: TriggerTemplate("name", "namespace",
				TriggerTemplateSpec(
					TriggerResourceTemplate(json.RawMessage(`{"rt1": "value"}`)),
					TriggerResourceTemplate(json.RawMessage(`{"rt2": "value"}`)),
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
					Params: []pipelinev1.ParamSpec{
						{
							Name:        "param1",
							Description: "description",
							Default: &pipelinev1.ArrayOrString{
								StringVal: "value1",
								Type:      pipelinev1.ParamTypeString,
							},
						},
						{
							Name:        "param2",
							Description: "description",
							Default: &pipelinev1.ArrayOrString{
								StringVal: "value2",
								Type:      pipelinev1.ParamTypeString,
							},
						},
					},
					ResourceTemplates: []v1alpha1.TriggerResourceTemplate{
						{
							RawMessage: json.RawMessage(`{"rt1": "value"}`),
						},
						{
							RawMessage: json.RawMessage(`{"rt2": "value"}`),
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
					TriggerResourceTemplate(json.RawMessage(`{"rt1": "value"}`)),
					TriggerResourceTemplate(json.RawMessage(`{"rt2": "value"}`)),
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
