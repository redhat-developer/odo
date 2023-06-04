package exec

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"k8s.io/klog"
	"k8s.io/kubectl/pkg/util/term"

	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/platform"
)

type ExecClient struct {
	platformClient platform.Client
}

func NewExecClient(platformClient platform.Client) *ExecClient {
	return &ExecClient{
		platformClient: platformClient,
	}
}

// ExecuteCommand executes the given command in the pod's container,
// writing the output to the specified respective pipe writers
// when directRun is true, will execute the command with terminal in Raw mode and connected to local standard I/Os
// so input, including Ctrl-c, is sent to the remote process
func (o ExecClient) ExecuteCommand(ctx context.Context, command []string, podName string, containerName string, directRun bool, stdoutWriter *io.PipeWriter, stderrWriter *io.PipeWriter) (stdout []string, stderr []string, err error) {
	if !directRun {
		soutReader, soutWriter := io.Pipe()
		serrReader, serrWriter := io.Pipe()

		klog.V(2).Infof("Executing command %v for pod: %v in container: %v", command, podName, containerName)

		// Read stdout and stderr, store their output in cmdOutput, and also pass output to consoleOutput Writers (if non-nil)
		stdoutCompleteChannel := startReaderGoroutine(os.Stdout, soutReader, directRun, &stdout, stdoutWriter)
		stderrCompleteChannel := startReaderGoroutine(os.Stderr, serrReader, directRun, &stderr, stderrWriter)

		err = o.platformClient.ExecCMDInContainer(ctx, containerName, podName, command, soutWriter, serrWriter, nil, false)

		// Block until we have received all the container output from each stream
		_ = soutWriter.Close()
		<-stdoutCompleteChannel
		_ = serrWriter.Close()
		<-stderrCompleteChannel

		// Details are displayed only if no outputs are displayed
		if err != nil && !directRun {
			// It is safe to read from stdout and stderr here, as the goroutines are guaranteed to have terminated at this point.
			klog.V(2).Infof("ExecuteCommand returned an an err: %v. for command '%v'\nstdout: %v\nstderr: %v",
				err, command, stdout, stderr)

			msg := fmt.Sprintf("unable to exec command %v", command)
			if len(stdout) != 0 {
				msg += fmt.Sprintf("\n=== stdout===\n%s", strings.Join(stdout, "\n"))
			}
			if len(stderr) != 0 {
				msg += fmt.Sprintf("\n=== stderr===\n%s", strings.Join(stderr, "\n"))
			}
			return stdout, stderr, fmt.Errorf("%s: %w", msg, err)
		}

		return stdout, stderr, err
	}

	tty := term.TTY{
		Raw: true,
		In:  os.Stdin,
		Out: os.Stdout,
	}

	fn := func() error {
		return o.platformClient.ExecCMDInContainer(ctx, containerName, podName, command, tty.Out, nil, tty.In, true)
	}

	return nil, nil, tty.Safe(fn)
}

// This goroutine will automatically pipe the output from the writer (passed into ExecCMDInContainer) to
// the loggers.
// The returned channel will contain a single nil entry once the reader has closed.
func startReaderGoroutine(logWriter io.Writer, reader io.Reader, show bool, cmdOutput *[]string, consoleOutput *io.PipeWriter) chan interface{} {
	result := make(chan interface{})

	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()

			if show {
				_, err := fmt.Fprintln(logWriter, line)
				if err != nil {
					log.Errorf("Unable to print to stdout: %s", err.Error())
				}
			} else {
				klog.V(2).Infof(line)
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
