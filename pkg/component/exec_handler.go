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
	kubeClient kclient.ClientInterface
	podName    string
	msg        string
	show       bool
}

const ShellExecutable string = "/bin/sh"

func NewExecHandler(kubeClient kclient.ClientInterface, podName string, msg string, show bool) *execHandler {
	return &execHandler{
		kubeClient: kubeClient,
		podName:    podName,
		msg:        msg,
		show:       show,
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
	}
	spinner := log.Spinner(msg)
	defer spinner.End(false)

	logger := machineoutput.NewMachineEventLoggingClient()
	stdoutWriter, stdoutChannel, stderrWriter, stderrChannel := logger.CreateContainerOutputWriter()

	cmdline := getCmdline(command)
	err := remotecmd.ExecuteCommand(cmdline, o.kubeClient, o.podName, command.Exec.Component, o.show, stdoutWriter, stderrWriter)

	closeWriterAndWaitForAck(stdoutWriter, stdoutChannel, stderrWriter, stderrChannel)

	spinner.End(true)
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
	var cmd []string
	if command.Exec.WorkingDir != "" {
		// since we are using /bin/sh -c, the command needs to be within a single double quote instance, for example "cd /tmp && pwd"
		cmd = []string{ShellExecutable, "-c", "cd " + command.Exec.WorkingDir + " && " + cmdLine}
	} else {
		cmd = []string{ShellExecutable, "-c", cmdLine}
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
