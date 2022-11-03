package podmandev

import (
	"context"
	"fmt"
	"io"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/library/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/dev"
	"github.com/redhat-developer/odo/pkg/dev/common"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/podman"
	"github.com/redhat-developer/odo/pkg/sync"
)

type DevClient struct {
	podmanClient podman.Client
	syncClient   sync.Client
	execClient   exec.Client
}

var _ dev.Client = (*DevClient)(nil)

func NewDevClient(
	podmanClient podman.Client,
	syncClient sync.Client,
	execClient exec.Client,
) *DevClient {
	return &DevClient{
		podmanClient: podmanClient,
		syncClient:   syncClient,
		execClient:   execClient,
	}
}

func (o *DevClient) Start(
	ctx context.Context,
	devfileObj parser.DevfileObj,
	componentName string,
	path string,
	devfilePath string,
	out io.Writer,
	errOut io.Writer,
	options dev.StartOptions,
) error {
	fmt.Printf("Deploying using Podman\n\n")

	pod, err := createPodFromComponent(
		devfileObj,
		componentName,
		"app",
		options.BuildCommand,
		options.RunCommand,
		"",
	)
	if err != nil {
		return err
	}

	err = o.podmanClient.PlayKube(pod)
	if err != nil {
		return err
	}

	containerName, syncFolder, err := common.GetFirstContainerWithSourceVolume(pod.Spec.Containers)
	if err != nil {
		return fmt.Errorf("error while retrieving container from pod %s with a mounted project volume: %w", pod.GetName(), err)
	}

	compInfo := sync.ComponentInfo{
		ComponentName: componentName,
		ContainerName: containerName,
		PodName:       pod.GetName(),
		SyncFolder:    syncFolder,
	}

	syncParams := sync.SyncParameters{
		Path:                     path,
		WatchFiles:               nil,
		WatchDeletedFiles:        nil,
		IgnoredFiles:             options.IgnorePaths,
		DevfileScanIndexForWatch: true,

		CompInfo:  compInfo,
		ForcePush: true,
		Files:     map[string]string{}, // ??? TODO
	}
	execRequired, err := o.syncClient.SyncFiles(syncParams)
	if err != nil {
		return err
	}
	if execRequired {
		doExecuteBuildCommand := func() error {
			execHandler := component.NewExecHandler(
				nil, /* TODO */
				o.execClient,
				"app", /* TODO */
				componentName,
				pod.Name,
				"Building your application in container on cluster",
				false, /* TODO */
			)
			return libdevfile.Build(devfileObj, options.BuildCommand, execHandler)
		}
		err = doExecuteBuildCommand()
		if err != nil {
			return err
		}

		cmdKind := devfilev1.RunCommandGroupKind
		cmdName := options.RunCommand
		if options.Debug {
			cmdKind = devfilev1.DebugCommandGroupKind
			cmdName = options.DebugCommand
		}
		cmdHandler := commandHandler{
			execClient:      o.execClient,
			platformClient:  o.podmanClient,
			componentExists: false, // TODO
			podName:         pod.Name,
			appName:         "app", // TODO
			componentName:   componentName,
		}
		err = libdevfile.ExecuteCommandByNameAndKind(devfileObj, cmdName, cmdKind, &cmdHandler, false)
		if err != nil {
			return err
		}
	}

	<-ctx.Done()

	fmt.Printf("Cleaning up resources\n")
	err = o.podmanClient.PodStop(pod.GetName())
	if err != nil {
		return err
	}
	err = o.podmanClient.PodRm(pod.GetName())
	if err != nil {
		return err
	}

	for _, volume := range pod.Spec.Volumes {
		err = o.podmanClient.VolumeRm(volume.Name)
		if err != nil {
			return err
		}
	}

	return nil
}
