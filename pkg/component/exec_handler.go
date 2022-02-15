package component

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/machineoutput"
	"github.com/redhat-developer/odo/pkg/util"
	"k8s.io/klog"
)

type execHandler struct {
	kubeClient kclient.ClientInterface
	podName    string
	show       bool
}

const ShellExecutable string = "/bin/sh"

func NewExecHandler(kubeClient kclient.ClientInterface, podName string, show bool) *execHandler {
	return &execHandler{
		kubeClient: kubeClient,
		podName:    podName,
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
	msg := fmt.Sprintf("Executing %s command %q on container %q", command.Id, command.Exec.CommandLine, command.Exec.Component)
	spinner := log.Spinner(msg)
	defer spinner.End(false)

	logger := machineoutput.NewMachineEventLoggingClient()
	stdoutWriter, stdoutChannel, stderrWriter, stderrChannel := logger.CreateContainerOutputWriter()

	cmdline := getCmdline(command)
	err := executeCommand(o.kubeClient, command.Exec.Component, o.podName, cmdline, o.show, stdoutWriter, stderrWriter)

	closeWriterAndWaitForAck(stdoutWriter, stdoutChannel, stderrWriter, stderrChannel)

	spinner.End(true)
	return err
}

func getCmdline(command v1alpha2.Command) []string {
	exe := command.Exec

	// deal with environment variables
	var cmdLine string
	setEnvVariable := util.GetCommandStringFromEnvs(exe.Env)

	if setEnvVariable == "" {
		cmdLine = exe.CommandLine
	} else {
		cmdLine = setEnvVariable + " && " + exe.CommandLine
	}

	// Change to the workdir and execute the command
	var cmd []string
	if exe.WorkingDir != "" {
		// since we are using /bin/sh -c, the command needs to be within a single double quote instance, for example "cd /tmp && pwd"
		cmd = []string{ShellExecutable, "-c", "cd " + exe.WorkingDir + " && " + cmdLine}
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

// ExecuteCommand executes the given command in the pod's container
func executeCommand(client kclient.ClientInterface, containerName string, podName string, command []string, show bool, consoleOutputStdout *io.PipeWriter, consoleOutputStderr *io.PipeWriter) (err error) {
	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	var cmdOutput string

	klog.V(2).Infof("Executing command %v for pod: %v in container: %v", command, podName, containerName)

	// Read stdout and stderr, store their output in cmdOutput, and also pass output to consoleOutput Writers (if non-nil)
	stdoutCompleteChannel := startReaderGoroutine(stdoutReader, show, &cmdOutput, consoleOutputStdout)
	stderrCompleteChannel := startReaderGoroutine(stderrReader, show, &cmdOutput, consoleOutputStderr)

	err = client.ExecCMDInContainer(containerName, podName, command, stdoutWriter, stderrWriter, nil, false)

	// Block until we have received all the container output from each stream
	_ = stdoutWriter.Close()
	<-stdoutCompleteChannel
	_ = stderrWriter.Close()
	<-stderrCompleteChannel

	if err != nil {
		// It is safe to read from cmdOutput here, as the goroutines are guaranteed to have terminated at this point.
		klog.V(2).Infof("ExecuteCommand returned an an err: %v. for command '%v'. output: %v", err, command, cmdOutput)

		return errors.Wrapf(err, "unable to exec command %v: \n%v", command, cmdOutput)
	}

	return
}

// This goroutine will automatically pipe the output from the writer (passed into ExecCMDInContainer) to
// the loggers.
// The returned channel will contain a single nil entry once the reader has closed.
func startReaderGoroutine(reader io.Reader, show bool, cmdOutput *string, consoleOutput *io.PipeWriter) chan interface{} {

	result := make(chan interface{})

	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()

			if log.IsDebug() || show {
				_, err := fmt.Fprintln(os.Stdout, line)
				if err != nil {
					log.Errorf("Unable to print to stdout: %s", err.Error())
				}
			}

			*cmdOutput += fmt.Sprintln(line)

			if consoleOutput != nil {
				_, err := consoleOutput.Write([]byte(line + "\n"))
				if err != nil {
					log.Errorf("Error occurred on writing string to consoleOutput writer: %s", err.Error())
				}
			}
		}
		result <- nil
	}()

	return result

}
