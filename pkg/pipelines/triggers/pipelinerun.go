package triggers

import (
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"

	"github.com/openshift/odo/pkg/pipelines/meta"
)

var (
	pipelineRunTypeMeta = meta.TypeMeta("PipelineRun", "tekton.dev/v1alpha1")
)

func createDevCDPipelineRun(saName, imageRepo string) pipelinev1.PipelineRun {
	return pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("", "app-cd-pipeline-run-$(uid)")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: saName,
			PipelineRef:        createPipelineRef("app-cd-pipeline"),
			Resources:          createDevResource(imageRepo+":$(params.gitsha)", "$(params.gitsha)"),
		},
	}
}

func createDevCIPipelineRun(saName, imageRepo string) pipelinev1.PipelineRun {
	return pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("", "app-ci-pipeline-run-$(uid)")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: saName,
			PipelineRef:        createPipelineRef("app-ci-pipeline"),
			Params: []pipelinev1.Param{
				createBindingParam("REPO", "$(params.fullname)"),
				createBindingParam("COMMIT_SHA", "$(params.gitsha)"),
			},
			Resources: createDevResource(imageRepo+":$(params.gitref)-$(params.gitsha)", "$(params.gitref)"),
		},
	}

}

func createCDPipelineRun(saName string) pipelinev1.PipelineRun {
	return pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("", "cd-deploy-from-push-pipeline-$(uid)")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: saName,
			PipelineRef:        createPipelineRef("cd-deploy-from-push-pipeline"),
			Resources:          createResources(),
		},
	}
}

func createCIPipelineRun(saName string) pipelinev1.PipelineRun {
	return pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("", "ci-dryrun-from-pr-pipeline-$(uid)")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: saName,
			PipelineRef:        createPipelineRef("ci-dryrun-from-pr-pipeline"),
			Resources:          createResources(),
		},
	}

}

func createDevResource(imageRepo string, param string) []pipelinev1.PipelineResourceBinding {
	return []pipelinev1.PipelineResourceBinding{
		{
			Name: "source-repo",
			ResourceSpec: &pipelinev1.PipelineResourceSpec{
				Type: "git",
				Params: []pipelinev1.ResourceParam{
					createResourceParams("revision", param),
					createResourceParams("url", "$(params.gitrepositoryurl)"),
				},
			},
		},
		{
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

func createResources() []pipelinev1.PipelineResourceBinding {
	return []pipelinev1.PipelineResourceBinding{
		{
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
