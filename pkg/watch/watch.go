package watch

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/devfile/library/pkg/devfile/parser"
	_delete "github.com/redhat-developer/odo/pkg/component/delete"

	"github.com/fsnotify/fsnotify"

	"github.com/redhat-developer/odo/pkg/devfile/adapters/common"
	"github.com/redhat-developer/odo/pkg/envinfo"
	"github.com/redhat-developer/odo/pkg/log"
	"github.com/redhat-developer/odo/pkg/util"

	dfutil "github.com/devfile/library/pkg/util"

	"k8s.io/klog"
)

const (
	// PushErrorString is the string that is printed when an error occurs during watch's Push operation
	PushErrorString = "Error occurred on Push"
)

type WatchClient struct {
	deleteClient _delete.Client
}

func NewWatchClient(deleteClient _delete.Client) *WatchClient {
	return &WatchClient{
		deleteClient: deleteClient,
	}
}

// WatchParameters is designed to hold the controllables and attributes that the watch function works on
type WatchParameters struct {
	// Name of component that is to be watched
	ComponentName string
	// Name of application, the component is part of
	ApplicationName string
	// The path to the source of component(local or binary)
	Path string
	// List/Slice of files/folders in component source, the updates to which need not be pushed to component deployed pod
	FileIgnores []string
	// Custom function that can be used to push detected changes to remote pod. For more info about what each of the parameters to this function, please refer, pkg/component/component.go#PushLocal
	// WatchHandler func(kclient.ClientInterface, string, string, string, io.Writer, []string, []string, bool, []string, bool) error
	// Custom function that can be used to push detected changes to remote devfile pod. For more info about what each of the parameters to this function, please refer, pkg/devfile/adapters/interface.go#PlatformAdapter
	DevfileWatchHandler func(common.PushParameters, WatchParameters) error
	// Parameter whether or not to show build logs
	Show bool
	// EnvSpecificInfo contains information of env.yaml file
	EnvSpecificInfo *envinfo.EnvSpecificInfo
	// DevfileBuildCmd takes the build command through the command line and overwrites devfile build command
	DevfileBuildCmd string
	// DevfileRunCmd takes the run command through the command line and overwrites devfile run command
	DevfileRunCmd string
	// DevfileDebugCmd takes the debug command through the command line and overwrites the devfile debug command
	DevfileDebugCmd string
	// InitialDevfileObj is used to compare the devfile between the very first run of odo dev and subsequent ones
	InitialDevfileObj parser.DevfileObj
}

// evaluateChangesFunc evaluates any file changes for the events by ignoring the files in fileIgnores slice and removes
// any deleted paths from the watcher. It returns a slice of changed files (if any) and paths that are deleted (if any)
// by the events
type evaluateChangesFunc func(events []fsnotify.Event, fileIgnores []string, watcher *fsnotify.Watcher) (changedFiles, deletedPaths []string)

// processEventsFunc processes the events received on the watcher. It uses the WatchParameters to trigger watch handler and writes to out
type processEventsFunc func(changedFiles, deletedPaths []string, parameters WatchParameters, out io.Writer)

// cleanupFunc deletes the component created using the devfileObj and writes any outputs to out
type cleanupFunc func(devfileObj parser.DevfileObj, out io.Writer) error

