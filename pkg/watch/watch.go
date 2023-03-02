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

	"github.com/devfile/library/v2/pkg/devfile/parser"

	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"

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
}

var _ Client = (*WatchClient)(nil)

func NewWatchClient(kubeClient kclient.ClientInterface) *WatchClient {
	return &WatchClient{
		kubeClient: kubeClient,
	}
}

// WatchParameters is designed to hold the controllables and attributes that the watch function works on
type WatchParameters struct {
	// Name of component that is to be watched
	ComponentName string
	// Name of application, the component is part of
	ApplicationName string
	// DevfilePath is the path of the devfile
	DevfilePath string
	// The path to the source of component(local or binary)
	Path string
	// List/Slice of files/folders in component source, the updates to which need not be pushed to component deployed pod
	FileIgnores []string
	// Custom function that can be used to push detected changes to remote pod. For more info about what each of the parameters to this function, please refer, pkg/component/component.go#PushLocal
	// WatchHandler func(kclient.ClientInterface, string, string, string, io.Writer, []string, []string, bool, []string, bool) error
	// Custom function that can be used to push detected changes to remote devfile pod. For more info about what each of the parameters to this function, please refer, pkg/devfile/adapters/interface.go#PlatformAdapter
	DevfileWatchHandler func(context.Context, adapters.PushParameters, WatchParameters, *ComponentStatus) error
	// Parameter whether or not to show build logs
	Show bool
	// DevfileBuildCmd takes the build command through the command line and overwrites devfile build command
	DevfileBuildCmd string
	// DevfileRunCmd takes the run command through the command line and overwrites devfile run command
	DevfileRunCmd string
	// DevfileDebugCmd takes the debug command through the command line and overwrites the devfile debug command
	DevfileDebugCmd string
	// InitialDevfileObj is used to compare the devfile between the very first run of odo dev and subsequent ones
	InitialDevfileObj parser.DevfileObj
	// Debug indicates if the debug command should be started after sync, or the run command by default
	Debug bool
	// DebugPort indicates which debug port to use for pushing after sync
	DebugPort int
	// Variables override Devfile variables
	Variables map[string]string
	// RandomPorts is true to forward containers ports on local random ports
	RandomPorts bool
	// IgnoreLocalhost indicates whether to proceed with port-forwarding regardless of any container ports being bound to the container loopback interface.
	// Applicable to Podman only.
	IgnoreLocalhost bool
	// WatchFiles indicates to watch for file changes and sync changes to the container
	WatchFiles bool
	// ForwardLocalhost indicates whether to try to make port-forwarding work with container apps listening on the loopback interface.
	ForwardLocalhost bool
	// WatchCluster indicates to watch Cluster-related objects (Deployment, Pod, etc)
	WatchCluster bool
	// ErrOut is a Writer to output forwarded port information
	Out io.Writer
	// ErrOut is a Writer to output forwarded port information
	ErrOut io.Writer
	// PromptMessage
	PromptMessage string
}

// evaluateChangesFunc evaluates any file changes for the events by ignoring the files in fileIgnores slice and removes
// any deleted paths from the watcher. It returns a slice of changed files (if any) and paths that are deleted (if any)
// by the events
type evaluateChangesFunc func(events []fsnotify.Event, path string, fileIgnores []string, watcher *fsnotify.Watcher) (changedFiles, deletedPaths []string)

// processEventsFunc processes the events received on the watcher. It uses the WatchParameters to trigger watch handler and writes to out
// It returns a Duration after which to recall in case of error
type processEventsFunc func(ctx context.Context, changedFiles, deletedPaths []string, parameters WatchParameters, out io.Writer, componentStatus *ComponentStatus, backoff *ExpBackoff) (*time.Duration, error)

func (o *WatchClient) WatchAndPush(out io.Writer, parameters WatchParameters, ctx context.Context, componentStatus ComponentStatus) error {
	klog.V(4).Infof("starting WatchAndPush, path: %s, component: %s, ignores %s", parameters.Path, parameters.ComponentName, parameters.FileIgnores)

	var err error
	if parameters.WatchFiles {
		o.sourcesWatcher, err = getFullSourcesWatcher(parameters.Path, parameters.FileIgnores)
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
		selector := labels.GetSelector(parameters.ComponentName, parameters.ApplicationName, labels.ComponentDevMode, true)
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
	if parameters.WatchFiles {
		var devfileFiles []string
		devfileFiles, err = libdevfile.GetReferencedLocalFiles(parameters.InitialDevfileObj)
		if err != nil {
			return err
		}
		devfileFiles = append(devfileFiles, parameters.DevfilePath)
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
			log.Fwarning(out, "Unable to watch Events resource, warning Events won't be displayed")
		}
	} else {
		o.warningsWatcher = NewNoOpWatcher()
	}

	o.keyWatcher = getKeyWatcher(ctx, out)
	return o.eventWatcher(ctx, parameters, out, evaluateFileChanges, o.processEvents, componentStatus)
}

