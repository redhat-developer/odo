package tasks

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testNS = "testing-ns"

func TestDeployFromSourceTask(t *testing.T) {
	wantedTask := pipelinev1.Task{
		TypeMeta: v1.TypeMeta{
			Kind:       "Task",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "deploy-from-source-task",
			Namespace: testNS,
		},
		Spec: pipelinev1.TaskSpec{
			Inputs: createInputsForDeployFromSourceTask(),
			TaskSpec: v1alpha2.TaskSpec{
				Steps: []pipelinev1.Step{
					pipelinev1.Step{
						Container: corev1.Container{
							Name:       "run-kubectl",
							Image:      "quay.io/kmcdermo/k8s-kubectl:latest",
							WorkingDir: "/workspace/source",
							Command:    []string{"kubectl"},
							Args:       argsForRunKubectlStep,
						},
					},
				},
			},
		},
	}
	deployFromSourceTask := generateDeployFromSourceTask(testNS)
	if diff := cmp.Diff(wantedTask, deployFromSourceTask); diff != "" {
		t.Fatalf("GenerateDeployFromSourceTask() failed \n%s", diff)
	}
}

func TestDeployUsingKubectlTask(t *testing.T) {
	validTask := pipelinev1.Task{
		TypeMeta: v1.TypeMeta{
			Kind:       "Task",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "deploy-using-kubectl-task",
			Namespace: testNS,
		},
		Spec: pipelinev1.TaskSpec{
			Inputs: createInputsForDeployKubectlTask(),
			TaskSpec: v1alpha2.TaskSpec{
				Steps: []pipelinev1.Step{
					pipelinev1.Step{
						Container: corev1.Container{
							Name:       "replace-image",
							Image:      "mikefarah/yq",
							WorkingDir: "/workspace/source",
							Command:    []string{"yq"},
							Args:       argsForReplaceImageStep,
						},
					},
					pipelinev1.Step{
						Container: corev1.Container{
							Name:       "run-kubectl",
							Image:      "quay.io/kmcdermo/k8s-kubectl:latest",
							WorkingDir: "/workspace/source",
							Command: []string{
								"kubectl",
							},
							Args: argsForKubectlStep,
						},
					},
				},
			},
		},
	}
	task := generateDeployUsingKubectlTask(testNS)
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

func TestCreateEnvFromSecret(t *testing.T) {
	validEnv := corev1.EnvVar{
		Name: "sampleName",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "sampleSec",
				},
				Key: "sampleKey",
			},
		},
	}
	env := createEnvFromSecret("sampleName", "sampleSec", "sampleKey")
	if diff := cmp.Diff(validEnv, env); diff != "" {
		t.Fatalf("createEnvFromSecret() failed:\n%s", diff)
	}
}

func TestGenerateBuildahTask(t *testing.T) {
	validTask := pipelinev1.Task{
		TypeMeta: v1.TypeMeta{
			Kind:       "Task",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "buildah-task",
		},
		Spec: pipelinev1.TaskSpec{
			Inputs:  createInputsForBuildah(false),
			Outputs: createOutputsForBuildah(),
			TaskSpec: v1alpha2.TaskSpec{
				Steps:   createStepsForBuildah(),
				Volumes: createVolumes("varlibcontainers"),
			},
		},
	}
	buildahTask := generateBuildahTask("", false)
	if diff := cmp.Diff(validTask, buildahTask); diff != "" {
		t.Fatalf("generateBuildahTask() failed:\n%s", diff)
	}
}

func TestCreateVolumes(t *testing.T) {
	validVolume := []corev1.Volume{
		corev1.Volume{
			Name: "sample",
		},
	}
	volume := createVolumes("sample")
	if diff := cmp.Diff(validVolume, volume); diff != "" {
		t.Fatalf("createVolumes() failed:\n%s", diff)
	}
}

func TestCreateSecurityContext(t *testing.T) {
	samplePrivilege := true
	validSecuirtyContext := &corev1.SecurityContext{
		Privileged: &samplePrivilege,
	}
	privilege := createSecurityContext(true)
	if diff := cmp.Diff(validSecuirtyContext, privilege); diff != "" {
		t.Fatalf("createSecurityContext() failed:\n%s", diff)
	}
}
