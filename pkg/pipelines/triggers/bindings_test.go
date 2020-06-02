package triggers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateBindingParam(t *testing.T) {
	validParam := pipelinev1.Param{
		Name: "gitref",
		Value: pipelinev1.ArrayOrString{
			StringVal: "$(body.ref)",
			Type:      pipelinev1.ParamTypeString,
		},
	}
	bindingParam := createPipelineBindingParam("gitref", "$(body.ref)")
	if diff := cmp.Diff(validParam, bindingParam); diff != "" {
		t.Fatalf("createPipelineBindingParam() failed\n%s", diff)
	}
}

func TestCreateImageRepoBinding(t *testing.T) {
	imageRepoBinding := triggersv1.TriggerBinding{
		TypeMeta: triggerBindingTypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-binding",
			Namespace: "testns",
		},
		Spec: triggersv1.TriggerBindingSpec{
			Params: []triggersv1.Param{
				{
					Name:  "imageRepo",
					Value: "quay.io/user/testing",
				},
				{
					Name:  "tlsVerify",
					Value: "true",
				},
			},
		},
	}
	binding := CreateImageRepoBinding("testns", "test-binding", "quay.io/user/testing", "true")
	if diff := cmp.Diff(imageRepoBinding, binding); diff != "" {
		t.Fatalf("CreateImageRepoBinding() failed:\n%s", diff)
	}
}
