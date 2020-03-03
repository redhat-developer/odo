package triggers

import (
	"encoding/json"

	"github.com/openshift/odo/pkg/pipelines/meta"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	triggerTemplateTypeMeta = v1.TypeMeta{
		Kind:       "TriggerTemplate",
		APIVersion: "tekton.dev/v1alpha1",
	}
)

// GenerateTemplates will return a slice of trigger templates
func GenerateTemplates(ns, saName, imageRepo string) []triggersv1.TriggerTemplate {
	return []triggersv1.TriggerTemplate{
		createDevCDDeployTemplate(ns, saName, imageRepo),
		createDevCIBuildPRTemplate(ns, saName, imageRepo),
		createStageCDPushTemplate(ns, saName),
		createStageCIdryrunptemplate(ns, saName),
	}
}

func createDevCDDeployTemplate(ns, saName, imageRepo string) triggersv1.TriggerTemplate {
	return triggersv1.TriggerTemplate{
		TypeMeta:   triggerTemplateTypeMeta,
		ObjectMeta: meta.CreateObjectMeta(ns, "dev-cd-deploy-from-master-template"),
		Spec: triggersv1.TriggerTemplateSpec{
			Params: []pipelinev1.ParamSpec{
				createTemplateParamSpecDefault("gitref", "The git revision", "master"),
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
		TypeMeta:   triggerTemplateTypeMeta,
		ObjectMeta: meta.CreateObjectMeta(ns, "dev-ci-build-from-pr-template"),
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

func createStageCDPushTemplate(ns, saName string) triggersv1.TriggerTemplate {
	return triggersv1.TriggerTemplate{
		TypeMeta:   triggerTemplateTypeMeta,
		ObjectMeta: meta.CreateObjectMeta(ns, "stage-cd-deploy-from-push-template"),
		Spec: triggersv1.TriggerTemplateSpec{
			Params: []pipelinev1.ParamSpec{

				createTemplateParamSpecDefault("gitref", "The git revision", "master"),
				createTemplateParamSpec("gitrepositoryurl", "The git repository url"),
			},
			ResourceTemplates: []triggersv1.TriggerResourceTemplate{
				triggersv1.TriggerResourceTemplate{
					RawMessage: createStageCDResourceTemplate(saName),
				},
			},
		},
	}
}

func createStageCIdryrunptemplate(ns, saName string) triggersv1.TriggerTemplate {
	return triggersv1.TriggerTemplate{
		TypeMeta:   triggerTemplateTypeMeta,
		ObjectMeta: meta.CreateObjectMeta(ns, "stage-ci-dryrun-from-pr-template"),
		Spec: triggersv1.TriggerTemplateSpec{
			Params: []pipelinev1.ParamSpec{

				createTemplateParamSpecDefault("gitref", "The git revision", "master"),
				createTemplateParamSpec("gitrepositoryurl", "The git repository url"),
			},
			ResourceTemplates: []triggersv1.TriggerResourceTemplate{
				triggersv1.TriggerResourceTemplate{
					RawMessage: createStageCIResourceTemplate(saName),
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

func createStageCDResourceTemplate(saName string) []byte {
	byteStageCD, _ := json.Marshal(createStageCDPipelineRun(saName))
	return []byte(string(byteStageCD))
}

func createStageCIResourceTemplate(saName string) []byte {
	byteStageCI, _ := json.Marshal(createStageCIPipelineRun(saName))
	return []byte(string(byteStageCI))
}
