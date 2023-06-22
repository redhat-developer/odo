package watch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/redhat-developer/odo/pkg/dev"
	"github.com/redhat-developer/odo/pkg/dev/common"

	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	odocontext "github.com/redhat-developer/odo/pkg/odo/context"

	"github.com/fsnotify/fsnotify"
	gitignore "github.com/sabhiram/go-gitignore"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog"
)

const (
	// PushErrorString is the string that is printed when an error occurs during watch's Push operation
	PushErrorString = "Error occurred on Push"
)

type WatchClient struct {
	kubeClient kclient.ClientInterface

	sourcesWatcher    *fsnotify.Watcher
	deploymentWatcher watch.Interface
	devfileWatcher    *fsnotify.Watcher
	podWatcher        watch.Interface
	warningsWatcher   watch.Interface
	keyWatcher        <-chan byte

	// true to force sync, used when manual sync
	forceSync bool

	// deploymentGeneration indicates the generation of the latest observed Deployment
	deploymentGeneration int64
	readyReplicas        int32
}

var _ Client = (*WatchClient)(nil)

func NewWatchClient(kubeClient kclient.ClientInterface) *WatchClient {
	return &WatchClient{
		kubeClient: kubeClient,
	}
}

// WatchParameters is designed to hold the controllables and attributes that the watch function works on
type WatchParameters struct {
	StartOptions dev.StartOptions

	// Custom function that can be used to push detected changes to remote pod. For more info about what each of the parameters to this function, please refer, pkg/component/component.go#PushLocal
	// WatchHandler func(kclient.ClientInterface, string, string, string, io.Writer, []string, []string, bool, []string, bool) error
	// Custom function that can be used to push detected changes to remote devfile pod. For more info about what each of the parameters to this function, please refer, pkg/devfile/adapters/interface.go#PlatformAdapter
	DevfileWatchHandler func(context.Context, common.PushParameters, *ComponentStatus) error
	// Parameter whether or not to show build logs
	Show bool
	// DebugPort indicates which debug port to use for pushing after sync
	DebugPort int

	// WatchCluster indicates to watch Cluster-related objects (Deployment, Pod, etc)
	WatchCluster bool
	// PromptMessage
	PromptMessage string
}

// evaluateChangesFunc evaluates any file changes for the events by ignoring the files in fileIgnores slice and removes
// any deleted paths from the watcher. It returns a slice of changed files (if any) and paths that are deleted (if any)
// by the events
type evaluateChangesFunc func(events []fsnotify.Event, path string, fileIgnores []string, watcher *fsnotify.Watcher) (changedFiles, deletedPaths []string)

// processEventsFunc processes the events received on the watcher. It uses the WatchParameters to trigger watch handler and writes to out
// It returns a Duration after which to recall in case of error
type processEventsFunc func(ctx context.Context, parameters WatchParameters, changedFiles, deletedPaths []string, componentStatus *ComponentStatus) error

func (o *WatchClient) WatchAndPush(ctx context.Context, parameters WatchParameters, componentStatus ComponentStatus) error {
	var (
		devfileObj    = odocontext.GetEffectiveDevfileObj(ctx)
		devfilePath   = odocontext.GetDevfilePath(ctx)
		path          = filepath.Dir(devfilePath)
		componentName = odocontext.GetComponentName(ctx)
		appName       = odocontext.GetApplication(ctx)
	)

	klog.V(4).Infof("starting WatchAndPush, path: %s, component: %s, ignores %s", path, componentName, parameters.StartOptions.IgnorePaths)

	var err error
	if parameters.StartOptions.WatchFiles {
		o.sourcesWatcher, err = getFullSourcesWatcher(path, parameters.StartOptions.IgnorePaths)
		if err != nil {
			return err
		}
	} else {
		o.sourcesWatcher, err = fsnotify.NewWatcher()
		if err != nil {
			return err
		}
	}
	defer o.sourcesWatcher.Close()

	if parameters.WatchCluster {
		selector := labels.GetSelector(componentName, appName, labels.ComponentDevMode, true)
		o.deploymentWatcher, err = o.kubeClient.DeploymentWatcher(ctx, selector)
		if err != nil {
			return fmt.Errorf("error watching deployment: %v", err)
		}

		o.podWatcher, err = o.kubeClient.PodWatcher(ctx, selector)
		if err != nil {
			return err
		}
	} else {
		o.deploymentWatcher = NewNoOpWatcher()
		o.podWatcher = NewNoOpWatcher()
	}

	o.devfileWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if parameters.StartOptions.WatchFiles {
		var devfileFiles []string
		devfileFiles, err = libdevfile.GetReferencedLocalFiles(*devfileObj)
		if err != nil {
			return err
		}
		devfileFiles = append(devfileFiles, devfilePath)
		for _, f := range devfileFiles {
			err = o.devfileWatcher.Add(f)
			if err != nil {
				klog.V(4).Infof("error adding watcher for path %s: %v", f, err)
			}
		}
	}

	if parameters.WatchCluster {
		var isForbidden bool
		o.warningsWatcher, isForbidden, err = o.kubeClient.PodWarningEventWatcher(ctx)
		if err != nil {
			return err
		}
		if isForbidden {
			log.Fwarning(parameters.StartOptions.Out, "Unable to watch Events resource, warning Events won't be displayed")
		}
	} else {
		o.warningsWatcher = NewNoOpWatcher()
	}

	o.keyWatcher = getKeyWatcher(ctx, parameters.StartOptions.Out)

	err = o.processEvents(ctx, parameters, nil, nil, &componentStatus)
	if err != nil {
		return err
	}

	return o.eventWatcher(ctx, parameters, evaluateFileChanges, o.processEvents, componentStatus)
}

