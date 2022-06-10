package remotecmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/util"
)

// kubeExecProcessHandler implements RemoteProcessHandler by executing Devfile commands right away in the container
// (like a 'kubectl exec'-like approach). Command execution is done in the background, in a separate goroutine.
// It works by storing the parent process PID in a file (_startCmdProcessPidFile) in the container,
// then fires the exec command in the background.
// The goroutine started can then be stopped by killing the process stored in the state file (_startCmdProcessPidFile)
// in the container.
type kubeExecProcessHandler struct{}

//This allows to verify interface compliance at compile-time.
//See https://github.com/redhat-developer/odo/wiki/Dev:-Coding-Conventions#verify-interface-compliance
var _ RemoteProcessHandler = (*kubeExecProcessHandler)(nil)

func NewKubeExecProcessHandler() *kubeExecProcessHandler {
	return &kubeExecProcessHandler{}
}

// GetProcessInfoForCommand returns information about the process representing the given Devfile command.
// A PID of 0 denotes a Stopped process.
func (k *kubeExecProcessHandler) GetProcessInfoForCommand(
	devfileCmd devfilev1.Command,
	kclient kclient.ClientInterface,
	podName string,
	containerName string,
) (RemoteProcessInfo, error) {
	pid, err := getRemoteProcessPID(kclient, devfileCmd, podName, containerName)
	if err != nil {
		return RemoteProcessInfo{}, err
	}

	process := RemoteProcessInfo{Pid: pid}

	if pid < 0 {
		process.Status = Unknown
		return process, fmt.Errorf("invalid PID value for remote process: %d", pid)
	}
	if pid == 0 {
		process.Status = Stopped
		return process, nil
	}

	//Now check that the PID value is a valid process
	stdout, _, err := ExecuteCommandAndGetOutput(kclient, podName, containerName, false,
		common.ShellExecutable, "-c", fmt.Sprintf("kill -0 %d; echo $?", pid))

	if err != nil {
		process.Status = Unknown
		return process, err
	}

	var killStatus int
	killStatus, err = strconv.Atoi(strings.TrimSpace(stdout[0]))
	if err != nil {
		process.Status = Unknown
		return process, err
	}

	if killStatus == 0 {
		process.Status = Running
	} else {
		process.Status = Stopped
	}

	return process, nil
}

// StartProcessForCommand runs the (potentially never finishing) Devfile command in the background.
// The goroutine spawned here can get stopped either by stopping the parent process (e.g., 'odo dev'),
// or by stopping the underlying remote process by calling the StopProcessForCommand method.
func (k *kubeExecProcessHandler) StartProcessForCommand(
	devfileCmd devfilev1.Command,
	kclient kclient.ClientInterface,
	podName string,
	containerName string,
	outputHandler CommandOutputHandler,
) error {
	if devfileCmd.Exec == nil {
		return errors.New(" only Exec commands are supported")
	}

	// deal with environment variables
	cmdLine := devfileCmd.Exec.CommandLine
	setEnvVariable := util.GetCommandStringFromEnvs(devfileCmd.Exec.Env)
	if setEnvVariable != "" {
		cmdLine = setEnvVariable + " && " + devfileCmd.Exec.CommandLine
	}

	// Change to the workdir and execute the command
	pidFile := getPidFileForCommand(devfileCmd)
	cmd := []string{common.ShellExecutable, "-c"}
	pidWriterCmd := fmt.Sprintf("echo $$ > %s", pidFile)
	if devfileCmd.Exec.WorkingDir != "" {
		// since we are using /bin/sh -c, the command needs to be within a single double quote instance,
		// for example "cd /tmp && pwd"
		// Full command is: /bin/sh -c "echo $$ > $pidFile && cd $workingDir && $cmdLine"
		// We are deleting the PID file if the main command does not succeed, so its status could be detected accordingly.
		cmd = append(cmd, fmt.Sprintf("%s && cd %s && %s",
			pidWriterCmd, devfileCmd.Exec.WorkingDir, cmdLine))
	} else {
		// Full command is: /bin/sh -c "echo $$ > $pidFile && $cmdLine"
		// We are deleting the PID file if the main command does not succeed, so its status could be detected accordingly.
		cmd = append(cmd, fmt.Sprintf("%s && %s", pidWriterCmd, cmdLine))
	}

	go func() {
		_ = log.SpinnerNoSpin("Executing the application")
		stdout, stderr, err := ExecuteCommandAndGetOutput(kclient, podName, containerName, false, cmd...)
		if err != nil {
			klog.V(2).Infof("error while running background command: %v", err)
		}
		if outputHandler != nil {
			outputHandler(stdout, stderr, err)
		}
	}()

	return nil
}

func (k *kubeExecProcessHandler) StopProcessForCommand(
	devfileCmd devfilev1.Command,
	kclient kclient.ClientInterface,
	podName string,
	containerName string,
	outputHandler CommandOutputHandler,
) error {
	stdout, stderr, err := ExecuteCommandAndGetOutput(kclient, podName, containerName, false,
		common.ShellExecutable, "-c", fmt.Sprintf("kill $(cat %[1]s); rm -f %[1]s", getPidFileForCommand(devfileCmd)))
	if outputHandler != nil {
		outputHandler(stdout, stderr, err)
	}
	return err
}

func getRemoteProcessPID(kclient kclient.ClientInterface, devfileCmd devfilev1.Command, podName string, containerName string) (int, error) {
	pidFile := getPidFileForCommand(devfileCmd)
	stdout, stderr, err := ExecuteCommandAndGetOutput(kclient, podName, containerName, false,
		common.ShellExecutable, "-c", fmt.Sprintf("cat %s || true", pidFile))

	if err != nil {
		return 0, err
	}

	if len(stdout) == 0 {
		//The file does not exist. We assume the process has not run yet.
		return 0, nil
	}
	if len(stdout) != 1 {
		return 0, fmt.Errorf(
			"unable to determine command status, due to unexpected number of lines for %s, output: %v %v",
			pidFile, stdout, stderr)
	}

	line := stdout[0]
	var pid int
	pid, err = strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		klog.V(2).Infof("error while trying to retrieve status of command: %+v", err)
		return 0, fmt.Errorf("unable to determine command status, due to unexpected content for %s: %s",
			pidFile, line)
	}
	return pid, nil
}

// getPidFileForCommand returns the path to the PID file in the remote container.
// The parent folder is supposed to be existing, because it should be mounted in the container using the mandatory
// shared volume (more info in the AddOdoMandatoryVolume function from the utils package).
func getPidFileForCommand(devfileCmd devfilev1.Command) string {
	return fmt.Sprintf("%s/.odo_devfile_cmd_%s.pid", strings.TrimSuffix(storage.SharedDataMountPath, "/"), devfileCmd.Id)
}
