package podmandev

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

	pod, err := o.podmanClient.GetPodUsingComponentName(componentName)
	if err != nil {
		return fmt.Errorf("unable to get pod for component %s: %w", componentName, err)
	}

	handler := component.NewRunHandler(
		ctx,
		o.podmanClient,
		o.execClient,
		nil, // TODO(feloy) set when running on new container is supported on podman
		pod.Name,
		false,
		component.GetContainersNames(pod),
		"Executing command in container",

		o.fs,
		image.SelectBackend(ctx),
		*devfileObj,
		devfilePath,
	)

	return libdevfile.ExecuteCommandByName(ctx, *devfileObj, commandName, handler, false)
}
