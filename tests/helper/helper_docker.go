package helper

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

type DockerRunner struct {
	// path to docker binary
	path string
}

// NewDockerRunner initializes new DockerRunner
func NewDockerRunner(ocPath string) DockerRunner {
	return DockerRunner{
		path: ocPath,
	}
}

// Run dpcler with given arguments
func (d *DockerRunner) Run(cmd string) *gexec.Session {
	session := CmdRunner(cmd)
	Eventually(session).Should(gexec.Exit(0))
	return session
}

// ListRunningImages runs 'docker ps' to list all running images
func (d *DockerRunner) ListRunningImages() string {
	fmt.Fprintf(GinkgoWriter, "Listing locally running Docker images")
	output := CmdShouldPass(d.path, "ps")

	return output
}

// GetRunningImagesByLabel lists all running images with the label (of the form "key=value")
func (d *DockerRunner) GetRunningImagesByLabel(label string) string {
	fmt.Fprintf(GinkgoWriter, "Listing locally running Docker images with label %s", label)
	output := CmdShouldPass(d.path, "ps", "-f", "label=\"", label, "\"")
	return output
}

// ListVolumes lists all volumes on the cluster
func (d *DockerRunner) ListVolumes() string {
	session := CmdRunner(d.path, "volume", "ls", "-q")
	session.Wait()
	if session.ExitCode() == 0 {
		return strings.TrimSpace(string(session.Out.Contents()))
	}
	return ""
}
