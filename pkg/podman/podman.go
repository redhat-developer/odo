package podman

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	corev1 "k8s.io/api/core/v1"
	jsonserializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/klog"
	"k8s.io/kubectl/pkg/scheme"
)

type PodmanCli struct{}

func NewPodmanCli() *PodmanCli {
	return &PodmanCli{}
}

func (o *PodmanCli) PlayKube(pod *corev1.Pod) error {
	serializer := jsonserializer.NewSerializerWithOptions(
		jsonserializer.SimpleMetaFactory{},
		scheme.Scheme,
		scheme.Scheme,
		jsonserializer.SerializerOptions{
			Yaml: true,
		},
	)

	err := serializer.Encode(pod, os.Stdin)
	if err != nil {
		return err
	}

	cmd := exec.Command("podman", "play", "kube", "-")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout

	if err = cmd.Start(); err != nil {
		return err
	}

	err = serializer.Encode(pod, stdin)
	if err != nil {
		return err
	}
	stdin.Close()

	go func() {
		for {
			tmp := make([]byte, 1024)
			_, err = stdout.Read(tmp)
			fmt.Print(string(tmp))
			if err != nil {
				break
			}
		}
	}()
	if err = cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func (o *PodmanCli) PodStop(podname string) error {
	out, err := exec.Command("podman", "pod", "stop", podname).Output()
	fmt.Printf("%s\n", string(out))
	return err
}

func (o *PodmanCli) PodRm(podname string) error {
	out, err := exec.Command("podman", "pod", "rm", podname).Output()
	fmt.Printf("%s\n", string(out))
	return err
}

func (o *PodmanCli) VolumeRm(volumeName string) error {
	out, err := exec.Command("podman", "volume", "rm", volumeName).Output()
	fmt.Printf("%s\n", string(out))
	return err
}

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

	command := exec.Command("podman", args...)
	command.Stdin = stdin

	klog.V(4).Infof("exec podman %v\n", args)
	out, err := command.Output()
	if err != nil {
		return err
	}
	_, err = stdout.Write(out)
	return err
}
