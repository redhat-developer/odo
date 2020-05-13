package exec

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"k8s.io/klog"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/log"
)

type ExecClient interface {
	ExecCMDInContainer(common.ComponentInfo, []string, io.Writer, io.Writer, io.Reader, bool) error
}

// ExecuteCommand executes the given command in the pod's container
func ExecuteCommand(client ExecClient, compInfo common.ComponentInfo, command []string, show bool) (err error) {
	reader, writer := io.Pipe()
	var cmdOutput string

	klog.V(3).Infof("Executing command %v for pod: %v in container: %v", command, compInfo.PodName, compInfo.ContainerName)

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

			cmdOutput += fmt.Sprintln(line)
		}
	}()

	err = client.ExecCMDInContainer(compInfo, command, writer, writer, nil, false)
	if err != nil {
		log.Errorf("\nUnable to exec command %v: \n%v", command, cmdOutput)
		return err
	}

	return
}
