package podmandev

import (
	"context"
	"fmt"
	"strings"

	devfilev1 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/dev"
	"github.com/redhat-developer/odo/pkg/dev/common"
	"github.com/redhat-developer/odo/pkg/devfile"
	"github.com/redhat-developer/odo/pkg/devfile/location"
	"github.com/redhat-developer/odo/pkg/exec"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"
	"github.com/redhat-developer/odo/pkg/podman"
	"github.com/redhat-developer/odo/pkg/portForward"
	"github.com/redhat-developer/odo/pkg/preference"
	"github.com/redhat-developer/odo/pkg/state"
	"github.com/redhat-developer/odo/pkg/sync"
	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/watch"

	corev1 "k8s.io/api/core/v1"
)

type DevClient struct {
	fs filesystem.Filesystem

	podmanClient      podman.Client
	prefClient        preference.Client
	portForwardClient portForward.Client
	syncClient        sync.Client
	execClient        exec.Client
	stateClient       state.Client
	watchClient       watch.Client

	deployedPod *corev1.Pod
	usedPorts   []int
}

var _ dev.Client = (*DevClient)(nil)

func NewDevClient(
	fs filesystem.Filesystem,
	podmanClient podman.Client,
	prefClient preference.Client,
	portForwardClient portForward.Client,
	syncClient sync.Client,
	execClient exec.Client,
	stateClient state.Client,
	watchClient watch.Client,
) *DevClient {
	return &DevClient{
		fs:                fs,
		podmanClient:      podmanClient,
		prefClient:        prefClient,
		portForwardClient: portForwardClient,
		syncClient:        syncClient,
		execClient:        execClient,
		stateClient:       stateClient,
		watchClient:       watchClient,
	}
}

func (o *DevClient) Start(
	ctx context.Context,
	options dev.StartOptions,
) error {
	klog.V(4).Infoln("Creating new adapter")

	var (
		componentStatus = watch.ComponentStatus{
			ImageComponentsAutoApplied: make(map[string]devfilev1.ImageComponent),
		}
	)

	klog.V(4).Infoln("Creating inner-loop resources for the component")

	watchParameters := watch.WatchParameters{
		StartOptions:        options,
		DevfileWatchHandler: o.watchHandler,
		WatchCluster:        false,
	}

	return o.watchClient.WatchAndPush(ctx, watchParameters, componentStatus)
}

// syncFiles syncs the local source files in path into the pod's source volume
func (o *DevClient) syncFiles(ctx context.Context, options dev.StartOptions, pod *corev1.Pod, path string) (bool, error) {
	var (
		devfileObj    = odocontext.GetEffectiveDevfileObj(ctx)
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
	s := log.Spinner("Syncing files into the container")
	defer s.End(false)

	syncFilesMap := make(map[string]string)
	var devfileCmd devfilev1.Command
	innerLoopWithCommands := !options.SkipCommands
	if innerLoopWithCommands {
		var (
			cmdKind = devfilev1.RunCommandGroupKind
			cmdName = options.RunCommand
		)
		if options.Debug {
			cmdKind = devfilev1.DebugCommandGroupKind
			cmdName = options.DebugCommand
		}
		var hasCmd bool
		devfileCmd, hasCmd, err = libdevfile.GetCommand(*devfileObj, cmdName, cmdKind)
		if err != nil {
			return false, err
		}
		if hasCmd {
			syncFilesMap = common.GetSyncFilesFromAttributes(devfileCmd)
		} else {
			klog.V(2).Infof("no command found with name %q and kind %v, syncing files without command attributes", cmdName, cmdKind)
		}
	}

	syncParams := sync.SyncParameters{
		Path:                     path,
		WatchFiles:               nil,
		WatchDeletedFiles:        nil,
		IgnoredFiles:             options.IgnorePaths,
		DevfileScanIndexForWatch: true,

		CompInfo:  compInfo,
		ForcePush: true,
		Files:     syncFilesMap,
	}
	execRequired, err := o.syncClient.SyncFiles(ctx, syncParams)
	if err != nil {
		return false, err
	}
	s.End(true)
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

func (o *DevClient) watchHandler(ctx context.Context, pushParams common.PushParameters, componentStatus *watch.ComponentStatus) error {

	devObj, err := devfile.ParseAndValidateFromFileWithVariables(location.DevfileLocation(o.fs, ""), pushParams.StartOptions.Variables, o.prefClient.GetImageRegistry(), true)
	if err != nil {
		return fmt.Errorf("unable to read devfile: %w", err)
	}
	pushParams.Devfile = devObj

	return o.reconcile(ctx, pushParams, componentStatus)
}
