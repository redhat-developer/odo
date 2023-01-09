package helper

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	jsonserializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/kubectl/pkg/scheme"
)

// PodmanComponent is an abstraction for a Devfile Component deployed on podman
type PodmanComponent struct {
	componentName string
	app           string
	containerName string
}

func NewPodmanComponent(componentName string, app string, containerName string) *PodmanComponent {
	return &PodmanComponent{
		componentName: componentName,
		app:           app,
		containerName: containerName,
	}
}

func (o *PodmanComponent) ExpectIsDeployed() {
	podName := fmt.Sprintf("%s-%s", o.componentName, o.app)
	cmd := exec.Command("podman", "pod", "list", "--format", "{{.Name}}", "--noheading")
	stdout, err := cmd.Output()
	Expect(err).ToNot(HaveOccurred())
	Expect(string(stdout)).To(ContainSubstring(podName))
}

func (o *PodmanComponent) ExpectIsNotDeployed() {
	podName := fmt.Sprintf("%s-%s", o.componentName, o.app)
	cmd := exec.Command("podman", "pod", "list", "--format", "{{.Name}}", "--noheading")
	stdout, err := cmd.Output()
	Expect(err).ToNot(HaveOccurred())
	Expect(string(stdout)).ToNot(ContainSubstring(podName))
}

func (o *PodmanComponent) Exec(container string, args ...string) string {
	containerName := fmt.Sprintf("%s-%s-%s", o.componentName, o.app, container)
	cmdargs := []string{"exec", "--interactive"}
	cmdargs = append(cmdargs, "--tty")
	cmdargs = append(cmdargs, containerName)
	cmdargs = append(cmdargs, args...)

	command := exec.Command("podman", cmdargs...)
	out, err := command.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
	}
	Expect(err).ToNot(HaveOccurred())
	return string(out)
}

func (o *PodmanComponent) GetEnvVars() map[string]string {
	envs := o.Exec(o.containerName, "env")
	return splitLines(envs)
}

func splitLines(str string) map[string]string {
	result := map[string]string{}
	sc := bufio.NewScanner(strings.NewReader(str))
	for sc.Scan() {
		line := sc.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) < 2 {
			continue
		}
		result[parts[0]] = parts[1]
	}
	return result
}

func GetPodDef(podname string) *corev1.Pod {
	serializer := jsonserializer.NewSerializerWithOptions(
		jsonserializer.SimpleMetaFactory{},
		scheme.Scheme,
		scheme.Scheme,
		jsonserializer.SerializerOptions{
			Yaml: true,
		},
	)

	cmd := exec.Command("podman", "generate", "kube", podname)
	resultBytes, err := cmd.Output()
	Expect(err).ToNot(HaveOccurred())
	var pod corev1.Pod
	_, _, err = serializer.Decode(resultBytes, nil, &pod)
	Expect(err).ToNot(HaveOccurred())
	return &pod
}
