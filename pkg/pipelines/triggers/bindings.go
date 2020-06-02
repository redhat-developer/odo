package triggers

import (
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"

	"github.com/openshift/odo/pkg/pipelines/meta"
)

var (
	TriggerBindingTypeMeta = meta.TypeMeta("TriggerBinding", "triggers.tekton.dev/v1alpha1")
)

// CreateImageRepoBinding returns a TriggerBinding with the imageRepo.
func CreateImageRepoBinding(ns, bindingName, imageRepo, tlsVerify string) triggersv1.TriggerBinding {
	return triggersv1.TriggerBinding{
		TypeMeta:   TriggerBindingTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, bindingName)),
		Spec: triggersv1.TriggerBindingSpec{
			Params: []triggersv1.Param{
				createBindingParam("imageRepo", imageRepo),
				createBindingParam("tlsVerify", tlsVerify),
			},
		},
	}
}

func createBindingParam(name string, value string) triggersv1.Param {
	return triggersv1.Param{
		Name:  name,
		Value: value,
	}
}
