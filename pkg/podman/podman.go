package podman

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"

	envcontext "github.com/redhat-developer/odo/pkg/config/context"
	"github.com/redhat-developer/odo/pkg/platform"

	corev1 "k8s.io/api/core/v1"
	jsonserializer "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/klog"
	"k8s.io/kubectl/pkg/scheme"
)

type PodmanCli struct {
	podmanCmd                   string
	containerRunGlobalExtraArgs []string
	containerRunExtraArgs       []string
}

var _ Client = (*PodmanCli)(nil)
var _ platform.Client = (*PodmanCli)(nil)

// NewPodmanCli returns a new podman client, or nil if the podman command is not accessible in the system
func NewPodmanCli(ctx context.Context) (*PodmanCli, error) {
	// Check if podman is available in the system
	cli := &PodmanCli{
		podmanCmd:                   envcontext.GetEnvConfig(ctx).PodmanCmd,
		containerRunGlobalExtraArgs: envcontext.GetEnvConfig(ctx).OdoContainerBackendGlobalArgs,
		containerRunExtraArgs:       envcontext.GetEnvConfig(ctx).OdoContainerRunArgs,
	}
	version, err := cli.Version()
	if err != nil {
		return nil, err
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

	// +3 because of "play kube -"
	args := make([]string, 0, len(o.containerRunGlobalExtraArgs)+len(o.containerRunExtraArgs)+3)
	args = append(args, o.containerRunGlobalExtraArgs...)
	args = append(args, "play", "kube")
	args = append(args, o.containerRunExtraArgs...)
	args = append(args, "-")

	cmd := exec.Command(o.podmanCmd, args...)
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

	if klog.V(4) {
		var sb strings.Builder
		_ = serializer.Encode(pod, &sb)
		klog.Infof("Pod spec to play: \n---\n%s\n---\n", sb.String())
	}

	err = serializer.Encode(pod, stdin)
	if err != nil {
		return err
	}
	stdin.Close()
	var podmanOut string
	go func() {
		for {
			tmp := make([]byte, 1024)
			_, err = stdout.Read(tmp)
			podmanOut += string(tmp)
			klog.V(4).Info(string(tmp))
			if err != nil {
				break
			}
		}
	}()
	if err = cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s\nComplete Podman output:\n%s", err, string(exiterr.Stderr), podmanOut)
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

	cmd := exec.Command(o.podmanCmd, append(o.containerRunGlobalExtraArgs, "generate", "kube", name)...)
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
	cmd := exec.Command(o.podmanCmd, append(o.containerRunGlobalExtraArgs, "pod", "stop", podname)...)
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
	cmd := exec.Command(o.podmanCmd, append(o.containerRunGlobalExtraArgs, "pod", "rm", podname)...)
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
	cmd := exec.Command(o.podmanCmd, append(o.containerRunGlobalExtraArgs, "pod", "list", "--format", "{{.Name}}", "--noheading")...)
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
	cmd := exec.Command(o.podmanCmd, append(o.containerRunGlobalExtraArgs, "volume", "rm", volumeName)...)
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
	cmd := exec.Command(o.podmanCmd, append(o.containerRunGlobalExtraArgs, "volume", "ls", "--format", "{{.Name}}", "--noheading")...)
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
