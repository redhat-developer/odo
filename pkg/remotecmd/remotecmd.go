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

	var cmdOutput []string

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
	soutReader, soutWriter := io.Pipe()
	serrReader, serrWriter := io.Pipe()

	klog.V(2).Infof("Executing command %v for pod: %v in container: %v", cmd, podName, containerName)

	// Read stdout and stderr, store their output in cmdOutput, and also pass output to consoleOutput Writers (if non-nil)
	stdoutCompleteChannel := startReaderGoroutine(soutReader, show, &stdout, nil)
	stderrCompleteChannel := startReaderGoroutine(serrReader, show, &stderr, nil)

	err = kclient.ExecCMDInContainer(containerName, podName, cmd, soutWriter, serrWriter, nil, false)

	// Block until we have received all the container output from each stream
	_ = soutWriter.Close()
	<-stdoutCompleteChannel
	_ = serrWriter.Close()
	<-stderrCompleteChannel

	if err != nil {
		// It is safe to read from cmdOutput here, as the goroutines are guaranteed to have terminated at this point.
		klog.V(2).Infof("ExecuteCommand returned an an err: %v. for command '%v'\nstdout: %v\nstderr: %v",
			err, cmd, stdout, stderr)

		return stdout, stderr, fmt.Errorf("unable to exec command %v: \n=== stdout===\n%v\n=== stderr===\n%v: %w",
			cmd, stdout, stderr, err)
	}

	return stdout, stderr, err
}

// This goroutine will automatically pipe the output from the writer (passed into ExecCMDInContainer) to
// the loggers.
// The returned channel will contain a single nil entry once the reader has closed.
func startReaderGoroutine(reader io.Reader, show bool, cmdOutput *[]string, consoleOutput *io.PipeWriter) chan interface{} {
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
				*cmdOutput = append(*cmdOutput, line)
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
