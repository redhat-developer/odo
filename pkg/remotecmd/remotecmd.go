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

// ExecuteCommand executes the given command in the pod's container, writing the command output to the specified stdout and stderr writers
func ExecuteCommand(
	client kclient.ClientInterface,
	containerName string,
	podName string,
	command []string,
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
	if err != nil {
		// It is safe to read from cmdOutput here, as the goroutines are guaranteed to have terminated at this point.
		klog.V(2).Infof("ExecuteCommand returned an err: %v. for command '%v'. output: %v", err, command, cmdOutput)
		err = fmt.Errorf("unable to exec command %v: \n%v: %w", command, cmdOutput, err)
	}

	// Block until we have received all the container output from each stream
	_ = soutWriter.Close()
	<-stdoutCompleteChannel
	_ = serrWriter.Close()
	<-stderrCompleteChannel

	return err
}

// Execute executes the given command in the pod's container, and returns the command stdout and stderr content
func Execute(
	kclient kclient.ClientInterface,
	podName string,
	containerName string,
	show bool,
	cmd ...string,
) (stdout []string, stderr []string, err error) {
	stdoutWriter, stdoutOutputChannel := createConsoleOutputWriterAndChannel()
	stderrWriter, stderrOutputChannel := createConsoleOutputWriterAndChannel()

	err = ExecuteCommand(kclient, containerName, podName, cmd, show, stdoutWriter, stderrWriter)

	// Close the writer and wait for the console output
	_ = stdoutWriter.Close()
	stdout = <-stdoutOutputChannel

	_ = stderrWriter.Close()
	stderr = <-stderrOutputChannel

	return stdout, stderr, err
}

// createConsoleOutputWriterAndChannel is a utility function that returns a pipeWriter and a channel;
// any strings written to that PipeWriter will be output to the channel (as lines) when the
// writer closes. This is used to retrieve the stdout/stderr output from the container exec commands.
//
// The io.PipeWriter can be passed to ExecuteCommand(...) above, in order to receive the full
// stderr/stdout output from the process.
// See calling functions of CreateConsoleOutputWriterAndChannel for examples of usage.
func createConsoleOutputWriterAndChannel() (*io.PipeWriter, chan []string) {
	reader, writer := io.Pipe()
	closeChannel := make(chan []string)

	go func() {
		var consoleContents []string
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

// This goroutine will automatically pipe the output from the writer (passed into ExecCMDInContainer) to
// the loggers.
// The returned channel will contain a single nil entry once the reader has closed.
func startReaderGoroutine(reader io.Reader, show bool, cmdOutput *string, consoleOutput *io.PipeWriter) chan []string {

	result := make(chan []string)

	go func() {
		var lines []string
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			lines = append(lines, line)

			if show || log.IsDebug() {
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
		result <- lines
	}()

	return result
}
