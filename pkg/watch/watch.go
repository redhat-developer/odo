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

	"github.com/devfile/library/pkg/devfile/parser"
	tty "github.com/mattn/go-tty"

	_delete "github.com/redhat-developer/odo/pkg/component/delete"
	"github.com/redhat-developer/odo/pkg/devfile/adapters"
	"github.com/redhat-developer/odo/pkg/kclient"
	"github.com/redhat-developer/odo/pkg/labels"
	"github.com/redhat-developer/odo/pkg/libdevfile"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/state"

	"github.com/fsnotify/fsnotify"
	gitignore "github.com/sabhiram/go-gitignore"

	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/util"

	dfutil "github.com/devfile/library/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog"
)

const (
	// PushErrorString is the string that is printed when an error occurs during watch's Push operation
	PushErrorString = "Error occurred on Push"
	PromptMessages  = `
 [Ctrl+c] - Exit and delete resources from the cluster
 [p]      - Manually sync / push files to the cluster
`
)

type WatchClient struct {
	kubeClient   kclient.ClientInterface
	deleteClient _delete.Client
	stateClient  state.Client
}

var _ Client = (*WatchClient)(nil)

func NewWatchClient(kubeClient kclient.ClientInterface, deleteClient _delete.Client, stateClient state.Client) *WatchClient {
	return &WatchClient{
		kubeClient:   kubeClient,
		deleteClient: deleteClient,
		stateClient:  stateClient,
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
	DevfileWatchHandler func(adapters.PushParameters, WatchParameters, *ComponentStatus) error
	// Parameter whether or not to show build logs
	Show bool
	// EnvSpecificInfo contains information about devfile
	EnvSpecificInfo *envinfo.EnvSpecificInfo
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
	// WatchFiles indicates to watch for file changes and sync changes to the container
	WatchFiles bool
	// ErrOut is a Writer to output forwarded port information
	ErrOut io.Writer
}

// evaluateChangesFunc evaluates any file changes for the events by ignoring the files in fileIgnores slice and removes
// any deleted paths from the watcher. It returns a slice of changed files (if any) and paths that are deleted (if any)
// by the events
type evaluateChangesFunc func(events []fsnotify.Event, path string, fileIgnores []string, watcher *fsnotify.Watcher) (changedFiles, deletedPaths []string)

// processEventsFunc processes the events received on the watcher. It uses the WatchParameters to trigger watch handler and writes to out
// It returns a Duration after which to recall in case of error
type processEventsFunc func(changedFiles, deletedPaths []string, parameters WatchParameters, out io.Writer, componentStatus *ComponentStatus, backoff *ExpBackoff) (*time.Duration, error)

// addRecursiveWatch handles adding watches recursively for the path provided
// and its subdirectories.  If a non-directory is specified, this call is a no-op.
// Files matching glob pattern defined in ignores will be ignored.
// Taken from https://github.com/openshift/origin/blob/85eb37b34f0657631592356d020cef5a58470f8e/pkg/util/fsnotification/fsnotification.go
// rootPath is the root path of the file or directory,
// path is the recursive path of the file or the directory,
// ignores contains the glob rules for matching
func addRecursiveWatch(watcher *fsnotify.Watcher, rootPath string, path string, ignores []string) error {

	file, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error introspecting path %s: %v", path, err)
	}

	ignoreMatcher := gitignore.CompileIgnoreLines(ignores...)

	mode := file.Mode()
	if mode.IsRegular() {
		var rel string
		rel, err = filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}
		matched := ignoreMatcher.MatchesPath(rel)
		if !matched {
			klog.V(4).Infof("adding watch on path %s", path)

			// checking if the file exits before adding the watcher to it
			if !util.CheckPathExists(path) {
				return nil
			}

			err = watcher.Add(path)
			if err != nil {
				klog.V(4).Infof("error adding watcher for path %s: %v", path, err)
			}
			return nil
		}
	}

	folders := []string{}
	err = filepath.Walk(path, func(newPath string, info os.FileInfo, err error) error {
		if err != nil {
			// Ignore the error if it's a 'path does not exist' error, no need to walk a non-existent path
			if !util.CheckPathExists(newPath) {
				klog.V(4).Infof("Walk func received an error for path %s, but the path doesn't exist so this is likely not an error. err: %v", path, err)
				return nil
			}
			return fmt.Errorf("unable to walk path: %s: %w", newPath, err)
		}

		if info.IsDir() {
			// If the current directory matches any of the ignore patterns, ignore them so that their contents are also not ignored
			rel, err := filepath.Rel(rootPath, newPath)
			if err != nil {
				return err
			}
			matched := ignoreMatcher.MatchesPath(rel)
			if err != nil {
				return fmt.Errorf("unable to addRecursiveWatch on %s: %w", newPath, err)
			}
			if matched {
				klog.V(4).Infof("ignoring watch on path %s", newPath)
				return filepath.SkipDir
			}
			// Append the folder we just walked on
			folders = append(folders, newPath)
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, folder := range folders {

		rel, err := filepath.Rel(rootPath, folder)
		if err != nil {
			return err
		}
		matched := ignoreMatcher.MatchesPath(rel)

		if matched {
			klog.V(4).Infof("ignoring watch for %s", folder)
			continue
		}

		// checking if the file exits before adding the watcher to it
		if !util.CheckPathExists(path) {
			continue
		}

		klog.V(4).Infof("adding watch on path %s", folder)
		err = watcher.Add(folder)
		if err != nil {
			// Linux "no space left on device" issues are usually resolved via
			// $ sudo sysctl fs.inotify.max_user_watches=65536
			// BSD / OSX: "too many open files" issues are ussualy resolved via
			// $ sysctl variables "kern.maxfiles" and "kern.maxfilesperproc",
			klog.V(4).Infof("error adding watcher for path %s: %v", folder, err)
		}
	}
	return nil
}

func (o *WatchClient) WatchAndPush(out io.Writer, parameters WatchParameters, ctx context.Context, componentStatus ComponentStatus) error {
	klog.V(4).Infof("starting WatchAndPush, path: %s, component: %s, ignores %s", parameters.Path, parameters.ComponentName, parameters.FileIgnores)

	var sourcesWatcher *fsnotify.Watcher
	var err error
	if parameters.WatchFiles {
		sourcesWatcher, err = getFullSourcesWatcher(parameters.Path, parameters.FileIgnores)
		if err != nil {
			return err
		}
	} else {
		sourcesWatcher, err = fsnotify.NewWatcher()
		if err != nil {
			return err
		}
	}
	defer sourcesWatcher.Close()

	selector := labels.GetSelector(parameters.ComponentName, parameters.ApplicationName, labels.ComponentDevMode, true)
	deploymentWatcher, err := o.kubeClient.DeploymentWatcher(ctx, selector)
	if err != nil {
		return fmt.Errorf("error watching deployment: %v", err)
	}

	devfileWatcher, err := fsnotify.NewWatcher()
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
			err = devfileWatcher.Add(f)
			if err != nil {
				klog.V(4).Infof("error adding watcher for path %s: %v", f, err)
			}
		}
	}

	podWatcher, err := o.kubeClient.PodWatcher(ctx, selector)
	if err != nil {
		return err
	}

	warningsWatcher, isForbidden, err := o.kubeClient.PodWarningEventWatcher(ctx)
	if err != nil {
		return err
	}
	if isForbidden {
		log.Fwarning(out, "Unable to watch Events resource, warning Events won't be displayed")
	}

	keyboardInput := createKeyWatcher(ctx)

	return o.eventWatcher(ctx, sourcesWatcher, deploymentWatcher, devfileWatcher, podWatcher, warningsWatcher, parameters, out, evaluateFileChanges, processEvents, componentStatus, keyboardInput)
}

