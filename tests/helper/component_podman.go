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
