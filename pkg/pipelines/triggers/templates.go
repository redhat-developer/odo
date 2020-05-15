package triggers

import (
	"encoding/json"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/odo/pkg/pipelines/meta"
)

var (
	triggerTemplateTypeMeta = meta.TypeMeta("TriggerTemplate", "tekton.dev/v1alpha1")
)

// CreateDevCDDeployTemplate creates DevCDDeployTemplate
func CreateDevCDDeployTemplate(ns, saName string) triggersv1.TriggerTemplate {
	return triggersv1.TriggerTemplate{
		TypeMeta:   triggerTemplateTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, "app-cd-template")),
		Spec: triggersv1.TriggerTemplateSpec{
			Params: []pipelinev1.ParamSpec{
				createTemplateParamSpec("gitsha", "The specific commit SHA."),
				createTemplateParamSpec("gitrepositoryurl", "The git repository url"),
			},
			ResourceTemplates: []triggersv1.TriggerResourceTemplate{
				{
					RawMessage: createDevCDResourcetemplate(saName),
				},
			},
		},
	}
}

// CreateDevCIBuildPRTemplate creates DevCIBuildPRTemplate
func CreateDevCIBuildPRTemplate(ns, saName string) triggersv1.TriggerTemplate {
	return triggersv1.TriggerTemplate{
		TypeMeta: triggerTemplateTypeMeta,
		ObjectMeta: meta.ObjectMeta(
			meta.NamespacedName(ns, "app-ci-template"),
			statusTrackerAnnotations("dev-ci-build-from-pr", "Dev CI Build")),
		Spec: triggersv1.TriggerTemplateSpec{
			Params: []pipelinev1.ParamSpec{
				createTemplateParamSpec("gitref", "The git branch for this PR."),
				createTemplateParamSpec("gitsha", "the specific commit SHA."),
				createTemplateParamSpec("gitrepositoryurl", "The git repository URL."),
				createTemplateParamSpec("fullname", "The GitHub repository for this PullRequest."),
				createTemplateParamSpec("imageRepo", "The repository to push built images to."),
			},
			ResourceTemplates: []triggersv1.TriggerResourceTemplate{
				{
					RawMessage: createDevCIResourceTemplate(saName),
				},
			},
		},
	}

}

// CreateCDPushTemplate returns TriggerTemplate for CD Push Request
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
				{
					RawMessage: createCDResourceTemplate(saName),
				},
			},
		},
	}
}

// CreateCIDryRunTemplate returns TriggerTemplate for CI Dry Try
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
				{
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

func createDevCDResourcetemplate(saName string) []byte {
	byteTemplate, _ := json.Marshal(createDevCDPipelineRun(saName))
	return []byte(string(byteTemplate))
}

func createDevCIResourceTemplate(saName string) []byte {
	byteTemplateCI, _ := json.Marshal(createDevCIPipelineRun(saName))
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
