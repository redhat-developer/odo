package component

import (
	"context"
	"io"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"

	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/platform"
	"github.com/redhat-developer/odo/pkg/util"
)

type execHandler struct {
	platformClient  platform.Client
	execClient      exec.Client
	appName         string
	componentName   string
	podName         string
	msg             string
	show            bool
	componentExists bool
}

var _ libdevfile.Handler = (*execHandler)(nil)

const ShellExecutable string = "/bin/sh"

func NewExecHandler(platformClient platform.Client, execClient exec.Client, appName, cmpName, podName, msg string, show bool, componentExists bool) *execHandler {
	return &execHandler{
		platformClient:  platformClient,
		execClient:      execClient,
		appName:         appName,
		componentName:   cmpName,
		podName:         podName,
		msg:             msg,
		show:            show,
		componentExists: componentExists,
	}
}

func (o *execHandler) ApplyImage(image v1alpha2.Component) error {
	return nil
}

func (o *execHandler) ApplyKubernetes(kubernetes v1alpha2.Component) error {
	return nil
}

func (o *execHandler) ApplyOpenShift(openshift v1alpha2.Component) error {
	return nil
}

func (o *execHandler) ExecuteNonTerminatingCommand(ctx context.Context, command v1alpha2.Command) error {
	return nil
}

func (o *execHandler) ExecuteTerminatingCommand(ctx context.Context, command v1alpha2.Command) error {
	return ExecuteTerminatingCommand(ctx, o.execClient, o.platformClient, command, o.componentExists, o.podName, o.appName, o.componentName, o.msg, o.show)
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
