package kubedev

import (
	"context"

	"github.com/redhat-developer/odo/pkg/dev/common"
	"k8s.io/klog"
)

func (o *DevClient) Run(
	ctx context.Context,
	commandName string,
) error {
	klog.V(4).Infof("running command %q on cluster", commandName)
	return common.Run(
		ctx,
		commandName,
		o.kubernetesClient,
		o.execClient,
		o.configAutomountClient,
		o.filesystem,
	)
}
