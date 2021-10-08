package common

import (
	"fmt"

	v1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
)

// IsContainer checks if the component is a container
func IsContainer(component v1.Component) bool {
	return component.Container != nil
}

// IsVolume checks if the component is a volume
func IsVolume(component v1.Component) bool {
	return component.Volume != nil
}

// GetComponentType returns the component type of a given component
func GetComponentType(component v1.Component) (v1.ComponentType, error) {
	switch {
	case component.Container != nil:
		return v1.ContainerComponentType, nil
	case component.Volume != nil:
		return v1.VolumeComponentType, nil
	case component.Plugin != nil:
		return v1.PluginComponentType, nil
	case component.Kubernetes != nil:
		return v1.KubernetesComponentType, nil
	case component.Openshift != nil:
		return v1.OpenshiftComponentType, nil
	case component.Image != nil:
		return v1.ImageComponentType, nil
	case component.Custom != nil:
		return v1.CustomComponentType, nil

	default:
		return "", fmt.Errorf("unknown component type")
	}
}