// eventWatcher loops till the context's Done channel indicates it to stop looping, at which point it performs cleanup.
// While looping, it listens for filesystem events and processes these events using the WatchParameters to push to the remote pod.
// It outputs any logs to the out io Writer
func (o *WatchClient) eventWatcher(
	ctx context.Context,
	parameters WatchParameters,
	out io.Writer,
	evaluateChangesHandler evaluateChangesFunc,
	processEventsHandler processEventsFunc,
	componentStatus ComponentStatus,
) error {

	expBackoff := NewExpBackoff()

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

	// retryTimer is a timer used to retry later a sync that has failed
	retryTimer := time.NewTimer(time.Millisecond)
	<-retryTimer.C

	podsPhases := NewPodPhases()

	for {
		select {
		case event := <-o.sourcesWatcher.Events:
			events = append(events, event)
			// We are waiting for more events in this interval
			sourcesTimer.Reset(100 * time.Millisecond)

		case <-sourcesTimer.C:
			// timer has fired
			if !componentCanSyncFile(componentStatus.State) {
				klog.V(4).Infof("State of component is %q, don't sync sources", componentStatus.State)
				continue
			}

			var changedFiles, deletedPaths []string
			if !o.forceSync {
				// first find the files that have changed (also includes the ones newly created) or deleted
				changedFiles, deletedPaths = evaluateChangesHandler(events, parameters.Path, parameters.FileIgnores, o.sourcesWatcher)
				// process the changes and sync files with remote pod
				if len(changedFiles) == 0 && len(deletedPaths) == 0 {
					continue
				}
			}

			componentStatus.State = StateSyncOutdated
			fmt.Fprintf(out, "Pushing files...\n\n")
			retry, err := processEventsHandler(ctx, changedFiles, deletedPaths, parameters, out, &componentStatus, expBackoff)
			o.forceSync = false
			if err != nil {
				return err
			}
			// empty the events to receive new events
			if componentStatus.State == StateReady {
				events = []fsnotify.Event{} // empty the events slice to capture new events
			}

			if retry != nil {
				retryTimer.Reset(*retry)
			} else {
				retryTimer.Reset(time.Millisecond)
				<-retryTimer.C
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
				klog.V(4).Infof("deployment watcher Event: Type: %s, name: %s, rv: %s, pods: %d\n",
					ev.Type, obj.GetName(), obj.GetResourceVersion(), obj.Status.ReadyReplicas)
				deployTimer.Reset(300 * time.Millisecond)

			case *metav1.Status:
				klog.V(4).Infof("Status: %+v\n", obj)
			}

		case <-deployTimer.C:
			retry, err := processEventsHandler(ctx, nil, nil, parameters, out, &componentStatus, expBackoff)
			if err != nil {
				return err
			}
			if retry != nil {
				retryTimer.Reset(*retry)
			} else {
				retryTimer.Reset(time.Millisecond)
				<-retryTimer.C
			}

		case <-o.devfileWatcher.Events:
			devfileTimer.Reset(100 * time.Millisecond)

		case <-devfileTimer.C:
			fmt.Fprintf(out, "Updating Component...\n\n")
			retry, err := processEventsHandler(ctx, nil, nil, parameters, out, &componentStatus, expBackoff)
			if err != nil {
				return err
			}
			if retry != nil {
				retryTimer.Reset(*retry)
			} else {
				retryTimer.Reset(time.Millisecond)
				<-retryTimer.C
			}

		case <-retryTimer.C:
			retry, err := processEventsHandler(ctx, nil, nil, parameters, out, &componentStatus, expBackoff)
			if err != nil {
				return err
			}
			if retry != nil {
				retryTimer.Reset(*retry)
			} else {
				retryTimer.Reset(time.Millisecond)
				<-retryTimer.C
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
				selector := labels.GetSelector(parameters.ComponentName, parameters.ApplicationName, labels.ComponentDevMode, true)
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
			return errors.New("Dev mode interrupted by user")
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
	changedFiles, deletedPaths []string,
	parameters WatchParameters,
	out io.Writer,
	componentStatus *ComponentStatus,
	backoff *ExpBackoff,
) (*time.Duration, error) {
	for _, file := range removeDuplicates(append(changedFiles, deletedPaths...)) {
		fmt.Fprintf(out, "\nFile %s changed\n", file)
	}

	var hasFirstSuccessfulPushOccurred bool

	klog.V(4).Infof("Copying files %s to pod", changedFiles)

	pushParams := adapters.PushParameters{
		Path:                     parameters.Path,
		WatchFiles:               changedFiles,
		WatchDeletedFiles:        deletedPaths,
		IgnoredFiles:             parameters.FileIgnores,
		DevfileBuildCmd:          parameters.DevfileBuildCmd,
		DevfileRunCmd:            parameters.DevfileRunCmd,
		DevfileDebugCmd:          parameters.DevfileDebugCmd,
		DevfileScanIndexForWatch: !hasFirstSuccessfulPushOccurred,
		Debug:                    parameters.Debug,
		RandomPorts:              parameters.RandomPorts,
		ErrOut:                   parameters.ErrOut,
	}
	oldStatus := *componentStatus
	err := parameters.DevfileWatchHandler(ctx, pushParams, parameters, componentStatus)
	if err != nil {
		if isFatal(err) {
			return nil, err
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
			if parameters.WatchFiles {
				fmt.Fprintf(out, "%s - %s\n\n", PushErrorString, err.Error())
			} else {
				return nil, err
			}
		}
		wait := backoff.Delay()
		return &wait, nil
	}
	backoff.Reset()
	if oldStatus.State != StateReady && componentStatus.State == StateReady ||
		!reflect.DeepEqual(oldStatus.EndpointsForwarded, componentStatus.EndpointsForwarded) {

		PrintInfoMessage(out, parameters.Path, parameters.WatchFiles, parameters.PromptMessage)
	}
	return nil, nil
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
	return errors.As(err, &adapters.ErrPortForward{})
}
