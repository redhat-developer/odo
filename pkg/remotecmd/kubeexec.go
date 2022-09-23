package remotecmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/remotecmd/kube"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/task"
)

// kubeExecProcessHandler implements RemoteProcessHandler by executing Devfile commands right away in the container
// (like a 'kubectl exec'-like approach). Command execution is done in the background, in a separate goroutine.
// It works by storing the parent process PID in a file in the container,
// then fires the exec command in the background.
// The goroutine started can then be stopped by killing the process stored in the state file (_startCmdProcessPidFile)
// in the container.
type kubeExecProcessHandler struct{}

// This allows to verify interface compliance at compile-time.
// See https://github.com/redhat-developer/odo/wiki/Dev:-Coding-Conventions#verify-interface-compliance
var _ RemoteProcessHandler = (*kubeExecProcessHandler)(nil)

func NewKubeExecProcessHandler() *kubeExecProcessHandler {
	return &kubeExecProcessHandler{}
}

// GetProcessInfoForCommand returns information about the process representing the given Devfile command.
// A PID of 0 denotes a Stopped process.
func (k *kubeExecProcessHandler) GetProcessInfoForCommand(
	def CommandDefinition,
	podName string,
	containerName string,
) (RemoteProcessInfo, error) {
	klog.V(4).Infof("GetProcessInfoForCommand for %q", def.Id)

	pid, exitStatus, err := getRemoteProcessPID(def, podName, containerName)
	if err != nil {
		return RemoteProcessInfo{}, err
	}

	return k.getProcessInfoFromPid(pid, exitStatus, podName, containerName)
}

// StartProcessForCommand runs the (potentially never finishing) Devfile command in the background.
// The goroutine spawned here can be stopped either by stopping the parent process (e.g., 'odo dev'),
// or by stopping the underlying remote process by calling the StopProcessForCommand method.
func (k *kubeExecProcessHandler) StartProcessForCommand(
	def CommandDefinition,
	podName string,
	containerName string,
	outputHandler CommandOutputHandler,
) error {
	klog.V(4).Infof("StartProcessForCommand for %q", def.Id)

	// deal with environment variables
	cmdLine := def.CmdLine
	envCommands := make([]string, 0, len(def.EnvVars))
	for _, envVar := range def.EnvVars {
		envCommands = append(envCommands, fmt.Sprintf("%s='%s'", envVar.Key, envVar.Value))
	}
	var setEnvCmd string
	if len(envCommands) != 0 {
		setEnvCmd = fmt.Sprintf("export %s &&", strings.Join(envCommands, " "))
	}

	var cdCmd string
	if def.WorkingDir != "" {
		// Change to the workdir and execute the command
		cdCmd = fmt.Sprintf("cd %s &&", def.WorkingDir)
	}

	// since we are using /bin/sh -c, the command needs to be within a single double quote instance,
	// for example "cd /tmp && pwd"
	// Full command is: /bin/sh -c "[cd $workingDir && ] echo $$ > $pidFile && (envVar1='value1' envVar2='value2' $cmdLine) 1>>/proc/1/fd/1 2>>/proc/1/fd/2"
	//
	// "echo $$ > $pidFile" allows to store the /bin/sh parent process PID. It will allow to determine its children later on and kill them when a stop is requested.
	// ($cmdLine) runs the command passed in a subshell (to handle cases where the command does more complex things like running processes in the background),
	// which will be the child process of the /bin/sh one.
	//
	// Redirecting to /proc/1/fd/* allows to redirect the process output to the output streams of PID 1 process inside the container.
	// This way, returning the container logs with 'odo logs' or 'kubectl logs' would work seamlessly.
	// See https://stackoverflow.com/questions/58716574/where-exactly-do-the-logs-of-kubernetes-pods-come-from-at-the-container-level
	pidFile := getPidFileForCommand(def)
	cmd := []string{
		ShellExecutable, "-c",
		fmt.Sprintf("echo $$ > %[1]s && %s %s (%s) 1>>/proc/1/fd/1 2>>/proc/1/fd/2; echo $? >> %[1]s", pidFile, cdCmd, setEnvCmd, cmdLine),
	}

	go func() {
		if outputHandler != nil {
			outputHandler(Starting, nil, nil, nil)
		}

		stdout, stderr, err := kube.ExecuteCommand(cmd, podName, containerName, false, nil, nil)
		if err != nil {
			klog.V(2).Infof("error while running background command: %v", err)
		}

		if outputHandler != nil {
			processInfo, infoErr := k.GetProcessInfoForCommand(def, podName, containerName)
			if infoErr != nil {
				outputHandler(Errored, stdout, stderr, err)
				return
			}
			outputHandler(processInfo.Status, stdout, stderr, err)
		}
	}()

	return nil
}

