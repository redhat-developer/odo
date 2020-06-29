package tasks

import (
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	"github.com/openshift/odo/pkg/pipelines/meta"
)

// CreateDeployFromSourceTask creates DeployFromSourceTask
func CreateDeployFromSourceTask(ns, script string) pipelinev1.Task {
	task := pipelinev1.Task{
		TypeMeta:   taskTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, "deploy-from-source-task")),
		Spec: pipelinev1.TaskSpec{
			Params:    paramsForDeploymentFromSourceTask(),
			Resources: createResourcesForDeployFromSourceTask(),
			Steps:     createStepsForDeployFromSourceTask(script),
		},
	}
	return task
}

func createStepsForDeployFromSourceTask(script string) []pipelinev1.Step {
	return []pipelinev1.Step{
		{
			Container: createContainer(
				"run-kubectl",
				"quay.io/redhat-developer/k8s-kubectl",
				"/workspace/source",
				nil,
				nil,
			),
			Script: script,
		},
	}
}

func paramsForDeploymentFromSourceTask() []pipelinev1.ParamSpec {
	return []pipelinev1.ParamSpec{
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
