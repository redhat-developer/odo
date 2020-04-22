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

// Run dpcler with given arguments
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

// GetRunningContainersByCompAlias returns the list of containers labeled with the specified component and alias
func (d *DockerRunner) GetRunningContainersByCompAlias(comp string, alias string) []string {
	fmt.Fprintf(GinkgoWriter, "Listing locally running Docker images with comp %s and alias %s", comp, alias)
	compLabel := "label=component=" + comp
	aliasLabel := "label=alias=" + alias
	output := strings.TrimSpace(CmdShouldPass(d.path, "ps", "-q", "--filter", compLabel, "--filter", aliasLabel))

	containers := strings.Fields(output)
	return containers
}

// ListVolumes lists all volumes on the cluster
func (d *DockerRunner) ListVolumes() []string {
	fmt.Fprintf(GinkgoWriter, "Listing locally running Docker volumes")
	output := CmdShouldPass(d.path, "volume", "ls", "-q")

	volumes := strings.Fields(output)
	return volumes
}

// GetVolumesByLabel returns a list of volumes with the label (of the form "key=value")
func (d *DockerRunner) GetVolumesByLabel(label string) []string {
	fmt.Fprintf(GinkgoWriter, "Listing Docker volumes with label %s", label)
	filterLabel := "label=" + label
	output := strings.TrimSpace(CmdShouldPass(d.path, "volume", "ls", "-q", "--filter", filterLabel))

	// Split the strings and remove any whitespace
	containers := strings.Fields(output)
	return containers
}

// GetVolumesByCompStorageName returns the list of volumes associated with a specific devfile volume in a component
func (d *DockerRunner) GetVolumesByCompStorageName(component string, storageName string) []string {
	fmt.Fprintf(GinkgoWriter, "Listing Docker volumes with comp %s and storage name %s", component, storageName)
	compLabel := "label=component=" + component
	storageLabel := "label=storage-name=" + storageName
	output := strings.TrimSpace(CmdShouldPass(d.path, "volume", "ls", "-q", "--filter", compLabel, "--filter", storageLabel))

	// Split the strings and remove any whitespace
	containers := strings.Fields(output)
	return containers
}

// IsVolumeMountedInContainer returns true if the specified volume is moutned in the container associated with specified component and alias
func (d *DockerRunner) IsVolumeMountedInContainer(volumeName string, component string, alias string) bool {
	// Get the container ID of the specified component and alias
	containers := d.GetRunningContainersByCompAlias(component, alias)
	Expect(len(containers)).To(Equal(1))

	containerID := containers[0]

	mounts := CmdShouldPass(d.path, "inspect", containerID, "--format", "'{{ .Mounts }}'")
	return strings.Contains(mounts, volumeName)
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
