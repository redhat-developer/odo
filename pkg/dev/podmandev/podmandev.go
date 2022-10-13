package podmandev

import (
	"context"
	"fmt"
	"io"

	"github.com/devfile/library/pkg/devfile/parser"
	"github.com/redhat-developer/odo/pkg/dev"
	"github.com/redhat-developer/odo/pkg/dev/common"
	"github.com/redhat-developer/odo/pkg/exec"
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
	_ = execRequired

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
