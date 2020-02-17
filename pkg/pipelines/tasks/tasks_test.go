package tasks

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGithubStatusTask(t *testing.T) {
	wantedTask := pipelinev1.Task{
		TypeMeta: v1.TypeMeta{
			Kind:       "Task",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "create-github-status-task",
		},
		Spec: pipelinev1.TaskSpec{
			Inputs: &pipelinev1.Inputs{
				Params: []pipelinev1.ParamSpec{
					pipelinev1.ParamSpec{
						Name:        "REPO",
						Type:        pipelinev1.ParamTypeString,
						Description: "The repo to publish the status update for e.g. tektoncd/triggers",
					},
					pipelinev1.ParamSpec{
						Name:        "COMMIT_SHA",
						Type:        pipelinev1.ParamTypeString,
						Description: "The specific commit to report a status for.",
					},
					pipelinev1.ParamSpec{
						Name:        "STATE",
						Type:        pipelinev1.ParamTypeString,
						Description: "The state to report error, failure, pending, or success.",
					},
					pipelinev1.ParamSpec{
						Name:        "TARGET_URL",
						Type:        pipelinev1.ParamTypeString,
						Description: "The target URL to associate with this status.",
						Default: &pipelinev1.ArrayOrString{
							StringVal: "",
							Type:      "string",
						},
					},
					pipelinev1.ParamSpec{
						Name:        "DESCRIPTION",
						Type:        pipelinev1.ParamTypeString,
						Description: "A short description of the status.",
					},
					pipelinev1.ParamSpec{
						Name:        "CONTEXT",
						Type:        pipelinev1.ParamTypeString,
						Description: "A string label to differentiate this status from the status of other systems.",
					},
				},
			},
			TaskSpec: v1alpha2.TaskSpec{
				Steps: []pipelinev1.Step{
					pipelinev1.Step{
						Container: corev1.Container{
							Name:       "start-status",
							Image:      "quay.io/kmcdermo/github-tool:latest",
							WorkingDir: "/workspace/source",
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name: "GITHUB_TOKEN",
									ValueFrom: &corev1.EnvVarSource{
										SecretKeyRef: &corev1.SecretKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: "github-auth",
											},
											Key: "token",
										},
									},
								},
							},
							Command: []string{"github-tools"},
							Args: []string{
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
							},
						},
					},
				},
			},
		},
	}

	githubStatusTask := generateGithubStatusTask()
	if diff := cmp.Diff(wantedTask, githubStatusTask); diff != "" {
		t.Fatalf("GenerateGithubStatusTask() failed:\n%s", diff)
	}
}

func TestDeployFromSourceTask(t *testing.T) {
	wantedTask := pipelinev1.Task{
		TypeMeta: v1.TypeMeta{
			Kind:       "Task",
			APIVersion: "tekton.dev/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "deploy-from-source-task",
		},
		Spec: pipelinev1.TaskSpec{
			Inputs: &pipelinev1.Inputs{
				Resources: []pipelinev1.TaskResource{
					pipelinev1.TaskResource{
						ResourceDeclaration: pipelinev1.ResourceDeclaration{
							Name: "source",
							Type: "git",
						},
					},
				},
				Params: []pipelinev1.ParamSpec{
					pipelinev1.ParamSpec{
						Name:        "PATHTODEPLOYMENT",
						Description: "Path to the manifest to apply",
						Type:        pipelinev1.ParamTypeString,
						Default: &pipelinev1.ArrayOrString{
							StringVal: "deploy",
							Type:      "string",
						},
					},
					pipelinev1.ParamSpec{
						Name:        "NAMESPACE",
						Type:        pipelinev1.ParamTypeString,
						Description: "Namespace to deploy into",
					},
					pipelinev1.ParamSpec{
						Name:        "DRYRUN",
						Type:        pipelinev1.ParamTypeString,
						Description: "If true run a server-side dryrun.",
						Default: &pipelinev1.ArrayOrString{
							StringVal: "false",
							Type:      "string",
						},
					},
				},
			},
			TaskSpec: v1alpha2.TaskSpec{
				Steps: []pipelinev1.Step{
					pipelinev1.Step{
						Container: corev1.Container{
							Name:       "run-kubectl",
							Image:      "quay.io/kmcdermo/k8s-kubectl:latest",
							WorkingDir: "/workspace/source",
							Command:    []string{"kubectl"},
							Args: []string{
								"apply",
								"--dry-run=$(inputs.params.DRYRUN)",
								"-n",
								"$(inputs.params.NAMESPACE)",
								"-k",
								"$(inputs.params.PATHTODEPLOYMENT)",
							},
						},
					},
				},
			},
		},
	}
	deployFromSourceTask := generateDeployFromSourceTask()
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
			Name: "deploy-using-kubectl-task",
		},
		Spec: pipelinev1.TaskSpec{
			Inputs: &pipelinev1.Inputs{
				Resources: []pipelinev1.TaskResource{
					pipelinev1.TaskResource{
						ResourceDeclaration: pipelinev1.ResourceDeclaration{
							Name: "source",
							Type: "git",
						},
					},
					pipelinev1.TaskResource{
						ResourceDeclaration: pipelinev1.ResourceDeclaration{
							Name: "image",
							Type: "image",
						},
					},
				},
				Params: []pipelinev1.ParamSpec{
					pipelinev1.ParamSpec{
						Name:        "PATHTODEPLOYMENT",
						Type:        "string",
						Description: "Path to the manifest to apply",
						Default: &pipelinev1.ArrayOrString{
							StringVal: "deploy",
							Type:      "string",
						},
					},
					pipelinev1.ParamSpec{
						Name:        "NAMESPACE",
						Type:        "string",
						Description: "Namespace to deploy into",
					},
					pipelinev1.ParamSpec{
						Name:        "DRYRUN",
						Type:        "string",
						Description: "If true run a server-side dryrun.",
						Default: &pipelinev1.ArrayOrString{
							StringVal: "false",
							Type:      "string",
						},
					},
					pipelinev1.ParamSpec{
						Name:        "YAMLPATHTOIMAGE",
						Type:        "string",
						Description: "The path to the image to replace in the yaml manifest (arg to yq)",
					},
				},
			},
			TaskSpec: v1alpha2.TaskSpec{
				Steps: []pipelinev1.Step{
					pipelinev1.Step{
						Container: corev1.Container{
							Name:       "replace-image",
							Image:      "mikefarah/yq",
							WorkingDir: "/workspace/source",
							Command:    []string{"yq"},
							Args: []string{
								"w",
								"-i",
								"$(inputs.params.PATHTODEPLOYMENT)/deployment.yaml",
								"$(inputs.params.YAMLPATHTOIMAGE)",
								"$(inputs.resources.image.url)",
							},
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
							Args: []string{
								"apply",
								"-n",
								"$(inputs.params.NAMESPACE)",
								"-k",
								"$(inputs.params.PATHTODEPLOYMENT)",
							},
						},
					},
				},
			},
		},
	}
	task := generateDeployUsingKubectlTask()
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
