package kubedev

import (
	"context"

	"k8s.io/klog"
)

func (o *DevClient) Run(
	ctx context.Context,
	commandName string,
) error {
	klog.V(4).Infof("running command %q on cluster", commandName)
	return nil
}
