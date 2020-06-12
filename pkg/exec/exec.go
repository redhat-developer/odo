package exec

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"

	"k8s.io/klog"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/log"
)

// ExecClient  is a wrapper around ExecCMDInContainer which executes a command in a specific container of a pod.
type ExecClient interface {
	ExecCMDInContainer(common.ComponentInfo, []string, io.Writer, io.Writer, io.Reader, bool) error
}

// ExecuteCommand executes the given command in the pod's container
func ExecuteCommand(client ExecClient, compInfo common.ComponentInfo, command []string, show bool) (err error) {

	// Create a mutex so we don't run into the issue of go routines trying to write to stdout
	var mu sync.Mutex

	reader, writer := io.Pipe()
	var cmdOutput string

	klog.V(4).Infof("Executing command %v for pod: %v in container: %v", command, compInfo.PodName, compInfo.ContainerName)

	// This Go routine will automatically pipe the output from ExecCMDInContainer to
	// our logger.
	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()

			if log.IsDebug() || show {
				_, err := fmt.Fprintln(os.Stdout, line)
				if err != nil {
					log.Errorf("Unable to print to stdout: %v", err)
				}
			}

			mu.Lock()
			cmdOutput += fmt.Sprintln(line)
			mu.Unlock()
		}
	}()

	err = client.ExecCMDInContainer(compInfo, command, writer, writer, nil, false)
	if err != nil {
		mu.Lock()
		log.Errorf("\nUnable to exec command %v: \n%v", command, cmdOutput)
		mu.Unlock()
		return err
	}

	return
}
