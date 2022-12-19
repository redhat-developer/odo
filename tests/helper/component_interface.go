package helper

import (
	. "github.com/onsi/ginkgo/v2"
)

// Component is an abstraction for a Devfile Component deployed on a specific platform
type Component interface {
	// ExpectIsNotDeployed checks that the component is not deployed
	ExpectIsNotDeployed()
}

func NewComponent(name string, app string, namespace string, cli CliRunner) Component {
	if NeedsCluster(CurrentSpecReport().Labels()) {
		return NewClusterComponent(name, app, namespace, cli)
	} else {
		return NewPodmanComponent(name, app)
	}
}
