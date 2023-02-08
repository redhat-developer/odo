package helper

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	jsonserializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/utils/pointer"

	"github.com/redhat-developer/odo/pkg/podman"
)

// PodmanComponent is an abstraction for a Devfile Component deployed on podman
type PodmanComponent struct {
	componentName string
	app           string
}

func NewPodmanComponent(componentName string, app string) *PodmanComponent {
	return &PodmanComponent{
		componentName: componentName,
		app:           app,
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

func (o *PodmanComponent) Exec(container string, args []string, expectedSuccess *bool) (string, string) {
	containerName := fmt.Sprintf("%s-%s-%s", o.componentName, o.app, container)
	cmdargs := []string{"exec", "--interactive"}
	cmdargs = append(cmdargs, "--tty")
	cmdargs = append(cmdargs, containerName)
	cmdargs = append(cmdargs, args...)

	command := exec.Command("podman", cmdargs...)
	out, err := command.CombinedOutput()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		fmt.Fprintln(GinkgoWriter, err)
	}
	if expectedSuccess != nil {
		if *expectedSuccess {
			Expect(err).ToNot(HaveOccurred())
		} else {
			Expect(err).Should(HaveOccurred())
		}
	}
	return string(out), ""
}

func (o *PodmanComponent) GetEnvVars(container string) map[string]string {
	envs, _ := o.Exec(container, []string{"env"}, pointer.Bool(true))
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

func (o *PodmanComponent) GetPodDef() *corev1.Pod {
	podname := fmt.Sprintf("%s-%s", o.componentName, o.app)
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

func (o *PodmanComponent) GetLabels() map[string]string {
	podName := fmt.Sprintf("%s-%s", o.componentName, o.app)
	cmd := exec.Command("podman", "pod", "inspect", podName, "--format", "json")
	stdout, err := cmd.Output()
	Expect(err).ToNot(HaveOccurred(), func() {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		fmt.Fprintln(GinkgoWriter, err)
	})

	var result podman.PodInspectData

	err = json.Unmarshal(stdout, &result)
	Expect(err).ToNot(HaveOccurred())

	return result.Labels
}

func (o *PodmanComponent) GetPodLogs() string {
	podName := fmt.Sprintf("%s-%s", o.componentName, o.app)
	cmd := exec.Command("podman", "pod", "logs", podName)
	stdout, err := cmd.Output()
	Expect(err).ToNot(HaveOccurred(), func() {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		fmt.Fprintln(GinkgoWriter, err)
	})
	return string(stdout)
}

func (o *PodmanComponent) ListImages() string {
	cmd := exec.Command("podman", "images", "--format", "{{.Repository}}:{{.Tag}}", "--noheading")
	stdout, err := cmd.Output()
	Expect(err).ToNot(HaveOccurred(), func() {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		fmt.Fprintln(GinkgoWriter, err)
	})
	return string(stdout)
}
