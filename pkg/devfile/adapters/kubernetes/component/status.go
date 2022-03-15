package component

import (
	"fmt"
	"strings"

	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/machineoutput"
)

// getSupervisordStatusInContainer executes 'supervisord ctl status' within the pod and container, parses the output,
// and returns the status for the container
func getSupervisordStatusInContainer(podName string, containerName string, a Adapter) []supervisordStatus {

	command := []string{common.SupervisordBinaryPath, common.SupervisordCtlSubCommand, "status"}
	compInfo := common.ComponentInfo{
		ContainerName: containerName,
		PodName:       podName,
	}

	stdoutWriter, stdoutOutputChannel := common.CreateConsoleOutputWriterAndChannel()
	stderrWriter, stderrOutputChannel := common.CreateConsoleOutputWriterAndChannel()

	err := common.ExecuteCommand(&a, compInfo, command, false, stdoutWriter, stderrWriter)

	// Close the writer and wait the console output
	stdoutWriter.Close()
	consoleResult := <-stdoutOutputChannel

	stderrWriter.Close()
	consoleStderrResult := <-stderrOutputChannel

	if err != nil {
		a.Logger().ReportError(fmt.Errorf("unable to execute command on %s within container %s, %v, output: %v %v: %w", podName, containerName, err, consoleResult, consoleStderrResult, err), machineoutput.TimestampNow())
		return nil
	}

	result := []supervisordStatus{}

	for _, line := range consoleResult {

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		result = append(result, supervisordStatus{program: fields[0], status: fields[1]})
	}

	return result
}

// supervisordStatus corresponds to the statuses reported by 'supervisord ctl status', example:
// - debugrun                         STOPPED
// - devrun                           RUNNING   pid 5640, uptime 11 days, 21:56:20
// Only the first and second fields are included (no pod, uptime, etc)
type supervisordStatus struct {
	program string
	status  string
}
