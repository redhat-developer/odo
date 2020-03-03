package tasks

import (
	"github.com/openshift/odo/pkg/pipelines/meta"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha2"
	corev1 "k8s.io/api/core/v1"
)

var (
	buildCommands = []string{
		"buildah", "bud", "--tls-verify=$(inputs.params.TLSVERIFY)", "--layers", "-f", "$(inputs.params.DOCKERFILE)", "-t", "$(outputs.resources.image.url)", ".",
	}
	pushCommands = []string{
		"buildah", "push", "--tls-verify=$(inputs.params.TLSVERIFY)", "$(outputs.resources.image.url)", "docker://$(outputs.resources.image.url)",
	}
)

func generateBuildahTask(ns string, usingInternalRegistry bool) pipelinev1.Task {
	return pipelinev1.Task{
		TypeMeta:   createTaskTypeMeta(),
		ObjectMeta: meta.CreateObjectMeta(ns, "buildah-task"),
		Spec: pipelinev1.TaskSpec{
			Inputs:  createInputsForBuildah(usingInternalRegistry),
			Outputs: createOutputsForBuildah(),
			TaskSpec: v1alpha2.TaskSpec{
				Steps:   createStepsForBuildah(),
				Volumes: createVolumes("varlibcontainers"),
			},
		},
	}
}

func createVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		corev1.VolumeMount{
			Name:      "varlibcontainers",
			MountPath: "/var/lib/containers",
		},
	}
}

func createVolumes(name string) []corev1.Volume {
	return []corev1.Volume{
		corev1.Volume{
			Name: name,
		},
	}
}

func createSecurityContext(buildPrivilege bool) *corev1.SecurityContext {
	return &corev1.SecurityContext{
		Privileged: &buildPrivilege,
	}
}

func createOutputsForBuildah() *pipelinev1.Outputs {
	return &pipelinev1.Outputs{
		Resources: []pipelinev1.TaskResource{
			createTaskResource("image", "image"),
		},
	}
}

// Returns string value for TLSVERIFY parameter based on usingInternalRegistry boolean
// If internal registry is used, we need to disable TLS verification
func getTLSVerify(usingInternalRegistry bool) string {
	if usingInternalRegistry {
		return "false"
	}
	return "true"
}

func createInputsForBuildah(usingInternalRegistry bool) *pipelinev1.Inputs {

	return &pipelinev1.Inputs{
		Params: []pipelinev1.ParamSpec{
			createTaskParamWithDefault(
				"BUILDER_IMAGE",
				"The location of the buildah builder image.",
				pipelinev1.ParamTypeString,
				"quay.io/buildah/stable:v1.11.3",
			),
			createTaskParamWithDefault(
				"DOCKERFILE",
				"Path to the Dockerfile to build.",
				pipelinev1.ParamTypeString,
				"./Dockerfile",
			),
			createTaskParamWithDefault(
				"TLSVERIFY",
				"Verify the TLS on the registry endpoint (for push/pull to a non-TLS registry)",
				pipelinev1.ParamTypeString,
				getTLSVerify(usingInternalRegistry),
			),
		},
		Resources: []pipelinev1.TaskResource{
			createTaskResource("source", "git"),
		},
	}
}

func createStepsForBuildah() []pipelinev1.Step {
	return []pipelinev1.Step{
		pipelinev1.Step{
			Container: corev1.Container{
				Name:            "build",
				Image:           "$(inputs.params.BUILDER_IMAGE)",
				WorkingDir:      "/workspace/source",
				Command:         buildCommands,
				VolumeMounts:    createVolumeMounts(),
				SecurityContext: createSecurityContext(true),
			},
		},
		pipelinev1.Step{
			Container: corev1.Container{
				Name:            "push",
				Image:           "$(inputs.params.BUILDER_IMAGE)",
				WorkingDir:      "/workspace/source",
				Command:         pushCommands,
				VolumeMounts:    createVolumeMounts(),
				SecurityContext: createSecurityContext(true),
			},
		},
	}
}