// eventWatcher loops till the context's Done channel indicates it to stop looping, at which point it performs cleanup.
// While looping, it listens for filesystem events and processes these events using the WatchParameters to push to the remote pod.
// It outputs any logs to the out io Writer
func (o *WatchClient) eventWatcher(
	ctx context.Context,
	parameters WatchParameters,
	evaluateChangesHandler evaluateChangesFunc,
	processEventsHandler processEventsFunc,
	componentStatus ComponentStatus,
) error {

	var (
		devfilePath   = odocontext.GetDevfilePath(ctx)
		path          = filepath.Dir(devfilePath)
		componentName = odocontext.GetComponentName(ctx)
		appName       = odocontext.GetApplication(ctx)
		out           = parameters.StartOptions.Out
	)

	var events []fsnotify.Event

	// sourcesTimer helps collect multiple events that happen in a quick succession. We start with 1ms as we don't care much
	// at this point. In the select block, however, every time we receive an event, we reset the sourcesTimer to watch for
	// 100ms since receiving that event. This is done because a single filesystem event by the user triggers multiple
	// events for fsnotify. It's a known-issue, but not really bug. For more info look at below issues:
	//    - https://github.com/fsnotify/fsnotify/issues/122
	//    - https://github.com/fsnotify/fsnotify/issues/344
	sourcesTimer := time.NewTimer(time.Millisecond)
	<-sourcesTimer.C

	// devfileTimer has the same usage as sourcesTimer, for file events coming from devfileWatcher
	devfileTimer := time.NewTimer(time.Millisecond)
	<-devfileTimer.C

	// deployTimer has the same usage as sourcesTimer, for events coming from watching Deployments, from deploymentWatcher
	deployTimer := time.NewTimer(time.Millisecond)
	<-deployTimer.C

	podsPhases := NewPodPhases()

	for {
		select {
		case event := <-o.sourcesWatcher.Events:
			events = append(events, event)
			// We are waiting for more events in this interval
			sourcesTimer.Reset(100 * time.Millisecond)

		case <-sourcesTimer.C:
			// timer has fired
			if !componentCanSyncFile(componentStatus.GetState()) {
				klog.V(4).Infof("State of component is %q, don't sync sources", componentStatus.GetState())
				continue
			}

			var changedFiles, deletedPaths []string
			if !o.forceSync {
				// first find the files that have changed (also includes the ones newly created) or deleted
				changedFiles, deletedPaths = evaluateChangesHandler(events, path, parameters.StartOptions.IgnorePaths, o.sourcesWatcher)
				// process the changes and sync files with remote pod
				if len(changedFiles) == 0 && len(deletedPaths) == 0 {
					continue
				}
			}

			componentStatus.SetState(StateSyncOutdated)
			fmt.Fprintf(out, "Pushing files...\n\n")
			err := processEventsHandler(ctx, parameters, changedFiles, deletedPaths, &componentStatus)
			o.forceSync = false
			if err != nil {
				return err
			}
			// empty the events to receive new events
			if componentStatus.GetState() == StateReady {
				events = []fsnotify.Event{} // empty the events slice to capture new events
			}

		case watchErr := <-o.sourcesWatcher.Errors:
			return watchErr

		case key := <-o.keyWatcher:
			if key == 'p' {
				o.forceSync = true
				sourcesTimer.Reset(100 * time.Millisecond)
			}

		case ev := <-o.deploymentWatcher.ResultChan():
			switch obj := ev.Object.(type) {
			case *appsv1.Deployment:
				klog.V(4).Infof("deployment watcher Event: Type: %s, name: %s, rv: %s, generation: %d, pods: %d\n",
					ev.Type, obj.GetName(), obj.GetResourceVersion(), obj.GetGeneration(), obj.Status.ReadyReplicas)
				if obj.GetGeneration() > o.deploymentGeneration || obj.Status.ReadyReplicas != o.readyReplicas {
					o.deploymentGeneration = obj.GetGeneration()
					o.readyReplicas = obj.Status.ReadyReplicas
					deployTimer.Reset(300 * time.Millisecond)
				}

			case *metav1.Status:
				klog.V(4).Infof("Status: %+v\n", obj)
			}

		case <-deployTimer.C:
			err := processEventsHandler(ctx, parameters, nil, nil, &componentStatus)
			if err != nil {
				return err
			}

		case <-o.devfileWatcher.Events:
			devfileTimer.Reset(100 * time.Millisecond)

		case <-devfileTimer.C:
			fmt.Fprintf(out, "Updating Component...\n\n")
			err := processEventsHandler(ctx, parameters, nil, nil, &componentStatus)
			if err != nil {
				return err
			}

		case ev := <-o.podWatcher.ResultChan():
			switch ev.Type {
			case watch.Deleted:
				pod, ok := ev.Object.(*corev1.Pod)
				if !ok {
					return errors.New("unable to decode watch event")
				}
				podsPhases.Delete(out, pod)
			case watch.Added, watch.Modified:
				pod, ok := ev.Object.(*corev1.Pod)
				if !ok {
					return errors.New("unable to decode watch event")
				}
				podsPhases.Add(out, pod.GetCreationTimestamp(), pod)
			}

		case ev := <-o.warningsWatcher.ResultChan():
			switch kevent := ev.Object.(type) {
			case *corev1.Event:
				podName := kevent.InvolvedObject.Name
				selector := labels.GetSelector(componentName, appName, labels.ComponentDevMode, true)
				matching, err := o.kubeClient.IsPodNameMatchingSelector(ctx, podName, selector)
				if err != nil {
					return err
				}
				if matching {
					log.Fwarning(out, kevent.Message)
				}
			}

		case watchErr := <-o.devfileWatcher.Errors:
			return watchErr

		case <-ctx.Done():
			klog.V(2).Info("Dev mode interrupted by user")
			return nil
		}
	}
}

