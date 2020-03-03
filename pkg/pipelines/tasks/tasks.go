package tasks

import (
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Generate will return a slice of tasks
func Generate(secretName, ns string, usingInternalRegistry bool) []pipelinev1.Task {
	return []pipelinev1.Task{
		generateBuildahTask(ns, usingInternalRegistry),
		generateDeployFromSourceTask(ns),
		generateDeployUsingKubectlTask(ns),
	}
}

func createTaskTypeMeta() metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       "Task",
		APIVersion: "tekton.dev/v1alpha1",
	}
}

func createTaskResource(name string, resourceType string) pipelinev1.TaskResource {
	return pipelinev1.TaskResource{
		ResourceDeclaration: pipelinev1.ResourceDeclaration{
			Name: name,
			Type: resourceType,
		},
	}
}

func createEnvFromSecret(name string, localObjectReference string, key string) corev1.EnvVar {
	return corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: localObjectReference,
				},
				Key: key,
			},
		},
	}
}

func createTaskParam(name string, description string, paramType pipelinev1.ParamType) pipelinev1.ParamSpec {
	return pipelinev1.ParamSpec{
		Name:        name,
		Type:        paramType,
		Description: description,
	}
}

func createTaskParamWithDefault(name string, description string, paramType pipelinev1.ParamType, paramDefault string) pipelinev1.ParamSpec {
	return pipelinev1.ParamSpec{
		Name:        name,
		Type:        paramType,
		Description: description,
		Default: &pipelinev1.ArrayOrString{
			Type:      pipelinev1.ParamTypeString,
			StringVal: paramDefault,
		},
	}
}

func createContainer(name string, image string, workingDir string, cmd []string, args []string) corev1.Container {
	return corev1.Container{
		Name:       name,
		Image:      image,
		WorkingDir: workingDir,
		Command:    cmd,
		Args:       args,
	}
}
