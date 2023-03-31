package remotecmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/storage"
	"github.com/redhat-developer/odo/pkg/task"
)

// kubeExecProcessHandler implements RemoteProcessHandler by executing Devfile commands right away in the container
// (like a 'kubectl exec'-like approach). Command execution is done in the background, in a separate goroutine.
// It works by storing the parent process PID in a file in the container,
// then fires the exec command in the background.
// The goroutine started can then be stopped by killing the process stored in the state file (_startCmdProcessPidFile)
// in the container.
type kubeExecProcessHandler struct {
	execClient exec.Client
}

// This allows to verify interface compliance at compile-time.
// See https://github.com/redhat-developer/odo/wiki/Dev:-Coding-Conventions#verify-interface-compliance
var _ RemoteProcessHandler = (*kubeExecProcessHandler)(nil)

func NewKubeExecProcessHandler(execClient exec.Client) *kubeExecProcessHandler {
	return &kubeExecProcessHandler{
		execClient: execClient,
	}
}

// GetProcessInfoForCommand returns information about the process representing the given Devfile command.
// A PID of 0 denotes a Stopped process.
func (k *kubeExecProcessHandler) GetProcessInfoForCommand(
	def CommandDefinition,
	podName string,
	containerName string,
) (RemoteProcessInfo, error) {
	klog.V(4).Infof("GetProcessInfoForCommand for %q", def.Id)

	pid, exitStatus, err := k.getRemoteProcessPID(def, podName, containerName)
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

	//Monitoring go-routine
	type event struct {
		status RemoteProcessStatus
		stdout []string
		stderr []string
		err    error
	}
	eventsChan := make(chan event)
	eventsReceived := make(map[RemoteProcessStatus]struct{})
	go func() {
		for e := range eventsChan {
			klog.V(5).Infof("event received for %q: %v, %v", def.Id, e.status, e.err)
			if _, ok := eventsReceived[e.status]; ok {
				continue
			}
			if outputHandler != nil {
				outputHandler(e.status, e.stdout, e.stderr, e.err)
			}
			eventsReceived[e.status] = struct{}{}
		}
	}()

	eventsChan <- event{status: Starting}

	go func() {
		eventsChan <- event{status: Running}
		stdout, stderr, err := k.execClient.ExecuteCommand(cmd, podName, containerName, false, nil, nil)
		if err != nil {
			klog.V(2).Infof("error while running background command: %v", err)
		}

		processInfo, infoErr := k.GetProcessInfoForCommand(def, podName, containerName)
		var status RemoteProcessStatus
		if infoErr != nil {
			status = Errored
		} else {
			status = processInfo.Status
		}

		eventsChan <- event{
			status: status,
			stdout: stdout,
			stderr: stderr,
			err:    err,
		}

		close(eventsChan)
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
		_, _, err := k.execClient.ExecuteCommand([]string{ShellExecutable, "-c", fmt.Sprintf("rm -f %s", pidFile)},
			podName, containerName, false, nil, nil)
		if err != nil {
			klog.V(2).Infof("Could not remove file %q: %v", pidFile, err)
		}
	}()

	kill := func(p int) error {
		_, _, err := k.execClient.ExecuteCommand([]string{ShellExecutable, "-c", fmt.Sprintf("kill %d || true", p)},
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

	ppid, _, err := k.getRemoteProcessPID(def, podName, containerName)
	if err != nil {
		return err
	}
	if ppid == 0 {
		return nil
	}
	defer func() {
		if kErr := kill(ppid); kErr != nil {
			klog.V(3).Infof("could not kill parent process %d: %v", ppid, kErr)
		}
	}()

	children, err := k.getProcessChildren(ppid, podName, containerName)
	if err != nil {
		return err
	}

	klog.V(3).Infof("Found %d children (either direct and indirect) for parent process %d: %v", len(children), ppid, children)

	for _, child := range children {
		if err = kill(child); err != nil {
			return err
		}
	}

	return nil
}

func (k *kubeExecProcessHandler) getRemoteProcessPID(def CommandDefinition, podName string, containerName string) (int, int, error) {
	pidFile := getPidFileForCommand(def)
	stdout, stderr, err := k.execClient.ExecuteCommand(
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
	stdout, _, err := k.execClient.ExecuteCommand(
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

// getProcessChildren returns all the children (either direct or indirect) of the specified process in the given container.
// It works by reading the /proc/<pid>/stat files, giving PPID for each PID.
// The overall result is an ordered list of children PIDs obtained via a recursive post-order traversal algorithm,
// so that the returned list can start with the deepest children processes.
func (k *kubeExecProcessHandler) getProcessChildren(pid int, podName string, containerName string) ([]int, error) {
	if pid <= 0 {
		return nil, fmt.Errorf("invalid pid: %d", pid)
	}

	allProcesses, err := k.getAllProcesses(podName, containerName)
	if err != nil {
		return nil, err
	}

	var getProcessChildrenRec func(int) []int
	getProcessChildrenRec = func(p int) []int {
		var result []int
		for _, child := range allProcesses[p] {
			result = append(result, getProcessChildrenRec(child)...)
		}
		if p != pid {
			// Do not include the root ppid, as we are getting only children.
			result = append(result, p)
		}
		return result
	}

	return getProcessChildrenRec(pid), nil
}

// getAllProcesses returns a map of all the processes and their direct children:
// i) the key is the process PID;
// ii) and the value is a list of all its direct children.
// It does so by reading all /proc/<pid>/stat files. More details on https://man7.org/linux/man-pages/man5/proc.5.html.
func (k *kubeExecProcessHandler) getAllProcesses(podName string, containerName string) (map[int][]int, error) {
	stdout, stderr, err := k.execClient.ExecuteCommand([]string{ShellExecutable, "-c", "cat /proc/*/stat || true"},
		podName, containerName, false, nil, nil)
	if err != nil {
		klog.V(7).Infof("stdout: %s\n", strings.Join(stdout, "\n"))
		klog.V(7).Infof("stderr: %s\n", strings.Join(stderr, "\n"))
		return nil, err
	}

	allProcesses := make(map[int][]int)
	for _, line := range stdout {
		var pid int
		_, err = fmt.Sscanf(line, "%d ", &pid)
		if err != nil {
			return nil, err
		}

		// Last index of the last ")" character to unambiguously parse the content.
		// See https://unix.stackexchange.com/questions/558239/way-to-unambiguously-parse-proc-pid-stat-given-arbitrary-contents-of-name-fi
		i := strings.LastIndex(line, ")")
		if i < 0 {
			continue
		}

		// At this point, "i" is the index of the last ")" character, and we have an additional space before the process state character, hence the "i+2".
		// For example:
		// 87 (main) S 0 81 81 0 -1 ...
		// This is required to scan the ppid correctly.
		var state byte
		var ppid int
		_, err = fmt.Sscanf(line[i+2:], "%c %d", &state, &ppid)
		if err != nil {
			return nil, err
		}

		allProcesses[ppid] = append(allProcesses[ppid], pid)
	}

	return allProcesses, nil
}

// getPidFileForCommand returns the path to the PID file in the remote container.
// The parent folder is supposed to be existing, because it should be mounted in the container using the mandatory
// shared volume (more info in the AddOdoMandatoryVolume function from the utils package).
func getPidFileForCommand(def CommandDefinition) string {
	parentDir := def.PidDirectory
	if parentDir == "" {
		parentDir = storage.SharedDataMountPath
	}
	return fmt.Sprintf("%s/.odo_cmd_%s.pid", strings.TrimSuffix(parentDir, "/"), def.Id)
}
