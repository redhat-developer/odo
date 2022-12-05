package podman

import (
	"fmt"
	"io"
	"os/exec"

	"k8s.io/klog"
)

func (o *PodmanCli) ExecCMDInContainer(containerName, podName string, cmd []string, stdout io.Writer, stderr io.Writer, stdin io.Reader, tty bool) error {
	options := []string{}
	if tty {
		options = append(options, "--tty")
	}

	name := fmt.Sprintf("%s-%s", podName, containerName)

	args := []string{"exec", "--interactive"}
	args = append(args, options...)
	args = append(args, name)
	args = append(args, cmd...)

	command := exec.Command(o.podmanCmd, args...)
	klog.V(3).Infof("executing %v", command.Args)
	command.Stdin = stdin

	out, err := command.Output()
	if err != nil {
		return err
	}
	_, err = stdout.Write(out)
	return err
}
