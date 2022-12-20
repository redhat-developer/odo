package helper

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	jsonserializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/kubectl/pkg/scheme"
)

// PodmanComponent is an abstraction for a Devfile Component deployed on podman
type PodmanComponent struct {
	name string
	app  string
}

func NewPodmanComponent(name string, app string) *PodmanComponent {
	return &PodmanComponent{
		name: name,
		app:  app,
	}
}

func (o *PodmanComponent) ExpectIsDeployed() {
	podName := fmt.Sprintf("%s-%s", o.name, o.app)
	cmd := exec.Command("podman", "pod", "list", "--format", "{{.Name}}", "--noheading")
	stdout, err := cmd.Output()
	Expect(err).ToNot(HaveOccurred())
	Expect(string(stdout)).To(ContainSubstring(podName))
}

func (o *PodmanComponent) ExpectIsNotDeployed() {
	podName := fmt.Sprintf("%s-%s", o.name, o.app)
	cmd := exec.Command("podman", "pod", "list", "--format", "{{.Name}}", "--noheading")
	stdout, err := cmd.Output()
	Expect(err).ToNot(HaveOccurred())
	Expect(string(stdout)).ToNot(ContainSubstring(podName))
}

func (o *PodmanComponent) Exec(container string, args ...string) string {
	containerName := fmt.Sprintf("%s-%s-%s", o.name, o.app, container)
	cmdargs := []string{"exec", "--interactive"}
	cmdargs = append(cmdargs, "--tty")
	cmdargs = append(cmdargs, containerName)
	cmdargs = append(cmdargs, args...)

	command := exec.Command("podman", cmdargs...)
	out, err := command.Output()
	Expect(err).ToNot(HaveOccurred())
	return string(out)
}

func (o *PodmanComponent) GetEnvVars() map[string]string {
	podName := fmt.Sprintf("%s-%s", o.name, o.app)
	podDef := getPodDef(podName)
	res := map[string]string{}
	for _, env := range podDef.Spec.Containers[0].Env {
		res[env.Name] = env.Value
	}
	return res
}

func getPodDef(podname string) *corev1.Pod {
	serializer := jsonserializer.NewSerializerWithOptions(
		jsonserializer.SimpleMetaFactory{},
		scheme.Scheme,
		scheme.Scheme,
		jsonserializer.SerializerOptions{
			Yaml: true,
		},
	)

	cmd := exec.Command("podman", "kube", "generate", podname)
	resultBytes, err := cmd.Output()
	Expect(err).ToNot(HaveOccurred())
	var pod corev1.Pod
	_, _, err = serializer.Decode(resultBytes, nil, &pod)
	Expect(err).ToNot(HaveOccurred())
	return &pod
}
