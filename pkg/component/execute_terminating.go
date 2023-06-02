package component

import (
	"context"
	"fmt"
	"io"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/platform"
	"github.com/redhat-developer/odo/pkg/util"
	"k8s.io/klog"
	"k8s.io/utils/pointer"
)

const ShellExecutable string = "/bin/sh"

func ExecuteTerminatingCommand(ctx context.Context, execClient exec.Client, platformClient platform.Client, command devfilev1.Command, componentExists bool, podName string, appName string, componentName string, msg string, showOutputs bool) error {

	if componentExists && command.Exec != nil && pointer.BoolDeref(command.Exec.HotReloadCapable, false) {
		klog.V(2).Infof("command is hot-reload capable, not executing %q again", command.Id)
		return nil
	}

	if msg == "" {
		msg = fmt.Sprintf("Executing %s command on container %q", command.Id, command.Exec.Component)
	} else {
		msg += " (command: " + command.Id + ")"
	}

	// Spinner is displayed only if no outputs are displayed
	var spinner *log.Status
	if !showOutputs {
		spinner = log.Spinner(msg)
		defer spinner.End(false)
	}

	logger := machineoutput.NewMachineEventLoggingClient()
	stdoutWriter, stdoutChannel, stderrWriter, stderrChannel := logger.CreateContainerOutputWriter()

	cmdline := getCmdline(command, !showOutputs)
	_, _, err := execClient.ExecuteCommand(ctx, cmdline, podName, command.Exec.Component, showOutputs, stdoutWriter, stderrWriter)

	closeWriterAndWaitForAck(stdoutWriter, stdoutChannel, stderrWriter, stderrChannel)

	if !showOutputs {
		spinner.End(err == nil)
	}
	// Complete logs are displayed only if no outputs are displayed
	if err != nil && !showOutputs {
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

func getCmdline(command v1alpha2.Command, redirectToPid1 bool) []string {
	// deal with environment variables
	var cmdLine string
	setEnvVariable := util.GetCommandStringFromEnvs(command.Exec.Env)

	if setEnvVariable == "" {
		cmdLine = command.Exec.CommandLine
	} else {
		cmdLine = setEnvVariable + " && " + command.Exec.CommandLine
	}

	// Change to the workdir and execute the command
	// Redirecting to /proc/1/fd/* allows to redirect the process output to the output streams of PID 1 process inside the container.
	// This way, returning the container logs with 'odo logs' or 'kubectl logs' would work seamlessly.
	// See https://stackoverflow.com/questions/58716574/where-exactly-do-the-logs-of-kubernetes-pods-come-from-at-the-container-level
	redirectString := ""
	if redirectToPid1 {
		redirectString = "1>>/proc/1/fd/1 2>>/proc/1/fd/2"
	}
	var cmd []string
	if command.Exec.WorkingDir != "" {
		// since we are using /bin/sh -c, the command needs to be within a single double quote instance, for example "cd /tmp && pwd"
		cmd = []string{ShellExecutable, "-c", "cd " + command.Exec.WorkingDir + " && (" + cmdLine + ") " + redirectString}
	} else {
		cmd = []string{ShellExecutable, "-c", "(" + cmdLine + ") " + redirectString}
	}
	return cmd
}

func closeWriterAndWaitForAck(stdoutWriter *io.PipeWriter, stdoutChannel chan interface{}, stderrWriter *io.PipeWriter, stderrChannel chan interface{}) {
	if stdoutWriter != nil {
		_ = stdoutWriter.Close()
		<-stdoutChannel
	}
	if stderrWriter != nil {
		_ = stderrWriter.Close()
		<-stderrChannel
	}
}
