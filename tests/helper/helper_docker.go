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
func NewDockerRunner(dockerPath string) DockerRunner {
	return DockerRunner{
		path: dockerPath,
	}
}

// Run docker with given arguments
func (d *DockerRunner) Run(cmd string) *gexec.Session {
	session := CmdRunner(cmd)
	Eventually(session).Should(gexec.Exit(0))
	return session
}

// ListRunningContainers runs 'docker ps' to list all running images
func (d *DockerRunner) ListRunningContainers() string {
	fmt.Fprintf(GinkgoWriter, "Listing locally running Docker images")
	output := CmdShouldPass(d.path, "ps")
	return output
}

// GetRunningContainersByLabel lists all running images with the label (of the form "key=value")
func (d *DockerRunner) GetRunningContainersByLabel(label string) []string {
	fmt.Fprintf(GinkgoWriter, "Listing locally running Docker images with label %s", label)
	filterLabel := "label=" + label
	output := strings.TrimSpace(CmdShouldPass(d.path, "ps", "-q", "--filter", filterLabel))

	// Split the strings and remove any whitespace
	containers := strings.Fields(output)
	return containers
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

// StopContainers kills and stops all running containers with the specified label (such as component=nodejs)
func (d *DockerRunner) StopContainers(label string) {
	fmt.Fprintf(GinkgoWriter, "Removing locally running Docker images with label %s", label)

	// Get the container IDs matching the specified label
	containerIDs := d.GetRunningContainersByLabel(label)

	// Loop over the containers to remove and run `docker stop` for each of them
	// We have to loop because `docker stop` does not allow us to remove all of the containers at once (e.g. docker stop con1 con2 con3 ... is not allowed)
	for _, container := range containerIDs {
		CmdShouldPass(d.path, "stop", container)
	}

}
