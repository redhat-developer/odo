package common

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"k8s.io/klog"

	"github.com/pkg/errors"
	"github.com/redhat-developer/odo/pkg/log"
)

// ExecClient  is a wrapper around ExecCMDInContainer which executes a command in a specific container of a pod.
type ExecClient interface {
	ExecCMDInContainer(ComponentInfo, []string, io.Writer, io.Writer, io.Reader, bool) error
}

// ExecuteCommand executes the given command in the pod's container
func ExecuteCommand(client ExecClient, compInfo ComponentInfo, command []string, show bool, consoleOutputStdout *io.PipeWriter, consoleOutputStderr *io.PipeWriter) (err error) {
	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	var cmdOutput string

	klog.V(2).Infof("Executing command %v for pod: %v in container: %v", command, compInfo.PodName, compInfo.ContainerName)

	// Read stdout and stderr, store their output in cmdOutput, and also pass output to consoleOutput Writers (if non-nil)
	stdoutCompleteChannel := startReaderGoroutine(stdoutReader, show, &cmdOutput, consoleOutputStdout)
	stderrCompleteChannel := startReaderGoroutine(stderrReader, show, &cmdOutput, consoleOutputStderr)

	err = client.ExecCMDInContainer(compInfo, command, stdoutWriter, stderrWriter, nil, false)

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

// CreateConsoleOutputWriterAndChannel is a utility function that returns a pipeWriter and a channel;
// any strings written to that PipeWriter will be output to the channel (as lines) when the
// writer closes. This is used to retrieve the stdout/stderr output from the container exec commands.
//
// The io.PipeWriter can be passed to ExecuteCommand(...) above, in order to receive the full
// stderr/stdout output from the process.
// See calling functions of CreateConsoleOutputWriterAndChannel for examples of usage.
func CreateConsoleOutputWriterAndChannel() (*io.PipeWriter, chan []string) {
	reader, writer := io.Pipe()

	closeChannel := make(chan []string)

	go func() {

		consoleContents := []string{}

		bufReader := bufio.NewReader(reader)
		for {
			line, _, err := bufReader.ReadLine()
			if err != nil {
				if err != io.EOF {
					klog.V(2).Infof("Unexpected error on reading container output reader: %v", err)
				}

				break
			}

			consoleContents = append(consoleContents, string(line))
		}

		// Output the final console contents to the channel
		closeChannel <- consoleContents
	}()

	return writer, closeChannel
}
