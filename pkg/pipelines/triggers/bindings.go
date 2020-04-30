package triggers

import (
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"

	"github.com/openshift/odo/pkg/pipelines/meta"
)

var (
	triggerBindingTypeMeta = meta.TypeMeta("TriggerBinding", "tekton.dev/v1alpha1")
)

// GenerateBindings returns a slice of trigger bindings
func GenerateBindings(ns string) []triggersv1.TriggerBinding {
	return []triggersv1.TriggerBinding{
		CreatePRBinding(ns),
		CreatePushBinding(ns),
	}
}

// CreatePRBinding returns TriggerBindings for Pull Request
func CreatePRBinding(ns string) triggersv1.TriggerBinding {
	return triggersv1.TriggerBinding{
		TypeMeta:   triggerBindingTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, "github-pr-binding")),
		Spec: triggersv1.TriggerBindingSpec{
			Params: []pipelinev1.Param{
				createBindingParam("gitref", "$(body.pull_request.head.ref)"),
				createBindingParam("gitsha", "$(body.pull_request.head.sha)"),
				createBindingParam("gitrepositoryurl", "$(body.repository.clone_url)"),
				createBindingParam("fullname", "$(body.repository.full_name)"),
			},
		},
	}
}

// CreatePushBinding returns TriggerBinding for Push Request
func CreatePushBinding(ns string) triggersv1.TriggerBinding {
	return triggersv1.TriggerBinding{
		TypeMeta:   triggerBindingTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, "github-push-binding")),
		Spec: triggersv1.TriggerBindingSpec{
			Params: []pipelinev1.Param{
				createBindingParam("gitref", "$(body.ref)"),
				createBindingParam("gitsha", "$(body.head_commit.id)"),
				createBindingParam("gitrepositoryurl", "$(body.repository.clone_url)"),
			},
		},
	}
}

func createBindingParam(name string, value string) pipelinev1.Param {
	return pipelinev1.Param{
		Name: name,
		Value: pipelinev1.ArrayOrString{
			StringVal: value,
			Type:      pipelinev1.ParamTypeString,
		},
	}
}