// addRecursiveWatch handles adding watches recursively for the path provided
// and its subdirectories.  If a non-directory is specified, this call is a no-op.
// Files matching glob pattern defined in ignores will be ignored.
// Taken from https://github.com/openshift/origin/blob/85eb37b34f0657631592356d020cef5a58470f8e/pkg/util/fsnotification/fsnotification.go
// path is the path of the file or the directory
// ignores contains the glob rules for matching
func addRecursiveWatch(watcher *fsnotify.Watcher, path string, ignores []string) error {

	file, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error introspecting path %s: %v", path, err)
	}

	mode := file.Mode()
	if mode.IsRegular() {
		matched, e := dfutil.IsGlobExpMatch(path, ignores)
		if e != nil {
			return fmt.Errorf("unable to watcher on %s: %w", path, e)
		}
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
			matched, err := dfutil.IsGlobExpMatch(newPath, ignores)
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

		if matched, _ := dfutil.IsGlobExpMatch(folder, ignores); matched {
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

func (o *WatchClient) WatchAndPush(out io.Writer, parameters WatchParameters, ctx context.Context) error {
	klog.V(4).Infof("starting WatchAndPush, path: %s, component: %s, ignores %s", parameters.Path, parameters.ComponentName, parameters.FileIgnores)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error setting up filesystem watcher: %v", err)
	}
	defer watcher.Close()

	// adding watch on the root folder and the sub folders recursively
	// so directory and the path in addRecursiveWatch() are the same
	err = addRecursiveWatch(watcher, parameters.Path, parameters.FileIgnores)
	if err != nil {
		return fmt.Errorf("error watching source path %s: %v", parameters.Path, err)
	}

	printInfoMessage(out, parameters.Path)

	return eventWatcher(ctx, watcher, parameters, out, evaluateFileChanges, processEvents, o.cleanupFunc)
}

// eventWatcher loops till the context's Done channel indicates it to stop looping, at which point it performs cleanup.
// While looping, it listens for filesystem events and processes these events using the WatchParameters to push to the remote pod.
// It outputs any logs to the out io Writer
func eventWatcher(ctx context.Context, watcher *fsnotify.Watcher, parameters WatchParameters, out io.Writer, evaluateChangesHandler evaluateChangesFunc, processEventsHandler processEventsFunc, cleanupHandler cleanupFunc) error {
	var events []fsnotify.Event

	// timer helps collect multiple events that happen in a quick succession. We start with 1ms as we don't care much
	// at this point. In the select block, however, every time we receive an event, we reset the timer to watch for
	// 100ms since receiving that event. This is done because a single filesystem event by the user triggers multiple
	// events for fsnotify. It's a known-issue, but not really bug. For more info look at below issues:
	//    - https://github.com/fsnotify/fsnotify/issues/122
	//    - https://github.com/fsnotify/fsnotify/issues/344
	timer := time.NewTimer(time.Millisecond)
	<-timer.C

	for {
		select {
		case event := <-watcher.Events:
			events = append(events, event)
			// We are waiting for more events in this interval
			timer.Reset(100 * time.Millisecond)
		case <-timer.C:
			// timer has fired
			// first find the files that have changed (also includes the ones newly created) or deleted
			changedFiles, deletedPaths := evaluateChangesHandler(events, parameters.FileIgnores, watcher)
			// process the changes and sync files with remote pod
			processEventsHandler(changedFiles, deletedPaths, parameters, out)
			// empty the events to receive new events
			events = []fsnotify.Event{} // empty the events slice to capture new events
		case watchErr := <-watcher.Errors:
			return watchErr
		case <-ctx.Done():
			return cleanupHandler(parameters.InitialDevfileObj, out)
		}
	}
}

// evaluateFileChanges evaluates any file changes for the events. It ignores the files in fileIgnores slice and removes
// any deleted paths from the watcher
func evaluateFileChanges(events []fsnotify.Event, fileIgnores []string, watcher *fsnotify.Watcher) ([]string, []string) {
	var changedFiles []string
	var deletedPaths []string

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
		matched, globErr := dfutil.IsGlobExpMatch(event.Name, fileIgnores)
		klog.V(4).Infof("Matching %s with %s. Matched %v, err: %v", event.Name, fileIgnores, matched, globErr)
		if globErr != nil {
			watchError = fmt.Errorf("unable to watch changes: %w", globErr)
		}
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
			if e := addRecursiveWatch(watcher, event.Name, fileIgnores); e != nil && watchError == nil {
				klog.V(4).Infof("Error occurred in addRecursiveWatch, setting watchError to %v", e)
				watchError = e
			}
		}
	}
	deletedPaths = removeDuplicates(deletedPaths)

	return changedFiles, deletedPaths
}

func processEvents(changedFiles, deletedPaths []string, parameters WatchParameters, out io.Writer) {
	if len(changedFiles) == 0 && len(deletedPaths) == 0 {
		return
	}

	for _, file := range removeDuplicates(append(changedFiles, deletedPaths...)) {
		fmt.Fprintf(out, "\nFile %s changed\n", file)
	}

	var hasFirstSuccessfulPushOccurred bool

	fmt.Fprintf(out, "Pushing files...\n\n")
	klog.V(4).Infof("Copying files %s to pod", changedFiles)

	pushParams := common.PushParameters{
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
		Debug:                    parameters.EnvSpecificInfo.GetRunMode() == envinfo.Debug,
		DebugPort:                parameters.EnvSpecificInfo.GetDebugPort(),
	}
	err := parameters.DevfileWatchHandler(pushParams, parameters)
	if err != nil {
		// Log and output, but intentionally not exiting on error here.
		// We don't want to break watch when push failed, it might be fixed with the next change.
		klog.V(4).Infof("Error from Push: %v", err)
		fmt.Fprintf(out, "%s - %s\n\n", PushErrorString, err.Error())
	} else {
		printInfoMessage(out, parameters.Path)
	}
}

func (o *WatchClient) cleanupFunc(devfileObj parser.DevfileObj, out io.Writer) error {
	isInnerLoopDeployed, resources, err := o.deleteClient.ListResourcesToDeleteFromDevfile(devfileObj, "app")
	if err != nil {
		fmt.Fprintf(out, "failed to delete inner loop resources: %v", err)
		return err
	}
	// if innerloop deployment resource is present, then execute preStop events
	if isInnerLoopDeployed {
		err = o.deleteClient.ExecutePreStopEvents(devfileObj, "app")
		if err != nil {
			fmt.Fprint(out, "Failed to execute preStop events")
		}
	}
	// delete all the resources
	failed := o.deleteClient.DeleteResources(resources, true)
	for _, fail := range failed {
		fmt.Fprintf(out, "Failed to delete the %q resource: %s\n", fail.GetKind(), fail.GetName())
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

func printInfoMessage(out io.Writer, path string) {
	log.Finfof(out, "\nWatching for changes in the current directory %s\n"+
		"Press Ctrl+c to exit `odo dev` and delete resources from the cluster\n", path)
}
