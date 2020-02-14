package builder

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTriggerBindingBuilder(t *testing.T) {
	tests := []struct {
		name    string
		normal  *v1alpha1.TriggerBinding
		builder *v1alpha1.TriggerBinding
	}{
		{
			name:    "Empty",
			normal:  &v1alpha1.TriggerBinding{},
			builder: TriggerBinding("", ""),
		},
		{
			name: "Name and Namespace",
			normal: &v1alpha1.TriggerBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
			},
			builder: TriggerBinding("name", "namespace"),
		},
		{
			name: "One Param",
			normal: &v1alpha1.TriggerBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: v1alpha1.TriggerBindingSpec{
					Params: []pipelinev1.Param{
						{
							Name: "param1",
							Value: pipelinev1.ArrayOrString{
								StringVal: "value1",
								Type:      pipelinev1.ParamTypeString,
							},
						},
					},
				},
			},
			builder: TriggerBinding("name", "namespace",
				TriggerBindingSpec(
					TriggerBindingParam("param1", "value1"),
				),
			),
		},
		{
			name: "Two Params",
			normal: &v1alpha1.TriggerBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
				},
				Spec: v1alpha1.TriggerBindingSpec{
					Params: []pipelinev1.Param{
						{
							Name: "param1",
							Value: pipelinev1.ArrayOrString{
								StringVal: "value1",
								Type:      pipelinev1.ParamTypeString,
							},
						},
						{
							Name: "param2",
							Value: pipelinev1.ArrayOrString{
								StringVal: "value2",
								Type:      pipelinev1.ParamTypeString,
							},
						},
					},
				},
			},
			builder: TriggerBinding("name", "namespace",
				TriggerBindingSpec(
					TriggerBindingParam("param1", "value1"),
					TriggerBindingParam("param2", "value2"),
				),
			),
		},
		{
			name: "Extra Meta",
			normal: &v1alpha1.TriggerBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "name",
					Namespace: "namespace",
					Labels: map[string]string{
						"key": "value",
					},
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "TriggerBinding",
					APIVersion: "v1alpha1",
				},
				Spec: v1alpha1.TriggerBindingSpec{},
			},
			builder: TriggerBinding("name", "namespace",
				TriggerBindingMeta(
					TypeMeta("TriggerBinding", "v1alpha1"),
					Label("key", "value"),
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
