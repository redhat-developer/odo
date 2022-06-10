package component

import (
	"fmt"
	"time"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"k8s.io/client-go/util/exec"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/remotecmd"
	"github.com/redhat-developer/odo/pkg/util"
)

const numberOfLinesToOutputLog = 100

type adapterHandler struct {
	Adapter
	cmdKind         devfilev1.CommandGroupKind
	parameters      common.PushParameters
	componentExists bool
}

var _ libdevfile.Handler = (*adapterHandler)(nil)

func (a *adapterHandler) ApplyImage(_ devfilev1.Component) error {
	klog.V(2).Info("this handler can only handle exec commands in container components, not image components")
	return nil
}

func (a *adapterHandler) ApplyKubernetes(_ devfilev1.Component) error {
	klog.V(2).Info("this handler can only handle exec commands in container components, not Kubernetes components")
	return nil
}

func (a *adapterHandler) Execute(devfileCmd devfilev1.Command) error {
	processName := "devrun"
	if a.parameters.Debug {
		processName = "debugrun"
	}

	doExecuteBuildCommand := func() error {
		execHandler := component.NewExecHandler(a.kubeClient, a.pod.Name, "Building your application in container on cluster", a.parameters.Show)
		return libdevfile.Build(a.Devfile, execHandler, true)
	}

	remoteProcessHandler := remotecmd.NewKubeExecProcessHandler()

	startHandler := func(stdout []string, stderr []string, err error) {
		if err != nil {
			klog.V(2).Infof("error while running background command: %v", err)

			var codeExitError exec.CodeExitError
			if errors.As(err, &codeExitError) && codeExitError.ExitStatus() == _sigtermExitCode {
				//process killed by SIGTERM signal (potentially via the stop command, upon a restart command)
				//=> displaying the logs might feel disturbing to the end-users
				klog.V(2).Infof("stdout: %s", strings.Join(stdout, "\n"))
				klog.V(2).Infof("stderr: %s", strings.Join(stderr, "\n"))
			} else {
				// Use GetStderr in order to make sure that colour output is correct
				// on non-TTY terminals
				log.Warningf("Last %d lines of log:", _numberOfLinesToOutputLog)
				_ = util.DisplayLogLines(stdout, log.GetStderr(), _numberOfLinesToOutputLog)
				_ = util.DisplayLogLines(stderr, log.GetStderr(), _numberOfLinesToOutputLog)
			}
		}
	}

	// if we need to restart, issue the remote process handler command to stop all running commands first.
	// We do not need to restart Hot reload capable commands.
	if a.componentExists {
		cmd, err := libdevfile.GetDefaultCommand(a.Devfile, a.cmdKind)
		if err != nil {
			return err
		}

		if a.parameters.RunModeChanged || cmd.Exec == nil || !util.SafeGetBool(cmd.Exec.HotReloadCapable) {
			klog.V(2).Info("restart required for command")
			if err = doExecuteBuildCommand(); err != nil {
				return err
			}

			err = remoteProcessHandler.StopProcessForCommand(devfileCmd, a.kubeClient, a.pod.Name, devfileCmd.Exec.Component, nil)
			if err != nil {
				return err
			}

			if err = remoteProcessHandler.StartProcessForCommand(devfileCmd, a.kubeClient, a.pod.Name, devfileCmd.Exec.Component, startHandler); err != nil {
				return err
			}
		} else {
			klog.V(2).Infof("command is hot-reload capable, not restarting %s", processName)
		}
	} else {
		if err := doExecuteBuildCommand(); err != nil {
			return err
		}

		if err := remoteProcessHandler.StartProcessForCommand(devfileCmd, a.kubeClient, a.pod.Name, devfileCmd.Exec.Component, startHandler); err != nil {
			return err
		}
	}

	retrySchedule := []time.Duration{
		5 * time.Second,
		6 * time.Second,
		9 * time.Second,
	}
	var isRunning bool
	var err error
	var totalWaitTime float64
	for _, s := range retrySchedule {
		time.Sleep(s)
		klog.V(4).Infof("checking if process for command %q is running (will timeout in %0.f seconds if not)",
			devfileCmd.Id, s.Seconds())
		isRunning, err = a.isRemoteProcessForCommandRunning(devfileCmd)
		if err == nil && isRunning {
			break
		}
		totalWaitTime += s.Seconds()
	}
	//All retries failed, but giving one last chance
	if err != nil {
		return err
	}

	return a.checkRemoteCommandStatus(devfileCmd,
		fmt.Sprintf("Devfile command %q exited with an error status in %.0f second(s)", devfileCmd.Id, totalWaitTime))
}

// isRemoteProcessForCommandRunning returns true if the command is running
func (a *adapterHandler) isRemoteProcessForCommandRunning(command devfilev1.Command) (bool, error) {
	remoteProcessHandler := remotecmd.NewKubeExecProcessHandler()
	remoteProcess, err := remoteProcessHandler.GetProcessInfoForCommand(command, a.kubeClient, a.pod.Name, command.Exec.Component)
	if err != nil {
		return false, err
	}

	return remoteProcess.Status == remotecmd.Running, nil
}

// checkRemoteCommandStatus checks if the command is running .
// if the command is not in a running state, we fetch the last 20 lines of the component's log and display it
func (a *adapterHandler) checkRemoteCommandStatus(command devfilev1.Command, notRunningMessage string) error {
	running, err := a.isRemoteProcessForCommandRunning(command)
	if err != nil {
		return err
	}

	if !running {
		log.Warningf(notRunningMessage)
		log.Warningf("Last %d lines of log:", numberOfLinesToOutputLog)

		rd, err := component.Log(a.kubeClient, a.ComponentName, a.AppName, false, command)
		if err != nil {
			return err
		}

		// Use GetStderr in order to make sure that colour output is correct
		// on non-TTY terminals
		err = util.DisplayLog(false, rd, log.GetStderr(), a.ComponentName, numberOfLinesToOutputLog)
		if err != nil {
			return err
		}
	}
	return nil
}
