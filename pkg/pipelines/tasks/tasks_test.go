package tasks

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testNS = "testing-ns"

func TestDeployFromSourceTask(t *testing.T) {
	wantedTask := pipelinev1.Task{
		TypeMeta: taskTypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name:      "deploy-from-source-task",
			Namespace: testNS,
		},
		Spec: pipelinev1.TaskSpec{
			Params:    paramsForDeploymentFromSourceTask("test"),
			Resources: createResourcesForDeployFromSourceTask(),
			Steps: []pipelinev1.Step{
				{
					Container: corev1.Container{
						Name:       "run-kubectl",
						Image:      "quay.io/redhat-developer/k8s-kubectl",
						WorkingDir: "/workspace/source",
						Command:    []string{"kubectl"},
						Args:       argsForRunKubectlStep,
					},
				},
			},
		},
	}
	deployFromSourceTask := CreateDeployFromSourceTask(testNS, "test")
	if diff := cmp.Diff(wantedTask, deployFromSourceTask); diff != "" {
		t.Fatalf("CreateDeployFromSourceTask() failed \n%s", diff)
	}
}

func TestDeployUsingKubectlTask(t *testing.T) {
	validTask := pipelinev1.Task{
		TypeMeta: taskTypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name:      "deploy-using-kubectl-task",
			Namespace: testNS,
		},
		Spec: pipelinev1.TaskSpec{
			Params: []pipelinev1.ParamSpec{
				{
					Name:        "PATHTODEPLOYMENT",
					Type:        "string",
					Description: "Path to the pipelines to apply",
					Default:     &pipelinev1.ArrayOrString{Type: "string", StringVal: "deploy"},
				},
				{Name: "NAMESPACE", Type: "string", Description: "Namespace to deploy into"},
				{
					Name:        "DRYRUN",
					Type:        "string",
					Description: "If true run a server-side dryrun.",
					Default:     &pipelinev1.ArrayOrString{Type: "string", StringVal: "false"},
				},
				{
					Name:        "YAMLPATHTOIMAGE",
					Type:        "string",
					Description: "The path to the image to replace in the yaml pipelines (arg to yq)",
				},
			},
			Resources: &pipelinev1.TaskResources{
				Inputs: []pipelinev1.TaskResource{
					{ResourceDeclaration: pipelinev1.ResourceDeclaration{Name: "source", Type: "git"}},
					{ResourceDeclaration: pipelinev1.ResourceDeclaration{Name: "image", Type: "image"}},
				},
			},

			Steps: []pipelinev1.Step{
				{
					Container: corev1.Container{
						Name:       "replace-image",
						Image:      "quay.io/redhat-developer/yq",
						WorkingDir: "/workspace/source",
						Command:    []string{"yq"},
						Args:       argsForReplaceImageStep,
					},
				},
				{
					Container: corev1.Container{
						Name:       "run-kubectl",
						Image:      "quay.io/redhat-developer/k8s-kubectl",
						WorkingDir: "/workspace/source",
						Command: []string{
							"kubectl",
						},
						Args: argsForKubectlStep,
					},
				},
			},
		},
	}
	task := CreateDeployUsingKubectlTask(testNS)
	if diff := cmp.Diff(validTask, task); diff != "" {
		t.Fatalf("GenerateDeployUsingKubectlTask() failed:\n%s", diff)
	}
}

func TestCreateTaskParamWithDefault(t *testing.T) {
	validTaskParam := pipelinev1.ParamSpec{
		Name:        "sample",
		Type:        pipelinev1.ParamTypeString,
		Description: "sample",
		Default: &pipelinev1.ArrayOrString{
			StringVal: "sample",
			Type:      "string",
		},
	}
	taskParam := createTaskParamWithDefault("sample", "sample", pipelinev1.ParamTypeString, "sample")
	if diff := cmp.Diff(validTaskParam, taskParam); diff != "" {
		t.Fatalf("createTaskParamWithDefault() failed:\n%s", diff)
	}
}

func TestCreateTaskParam(t *testing.T) {
	validTaskParam := pipelinev1.ParamSpec{
		Name:        "sample",
		Type:        pipelinev1.ParamTypeString,
		Description: "sample",
	}
	taskParam := createTaskParam("sample", "sample", pipelinev1.ParamTypeString)
	if diff := cmp.Diff(validTaskParam, taskParam); diff != "" {
		t.Fatalf("createTaskParam() failed:\n%s", diff)
	}
}

func TestCreateContainer(t *testing.T) {
	validContainer := corev1.Container{
		Name:       "sampleName",
		Image:      "sampleImage",
		WorkingDir: "sampleDir",
		Command:    []string{"sample"},
		Args:       []string{"sample"},
	}
	container := createContainer("sampleName", "sampleImage", "sampleDir", []string{"sample"}, []string{"sample"})
	if diff := cmp.Diff(validContainer, container); diff != "" {
		t.Fatalf("createContainer() failed:\n%s", diff)
	}
}

func TestCreateTaskResource(t *testing.T) {
	validTaskResource := pipelinev1.TaskResource{
		ResourceDeclaration: pipelinev1.ResourceDeclaration{
			Name: "sample",
			Type: "git",
		},
	}
	taskResource := createTaskResource("sample", "git")
	if diff := cmp.Diff(validTaskResource, taskResource); diff != "" {
		t.Fatalf("createTaskResource() failed:\n%s", diff)
	}
}
