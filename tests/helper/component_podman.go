package helper

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/gomega"
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
	fmt.Printf("exec %v\n", cmdargs)
	out, err := command.Output()
	Expect(err).ToNot(HaveOccurred())
	return string(out)
}