// evaluateFileChanges evaluates any file changes for the events. It ignores the files in fileIgnores slice related to path, and removes
// any deleted paths from the watcher
func evaluateFileChanges(events []fsnotify.Event, path string, fileIgnores []string, watcher *fsnotify.Watcher) ([]string, []string) {
	var changedFiles []string
	var deletedPaths []string

	ignoreMatcher := gitignore.CompileIgnoreLines(fileIgnores...)

	for _, event := range events {
		klog.V(4).Infof("filesystem watch event: %s", event)
		isIgnoreEvent := shouldIgnoreEvent(event)

		// add file name to changedFiles only once
		alreadyInChangedFiles := false
		for _, cfile := range changedFiles {
			if cfile == event.Name {
				alreadyInChangedFiles = true
				break
			}
		}

		// Filter out anything in ignores list from the list of changed files
		// This is important in spite of not watching the
		// ignores paths because, when a directory that is ignored, is deleted,
		// because its parent is watched, the fsnotify automatically raises an event
		// for it.
		var watchError error
		rel, err := filepath.Rel(path, event.Name)
		if err != nil {
			watchError = fmt.Errorf("unable to get relative path of %q on %q", event.Name, path)
		}
		matched := ignoreMatcher.MatchesPath(rel)
		if !alreadyInChangedFiles && !matched && !isIgnoreEvent {
			// Append the new file change event to changedFiles if and only if the event is not a file remove event
			if event.Op&fsnotify.Remove != fsnotify.Remove {
				changedFiles = append(changedFiles, event.Name)
			}
		}

		// Rename operation triggers RENAME event on old path + CREATE event for renamed path so delete old path in case of rename
		// Also weirdly, fsnotify raises a RENAME event for deletion of files/folders with space in their name so even that should be handled here
		if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
			// On remove/rename, stop watching the resource
			if e := watcher.Remove(event.Name); e != nil {
				klog.V(4).Infof("error removing watch for %s: %v", event.Name, e)
			}
			// Append the file to list of deleted files
			// When a file/folder is deleted, it raises 2 events:
			//	a. RENAME with event.Name empty
			//	b. REMOVE with event.Name as file name
			if !alreadyInChangedFiles && !matched && event.Name != "" {
				deletedPaths = append(deletedPaths, event.Name)
			}
		} else {
			// On other ops, recursively watch the resource (if applicable)
			if e := addRecursiveWatch(watcher, path, event.Name, fileIgnores); e != nil && watchError == nil {
				klog.V(4).Infof("Error occurred in addRecursiveWatch, setting watchError to %v", e)
				watchError = e
			}
		}
	}
	deletedPaths = removeDuplicates(deletedPaths)

	return changedFiles, deletedPaths
}

