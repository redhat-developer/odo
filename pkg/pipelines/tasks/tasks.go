package tasks

import (
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	corev1 "k8s.io/api/core/v1"

	"github.com/openshift/odo/pkg/pipelines/meta"
)

var (
	taskTypeMeta = meta.TypeMeta("Task", "tekton.dev/v1alpha1")
)

func createTaskResource(name string, resourceType string) pipelinev1.TaskResource {
	return pipelinev1.TaskResource{
		ResourceDeclaration: pipelinev1.ResourceDeclaration{
			Name: name,
			Type: resourceType,
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
