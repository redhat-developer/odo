package triggers

import (
	"github.com/openshift/odo/pkg/pipelines/meta"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

var (
	pipelineRunTypeMeta = meta.TypeMeta("PipelineRun", "tekton.dev/v1alpha1")
)

func createDevCDPipelineRun(saName, imageRepo string) pipelinev1.PipelineRun {
	return pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("", "dev-cd-pipeline-run-$(uid)")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: saName,
			PipelineRef:        createPipelineRef("dev-cd-pipeline"),
			Resources:          createDevResource(imageRepo+":$(params.gitsha)", "$(params.gitsha)"),
		},
	}
}

func createDevCIPipelineRun(saName, imageRepo string) pipelinev1.PipelineRun {
	return pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("", "dev-ci-pipeline-run-$(uid)")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: saName,
			PipelineRef:        createPipelineRef("dev-ci-pipeline"),
			Params: []pipelinev1.Param{
				createBindingParam("REPO", "$(params.fullname)"),
				createBindingParam("COMMIT_SHA", "$(params.gitsha)"),
			},
			Resources: createDevResource(imageRepo+":$(params.gitref)-$(params.gitsha)", "$(params.gitref)"),
		},
	}

}

func createStageCDPipelineRun(saName string) pipelinev1.PipelineRun {
	return pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("", "stage-cd-pipeline-run-$(uid)")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: saName,
			PipelineRef:        createPipelineRef("stage-cd-pipeline"),
			Resources:          createStageResources(),
		},
	}
}

func createStageCIPipelineRun(saName string) pipelinev1.PipelineRun {
	return pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("", "stage-ci-pipeline-run-$(uid)")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: saName,
			PipelineRef:        createPipelineRef("stage-ci-pipeline"),
			Resources:          createStageResources(),
		},
	}

}

func createDevResource(imageRepo string, param string) []pipelinev1.PipelineResourceBinding {
	return []pipelinev1.PipelineResourceBinding{
		pipelinev1.PipelineResourceBinding{
			Name: "source-repo",
			ResourceSpec: &pipelinev1.PipelineResourceSpec{
				Type: "git",
				Params: []pipelinev1.ResourceParam{
					createResourceParams("revision", param),
					createResourceParams("url", "$(params.gitrepositoryurl)"),
				},
			},
		},
		pipelinev1.PipelineResourceBinding{
			Name: "runtime-image",
			ResourceSpec: &pipelinev1.PipelineResourceSpec{
				Type: "image",
				Params: []pipelinev1.ResourceParam{
					createResourceParams("url", imageRepo),
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
