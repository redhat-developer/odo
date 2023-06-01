package common

import (
	"context"
	"fmt"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/configAutomount"
	"github.com/redhat-developer/odo/pkg/devfile/image"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/platform"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
)

func Run(
	ctx context.Context,
	commandName string,
	platformClient platform.Client,
	execClient exec.Client,
	configAutomountClient configAutomount.Client,
	filesystem filesystem.Filesystem,
) error {
	var (
		componentName = odocontext.GetComponentName(ctx)
		devfileObj    = odocontext.GetEffectiveDevfileObj(ctx)
		devfilePath   = odocontext.GetDevfilePath(ctx)
	)

	pod, err := platformClient.GetPodUsingComponentName(componentName)
	if err != nil {
		return fmt.Errorf("unable to get pod for component %s: %w", componentName, err)
	}

	handler := component.NewRunHandler(
		ctx,
		platformClient,
		execClient,
		configAutomountClient,
		pod.Name,
		false,
		component.GetContainersNames(pod),
		"Executing command in container",

		filesystem,
		image.SelectBackend(ctx),
		*devfileObj,
		devfilePath,
	)

	return libdevfile.ExecuteCommandByName(ctx, *devfileObj, commandName, handler, false)
}
