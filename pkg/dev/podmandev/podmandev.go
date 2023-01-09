package podmandev

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/devfile/library/pkg/devfile/parser"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/dev"
	"github.com/redhat-developer/odo/pkg/dev/common"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/log"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/podman"
	"github.com/redhat-developer/odo/pkg/state"
	"github.com/redhat-developer/odo/pkg/sync"
	"github.com/redhat-developer/odo/pkg/watch"

	corev1 "k8s.io/api/core/v1"
)

const (
	promptMessage = `
[Ctrl+c] - Exit and delete resources from podman
     [p] - Manually apply local changes to the application on podman
`
)

type DevClient struct {
	podmanClient podman.Client
	syncClient   sync.Client
	execClient   exec.Client
	stateClient  state.Client
	watchClient  watch.Client

	deployedPod *corev1.Pod
	usedPorts   []int
}

var _ dev.Client = (*DevClient)(nil)

func NewDevClient(
	podmanClient podman.Client,
	syncClient sync.Client,
	execClient exec.Client,
	stateClient state.Client,
	watchClient watch.Client,
) *DevClient {
	return &DevClient{
		podmanClient: podmanClient,
		syncClient:   syncClient,
		execClient:   execClient,
		stateClient:  stateClient,
		watchClient:  watchClient,
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

		componentStatus = watch.ComponentStatus{}
	)

	err := o.reconcile(ctx, out, errOut, options, &componentStatus)
	if err != nil {
		return err
	}

	watch.PrintInfoMessage(out, path, options.WatchFiles, promptMessage)

	watchParameters := watch.WatchParameters{
		DevfilePath:         devfilePath,
		Path:                path,
		ComponentName:       componentName,
		ApplicationName:     appName,
		InitialDevfileObj:   *devfileObj,
		DevfileWatchHandler: o.watchHandler,
		FileIgnores:         options.IgnorePaths,
		Debug:               options.Debug,
		DevfileBuildCmd:     options.BuildCommand,
		DevfileRunCmd:       options.RunCommand,
		Variables:           options.Variables,
		RandomPorts:         options.RandomPorts,
		WatchFiles:          options.WatchFiles,
		WatchCluster:        false,
		Out:                 out,
		ErrOut:              errOut,
		PromptMessage:       promptMessage,
	}

	return o.watchClient.WatchAndPush(out, watchParameters, ctx, componentStatus)
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

func (o *DevClient) watchHandler(ctx context.Context, pushParams adapters.PushParameters, watchParams watch.WatchParameters, componentStatus *watch.ComponentStatus) error {
	printWarningsOnDevfileChanges(ctx, watchParams)

	startOptions := dev.StartOptions{
		IgnorePaths:  watchParams.FileIgnores,
		Debug:        watchParams.Debug,
		BuildCommand: watchParams.DevfileBuildCmd,
		RunCommand:   watchParams.DevfileRunCmd,
		RandomPorts:  watchParams.RandomPorts,
		WatchFiles:   watchParams.WatchFiles,
		Variables:    watchParams.Variables,
	}
	return o.reconcile(ctx, watchParams.Out, watchParams.ErrOut, startOptions, componentStatus)
}

func printWarningsOnDevfileChanges(ctx context.Context, parameters watch.WatchParameters) {
	var warning string
	currentDevfile := odocontext.GetDevfileObj(ctx)
	newDevfile, err := devfile.ParseAndValidateFromFileWithVariables(location.DevfileLocation(""), parameters.Variables)
	if err != nil {
		warning = fmt.Sprintf("error while reading the Devfile. Please restart 'odo dev' if you made any changes to the Devfile. Error message is: %v", err)
	} else {
		devfileEquals := func(d1, d2 parser.DevfileObj) (bool, error) {
			// Compare two Devfile objects by comparing the result of their JSON encoding,
			// because reflect.DeepEqual does not work properly with the parser.DevfileObj structure.
			d1Json, jsonErr := json.Marshal(d1.Data)
			if jsonErr != nil {
				return false, jsonErr
			}
			d2Json, jsonErr := json.Marshal(d2.Data)
			if jsonErr != nil {
				return false, jsonErr
			}
			return bytes.Equal(d1Json, d2Json), nil
		}
		equal, eqErr := devfileEquals(*currentDevfile, newDevfile)
		if eqErr != nil {
			klog.V(5).Infof("error while checking if Devfile has changed: %v", eqErr)
		} else if !equal {
			warning = "Detected changes in the Devfile, but this is not supported yet on Podman. Please restart 'odo dev' for such changes to be applied."
		}
	}
	if warning != "" {
		log.Fwarning(parameters.Out, warning+"\n")
	}
}
