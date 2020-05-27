package tasks

import (
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	"github.com/openshift/odo/pkg/pipelines/meta"
)

var argsForRunKubectlStep = []string{
	"apply",
	"--dry-run=$(inputs.params.DRYRUN)",
	"-n",
	"$(inputs.params.NAMESPACE)",
	"-k",
	"$(inputs.params.PATHTODEPLOYMENT)",
}

// CreateDeployFromSourceTask creates DeployFromSourceTask
func CreateDeployFromSourceTask(ns, path string) pipelinev1.Task {
	task := pipelinev1.Task{
		TypeMeta:   taskTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, "deploy-from-source-task")),
		Spec: pipelinev1.TaskSpec{
			Params:    paramsForDeploymentFromSourceTask(path),
			Resources: createResourcesForDeployFromSourceTask(),
			Steps:     createStepsForDeployFromSourceTask(),
		},
	}
	return task
}

func createStepsForDeployFromSourceTask() []pipelinev1.Step {
	return []pipelinev1.Step{
		{
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

func paramsForDeploymentFromSourceTask(path string) []pipelinev1.ParamSpec {
	return []pipelinev1.ParamSpec{
		createTaskParamWithDefault(
			"PATHTODEPLOYMENT",
			"Path to the pipelines to apply",
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
	}
}

func createResourcesForDeployFromSourceTask() *pipelinev1.TaskResources {
	return &pipelinev1.TaskResources{
		Inputs: []pipelinev1.TaskResource{
			createTaskResource("source", "git"),
		},
	}
}
