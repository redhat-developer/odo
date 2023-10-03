package podman

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/api"
	odolabels "github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/odo/commonflags"
)

// ListPodsReport contains the result of the `podman pod ps --format json` command
type ListPodsReport struct {
	Name       string
	Labels     map[string]string
	Containers []ListPodsContainer `json:"Containers,omitempty"`
}

type ListPodsContainer struct {
	Names string `json:"Names,omitempty"`
}

func (o *PodmanCli) ListAllComponents() ([]api.ComponentAbstract, error) {
	cmd := exec.Command(o.podmanCmd, append(o.containerRunGlobalExtraArgs, "pod", "ps", "--format", "json", "--filter", "status=running")...)
	klog.V(3).Infof("executing %v", cmd.Args)
	out, err := cmd.Output()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf("%s: %s", err, string(exiterr.Stderr))
		}
		return nil, err
	}

	var list []ListPodsReport
	if err = json.Unmarshal(out, &list); err != nil {
		return nil, err
	}

	for _, pod := range list {
		klog.V(5).Infof("\npod name: %s", pod.Name)
		klog.V(5).Infof("labels:")
		for k, v := range pod.Labels {
			klog.V(5).Infof(" - %s: %s", k, v)
		}
	}

	var components []api.ComponentAbstract

	for _, pod := range list {

		labels := pod.Labels

		// Figure out the correct name to use
		// if there is no instance label (app.kubernetes.io/instance),
		// we SKIP the resource as it is not a component essential for Kubernetes.
		name := odolabels.GetComponentName(labels)
		if name == "" {
			continue
		}

		// Get the component type (if there is any..)
		componentType, err := odolabels.GetProjectType(labels, nil)
		if err != nil || componentType == "" {
			componentType = api.TypeUnknown
		}

		managedBy := odolabels.GetManagedBy(labels)
		managedByVersion := odolabels.GetManagedByVersion(labels)

		// Generate the appropriate "component" with all necessary information
		component := api.ComponentAbstract{
			Name:             name,
			ManagedBy:        managedBy,
			Type:             componentType,
			ManagedByVersion: managedByVersion,
			//lint:ignore SA1019 we need to output the deprecated value, before to remove it in a future release
			RunningOn: commonflags.PlatformPodman,
			Platform:  commonflags.PlatformPodman,
		}
		mode := odolabels.GetMode(labels)
		if mode != "" {
			component.RunningIn = api.NewRunningModes()
			component.RunningIn.AddRunningMode(api.RunningMode(strings.ToLower(mode)))
		}
		components = append(components, component)
	}

	return components, nil
}

func (o *PodmanCli) GetPodUsingComponentName(componentName string) (*corev1.Pod, error) {
	podSelector := fmt.Sprintf("component=%s", componentName)
	return o.GetRunningPodFromSelector(podSelector)
}
