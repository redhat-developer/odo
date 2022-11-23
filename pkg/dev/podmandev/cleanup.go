package podmandev

import (
	"context"
	"fmt"
	"io"

	"k8s.io/klog"
)

func (o *DevClient) CleanupResources(ctx context.Context, out io.Writer) error {
	fmt.Printf("Cleaning up resources\n")

	if o.deployedPod == nil {
		return nil
	}

	err := o.podmanClient.PodStop(o.deployedPod.GetName())
	if err != nil {
		return err
	}
	err = o.podmanClient.PodRm(o.deployedPod.GetName())
	if err != nil {
		return err
	}

	for _, volume := range o.deployedPod.Spec.Volumes {
		if volume.PersistentVolumeClaim == nil {
			continue
		}
		volumeName := volume.PersistentVolumeClaim.ClaimName
		klog.V(3).Infof("deleting podman volume %q", volumeName)
		err = o.podmanClient.VolumeRm(volumeName)
		if err != nil {
			return err
		}
	}
	return nil
}
