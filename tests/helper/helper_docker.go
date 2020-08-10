package helper

import (
	"encoding/json"
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
	fmt.Fprintf(GinkgoWriter, "Listing locally running Docker images\n")
	output := CmdShouldPass(d.path, "ps")
	return output
}

// GetRunningContainersByLabel lists all running images with the label (of the form "key=value")
func (d *DockerRunner) GetRunningContainersByLabel(label string) []string {
	fmt.Fprintf(GinkgoWriter, "Listing locally running Docker images with label %s\n", label)
	filterLabel := "label=" + label
	output := strings.TrimSpace(CmdShouldPass(d.path, "ps", "-q", "--filter", filterLabel))

	// Split the strings and remove any whitespace
	containers := strings.Fields(output)
	return containers
}

// GetRunningContainersByCompAlias returns the list of containers labeled with the specified component and alias
func (d *DockerRunner) GetRunningContainersByCompAlias(comp string, alias string) []string {
	fmt.Fprintf(GinkgoWriter, "Listing locally running Docker images with comp %s and alias %s\n", comp, alias)
	compLabel := "label=component=" + comp
	aliasLabel := "label=alias=" + alias
	output := strings.TrimSpace(CmdShouldPass(d.path, "ps", "-q", "--filter", compLabel, "--filter", aliasLabel))

	containers := strings.Fields(output)
	return containers
}

// ListVolumes lists all volumes on the cluster
func (d *DockerRunner) ListVolumes() []string {
	fmt.Fprintf(GinkgoWriter, "Listing locally running Docker volumes\n")
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

func (d *DockerRunner) GetEnvsDevFileDeployment(containerID, command string) map[string]string {
	var mapOutput = make(map[string]string)

	output := d.ExecContainer(containerID, command)

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimPrefix(line, "'")
		splits := strings.Split(line, "=")
		name := splits[0]
		value := strings.Join(splits[1:], "=")
		mapOutput[name] = value
	}
	return mapOutput
}

// GetVolumesByLabel returns a list of volumes with the label (of the form "key=value")
func (d *DockerRunner) GetVolumesByLabel(label string) []string {
	fmt.Fprintf(GinkgoWriter, "Listing Docker volumes with label %s\n", label)
	filterLabel := "label=" + label
	output := strings.TrimSpace(CmdShouldPass(d.path, "volume", "ls", "-q", "--filter", filterLabel))

	// Split the strings and remove any whitespace
	containers := strings.Fields(output)
	return containers
}

// VolumeExists returns true if a volume with the given name exists, false otherwise.
func (d *DockerRunner) VolumeExists(name string) bool {
	vols := d.ListVolumes()

	for _, vol := range vols {
		if vol == name {
			return true
		}
	}
	return false

}

// GetVolumesByCompStorageName returns the list of volumes associated with a specific devfile volume in a component
func (d *DockerRunner) GetVolumesByCompStorageName(component string, storageName string) []string {
	fmt.Fprintf(GinkgoWriter, "Listing Docker volumes with comp %s and storage name %s\n", component, storageName)
	compLabel := "label=component=" + component
	storageLabel := "label=storage-name=" + storageName
	output := strings.TrimSpace(CmdShouldPass(d.path, "volume", "ls", "-q", "--filter", compLabel, "--filter", storageLabel))

	// Split the strings and remove any whitespace
	containers := strings.Fields(output)
	return containers
}

// InspectVolume returns a map-representation of the JSON returned by the 'docker inspect volume' command
func (d *DockerRunner) InspectVolume(volumeName string) []map[string]interface{} {

	fmt.Fprintf(GinkgoWriter, "Inspecting volume %s\n", volumeName)
	output := CmdShouldPass(d.path, "inspect", volumeName)

	var result []map[string]interface{}
	err := json.Unmarshal([]byte(output), &result)
	Expect(err).NotTo(HaveOccurred())

	return result
}

