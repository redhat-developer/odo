package pipelines

import (
	"github.com/openshift/odo/pkg/pipelines/meta"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	pipelineTypeMeta = meta.TypeMeta("Pipeline", "tekton.dev/v1alpha1")
)

// createCIPipeline creates the dev-ci-pipeline.
func createDevCIPipeline(name types.NamespacedName, isInternalRegistry bool) *pipelinev1.Pipeline {
	return &pipelinev1.Pipeline{
		TypeMeta:   pipelineTypeMeta,
		ObjectMeta: meta.ObjectMeta(name),
		Spec: pipelinev1.PipelineSpec{
			Params: []pipelinev1.ParamSpec{
				createParamSpec("REPO", "string"),
				createParamSpec("COMMIT_SHA", "string"),
			},
			Resources: []pipelinev1.PipelineDeclaredResource{
				createPipelineDeclaredResource("source-repo", "git"),
				createPipelineDeclaredResource("runtime-image", "image"),
			},

			Tasks: []pipelinev1.PipelineTask{
				createBuildImageTask("build-image", isInternalRegistry),
			},
		},
	}
}

func createCIPipeline(name types.NamespacedName, stageNamespace string) *pipelinev1.Pipeline {
	return &pipelinev1.Pipeline{
		TypeMeta:   pipelineTypeMeta,
		ObjectMeta: meta.ObjectMeta(name),
		Spec: pipelinev1.PipelineSpec{

			Resources: []pipelinev1.PipelineDeclaredResource{
				createPipelineDeclaredResource("source-repo", "git"),
			},

			Tasks: []pipelinev1.PipelineTask{
				createCIPipelineTask("apply-source", stageNamespace),
			},
		},
	}
}

func createDevCDPipeline(name types.NamespacedName, deploymentPath, devNamespace string, isInternalRegistry bool) *pipelinev1.Pipeline {
	return &pipelinev1.Pipeline{
		TypeMeta:   pipelineTypeMeta,
		ObjectMeta: meta.ObjectMeta(name),
		Spec: pipelinev1.PipelineSpec{
			Resources: []pipelinev1.PipelineDeclaredResource{
				createPipelineDeclaredResource("source-repo", "git"),
				createPipelineDeclaredResource("runtime-image", "image"),
			},
			Tasks: []pipelinev1.PipelineTask{
				createDevCDBuildImageTask("build-image", isInternalRegistry),
				createDevCDDeployImageTask("deploy-image", devNamespace, deploymentPath),
			},
		},
	}
}

func createCDPipeline(name types.NamespacedName, stageNamespace string) *pipelinev1.Pipeline {
	return &pipelinev1.Pipeline{
		TypeMeta:   pipelineTypeMeta,
		ObjectMeta: meta.ObjectMeta(name),
		Spec: pipelinev1.PipelineSpec{
			Resources: []pipelinev1.PipelineDeclaredResource{
				createPipelineDeclaredResource("source-repo", "git"),
			},
			Tasks: []pipelinev1.PipelineTask{
				createCDPipelineTask("apply-source", stageNamespace),
			},
		},
	}
}

func createParamSpec(name string, paramType pipelinev1.ParamType) pipelinev1.ParamSpec {
	return pipelinev1.ParamSpec{Name: name, Type: paramType}
}

func createPipelineDeclaredResource(name string, resourceType string) pipelinev1.PipelineDeclaredResource {
	return pipelinev1.PipelineDeclaredResource{Name: name, Type: resourceType}
}

// Returns string value for TLSVERIFY parameter based on usingInternalRegistry boolean
// If internal registry is used, we need to disable TLS verification
func getTLSVerify(usingInternalRegistry bool) string {
	if usingInternalRegistry {
		return "false"
	}
	return "true"
}