func (o *WatchClient) processEvents(
	ctx context.Context,
	parameters WatchParameters,
	changedFiles, deletedPaths []string,
	componentStatus *ComponentStatus,
) error {
	var (
		devfilePath = odocontext.GetDevfilePath(ctx)
		path        = filepath.Dir(devfilePath)
		out         = parameters.StartOptions.Out
	)

	for _, file := range removeDuplicates(append(changedFiles, deletedPaths...)) {
		fmt.Fprintf(out, "\nFile %s changed\n", file)
	}

	var hasFirstSuccessfulPushOccurred bool

	klog.V(4).Infof("Copying files %s to pod", changedFiles)

	pushParams := common.PushParameters{
		StartOptions:             parameters.StartOptions,
		WatchFiles:               changedFiles,
		WatchDeletedFiles:        deletedPaths,
		DevfileScanIndexForWatch: !hasFirstSuccessfulPushOccurred,
	}
	oldStatus := *componentStatus
	err := parameters.DevfileWatchHandler(ctx, pushParams, componentStatus)
	if err != nil {
		if isFatal(err) {
			return err
		}
		klog.V(4).Infof("Error from Push: %v", err)
		// Log and output, but intentionally not exiting on error here.
		// We don't want to break watch when push failed, it might be fixed with the next push.
		if kerrors.IsUnauthorized(err) || kerrors.IsForbidden(err) {
			fmt.Fprintf(out, "Error connecting to the cluster. Please log in again\n\n")
			var refreshed bool
			refreshed, err = o.kubeClient.Refresh()
			if err != nil {
				fmt.Fprintf(out, "Error updating Kubernetes config: %s\n", err)
			} else if refreshed {
				fmt.Fprintf(out, "Updated Kubernetes config\n")
			}
		} else {
			fmt.Fprintf(out, "%s - %s\n\n", PushErrorString, err.Error())
			PrintInfoMessage(out, path, parameters.StartOptions.WatchFiles, parameters.PromptMessage)
		}
		return nil
	}
	if oldStatus.GetState() != StateReady && componentStatus.GetState() == StateReady ||
		!reflect.DeepEqual(oldStatus.EndpointsForwarded, componentStatus.EndpointsForwarded) {

		PrintInfoMessage(out, path, parameters.StartOptions.WatchFiles, parameters.PromptMessage)
	}
	return nil
}

func shouldIgnoreEvent(event fsnotify.Event) (ignoreEvent bool) {
	if !(event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename) {
		stat, err := os.Lstat(event.Name)
		if err != nil {
			// Some of the editors like vim and gedit, generate temporary buffer files during update to the file and deletes it soon after exiting from the editor
			// So, its better to log the error rather than feeding it to error handler via `watchError = fmt.Errorf("unable to watch changes: %w", err)`,
			// which will terminate the watch
			klog.V(4).Infof("Failed getting details of the changed file %s. Ignoring the change", event.Name)
		}
		// Some of the editors generate temporary buffer files during update to the file and deletes it soon after exiting from the editor
		// So, its better to log the error rather than feeding it to error handler via `watchError = fmt.Errorf("unable to watch changes: %w", err)`,
		// which will terminate the watch
		if stat == nil {
			klog.V(4).Infof("Ignoring event for file %s as details about the file couldn't be fetched", event.Name)
			ignoreEvent = true
		}

		// In windows, every new file created under a sub-directory of the watched directory, raises 2 events:
		// 1. Write event for the directory under which the file was created
		// 2. Create event for the file that was created
		// Ignore 1 to avoid duplicate events.
		if ignoreEvent || (stat.IsDir() && event.Op&fsnotify.Write == fsnotify.Write) {
			ignoreEvent = true
		}
	}
	return ignoreEvent
}

func removeDuplicates(input []string) []string {
	valueMap := map[string]string{}
	for _, str := range input {
		valueMap[str] = str
	}

	result := []string{}
	for str := range valueMap {
		result = append(result, str)
	}
	return result
}

func PrintInfoMessage(out io.Writer, path string, watchFiles bool, promptMessage string) {
	log.Sectionf("Dev mode")
	if watchFiles {
		fmt.Fprintf(
			out,
			" %s\n Watching for changes in the current directory %s\n\n",
			log.Sbold("Status:"),
			path,
		)
	}
	fmt.Fprintf(
		out,
		" %s%s",
		log.Sbold("Keyboard Commands:"),
		promptMessage,
	)
}

func isFatal(err error) bool {
	return errors.As(err, &common.ErrPortForward{})
}
