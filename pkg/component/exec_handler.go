package component

import (
	"fmt"
	"io"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/remotecmd"
	"github.com/redhat-developer/odo/pkg/util"
)

type execHandler struct {
	kubeClient    kclient.ClientInterface
	appName       string
	componentName string
	podName       string
	msg           string
	show          bool
}

const ShellExecutable string = "/bin/sh"

func NewExecHandler(kubeClient kclient.ClientInterface, appName, cmpName, podName, msg string, show bool) *execHandler {
	return &execHandler{
		kubeClient:    kubeClient,
		appName:       appName,
		componentName: cmpName,
		podName:       podName,
		msg:           msg,
		show:          show,
	}
}

func (o *execHandler) ApplyImage(image v1alpha2.Component) error {
	return nil
}

func (o *execHandler) ApplyKubernetes(kubernetes v1alpha2.Component) error {
	return nil
}

func (o *execHandler) Execute(command v1alpha2.Command) error {
	msg := o.msg
	if msg == "" {
		msg = fmt.Sprintf("Executing %s command %q on container %q", command.Id, command.Exec.CommandLine, command.Exec.Component)
	} else {
		msg += " (command: " + command.Id + ")"
	}
	spinner := log.Spinner(msg)
	defer spinner.End(false)

	logger := machineoutput.NewMachineEventLoggingClient()
	stdoutWriter, stdoutChannel, stderrWriter, stderrChannel := logger.CreateContainerOutputWriter()

	cmdline := getCmdline(command)
	_, _, err := remotecmd.ExecuteCommand(cmdline, o.kubeClient, o.podName, command.Exec.Component, o.show, stdoutWriter, stderrWriter)

	closeWriterAndWaitForAck(stdoutWriter, stdoutChannel, stderrWriter, stderrChannel)

	spinner.End(err == nil)
	if err != nil {
		rd, errLog := Log(o.kubeClient, o.componentName, o.appName, false, command)
		if errLog != nil {
			return fmt.Errorf("unable to log error %v: %w", err, errLog)
		}

		// Use GetStderr in order to make sure that colour output is correct
		// on non-TTY terminals
		errLog = util.DisplayLog(false, rd, log.GetStderr(), o.componentName, -1)
		if errLog != nil {
			return fmt.Errorf("unable to log error %v: %w", err, errLog)
		}
	}
	return err
}

func getCmdline(command v1alpha2.Command) []string {
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
	redirectString := "1>>/proc/1/fd/1 2>>/proc/1/fd/2"
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
