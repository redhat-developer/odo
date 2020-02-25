package triggers

import (
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	pipelineRunTypeMeta = v1.TypeMeta{
		Kind:       "PipelineRun",
		APIVersion: "tekton.dev/v1alpha1",
	}
)

func createDevCDPipelineRun() pipelinev1.PipelineRun {
	return pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: createObjectMeta("dev-cd-pipeline-run-$(uid)"),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: "demo-sa",
			PipelineRef:        createPipelineRef("dev-cd-pipeline"),
			Resources:          createDevResource(),
		},
	}

}
func createDevCIPipelineRun() pipelinev1.PipelineRun {
	return pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: createObjectMeta("dev-ci-pipeline-run-$(uid)"),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: "demo-sa",
			PipelineRef:        createPipelineRef("dev-ci-pipeline"),
			Resources:          createDevResource(),
		},
	}

}

func createStageCDPipelineRun() pipelinev1.PipelineRun {
	return pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: createObjectMeta("stage-cd-pipeline-run-$(uid)"),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: "demo-sa",
			PipelineRef:        createPipelineRef("stage-ci-pipeline"),
			Resources:          createStageResources(),
		},
	}
}

func createStageCIPipelineRun() pipelinev1.PipelineRun {
	return pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: createObjectMeta("stage-ci-pipeline-run-$(uid)"),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: "demo-sa",
			PipelineRef:        createPipelineRef("stage-ci-pipeline"),
			Resources:          createStageResources(),
		},
	}

}

func createDevResource() []pipelinev1.PipelineResourceBinding {
	return []pipelinev1.PipelineResourceBinding{
		pipelinev1.PipelineResourceBinding{
			Name: "source-repo",
			ResourceSpec: &pipelinev1.PipelineResourceSpec{
				Type: "git",
				Params: []pipelinev1.ResourceParam{
					createResourceParams("revision", "$(params.gitref)"),
					createResourceParams("url", "$(params.gitrepositoryurl)"),
				},
			},
		},
		pipelinev1.PipelineResourceBinding{
			Name: "runtime-image",
			ResourceSpec: &pipelinev1.PipelineResourceSpec{
				Type: "image",
				Params: []pipelinev1.ResourceParam{
					createResourceParams("url", "REPLACE_IMAGE:$(params.gitref)"),
				},
			},
		},
	}
}

func createStageResources() []pipelinev1.PipelineResourceBinding {
	return []pipelinev1.PipelineResourceBinding{
		pipelinev1.PipelineResourceBinding{
			Name: "source-repo",
			ResourceSpec: &pipelinev1.PipelineResourceSpec{
				Type: "git",
				Params: []pipelinev1.ResourceParam{
					createResourceParams("revision", "$(params.gitref)"),
					createResourceParams("url", "$(params.gitrepositoryurl)"),
				},
			},
		},
	}
}

func createResourceParams(name string, value string) pipelinev1.ResourceParam {
	return pipelinev1.ResourceParam{
		Name:  name,
		Value: value,
	}

}
func createPipelineRef(name string) *pipelinev1.PipelineRef {
	return &pipelinev1.PipelineRef{
		Name: name,
	}
}
