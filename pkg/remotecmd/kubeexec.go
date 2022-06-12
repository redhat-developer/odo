package remotecmd

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/task"
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
	klog.V(4).Infof("GetProcessInfoForCommand for %q", devfileCmd.Id)

	pid, err := getRemoteProcessPID(kclient, devfileCmd, podName, containerName)
	if err != nil {
		return RemoteProcessInfo{}, err
	}

	return k.getProcessInfoFromPid(pid, kclient, podName, containerName)
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
	klog.V(4).Infof("StartProcessForCommand for %q", devfileCmd.Id)

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
	// Storing the /bin/sh parent process PID. It will allow to determine its children later on and kill them when a stop is request
	pidWriterCmd := fmt.Sprintf("echo $$ > %s", pidFile)
	// Redirecting to /proc/1/fd/* allows to redirect the process output to the output streams of PID 1 process inside the container.
	// This way, returning the container logs with 'odo logs' or 'kubectl logs' would work seamlessly.
	// See https://stackoverflow.com/questions/58716574/where-exactly-do-the-logs-of-kubernetes-pods-come-from-at-the-container-level
	outputRedirectCmd := "1>>/proc/1/fd/1 2>>/proc/1/fd/2"
	if devfileCmd.Exec.WorkingDir != "" {
		// since we are using /bin/sh -c, the command needs to be within a single double quote instance,
		// for example "cd /tmp && pwd"
		// Full command is: /bin/sh -c "echo $$ > $pidFile && cd $workingDir && ($cmdLine) 1>>/proc/1/fd/1 2>>/proc/1/fd/2"
		// We are deleting the PID file if the main command does not succeed, so its status could be detected accordingly.
		cmd = append(cmd, fmt.Sprintf("%s && cd %s && (%s) %s",
			pidWriterCmd, devfileCmd.Exec.WorkingDir, cmdLine, outputRedirectCmd))
	} else {
		// Full command is: /bin/sh -c "echo $$ > $pidFile && ($cmdLine) 1>>/proc/1/fd/1 2>>/proc/1/fd/2"
		// We are deleting the PID file if the main command does not succeed, so its status could be detected accordingly.
		cmd = append(cmd, fmt.Sprintf("%s && (%s) %s", pidWriterCmd, cmdLine, outputRedirectCmd))
	}

	go func() {
		if outputHandler != nil {
			outputHandler(Starting, nil, nil, nil)
		}

		stdout, stderr, err := ExecuteCommandAndGetOutput(kclient, podName, containerName, false, cmd...)
		if err != nil {
			klog.V(2).Infof("error while running background command: %v", err)
		}

		if outputHandler != nil {
			outputHandler(Stopped, stdout, stderr, err)
		}
	}()

	return nil
}

// StopProcessForCommand stops the process representing the specified Devfile command.
// Because of the way this process is launched and its PID stored (see StartProcessForCommand),
// we need to determine the process children (there should be only one child). Then killing those children
// will exit the parent 'sh' process.
func (k *kubeExecProcessHandler) StopProcessForCommand(
	devfileCmd devfilev1.Command,
	kclient kclient.ClientInterface,
	podName string,
	containerName string,
) error {
	klog.V(4).Infof("StopProcessForCommand for %q", devfileCmd.Id)
	defer func() {
		pidFile := getPidFileForCommand(devfileCmd)
		err := ExecuteCommand([]string{common.ShellExecutable, "-c", fmt.Sprintf("rm -f %s", pidFile)},
			kclient, podName, containerName, false, nil, nil)
		if err != nil {
			klog.V(2).Infof("Could not remove file %q: %v", pidFile, err)
		}
	}()

	ppid, err := getRemoteProcessPID(kclient, devfileCmd, podName, containerName)
	if err != nil {
		return err
	}
	if ppid == 0 {
		return nil
	}

	children, err := getProcessChildren(ppid, kclient, podName, containerName)
	if err != nil {
		return err
	}

	kill := func(p int) error {
		err = ExecuteCommand([]string{common.ShellExecutable, "-c", fmt.Sprintf("kill %d || true", p)},
			kclient, podName, containerName, false, nil, nil)

		//Because the process we just stopped might take longer to exit (it might have caught the signal and is performing additional cleanup),
		//retry detecting its actual state till it is stopped or timeout expires
		var processInfo interface{}
		processInfo, err = task.NewRetryable(fmt.Sprintf("status for remote process %d", p), func() (bool, interface{}, error) {
			pInfo, e := k.getProcessInfoFromPid(p, kclient, podName, containerName)
			return e == nil || pInfo.Status == Stopped, pInfo, e
		}, true).RetryWithSchedule(2*time.Second, 4*time.Second, 8*time.Second)
		if err != nil {
			return err
		}

		pInfo, ok := processInfo.(RemoteProcessInfo)
		if !ok {
			klog.V(2).Infof("invalid type for remote process (%d) info, expected RemoteProcessInfo", p)
			return fmt.Errorf("internal error while checking remote process status: %d", p)
		}
		if pInfo.Status != Stopped {
			return fmt.Errorf("invalid status for remote process %d: %+v", p, processInfo)
		}
		return nil
	}

	if len(children) == 0 {
		return kill(ppid)
	}

	for _, child := range children {
		if err = kill(child); err != nil {
			return err
		}
	}

	return nil
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

func (k *kubeExecProcessHandler) getProcessInfoFromPid(
	pid int,
	kclient kclient.ClientInterface,
	podName string,
	containerName string,
) (RemoteProcessInfo, error) {
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

// getProcessChildren returns the children of the specified process in the given container.
// It works by reading the /proc/<pid>/task/<pid>/children file, which is a space-separated list of children
func getProcessChildren(pid int, kclient kclient.ClientInterface, podName string, containerName string) ([]int, error) {
	if pid <= 0 {
		return nil, fmt.Errorf("invalid pid: %d", pid)
	}

	stdout, _, err := ExecuteCommandAndGetOutput(kclient, podName, containerName, false,
		common.ShellExecutable, "-c", fmt.Sprintf("cat /proc/%[1]d/task/%[1]d/children || true", pid))
	if err != nil {
		return nil, err
	}

	var children []int
	for _, line := range stdout {
		l := strings.Split(strings.TrimSpace(line), " ")
		for _, p := range l {
			c, err := strconv.Atoi(p)
			if err != nil {
				return children, err
			}
			children = append(children, c)
		}
	}

	return children, nil
}

// getPidFileForCommand returns the path to the PID file in the remote container.
// The parent folder is supposed to be existing, because it should be mounted in the container using the mandatory
// shared volume (more info in the AddOdoMandatoryVolume function from the utils package).
func getPidFileForCommand(devfileCmd devfilev1.Command) string {
	return fmt.Sprintf("%s/.odo_devfile_cmd_%s.pid", strings.TrimSuffix(storage.SharedDataMountPath, "/"), devfileCmd.Id)
}
