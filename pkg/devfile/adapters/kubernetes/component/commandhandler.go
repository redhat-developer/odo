package component

import (
	"errors"
	"fmt"
	"time"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/remotecmd"
	"github.com/redhat-developer/odo/pkg/sync"
	"github.com/redhat-developer/odo/pkg/task"
	"github.com/redhat-developer/odo/pkg/util"
)

const numberOfLinesToOutputLog = 100

type adapterHandler struct {
	Adapter
	parameters      common.PushParameters
	componentExists bool
}

var _ libdevfile.Handler = (*adapterHandler)(nil)
var _ common.ComponentAdapter = (*adapterHandler)(nil)
var _ sync.SyncClient = (*adapterHandler)(nil)

func (a *adapterHandler) ApplyImage(_ devfilev1.Component) error {
	klog.V(2).Info("this handler can only handle exec commands in container components, not image components")
	return nil
}

func (a *adapterHandler) ApplyKubernetes(_ devfilev1.Component) error {
	klog.V(2).Info("this handler can only handle exec commands in container components, not Kubernetes components")
	return nil
}

func (a *adapterHandler) Execute(devfileCmd devfilev1.Command) error {
	remoteProcessHandler := remotecmd.NewKubeExecProcessHandler()

	statusHandlerFunc := func(s *log.Status) remotecmd.CommandOutputHandler {
		return func(status remotecmd.RemoteProcessStatus, stdout []string, stderr []string, err error) {
			switch status {
			case remotecmd.Starting:
				// Creating with no spin because the command could be long-running, and we cannot determine when it will end.
				s.Start(fmt.Sprintf("Executing the application (command: %s)", devfileCmd.Id), true)
			case remotecmd.Stopped, remotecmd.Errored:
				s.End(status == remotecmd.Stopped)
				if err != nil {
					klog.V(2).Infof("error while running background command: %v", err)
				}
			}
		}
	}

	// Spinner created but not started yet.
	// It will be displayed when the statusHandlerFunc function is called with the "Starting" state.
	spinner := log.NewStatus(log.GetStdout())

	// if we need to restart, issue the remote process handler command to stop all running commands first.
	// We do not need to restart Hot reload capable commands.
	if a.componentExists {
		if a.parameters.RunModeChanged || devfileCmd.Exec == nil || !util.SafeGetBool(devfileCmd.Exec.HotReloadCapable) {
			klog.V(2).Infof("restart required for command %s", devfileCmd.Id)

			cmdDef, err := devfileCommandToRemoteCmdDefinition(devfileCmd)
			if err != nil {
				return err
			}

			err = remoteProcessHandler.StopProcessForCommand(cmdDef, a.kubeClient, a.pod.Name, devfileCmd.Exec.Component)
			if err != nil {
				return err
			}

			if err = remoteProcessHandler.StartProcessForCommand(cmdDef, a.kubeClient, a.pod.Name, devfileCmd.Exec.Component, statusHandlerFunc(spinner)); err != nil {
				return err
			}
		} else {
			klog.V(2).Infof("command is hot-reload capable, not restarting %s", devfileCmd.Id)
		}
	} else {
		cmdDef, err := devfileCommandToRemoteCmdDefinition(devfileCmd)
		if err != nil {
			return err
		}

		if err := remoteProcessHandler.StartProcessForCommand(cmdDef, a.kubeClient, a.pod.Name, devfileCmd.Exec.Component, statusHandlerFunc(spinner)); err != nil {
			return err
		}
	}

	retrySchedule := []time.Duration{
		5 * time.Second,
		6 * time.Second,
		9 * time.Second,
	}
	var totalWaitTime float64
	for _, s := range retrySchedule {
		totalWaitTime += s.Seconds()
	}

	_, err := task.NewRetryable(fmt.Sprintf("process for command %q", devfileCmd.Id), func() (bool, interface{}, error) {
		klog.V(4).Infof("checking if process for command %q is running", devfileCmd.Id)
		remoteProcess, err := remoteProcessHandler.GetProcessInfoForCommand(
			remotecmd.CommandDefinition{Id: devfileCmd.Id}, a.kubeClient, a.pod.Name, devfileCmd.Exec.Component)
		if err != nil {
			return false, nil, err
		}
		isRunningOrDone := remoteProcess.Status == remotecmd.Running ||
			remoteProcess.Status == remotecmd.Stopped ||
			remoteProcess.Status == remotecmd.Errored
		return isRunningOrDone, nil, err
	}).RetryWithSchedule(retrySchedule, false)
	if err != nil {
		return err
	}

	return a.checkRemoteCommandStatus(devfileCmd,
		fmt.Sprintf("Devfile command %q exited with an error status in %.0f second(s)", devfileCmd.Id, totalWaitTime))
}

// devfileCommandToRemoteCmdDefinition builds and returns a new remotecmd.CommandDefinition object from the specified devfileCmd.
// An error is returned for non-exec Devfile commands.
func devfileCommandToRemoteCmdDefinition(devfileCmd devfilev1.Command) (remotecmd.CommandDefinition, error) {
	if devfileCmd.Exec == nil {
		return remotecmd.CommandDefinition{}, errors.New(" only Exec commands are supported")
	}

	envVars := make([]remotecmd.CommandEnvVar, 0, len(devfileCmd.Exec.Env))
	for _, e := range devfileCmd.Exec.Env {
		envVars = append(envVars, remotecmd.CommandEnvVar{Key: e.Name, Value: e.Value})
	}

	return remotecmd.CommandDefinition{
		Id:         devfileCmd.Id,
		WorkingDir: devfileCmd.Exec.WorkingDir,
		EnvVars:    envVars,
		CmdLine:    devfileCmd.Exec.CommandLine,
	}, nil
}

// isRemoteProcessForCommandRunning returns true if the command is running
func (a *adapterHandler) isRemoteProcessForCommandRunning(command devfilev1.Command) (bool, error) {
	remoteProcess, err := remotecmd.NewKubeExecProcessHandler().GetProcessInfoForCommand(
		remotecmd.CommandDefinition{Id: command.Id}, a.kubeClient, a.pod.Name, command.Exec.Component)
	if err != nil {
		return false, err
	}

	return remoteProcess.Status == remotecmd.Running, nil
}

// checkRemoteCommandStatus checks if the command is running .
// if the command is not in a running state, we fetch the last 20 lines of the component's log and display it
func (a *adapterHandler) checkRemoteCommandStatus(command devfilev1.Command, notRunningMessage string) error {
	remoteProcessHandler := remotecmd.NewKubeExecProcessHandler()
	remoteProcess, err := remoteProcessHandler.GetProcessInfoForCommand(remotecmd.CommandDefinition{Id: command.Id}, a.kubeClient, a.pod.Name, command.Exec.Component)
	if err != nil {
		return err
	}

	if remoteProcess.Status != remotecmd.Running && remoteProcess.Status != remotecmd.Stopped {
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
