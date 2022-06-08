package component

import (
	"time"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/remotecmd"
	"github.com/redhat-developer/odo/pkg/util"
)

const _numberOfLinesToOutputLog = 100
const _processStatusWaitTime = 1 * time.Second

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
		return libdevfile.Build(a.Devfile, component.NewExecHandler(a.kubeClient, a.pod.Name, a.parameters.Show), true)
	}

	remoteProcessHandler := remotecmd.NewKubeExecProcessHandler()

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

			_, _, err = remoteProcessHandler.StopProcessForCommand(devfileCmd, a.kubeClient, a.pod.Name, devfileCmd.Exec.Component)
			if err != nil {
				return err
			}

			_, _, err = remoteProcessHandler.StartProcessForCommand(devfileCmd, a.kubeClient, a.pod.Name, devfileCmd.Exec.Component)
			if err != nil {
				return err
			}
		} else {
			klog.V(2).Infof("command is hot-reload capable, not restarting %s", processName)
		}
	} else {
		if err := doExecuteBuildCommand(); err != nil {
			return err
		}

		_, _, err := remoteProcessHandler.StartProcessForCommand(devfileCmd, a.kubeClient, a.pod.Name, devfileCmd.Exec.Component)
		if err != nil {
			return err
		}
	}

	//Need to wait a few seconds prior to checking the status,
	//as some implementations might take time to report a correct status
	time.Sleep(_processStatusWaitTime)

	return a.checkRemoteCommandStatus(devfileCmd)
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
func (a *adapterHandler) checkRemoteCommandStatus(command devfilev1.Command) error {
	running, err := a.isRemoteProcessForCommandRunning(command)
	if err != nil {
		return err
	}

	if !running {
		log.Warningf("Devfile command %q exited with an error status in %.0f sec", command.Id, _processStatusWaitTime.Seconds())

		rd, err := component.Log(a.kubeClient, a.ComponentName, a.AppName, false, command)
		if err != nil {
			return err
		}

		// Use GetStderr in order to make sure that colour output is correct
		// on non-TTY terminals
		err = util.DisplayLog(false, rd, log.GetStderr(), a.ComponentName, _numberOfLinesToOutputLog)
		if err != nil {
			return err
		}
	}
	return nil
}
