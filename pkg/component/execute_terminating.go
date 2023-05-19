package component

import (
	"context"
	"fmt"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/platform"
	"github.com/redhat-developer/odo/pkg/util"
	"k8s.io/klog"
	"k8s.io/utils/pointer"
)

func ExecuteTerminatingCommand(ctx context.Context, execClient exec.Client, platformClient platform.Client, command devfilev1.Command, componentExists bool, podName string, appName string, componentName string, msg string, show bool) error {

	if componentExists && command.Exec != nil && pointer.BoolDeref(command.Exec.HotReloadCapable, false) {
		klog.V(2).Infof("command is hot-reload capable, not executing %q again", command.Id)
		return nil
	}

	if msg == "" {
		msg = fmt.Sprintf("Executing %s command on container %q", command.Id, command.Exec.Component)
	} else {
		msg += " (command: " + command.Id + ")"
	}
	spinner := log.Spinner(msg)
	defer spinner.End(false)

	logger := machineoutput.NewMachineEventLoggingClient()
	stdoutWriter, stdoutChannel, stderrWriter, stderrChannel := logger.CreateContainerOutputWriter()

	cmdline := getCmdline(command)
	_, _, err := execClient.ExecuteCommand(ctx, cmdline, podName, command.Exec.Component, show, stdoutWriter, stderrWriter)

	closeWriterAndWaitForAck(stdoutWriter, stdoutChannel, stderrWriter, stderrChannel)

	spinner.End(err == nil)
	if err != nil {
		rd, errLog := Log(platformClient, componentName, appName, false, command)
		if errLog != nil {
			return fmt.Errorf("unable to log error %v: %w", err, errLog)
		}

		// Use GetStderr in order to make sure that colour output is correct
		// on non-TTY terminals
		errLog = util.DisplayLog(false, rd, log.GetStderr(), componentName, -1)
		if errLog != nil {
			return fmt.Errorf("unable to log error %v: %w", err, errLog)
		}
	}
	return err
}