func createBuildImageTask(name string, isInternalRegistry bool) pipelinev1.PipelineTask {
	return pipelinev1.PipelineTask{
		Name:    name,
		TaskRef: createTaskRef("buildah", pipelinev1.ClusterTaskKind),
		Resources: &pipelinev1.PipelineTaskResources{
			Inputs:  []pipelinev1.PipelineTaskInputResource{createInputTaskResource("source", "source-repo")},
			Outputs: []pipelinev1.PipelineTaskOutputResource{createOutputTaskResource("image", "runtime-image")},
		},
		Params: []pipelinev1.Param{
			createTaskParam("TLSVERIFY", getTLSVerify(isInternalRegistry)),
		},
	}

}

func createDevCDBuildImageTask(name string, isInternalRegistry bool) pipelinev1.PipelineTask {
	return pipelinev1.PipelineTask{
		Name:    name,
		TaskRef: createTaskRef("buildah", pipelinev1.ClusterTaskKind),
		Resources: &pipelinev1.PipelineTaskResources{
			Inputs:  []pipelinev1.PipelineTaskInputResource{createInputTaskResource("source", "source-repo")},
			Outputs: []pipelinev1.PipelineTaskOutputResource{createOutputTaskResource("image", "runtime-image")},
		},
		Params: []pipelinev1.Param{
			createTaskParam("TLSVERIFY", getTLSVerify(isInternalRegistry)),
		},
	}
}

func createDevCDDeployImageTask(name, devNamespace, deploymentPath string) pipelinev1.PipelineTask {
	return pipelinev1.PipelineTask{
		Name:     name,
		TaskRef:  createTaskRef("deploy-using-kubectl-task", pipelinev1.NamespacedTaskKind),
		RunAfter: []string{"build-image"},
		Resources: &pipelinev1.PipelineTaskResources{
			Inputs: []pipelinev1.PipelineTaskInputResource{
				createInputTaskResource("source", "source-repo"),
				createInputTaskResource("image", "runtime-image"),
			},
		},
		Params: []pipelinev1.Param{
			createTaskParam("PATHTODEPLOYMENT", deploymentPath),
			createTaskParam("YAMLPATHTOIMAGE", "spec.template.spec.containers[0].image"),
			createTaskParam("NAMESPACE", devNamespace),
		},
	}
}

func createCIPipelineTask(taskName, stageNamespace string) pipelinev1.PipelineTask {
	return pipelinev1.PipelineTask{
		Name:    taskName,
		TaskRef: createTaskRef("deploy-from-source-task", pipelinev1.NamespacedTaskKind),
		Resources: &pipelinev1.PipelineTaskResources{
			Inputs: []pipelinev1.PipelineTaskInputResource{createInputTaskResource("source", "source-repo")},
		},
		Params: []pipelinev1.Param{
			createTaskParam("NAMESPACE", stageNamespace),
			createTaskParam("DRYRUN", "true"),
		},
	}
}

func createCDPipelineTask(taskName, stageNamespace string) pipelinev1.PipelineTask {
	return pipelinev1.PipelineTask{
		Name:    taskName,
		TaskRef: createTaskRef("deploy-from-source-task", pipelinev1.NamespacedTaskKind),
		Resources: &pipelinev1.PipelineTaskResources{
			Inputs: []pipelinev1.PipelineTaskInputResource{createInputTaskResource("source", "source-repo")},
		},
		Params: []pipelinev1.Param{
			createTaskParam("NAMESPACE", stageNamespace),
		},
	}
}

func createTaskParam(name, value string) pipelinev1.Param {
	return pipelinev1.Param{
		Name: name,

		Value: pipelinev1.ArrayOrString{
			Type:      pipelinev1.ParamTypeString,
			StringVal: value,
		},
	}
}

func createTaskRef(name string, kind pipelinev1.TaskKind) *pipelinev1.TaskRef {
	return &pipelinev1.TaskRef{
		Name: name,
		Kind: kind,
	}
}

func createInputTaskResource(name string, resource string) pipelinev1.PipelineTaskInputResource {
	return pipelinev1.PipelineTaskInputResource{
		Name:     name,
		Resource: resource,
	}

}

func createOutputTaskResource(name string, resource string) pipelinev1.PipelineTaskOutputResource {
	return pipelinev1.PipelineTaskOutputResource{
		Name:     name,
		Resource: resource,
	}
}
