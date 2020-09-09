package common

import "k8s.io/klog"

// IsContainer checks if the component is a container
func (component DevfileComponent) IsContainer() bool {
	// Currently odo only uses devfile components of type container, since most of the Che registry devfiles use it
	if component.Container != nil {
		klog.V(2).Infof("Found component \"%v\" with name \"%v\"\n", ContainerComponentType, component.Name)
		return true
	}
	return false
}

// IsVolume checks if the component is a volume
func (component DevfileComponent) IsVolume() bool {
	if component.Volume != nil {
		klog.V(2).Infof("Found component \"%v\" with name \"%v\"\n", VolumeComponentType, component.Name)
		return true
	}
	return false
}