func getFullSourcesWatcher(path string, fileIgnores []string) (*fsnotify.Watcher, error) {
	absIgnorePaths := dfutil.GetAbsGlobExps(path, fileIgnores)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("error setting up filesystem watcher: %v", err)
	}

	// adding watch on the root folder and the sub folders recursively
	// so directory and the path in addRecursiveWatch() are the same
	err = addRecursiveWatch(watcher, path, path, absIgnorePaths)
	if err != nil {
		return nil, fmt.Errorf("error watching source path %s: %v", path, err)
	}
	return watcher, nil
}

// eventWatcher loops till the context's Done channel indicates it to stop looping, at which point it performs cleanup.
// While looping, it listens for filesystem events and processes these events using the WatchParameters to push to the remote pod.
// It outputs any logs to the out io Writer
func (o *WatchClient) eventWatcher(ctx context.Context, sourcesWatcher *fsnotify.Watcher, deploymentWatcher watch.Interface, devfileWatcher *fsnotify.Watcher, podWatcher watch.Interface, eventsWatcher watch.Interface, parameters WatchParameters, out io.Writer, evaluateChangesHandler evaluateChangesFunc, processEventsHandler processEventsFunc, componentStatus ComponentStatus, keyboardInput chan string) error {

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

	devfileTimer := time.NewTimer(time.Millisecond)
	<-devfileTimer.C

	deployTimer := time.NewTimer(time.Millisecond)
	<-deployTimer.C

	retryTimer := time.NewTimer(time.Millisecond)
	<-retryTimer.C

	// List of pod phases that will run as background processes
	podsPhases := NewPodPhases()

	for {

		select {
		case event := <-sourcesWatcher.Events:
			events = append(events, event)
			// We are waiting for more events in this interval
			sourcesTimer.Reset(100 * time.Millisecond)

			// Trigger manual sync
		case key := <-keyboardInput:
			if key == "p" {
				log.Section("Manually syncing files")
				log.Print("Updating component ...")

				// Before we push a sync, we *should* still check to see what files have changed
				changedFiles, deletedPaths := evaluateChangesHandler(events, parameters.Path, parameters.FileIgnores, sourcesWatcher)

				// Sync files REGARDLESS if changedFiles / deletedPaths == 0
				// Set the state to sync being outdated
				componentStatus.State = StateSyncOutdated

				// Push the files
				log.Print("Pushing files ...")
				retry, err := processEventsHandler(changedFiles, deletedPaths, parameters, out, &componentStatus, expBackoff)
				if err != nil {
					return err
				}

				// Empty all events
				if componentStatus.State == StateReady {
					events = []fsnotify.Event{} // empty the events slice to capture new events
				}

				// Reset the retry timer
				if retry != nil {
					retryTimer.Reset(*retry)
				} else {
					retryTimer.Reset(time.Millisecond)
					<-retryTimer.C
				}
			}

		case <-sourcesTimer.C:
			// timer has fired
			if !componentCanSyncFile(componentStatus.State) {
				continue
			}

			// first find the files that have changed (also includes the ones newly created) or deleted
			changedFiles, deletedPaths := evaluateChangesHandler(events, parameters.Path, parameters.FileIgnores, sourcesWatcher)
			// process the changes and sync files with remote pod
			if len(changedFiles) == 0 && len(deletedPaths) == 0 {
				continue
			}
			componentStatus.State = StateSyncOutdated
			log.Section("Automatically syncing files")
			log.Print("Pushing files ...")
			retry, err := processEventsHandler(changedFiles, deletedPaths, parameters, out, &componentStatus, expBackoff)
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

		case watchErr := <-sourcesWatcher.Errors:
			return watchErr

		case ev := <-deploymentWatcher.ResultChan():
			switch obj := ev.Object.(type) {
			case *appsv1.Deployment:
				klog.V(4).Infof("deployment watcher Event: Type: %s, name: %s, rv: %s, pods: %d\n",
					ev.Type, obj.GetName(), obj.GetResourceVersion(), obj.Status.ReadyReplicas)
				deployTimer.Reset(300 * time.Millisecond)

			case *metav1.Status:
				klog.V(4).Infof("Status: %+v\n", obj)
			}

		case <-deployTimer.C:
			retry, err := processEventsHandler(nil, nil, parameters, out, &componentStatus, expBackoff)
			if err != nil {
				return err
			}
			if retry != nil {
				retryTimer.Reset(*retry)
			} else {
				retryTimer.Reset(time.Millisecond)
				<-retryTimer.C
			}

		case <-devfileWatcher.Events:
			devfileTimer.Reset(100 * time.Millisecond)

		case <-devfileTimer.C:
			log.Print("Updating component ...")
			retry, err := processEventsHandler(nil, nil, parameters, out, &componentStatus, expBackoff)
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
			retry, err := processEventsHandler(nil, nil, parameters, out, &componentStatus, expBackoff)
			if err != nil {
				return err
			}
			if retry != nil {
				retryTimer.Reset(*retry)
			} else {
				retryTimer.Reset(time.Millisecond)
				<-retryTimer.C
			}

		case ev := <-podWatcher.ResultChan():
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

		case ev := <-eventsWatcher.ResultChan():
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

		case watchErr := <-devfileWatcher.Errors:
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

func processEvents(
	changedFiles, deletedPaths []string,
	parameters WatchParameters,
	out io.Writer,
	componentStatus *ComponentStatus,
	backoff *ExpBackoff,
) (*time.Duration, error) {
	for _, file := range removeDuplicates(append(changedFiles, deletedPaths...)) {
		log.Printf("File %s changed", file)
	}

	var hasFirstSuccessfulPushOccurred bool

	klog.V(4).Infof("Copying files %s to pod", changedFiles)

	pushParams := adapters.PushParameters{
		Path:                     parameters.Path,
		WatchFiles:               changedFiles,
		WatchDeletedFiles:        deletedPaths,
		IgnoredFiles:             parameters.FileIgnores,
		ForceBuild:               false,
		DevfileBuildCmd:          parameters.DevfileBuildCmd,
		DevfileRunCmd:            parameters.DevfileRunCmd,
		DevfileDebugCmd:          parameters.DevfileDebugCmd,
		DevfileScanIndexForWatch: !hasFirstSuccessfulPushOccurred,
		EnvSpecificInfo:          *parameters.EnvSpecificInfo,
		Debug:                    parameters.Debug,
		RandomPorts:              parameters.RandomPorts,
		ErrOut:                   parameters.ErrOut,
	}
	oldStatus := *componentStatus
	err := parameters.DevfileWatchHandler(pushParams, parameters, componentStatus)
	if err != nil {
		if isFatal(err) {
			return nil, err
		}
		klog.V(4).Infof("Error from Push: %v", err)
		if parameters.WatchFiles {
			// Log and output, but intentionally not exiting on error here.
			// We don't want to break watch when push failed, it might be fixed with the next change.
			log.Printf("%s - %s", PushErrorString, err.Error())
		} else {
			return nil, err
		}
		wait := backoff.Delay()
		return &wait, nil
	}
	backoff.Reset()
	if oldStatus.State != StateReady && componentStatus.State == StateReady ||
		!reflect.DeepEqual(oldStatus.EndpointsForwarded, componentStatus.EndpointsForwarded) {

		printInfoMessage(out, parameters.Path)
	}
	return nil, nil
}

func (o *WatchClient) CleanupDevResources(devfileObj parser.DevfileObj, componentName string, out io.Writer) error {
	log.Print("Cleaning resources, please wait")
	isInnerLoopDeployed, resources, err := o.deleteClient.ListResourcesToDeleteFromDevfile(devfileObj, "app", componentName, labels.ComponentDevMode)
	if err != nil {
		fmt.Fprintf(out, "failed to delete inner loop resources: %v", err)
		return err
	}
	// if innerloop deployment resource is present, then execute preStop events
	if isInnerLoopDeployed {
		err = o.deleteClient.ExecutePreStopEvents(devfileObj, "app", componentName)
		if err != nil {
			fmt.Fprint(out, "Failed to execute preStop events")
		}
	}
	// delete all the resources
	failed := o.deleteClient.DeleteResources(resources, true)
	for _, fail := range failed {
		fmt.Fprintf(out, "Failed to delete the %q resource: %s\n", fail.GetKind(), fail.GetName())
	}

	return o.stateClient.SaveExit()
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

func printInfoMessage(out io.Writer, path string) {
	log.Sectionf("Dev mode")
	fmt.Fprintf(out, " %s\n Watching for changes in the current directory %s\n\n %s%s", log.Sbold("Status:"), path, log.Sbold("Keyboard Commands:"), PromptMessages)
}

func isFatal(err error) bool {
	return errors.As(err, &adapters.ErrPortForward{})
}

// createKeyWatcher will create a go routine that will run in the background in order to
// watch user key inputs
func createKeyWatcher(ctx context.Context) chan string {
	keyInput := make(chan string)

	// Create a go routine that will read a character from the tty in the background without interruption
	go func() {

		// Create this in the background / go routine so there is no blocking, as well
		// as always open to be read from the user.
		// In order to capture the keyboard events, we'll use the "tty" library.
		// Open up a tty to read in character keys from the user, this will be done in the background
		tty, err := tty.Open()
		if err != nil {
			klog.Error(err)
		}
		defer tty.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case c := <-getRune(tty):
				keyInput <- string(c)
			}

		}
	}()

	return keyInput
}

func getRune(tty *tty.TTY) (ch chan rune) {
	ch = make(chan rune)
	go func() {
		char, err := tty.ReadRune()
		if err != nil {
			klog.Error(err)
		}
		ch <- char
		close(ch)
	}()
	return ch
}
