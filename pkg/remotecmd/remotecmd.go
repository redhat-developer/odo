package remotecmd

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/log"
)

// ExecuteCommand executes the given command in the pod's container,
// writing the output to the specified respective pipe writers
func ExecuteCommand(
	command []string,
	client kclient.ClientInterface,
	podName string,
	containerName string,
	show bool,
	stdoutWriter *io.PipeWriter,
	stderrWriter *io.PipeWriter,
) error {
	soutReader, soutWriter := io.Pipe()
	serrReader, serrWriter := io.Pipe()

	var cmdOutput string

	klog.V(2).Infof("Executing command %v for pod: %v in container: %v", command, podName, containerName)

	// Read stdout and stderr, store their output in cmdOutput, and also pass output to consoleOutput Writers (if non-nil)
	stdoutCompleteChannel := startReaderGoroutine(soutReader, show, &cmdOutput, stdoutWriter)
	stderrCompleteChannel := startReaderGoroutine(serrReader, show, &cmdOutput, stderrWriter)

	err := client.ExecCMDInContainer(containerName, podName, command, soutWriter, serrWriter, nil, false)

	// Block until we have received all the container output from each stream
	_ = soutWriter.Close()
	<-stdoutCompleteChannel
	_ = serrWriter.Close()
	<-stderrCompleteChannel

	if err != nil {
		// It is safe to read from cmdOutput here, as the goroutines are guaranteed to have terminated at this point.
		klog.V(2).Infof("ExecuteCommand returned an an err: %v. for command '%v'. output: %v", err, command, cmdOutput)

		return fmt.Errorf("unable to exec command %v: \n%v: %w", command, cmdOutput, err)
	}

	return err
}

// ExecuteCommandAndGetOutput executes the given command in the pod's container, and returns the command stdout and stderr content
func ExecuteCommandAndGetOutput(
	kclient kclient.ClientInterface,
	podName string,
	containerName string,
	show bool,
	cmd ...string,
) (stdout []string, stderr []string, err error) {
	stdoutWriter, stdoutChan := createConsoleOutputWriterAndChannel()
	stderrWriter, stderrChan := createConsoleOutputWriterAndChannel()

	err = ExecuteCommand(cmd, kclient, podName, containerName, show, stdoutWriter, stderrWriter)

	_ = stdoutWriter.Close()
	stdout = getDataFromChannel(stdoutChan)
	_ = stderrWriter.Close()
	stderr = getDataFromChannel(stderrChan)

	return stdout, stderr, err
}

func getDataFromChannel(c <-chan string) []string {
	var result []string
	for l := range c {
		result = append(result, l)
	}
	return result
}

// createConsoleOutputWriterAndChannel is a utility function that returns a pipeWriter and a channel;
// any strings written to that PipeWriter will be output to the channel (as lines) when the
// writer closes. This is used to retrieve the stdout/stderr output from the container exec commands.
//
// The io.PipeWriter can be passed to ExecuteCommand(...) above, in order to receive the full
// stderr/stdout output from the process.
// See calling functions of CreateConsoleOutputWriterAndChannel for examples of usage.
func createConsoleOutputWriterAndChannel() (*io.PipeWriter, chan string) {
	reader, writer := io.Pipe()
	closeChannel := make(chan string, 1)

	go func() {
		bufReader := bufio.NewReader(reader)
		for {
			line, _, err := bufReader.ReadLine()
			if err != nil {
				if err != io.EOF {
					klog.V(2).Infof("Unexpected error on reading container output reader: %v", err)
				}
				break
			}
			closeChannel <- string(line)
		}
		close(closeChannel)
	}()

	return writer, closeChannel
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

			if show || log.IsDebug() {
				_, err := fmt.Fprintln(os.Stdout, line)
				if err != nil {
					log.Errorf("Unable to print to stdout: %s", err.Error())
				}
			}

			if cmdOutput != nil {
				*cmdOutput += fmt.Sprintln(line)
			}

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
