package podmandev

import (
	"context"

	"github.com/redhat-developer/odo/pkg/dev/common"
	"k8s.io/klog"
)

func (o *DevClient) Run(
	ctx context.Context,
	commandName string,
) error {
	klog.V(4).Infof("running command %q on podman", commandName)
	return common.Run(
		ctx,
		commandName,
		o.podmanClient,
		o.execClient,
		nil, // TODO(feloy) set when running on new container is supported on podman
		o.fs,
	)
}
