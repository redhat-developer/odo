package tasks

import (
	"github.com/openshift/odo/pkg/pipelines/meta"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
)

func CreateDeployFromSourceTask(ns, path string) pipelinev1.Task {
	task := pipelinev1.Task{
		TypeMeta:   taskTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, "deploy-from-source-task")),
		Spec: pipelinev1.TaskSpec{
			Inputs: createInputsForDeployFromSourceTask(path),
			Steps:  createStepsForDeployFromSourceTask(),
		},
	}
	return task
}

func createStepsForDeployFromSourceTask() []pipelinev1.Step {
	return []pipelinev1.Step{
		pipelinev1.Step{
			Container: createContainer(
				"run-kubectl",
				"quay.io/redhat-developer/k8s-kubectl",
				"/workspace/source",
				[]string{"kubectl"},
				argsForRunKubectlStep,
			),
		},
	}
}

var argsForRunKubectlStep = []string{
	"apply",
	"--dry-run=$(inputs.params.DRYRUN)",
	"-n",
	"$(inputs.params.NAMESPACE)",
	"-k",
	"$(inputs.params.PATHTODEPLOYMENT)",
}

func createInputsForDeployFromSourceTask(path string) *pipelinev1.Inputs {
	return &pipelinev1.Inputs{
		Resources: []pipelinev1.TaskResource{
			createTaskResource("source", "git"),
		},
		Params: []pipelinev1.ParamSpec{
			createTaskParamWithDefault(
				"PATHTODEPLOYMENT",
				"Path to the manifest to apply",
				pipelinev1.ParamTypeString,
				path,
			),
			createTaskParam(
				"NAMESPACE",
				"Namespace to deploy into",
				pipelinev1.ParamTypeString,
			),
			createTaskParamWithDefault(
				"DRYRUN",
				"If true run a server-side dryrun.",
				pipelinev1.ParamTypeString,
				"false",
			),
		},
	}

}