// IsVolumeMountedInContainer returns true if the specified volume is mounted in the container associated with specified component and alias
func (d *DockerRunner) IsVolumeMountedInContainer(volumeName string, component string, alias string) bool {
	// Get the container ID of the specified component and alias
	containers := d.GetRunningContainersByCompAlias(component, alias)
	Expect(len(containers)).To(Equal(1))

	containerID := containers[0]

	mounts := CmdShouldPass(d.path, "inspect", containerID, "--format", "'{{ .Mounts }}'")
	return strings.Contains(mounts, volumeName)
}

// GetSourceAndStorageVolumesByComponent lists only the volumes that are associated with this component
// and contain either the 'type' or 'storage-name' fields.
func (d *DockerRunner) GetSourceAndStorageVolumesByComponent(componentLabel string) []string {

	result := []string{}

	volumeList := d.GetVolumesByLabel("component=" + componentLabel)
	if len(volumeList) == 0 {
		return result
	}

	fmt.Fprintf(GinkgoWriter, "Removing volumes with component label %s\n", componentLabel)

	for _, volumeName := range volumeList {

		// Only return volumes that contain the component label, and either 'type' or 'storage-name'
		volumeJSON := d.InspectVolume(volumeName)
		volumeLabels := (volumeJSON[0]["Labels"]).(map[string]interface{})

		match := false
		if typeValue, ok := volumeLabels["type"]; ok {
			match = match || typeValue == "projects"
		}
		if _, ok := volumeLabels["storage-name"]; ok {
			match = true
		}

		if match {
			result = append(result, volumeName)
		}
	}

	return result

}

// RemoveVolumeByName removes a specific volume by name
func (d *DockerRunner) RemoveVolumeByName(volumeName string) *gexec.Session {
	fmt.Fprintf(GinkgoWriter, "Removing volume with ID %s\n", volumeName)

	session := CmdRunner(d.path, "volume", "rm", "-f", volumeName)
	session.Wait()

	return session
}

// RemoveVolumesByComponent removes source/storage volumes that match specified component
func (d *DockerRunner) RemoveVolumesByComponent(componentLabel string) string {

	volumeList := d.GetSourceAndStorageVolumesByComponent(componentLabel)
	if len(volumeList) == 0 {
		return ""
	}

	fmt.Fprintf(GinkgoWriter, "Removing volumes with component label %s\n", componentLabel)

	output := ""

	for _, volumeName := range volumeList {

		session := d.RemoveVolumeByName(volumeName)

		sessionOut := strings.TrimSpace(string(session.Out.Contents()))

		if session.ExitCode() == 0 {
			output += sessionOut + " "
		} else {
			fmt.Fprintf(GinkgoWriter, "Non-zero error code on removing volume with component label %s, output: %s\n", componentLabel, sessionOut)
		}

	}

	return output
}

// StopContainers kills and stops all running containers with the specified label (such as component=nodejs)
func (d *DockerRunner) StopContainers(label string) {
	fmt.Fprintf(GinkgoWriter, "Removing locally running Docker images with label %s\n", label)

	// Get the container IDs matching the specified label
	containerIDs := d.GetRunningContainersByLabel(label)

	// Loop over the containers to remove and run `docker stop` for each of them
	// We have to loop because `docker stop` does not allow us to remove all of the containers at once (e.g. docker stop con1 con2 con3 ... is not allowed)
	for _, container := range containerIDs {
		CmdShouldPass(d.path, "stop", container)
	}

}

// CreateVolume creates an empty volume with the given name and labels
func (d *DockerRunner) CreateVolume(volumeName string, labels []string) {
	fmt.Fprintf(GinkgoWriter, "Creating volume %s with labels %v\n", volumeName, labels)

	args := []string{"volume", "create"}

	for _, label := range labels {
		args = append(args, "--label", label)
	}

	args = append(args, volumeName)

	CmdShouldPass(d.path, args...)

}
