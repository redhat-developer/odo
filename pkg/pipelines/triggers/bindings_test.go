package triggers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreateDevCDDeployBinding(t *testing.T) {
	validDevCDBinding := triggersv1.TriggerBinding{
		TypeMeta: v1.TypeMeta{
			Kind:       "TriggerBinding",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "dev-cd-deploy-from-master-binding",
			Namespace: "testns",
		},
		Spec: triggersv1.TriggerBindingSpec{
			Params: []pipelinev1.Param{
				pipelinev1.Param{
					Name: "gitref",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.head_commit.id)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				pipelinev1.Param{
					Name: "gitrepositoryurl",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.repository.clone_url)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
			},
		},
	}
	binding := createDevCDDeployBinding("testns")
	if diff := cmp.Diff(validDevCDBinding, binding); diff != "" {
		t.Fatalf("createDevCDDeployBinding() failed:\n%s", diff)
	}
}

func TestCreateDevCIBuildBinding(t *testing.T) {
	validDevCIBinding := triggersv1.TriggerBinding{
		TypeMeta: v1.TypeMeta{
			Kind:       "TriggerBinding",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "dev-ci-build-from-pr-binding",
			Namespace: "testns",
		},
		Spec: triggersv1.TriggerBindingSpec{
			Params: []pipelinev1.Param{
				pipelinev1.Param{
					Name: "gitref",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.pull_request.head.ref)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				pipelinev1.Param{
					Name: "gitsha",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.pull_request.head.sha)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				pipelinev1.Param{
					Name: "gitrepositoryurl",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.repository.clone_url)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				pipelinev1.Param{
					Name: "fullname",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.repository.full_name)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
			},
		},
	}
	binding := createDevCIBuildBinding("testns")
	if diff := cmp.Diff(validDevCIBinding, binding); diff != "" {
		t.Fatalf("createDevCIBuildBinding() failed:\n%s", diff)
	}
}

func TestCreateStageCDDeployBinding(t *testing.T) {
	validStageCDDeployBinding := triggersv1.TriggerBinding{
		TypeMeta: v1.TypeMeta{
			Kind:       "TriggerBinding",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "stage-cd-deploy-from-push-binding",
			Namespace: "testns",
		},
		Spec: triggersv1.TriggerBindingSpec{
			Params: []pipelinev1.Param{
				pipelinev1.Param{
					Name: "gitref",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.ref)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				pipelinev1.Param{
					Name: "gitsha",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.commits[0].id)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				pipelinev1.Param{
					Name: "gitrepositoryurl",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.repository.clone_url)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
			},
		},
	}
	binding := createStageCDDeployBinding("testns")
	if diff := cmp.Diff(validStageCDDeployBinding, binding); diff != "" {
		t.Fatalf("createDevCIBuildBinding() failed:\n%s", diff)
	}
}

func TestCreateStageCIDryRunBinding(t *testing.T) {
	validStageCIDryRunBinding := triggersv1.TriggerBinding{
		TypeMeta: v1.TypeMeta{
			Kind:       "TriggerBinding",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "stage-ci-dryrun-from-pr-binding",
			Namespace: "testns",
		},
		Spec: triggersv1.TriggerBindingSpec{
			Params: []pipelinev1.Param{
				pipelinev1.Param{
					Name: "gitref",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.pull_request.head.ref)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
				pipelinev1.Param{
					Name: "gitrepositoryurl",
					Value: pipelinev1.ArrayOrString{
						StringVal: "$(body.repository.clone_url)",
						Type:      pipelinev1.ParamTypeString,
					},
				},
			},
		},
	}
	binding := createStageCIDryRunBinding("testns")
	if diff := cmp.Diff(validStageCIDryRunBinding, binding); diff != "" {
		t.Errorf("createStageCIDryRunBinding() failed:\n%s", diff)
	}
}

func TestCreateBindingParam(t *testing.T) {
	validParam := pipelinev1.Param{
		Name: "gitref",
		Value: pipelinev1.ArrayOrString{
			StringVal: "$(body.ref)",
			Type:      pipelinev1.ParamTypeString,
		},
	}
	bindingParam := createBindingParam("gitref", "$(body.ref)")
	if diff := cmp.Diff(validParam, bindingParam); diff != "" {
		t.Fatalf("createBindingParam() failed\n%s", diff)
	}
}
