package triggers

import (
	"encoding/json"

	"github.com/openshift/odo/pkg/pipelines/meta"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	triggerTemplateTypeMeta = meta.TypeMeta("TriggerTemplate", "tekton.dev/v1alpha1")
)

// GenerateTemplates will return a slice of trigger templates
func GenerateTemplates(ns, saName, imageRepo string) []triggersv1.TriggerTemplate {
	return []triggersv1.TriggerTemplate{
		createDevCDDeployTemplate(ns, saName, imageRepo),
		createDevCIBuildPRTemplate(ns, saName, imageRepo),
		CreateCDPushTemplate(ns, saName),
		CreateCIDryRunTemplate(ns, saName),
	}
}

func createDevCDDeployTemplate(ns, saName, imageRepo string) triggersv1.TriggerTemplate {
	return triggersv1.TriggerTemplate{
		TypeMeta:   triggerTemplateTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, "dev-cd-deploy-from-master-template")),
		Spec: triggersv1.TriggerTemplateSpec{
			Params: []pipelinev1.ParamSpec{
				createTemplateParamSpec("gitsha", "The specific commit SHA."),
				createTemplateParamSpec("gitrepositoryurl", "The git repository url"),
			},
			ResourceTemplates: []triggersv1.TriggerResourceTemplate{
				triggersv1.TriggerResourceTemplate{
					RawMessage: createDevCDResourcetemplate(saName, imageRepo),
				},
			},
		},
	}
}

func createDevCIBuildPRTemplate(ns, saName, imageRepo string) triggersv1.TriggerTemplate {
	return triggersv1.TriggerTemplate{
		TypeMeta: triggerTemplateTypeMeta,
		ObjectMeta: meta.ObjectMeta(
			meta.NamespacedName(ns, "dev-ci-build-from-pr-template"),
			statusTrackerAnnotations("dev-ci-build-from-pr", "Dev CI Build")),
		Spec: triggersv1.TriggerTemplateSpec{
			Params: []pipelinev1.ParamSpec{
				createTemplateParamSpec("gitref", "The git branch for this PR"),
				createTemplateParamSpec("gitsha", "the specific commit SHA."),
				createTemplateParamSpec("gitrepositoryurl", "The git repository url"),
				createTemplateParamSpec("fullname", "The GitHub repository for this PullRequest."),
			},
			ResourceTemplates: []triggersv1.TriggerResourceTemplate{
				triggersv1.TriggerResourceTemplate{
					RawMessage: createDevCIResourceTemplate(saName, imageRepo),
				},
			},
		},
	}

}

func CreateCDPushTemplate(ns, saName string) triggersv1.TriggerTemplate {
	return triggersv1.TriggerTemplate{
		TypeMeta:   triggerTemplateTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, "cd-deploy-from-push-template")),
		Spec: triggersv1.TriggerTemplateSpec{
			Params: []pipelinev1.ParamSpec{

				createTemplateParamSpecDefault("gitref", "The git revision", "master"),
				createTemplateParamSpec("gitrepositoryurl", "The git repository url"),
			},
			ResourceTemplates: []triggersv1.TriggerResourceTemplate{
				triggersv1.TriggerResourceTemplate{
					RawMessage: createCDResourceTemplate(saName),
				},
			},
		},
	}
}

func CreateCIDryRunTemplate(ns, saName string) triggersv1.TriggerTemplate {
	return triggersv1.TriggerTemplate{
		TypeMeta: triggerTemplateTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, "ci-dryrun-from-pr-template"),
			statusTrackerAnnotations("ci-dryrun-from-pr-pipeline", "Stage CI Dry Run")),
		Spec: triggersv1.TriggerTemplateSpec{
			Params: []pipelinev1.ParamSpec{

				createTemplateParamSpecDefault("gitref", "The git revision", "master"),
				createTemplateParamSpec("gitrepositoryurl", "The git repository url"),
			},
			ResourceTemplates: []triggersv1.TriggerResourceTemplate{
				triggersv1.TriggerResourceTemplate{
					RawMessage: createCIResourceTemplate(saName),
				},
			},
		},
	}
}

func createTemplateParamSpecDefault(name string, description string, value string) pipelinev1.ParamSpec {
	return pipelinev1.ParamSpec{
		Name:        name,
		Description: description,
		Default: &pipelinev1.ArrayOrString{
			StringVal: value,
			Type:      pipelinev1.ParamTypeString,
		},
	}
}

func createTemplateParamSpec(name string, description string) pipelinev1.ParamSpec {
	return pipelinev1.ParamSpec{
		Name:        name,
		Description: description,
	}

}

func createDevCDResourcetemplate(saName, imageRepo string) []byte {
	byteTemplate, _ := json.Marshal(createDevCDPipelineRun(saName, imageRepo))
	return []byte(string(byteTemplate))
}

func createDevCIResourceTemplate(saName, imageRepo string) []byte {
	byteTemplateCI, _ := json.Marshal(createDevCIPipelineRun(saName, imageRepo))
	return []byte(string(byteTemplateCI))
}

func createCDResourceTemplate(saName string) []byte {
	byteStageCD, _ := json.Marshal(createCDPipelineRun(saName))
	return []byte(string(byteStageCD))
}

func createCIResourceTemplate(saName string) []byte {
	byteStageCI, _ := json.Marshal(createCIPipelineRun(saName))
	return []byte(string(byteStageCI))
}

func statusTrackerAnnotations(pipeline, description string) func(*v1.ObjectMeta) {
	return func(om *v1.ObjectMeta) {
		annotations := map[string]string{
			"tekton.dev/git-status":         "true",
			"tekton.dev/status-context":     pipeline,
			"tekton.dev/status-description": description,
		}
		if om.Annotations == nil {
			om.Annotations = map[string]string{}
		}
		for k, v := range annotations {
			om.Annotations[k] = v
		}
	}
}
