package podman

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"

	envcontext "github.com/redhat-developer/odo/pkg/config/context"

	corev1 "k8s.io/api/core/v1"
	jsonserializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/klog"
	"k8s.io/kubectl/pkg/scheme"
)

type PodmanCli struct {
	podmanCmd string
}

// NewPodmanCli returns a new podman client, or nil if the podman command is not accessible in the system
func NewPodmanCli(ctx context.Context) (*PodmanCli, error) {
	// Check if podman is available in the system
	cli := &PodmanCli{
		podmanCmd: envcontext.GetEnvConfig(ctx).PodmanCmd,
	}
	version, err := cli.Version()
	if err != nil {
		return nil, fmt.Errorf("executable %q not found", cli.podmanCmd)
	}
	if version.Client == nil {
		return nil, fmt.Errorf("executable %q not recognized as podman client", cli.podmanCmd)
	}

	return cli, nil
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

	cmd := exec.Command(o.podmanCmd, "play", "kube", "-")
	klog.V(3).Infof("executing %v", cmd.Args)
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
			klog.V(4).Info(string(tmp))
			if err != nil {
				break
			}
		}
	}()
	if err = cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		return err
	}

	return nil
}

func (o *PodmanCli) KubeGenerate(name string) (*corev1.Pod, error) {
	serializer := jsonserializer.NewSerializerWithOptions(
		jsonserializer.SimpleMetaFactory{},
		scheme.Scheme,
		scheme.Scheme,
		jsonserializer.SerializerOptions{
			Yaml: true,
		},
	)

	cmd := exec.Command(o.podmanCmd, "generate", "kube", name)
	klog.V(3).Infof("executing %v", cmd.Args)
	resultBytes, err := cmd.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		return nil, err
	}
	var pod corev1.Pod
	_, _, err = serializer.Decode(resultBytes, nil, &pod)
	if err != nil {
		return nil, err
	}
	return &pod, nil
}

func (o *PodmanCli) PodStop(podname string) error {
	cmd := exec.Command(o.podmanCmd, "pod", "stop", podname)
	klog.V(3).Infof("executing %v", cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		return err
	}
	klog.V(4).Infof("Stopped pod %s", string(out))
	return nil
}

func (o *PodmanCli) PodRm(podname string) error {
	cmd := exec.Command(o.podmanCmd, "pod", "rm", podname)
	klog.V(3).Infof("executing %v", cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		return err
	}
	klog.V(4).Infof("Deleted pod %s", string(out))
	return nil
}

func (o *PodmanCli) PodLs() (map[string]bool, error) {
	cmd := exec.Command(o.podmanCmd, "pod", "list", "--format", "{{.Name}}", "--noheading")
	klog.V(3).Infof("executing %v", cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		return nil, err
	}
	return SplitLinesAsSet(string(out)), nil
}

func (o *PodmanCli) VolumeRm(volumeName string) error {
	cmd := exec.Command(o.podmanCmd, "volume", "rm", volumeName)
	klog.V(3).Infof("executing %v", cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		return err
	}
	klog.V(4).Infof("Deleted volume %s", string(out))
	return nil
}

func (o *PodmanCli) VolumeLs() (map[string]bool, error) {
	cmd := exec.Command(o.podmanCmd, "volume", "ls", "--format", "{{.Name}}", "--noheading")
	klog.V(3).Infof("executing %v", cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		return nil, err
	}
	return SplitLinesAsSet(string(out)), nil
}

func (o *PodmanCli) CleanupPodResources(pod *corev1.Pod) error {
	err := o.PodStop(pod.GetName())
	if err != nil {
		return err
	}
	err = o.PodRm(pod.GetName())
	if err != nil {
		return err
	}

	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim == nil {
			continue
		}
		volumeName := volume.PersistentVolumeClaim.ClaimName
		klog.V(3).Infof("deleting podman volume %q", volumeName)
		err = o.VolumeRm(volumeName)
		if err != nil {
			return err
		}
	}
	return nil
}

func SplitLinesAsSet(s string) map[string]bool {
	lines := map[string]bool{}
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		lines[sc.Text()] = true
	}
	return lines
}
