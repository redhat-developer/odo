package kubedev

import (
	"context"
	"fmt"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/devfile/image"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"k8s.io/klog"
)

func (o *DevClient) Run(
	ctx context.Context,
	commandName string,
) error {
	var (
		componentName = odocontext.GetComponentName(ctx)
		devfileObj    = odocontext.GetEffectiveDevfileObj(ctx)
		devfilePath   = odocontext.GetDevfilePath(ctx)
	)

	klog.V(4).Infof("running command %q on cluster", commandName)

	pod, err := o.kubernetesClient.GetPodUsingComponentName(componentName)
	if err != nil {
		return fmt.Errorf("unable to get pod for component %s: %w", componentName, err)
	}

	handler := component.NewRunHandler(
		ctx,
		o.kubernetesClient,
		o.execClient,
		o.configAutomountClient,
		pod.Name,
		false,
		component.GetContainersNames(pod),
		"Executing command in container",

		o.filesystem,
		image.SelectBackend(ctx),
		*devfileObj,
		devfilePath,
	)

	return libdevfile.ExecuteCommandByName(ctx, *devfileObj, commandName, handler, false)
}
