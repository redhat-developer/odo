package helper

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	"github.com/redhat-developer/odo/pkg/labels"
)

// ClusterComponent is an abstraction for a Devfile Component deployed on a cluster (either Kubernetes or OpenShift)
type ClusterComponent struct {
	name      string
	app       string
	mode      string
	namespace string
	cli       CliRunner
}

func NewClusterComponent(name string, app string, mode string, namespace string, cli CliRunner) *ClusterComponent {
	return &ClusterComponent{
		name:      name,
		app:       app,
		mode:      mode,
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

func (o *ClusterComponent) Exec(container string, args []string, expectedSuccess *bool) (string, string) {
	podName := o.cli.GetRunningPodNameByComponent(o.name, o.namespace)
	return o.cli.Exec(podName, o.namespace, append([]string{"-c", container, "--"}, args...), expectedSuccess)
}

func (o *ClusterComponent) GetEnvVars(string) map[string]string {
	return o.cli.GetEnvsDevFileDeployment(o.name, o.app, o.namespace)
}

func (o *ClusterComponent) GetLabels() map[string]string {
	selector := labels.Builder().WithComponentName(o.name).WithAppName(o.app).WithMode(o.mode).SelectorFlag()
	stdout := o.cli.Run("get", "deployment", selector, "-n", o.namespace, "-o", "jsonpath={.items[0].metadata.labels}").Out.Contents()

	var result map[string]string
	err := json.Unmarshal(stdout, &result)
	Expect(err).ToNot(HaveOccurred())

	return result
}

func (o *ClusterComponent) GetPodDef() *corev1.Pod {
	var podDef corev1.Pod
	podName := o.cli.GetRunningPodNameByComponent(o.name, o.namespace)
	bufferOutput := o.cli.Run("get", "pods", podName, "-o", "json").Out.Contents()
	err := json.Unmarshal(bufferOutput, &podDef)
	Expect(err).ToNot(HaveOccurred())
	return &podDef
}

func (o *ClusterComponent) GetPodLogs() string {
	podName := o.cli.GetRunningPodNameByComponent(o.name, o.namespace)
	return string(o.cli.Run("-n", o.namespace, "logs", podName).Out.Contents())
}
