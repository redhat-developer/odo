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

// ExecContainer returns output after exec the command in the container
func (d *DockerRunner) ExecContainer(containerID, command string) string {
	stdOut := CmdShouldPass(d.path, "exec", containerID, "/bin/sh", "-c",
		command)
	return stdOut
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

// ListVolumesOfComponentAndType lists all volumes that match the expected component/type labels
func (d *DockerRunner) ListVolumesOfComponentAndType(componentLabel string, typeLabel string) []string {

	fmt.Fprintf(GinkgoWriter, "Listing volumes with component label %s and type label %s", componentLabel, typeLabel)
	session := CmdRunner(d.path, "volume", "ls", "-q", "--filter", "label=component="+componentLabel, "--filter", "label=type="+typeLabel)

	session.Wait()
	if session.ExitCode() == 0 {

		volumes := strings.Fields(strings.TrimSpace(string(session.Out.Contents())))

		return volumes
	}
	return []string{}
}

// RemoveVolumesByComponentAndType removes any volumes that match specified component and type labels
func (d *DockerRunner) RemoveVolumesByComponentAndType(componentLabel string, typeLabel string) string {

	volumes := d.ListVolumesOfComponentAndType(componentLabel, typeLabel)

	if len(volumes) == 0 {
		return ""
	}
	fmt.Fprintf(GinkgoWriter, "Removing volumes with component label %s and type label %s", componentLabel, typeLabel)

	output := ""

	for _, volume := range volumes {

		fmt.Fprintf(GinkgoWriter, "Removing volume with ID %s", volume)

		session := CmdRunner(d.path, "volume", "rm", "-f", volume)
		session.Wait()

		sessionOut := strings.TrimSpace(string(session.Out.Contents()))

		if session.ExitCode() == 0 {
			output += sessionOut + " "
		} else {
			fmt.Fprintf(GinkgoWriter, "Non-zero error code on removing volume with component label %s and type label %s, output: %s", componentLabel, typeLabel, sessionOut)
		}

	}

	return output
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
