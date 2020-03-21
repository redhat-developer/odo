package tasks

import (
	"github.com/openshift/odo/pkg/pipelines/meta"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

func generateDeployUsingKubectlTask(ns string) pipelinev1.Task {
	return pipelinev1.Task{
		TypeMeta:   taskTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, "deploy-using-kubectl-task")),
		Spec: pipelinev1.TaskSpec{
			Inputs: createInputsForDeployKubectlTask(),
			Steps:  createStepsForDeployKubectlTask(),
		},
	}
}

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

func createStepsForDeployKubectlTask() []pipelinev1.Step {
	return []pipelinev1.Step{
		pipelinev1.Step{
			Container: createContainer(
				"replace-image",
				"quay.io/redhat-developer/yq",
				"/workspace/source",
				[]string{"yq"},
				argsForReplaceImageStep,
			),
		},
		pipelinev1.Step{
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

func createInputsForDeployKubectlTask() *pipelinev1.Inputs {
	return &pipelinev1.Inputs{
		Resources: []pipelinev1.TaskResource{
			createTaskResource("source", "git"),
			createTaskResource("image", "image"),
		},
		Params: []pipelinev1.ParamSpec{
			createTaskParamWithDefault(
				"PATHTODEPLOYMENT",
				"Path to the manifest to apply",
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
				"The path to the image to replace in the yaml manifest (arg to yq)",
				"string",
			),
		},
	}
}
