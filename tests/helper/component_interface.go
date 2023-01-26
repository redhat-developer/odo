package helper

import (
	. "github.com/onsi/ginkgo/v2"
	corev1 "k8s.io/api/core/v1"
)

// Component is an abstraction for a Devfile Component deployed on a specific platform
type Component interface {
	// ExpectIsDeployed checks that the component is deployed
	ExpectIsDeployed()
	// ExpectIsNotDeployed checks that the component is not deployed
	ExpectIsNotDeployed()
	// Exec executes the command in specific container of the component.
	// If success is true, the command exit code is expected to be 0.
	// If success is false, the command exit code is expected to be non-zero.
	// If success is nil, the command is just supposed to run, with no assertion on its exit code.
	Exec(container string, success *bool, args ...string) (string, string)
	// GetEnvVars returns the environment variables defined for the container
	GetEnvVars(container string) map[string]string
	// GetLabels returns the labels defined for the component
	GetLabels() map[string]string
	// GetPodDef returns the definition of the pod
	GetPodDef() *corev1.Pod
	// GetPodLogs returns logs for the pod
	GetPodLogs() string
}

func NewComponent(componentName string, app string, mode string, namespace string, cli CliRunner) Component {
	if NeedsCluster(CurrentSpecReport().Labels()) {
		return NewClusterComponent(componentName, app, mode, namespace, cli)
	} else {
		return NewPodmanComponent(componentName, app)
	}
}
