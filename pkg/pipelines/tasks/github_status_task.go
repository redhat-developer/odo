package tasks

import (
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha2"
	corev1 "k8s.io/api/core/v1"
)

func generateGithubStatusTask(secretName, ns string) pipelinev1.Task {
	task := pipelinev1.Task{
		TypeMeta:   createTaskTypeMeta(),
		ObjectMeta: createTaskObjectMeta("create-github-status-task", ns),
		Spec: pipelinev1.TaskSpec{
			Inputs: createInputsForGithubStatusTask(),
			TaskSpec: v1alpha2.TaskSpec{
				Steps: createStepsForGithubStatusTask(secretName),
			},
		},
	}
	return task
}

var argsForStartStatusStep = []string{
	"create-status",
	"--repo",
	"$(inputs.params.REPO)",
	"--sha",
	"$(inputs.params.COMMIT_SHA)",
	"--state",
	"$(inputs.params.STATE)",
	"--target-url",
	"$(inputs.params.TARGET_URL)",
	"--description",
	"$(inputs.params.DESCRIPTION)",
	"--context",
	"$(inputs.params.CONTEXT)",
}

func createInputsForGithubStatusTask() *pipelinev1.Inputs {
	return &pipelinev1.Inputs{
		Params: []pipelinev1.ParamSpec{
			createTaskParam(
				"REPO",
				"The repo to publish the status update for e.g. tektoncd/triggers",
				pipelinev1.ParamTypeString,
			),
			createTaskParam(
				"COMMIT_SHA",
				"The specific commit to report a status for.",
				pipelinev1.ParamTypeString,
			),
			createTaskParam(
				"STATE",
				"The state to report error, failure, pending, or success.",
				pipelinev1.ParamTypeString,
			),
			createTaskParamWithDefault(
				"TARGET_URL",
				"The target URL to associate with this status.",
				pipelinev1.ParamTypeString,
				"",
			),
			createTaskParam(
				"DESCRIPTION",
				"A short description of the status.",
				pipelinev1.ParamTypeString,
			),
			createTaskParam(
				"CONTEXT",
				"A string label to differentiate this status from the status of other systems.",
				pipelinev1.ParamTypeString,
			),
		},
	}
}

func createStepsForGithubStatusTask(secretName string) []pipelinev1.Step {
	return []pipelinev1.Step{
		pipelinev1.Step{
			Container: corev1.Container{
				Name:       "start-status",
				Image:      "quay.io/kmcdermo/github-tool:latest",
				WorkingDir: "/workspace/source",
				Env: []corev1.EnvVar{
					createEnvFromSecret("GITHUB_TOKEN", secretName, "token"),
				},
				Command: []string{"github-tools"},
				Args:    argsForStartStatusStep,
			},
		},
	}
}
