package podmandev

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/fatih/color"

	"github.com/redhat-developer/odo/pkg/api"
	"github.com/redhat-developer/odo/pkg/component"
	"github.com/redhat-developer/odo/pkg/dev"
	"github.com/redhat-developer/odo/pkg/dev/common"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/podman"
	"github.com/redhat-developer/odo/pkg/state"
	"github.com/redhat-developer/odo/pkg/sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

const (
	PromptMessage = `
[Ctrl+c] - Exit and delete resources from podman
`
)

type DevClient struct {
	podmanClient podman.Client
	syncClient   sync.Client
	execClient   exec.Client
	stateClient  state.Client
}

var _ dev.Client = (*DevClient)(nil)

func NewDevClient(
	podmanClient podman.Client,
	syncClient sync.Client,
	execClient exec.Client,
	stateClient state.Client,
) *DevClient {
	return &DevClient{
		podmanClient: podmanClient,
		syncClient:   syncClient,
		execClient:   execClient,
		stateClient:  stateClient,
	}
}

func (o *DevClient) Start(
	ctx context.Context,
	out io.Writer,
	errOut io.Writer,
	options dev.StartOptions,
) error {
	var (
		appName       = odocontext.GetApplication(ctx)
		componentName = odocontext.GetComponentName(ctx)
		devfileObj    = odocontext.GetDevfileObj(ctx)
		devfilePath   = odocontext.GetDevfilePath(ctx)
		path          = filepath.Dir(devfilePath)
	)

	pod, fwPorts, err := o.deployPod(ctx, options)
	if err != nil {
		return err
	}

	for _, fwPort := range fwPorts {
		s := fmt.Sprintf("Forwarding from %s:%d -> %d", fwPort.LocalAddress, fwPort.LocalPort, fwPort.ContainerPort)
		fmt.Fprintf(out, " -  %s", log.SboldColor(color.FgGreen, s))
	}
	err = o.stateClient.SetForwardedPorts(fwPorts)
	if err != nil {
		return err
	}

	execRequired, err := o.syncFiles(ctx, options, pod, path)
	if err != nil {
		return err
	}

	// PostStart events from the devfile will only be executed when the component
	// didn't previously exist
	if libdevfile.HasPostStartEvents(*devfileObj) {
		execHandler := component.NewExecHandler(
			o.podmanClient,
			o.execClient,
			appName,
			componentName,
			pod.Name,
			"",
			false, /* TODO */
		)
		err = libdevfile.ExecPostStartEvents(*devfileObj, execHandler)
		if err != nil {
			return err
		}
	}

	if execRequired {
		doExecuteBuildCommand := func() error {
			execHandler := component.NewExecHandler(
				o.podmanClient,
				o.execClient,
				appName,
				componentName,
				pod.Name,
				"Building your application in container",
				false, /* TODO */
			)
			return libdevfile.Build(*devfileObj, options.BuildCommand, execHandler)
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
			appName:         appName,
			componentName:   componentName,
		}
		err = libdevfile.ExecuteCommandByNameAndKind(*devfileObj, cmdName, cmdKind, &cmdHandler, false)
		if err != nil {
			return err
		}
	}

	fmt.Fprintf(
		out,
		" %s%s",
		log.Sbold("Keyboard Commands:"),
		PromptMessage,
	)

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

// deployPod deploys the component as a Pod in podman
func (o *DevClient) deployPod(ctx context.Context, options dev.StartOptions) (*corev1.Pod, []api.ForwardedPort, error) {
	var (
		appName       = odocontext.GetApplication(ctx)
		componentName = odocontext.GetComponentName(ctx)
		devfileObj    = odocontext.GetDevfileObj(ctx)
	)

	spinner := log.Spinner("Deploying pod")
	defer spinner.End(false)

	pod, fwPorts, err := createPodFromComponent(
		*devfileObj,
		componentName,
		appName,
		options.BuildCommand,
		options.RunCommand,
		"",
	)
	if err != nil {
		return nil, nil, err
	}

	err = o.checkVolumesFree(pod)
	if err != nil {
		return nil, nil, err
	}

	err = o.podmanClient.PlayKube(pod)
	if err != nil {
		return nil, nil, err
	}

	spinner.End(true)
	return pod, fwPorts, nil
}

// syncFiles syncs the local source files in path into the pod's source volume
func (o *DevClient) syncFiles(ctx context.Context, options dev.StartOptions, pod *corev1.Pod, path string) (bool, error) {
	var (
		componentName = odocontext.GetComponentName(ctx)
	)

	containerName, syncFolder, err := common.GetFirstContainerWithSourceVolume(pod.Spec.Containers)
	if err != nil {
		return false, fmt.Errorf("error while retrieving container from pod %s with a mounted project volume: %w", pod.GetName(), err)
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
		return false, err
	}
	return execRequired, nil
}

// checkVolumesFree checks that all persistent volumes declared in pod
// are not using an existing volume
func (o *DevClient) checkVolumesFree(pod *corev1.Pod) error {
	existingVolumesSet, err := o.podmanClient.VolumeLs()
	if err != nil {
		return err
	}
	var problematicVolumes []string
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil && existingVolumesSet[volume.PersistentVolumeClaim.ClaimName] {
			problematicVolumes = append(problematicVolumes, volume.PersistentVolumeClaim.ClaimName)
		}
	}
	if len(problematicVolumes) > 0 {
		return fmt.Errorf("volumes already exist, please remove them before to run odo dev: %s", strings.Join(problematicVolumes, ", "))
	}
	return nil
}
