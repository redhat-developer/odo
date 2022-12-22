package helper

import (
	. "github.com/onsi/ginkgo/v2"
)

// Component is an abstraction for a Devfile Component deployed on a specific platform
type Component interface {
	// ExpectIsDeployed checks that the component is deployed
	ExpectIsDeployed()
	// ExpectIsNotDeployed checks that the component is not deployed
	ExpectIsNotDeployed()
	// Exec executes the command in specific container of the component
	Exec(container string, args ...string) string
	// GetEnvVars returns the environment variables defined for the component
	GetEnvVars() map[string]string
}

func NewComponent(name string, app string, namespace string, cli CliRunner) Component {
	if NeedsCluster(CurrentSpecReport().Labels()) {
		return NewClusterComponent(name, app, namespace, cli)
	} else {
		return NewPodmanComponent(name, app)
	}
}
