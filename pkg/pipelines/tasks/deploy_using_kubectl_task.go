package tasks

import (
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	"github.com/openshift/odo/pkg/pipelines/meta"
)

var argsForReplaceImageStep = []string{
	"w",
	"-i",
	"$(inputs.params.PATHTODEPLOYMENT)/deployment.yaml",
	"$(inputs.params.YAMLPATHTOIMAGE)",
	"$(inputs.resources.image.url)",
}

var argsForKubectlStep = []string{
	"apply",
	"-n",
	"$(inputs.params.NAMESPACE)",
	"-k",
	"$(inputs.params.PATHTODEPLOYMENT)",
}

// CreateDeployUsingKubectlTask creates DeployUsingKubectlTask
func CreateDeployUsingKubectlTask(ns string) pipelinev1.Task {
	return pipelinev1.Task{
		TypeMeta:   taskTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, "deploy-using-kubectl-task")),
		Spec: pipelinev1.TaskSpec{
			Params:    paramsForDeployKubectlTask(),
			Resources: createResourcesForDeployKubectlTask(),
			Steps:     createStepsForDeployKubectlTask(),
		},
	}
}

func createStepsForDeployKubectlTask() []pipelinev1.Step {
	return []pipelinev1.Step{
		{
			Container: createContainer(
				"replace-image",
				"quay.io/redhat-developer/yq",
				"/workspace/source",
				[]string{"yq"},
				argsForReplaceImageStep,
			),
		},
		{
			Container: createContainer(
				"run-kubectl",
				"quay.io/redhat-developer/k8s-kubectl",
				"/workspace/source",
				[]string{"kubectl"},
				argsForKubectlStep,
			),
		},
	}
}

func paramsForDeployKubectlTask() []pipelinev1.ParamSpec {
	return []pipelinev1.ParamSpec{
		createTaskParamWithDefault(
			"PATHTODEPLOYMENT",
			"Path to the pipelines to apply",
			"string",
			"deploy",
		),
		createTaskParam(
			"NAMESPACE",
			"Namespace to deploy into",
			"string",
		),
		createTaskParamWithDefault(
			"DRYRUN",
			"If true run a server-side dryrun.",
			"string",
			"false",
		),
		createTaskParam(
			"YAMLPATHTOIMAGE",
			"The path to the image to replace in the yaml pipelines (arg to yq)",
			"string",
		),
	}
}

func createResourcesForDeployKubectlTask() *pipelinev1.TaskResources {
	return &pipelinev1.TaskResources{
		Inputs: []pipelinev1.TaskResource{
			createTaskResource("source", "git"),
			createTaskResource("image", "image"),
		},
	}
}
