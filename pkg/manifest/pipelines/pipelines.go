package pipelines

import (
	"github.com/openshift/odo/pkg/manifest/meta"
	"k8s.io/apimachinery/pkg/types"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

var (
	pipelineTypeMeta = meta.TypeMeta("Pipeline", "tekton.dev/v1alpha1")
)

// CreateCDPipeline create CI pileline
func CreateCDPipeline(name types.NamespacedName, stageNamespace string) *pipelinev1.Pipeline {
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

// CreateCIPipeline creates CI pipeline
func CreateCIPipeline(name types.NamespacedName, stageNamespace string) *pipelinev1.Pipeline {
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

func createInputTaskResource(name string, resource string) pipelinev1.PipelineTaskInputResource {
	return pipelinev1.PipelineTaskInputResource{
		Name:     name,
		Resource: resource,
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

func createOutputTaskResource(name string, resource string) pipelinev1.PipelineTaskOutputResource {
	return pipelinev1.PipelineTaskOutputResource{
		Name:     name,
		Resource: resource,
	}
}

func createTaskRef(name string, kind pipelinev1.TaskKind) *pipelinev1.TaskRef {
	return &pipelinev1.TaskRef{
		Name: name,
		Kind: kind,
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

// Returns string value for TLSVERIFY parameter based on usingInternalRegistry boolean
// If internal registry is used, we need to disable TLS verification
func getTLSVerify(usingInternalRegistry bool) string {
	if usingInternalRegistry {
		return "false"
	}
	return "true"
}

func createPipelineDeclaredResource(name string, resourceType string) pipelinev1.PipelineDeclaredResource {
	return pipelinev1.PipelineDeclaredResource{Name: name, Type: resourceType}
}
