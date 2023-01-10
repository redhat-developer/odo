package helper

import (
	"fmt"

	. "github.com/onsi/gomega"
)

// ClusterComponent is an abstraction for a Devfile Component deployed on a cluster (either Kubernetes or OpenShift)
type ClusterComponent struct {
	name      string
	app       string
	namespace string
	cli       CliRunner
}

func NewClusterComponent(name string, app string, namespace string, cli CliRunner) *ClusterComponent {
	return &ClusterComponent{
		name:      name,
		app:       app,
		namespace: namespace,
		cli:       cli,
	}
}

func (o *ClusterComponent) ExpectIsDeployed() {
	deploymentName := fmt.Sprintf("%s-%s", o.name, o.app)
	stdout := o.cli.Run("get", "deployment", "-n", o.namespace).Out.Contents()
	Expect(string(stdout)).To(ContainSubstring(deploymentName))
}

func (o *ClusterComponent) ExpectIsNotDeployed() {
	deploymentName := fmt.Sprintf("%s-%s", o.name, o.app)
	stdout := o.cli.Run("get", "deployment", "-n", o.namespace).Out.Contents()
	Expect(string(stdout)).To(Not(ContainSubstring(deploymentName)))
}

func (o *ClusterComponent) Exec(container string, args ...string) string {
	podName := o.cli.GetRunningPodNameByComponent(o.name, o.namespace)
	return o.cli.Exec(podName, o.namespace, append([]string{"-c", container, "--"}, args...)...)
}

func (o *ClusterComponent) GetEnvVars(string) map[string]string {
	return o.cli.GetEnvsDevFileDeployment(o.name, o.app, o.namespace)
}