// StopProcessForCommand stops the process representing the specified Devfile command.
// Because of the way this process is launched and its PID stored (see StartProcessForCommand),
// we need to determine the process children (there should be only one child which is the sub-shell running the command passed to StartProcessForCommand).
// Then killing those children will exit the parent 'sh' process.
func (k *kubeExecProcessHandler) StopProcessForCommand(
	def CommandDefinition,
	podName string,
	containerName string,
) error {
	klog.V(4).Infof("StopProcessForCommand for %q", def.Id)
	defer func() {
		pidFile := getPidFileForCommand(def)
		_, _, err := kube.ExecuteCommand([]string{ShellExecutable, "-c", fmt.Sprintf("rm -f %s", pidFile)},
			podName, containerName, false, nil, nil)
		if err != nil {
			klog.V(2).Infof("Could not remove file %q: %v", pidFile, err)
		}
	}()

	kill := func(p int) error {
		_, _, err := kube.ExecuteCommand([]string{ShellExecutable, "-c", fmt.Sprintf("kill %d || true", p)},
			podName, containerName, false, nil, nil)
		if err != nil {
			return err
		}

		//Because the process we just stopped might take longer to exit (it might have caught the signal and is performing additional cleanup),
		//retry detecting its actual state till it is stopped or timeout expires
		var processInfo interface{}
		processInfo, err = task.NewRetryable(fmt.Sprintf("status for remote process %d", p), func() (bool, interface{}, error) {
			pInfo, e := k.getProcessInfoFromPid(p, 0, podName, containerName)
			return e == nil && (pInfo.Status == Stopped || pInfo.Status == Errored), pInfo, e
		}).RetryWithSchedule([]time.Duration{2 * time.Second, 4 * time.Second, 8 * time.Second}, true)
		if err != nil {
			return err
		}

		pInfo, ok := processInfo.(RemoteProcessInfo)
		if !ok {
			klog.V(2).Infof("invalid type for remote process (%d) info, expected RemoteProcessInfo", p)
			return fmt.Errorf("internal error while checking remote process status: %d", p)
		}
		if pInfo.Status != Stopped && pInfo.Status != Errored {
			return fmt.Errorf("invalid status for remote process %d: %+v", p, processInfo)
		}
		return nil
	}

	ppid, _, err := getRemoteProcessPID(def, podName, containerName)
	if err != nil {
		return err
	}
	if ppid == 0 {
		return nil
	}

	children, err := getProcessChildren(ppid, podName, containerName)
	if err != nil {
		return err
	}

	if len(children) == 0 {
		//TODO(rm3l): A length of 0 might indicate that there is no children file, which might happen if the host kernel
		//was not built with the CONFIG_PROC_CHILDREN config.
		//This happened for example with the Minikube VM when using its (non-default) VirtualBox driver.
		//In this case, we should find a fallback mechanism to identify those children processes and kill them.
		return kill(ppid)
	}

	for _, child := range children {
		if err = kill(child); err != nil {
			return err
		}
	}

	return nil
}

func getRemoteProcessPID(def CommandDefinition, podName string, containerName string) (int, int, error) {
	pidFile := getPidFileForCommand(def)
	stdout, stderr, err := kube.ExecuteCommand(
		[]string{ShellExecutable, "-c", fmt.Sprintf("cat %s || true", pidFile)},
		podName, containerName, false, nil, nil)

	if err != nil {
		return 0, 0, err
	}

	if len(stdout) == 0 {
		//The file does not exist. We assume the process has not run yet.
		return 0, 0, nil
	}
	if len(stdout) > 2 {
		return 0, 0, fmt.Errorf(
			"unable to determine command status, due to unexpected number of lines for %s, output: %v %v",
			pidFile, stdout, stderr)
	}

	line := stdout[0]
	var pid int
	pid, err = strconv.Atoi(strings.TrimSpace(line))
	if err != nil {
		klog.V(2).Infof("error while trying to retrieve status of command: %+v", err)
		return 0, 0, fmt.Errorf("unable to determine command status, due to unexpected PID content for %s: %s",
			pidFile, line)
	}
	if len(stdout) == 1 {
		return pid, 0, nil
	}

	line2 := stdout[1]
	var exitStatus int
	exitStatus, err = strconv.Atoi(strings.TrimSpace(line2))
	if err != nil {
		klog.V(2).Infof("error while trying to retrieve status of command: %+v", err)
		return pid, 0, fmt.Errorf("unable to determine command status, due to unexpected exit status content for %s: %s",
			pidFile, line2)
	}

	return pid, exitStatus, nil
}

func (k *kubeExecProcessHandler) getProcessInfoFromPid(
	pid int,
	lastKnownExitStatus int,
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
	stdout, _, err := kube.ExecuteCommand(
		[]string{ShellExecutable, "-c", fmt.Sprintf("kill -0 %d; echo $?", pid)},
		podName, containerName, false, nil, nil)

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
		if lastKnownExitStatus == 0 {
			process.Status = Stopped
		} else {
			process.Status = Errored
		}
	}

	return process, nil
}

// getProcessChildren returns the children of the specified process in the given container.
// It works by reading the /proc/<pid>/task/<pid>/children file, which is a space-separated list of children
func getProcessChildren(pid int, podName string, containerName string) ([]int, error) {
	if pid <= 0 {
		return nil, fmt.Errorf("invalid pid: %d", pid)
	}

	stdout, _, err := kube.ExecuteCommand(
		[]string{ShellExecutable, "-c", fmt.Sprintf("cat /proc/%[1]d/task/%[1]d/children || true", pid)},
		podName, containerName, false, nil, nil)
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
func getPidFileForCommand(def CommandDefinition) string {
	return fmt.Sprintf("%s/.odo_cmd_%s.pid", strings.TrimSuffix(storage.SharedDataMountPath, "/"), def.Id)
}
