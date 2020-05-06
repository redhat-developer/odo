package watch

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/openshift/odo/pkg/devfile/adapters/common"
	"github.com/openshift/odo/pkg/util"

	"github.com/openshift/odo/pkg/occlient"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

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
	WatchHandler func(*occlient.Client, string, string, string, io.Writer, []string, []string, bool, []string, bool) error
	// Custom function that can be used to push detected changes to remote devfile pod. For more info about what each of the parameters to this function, please refer, pkg/component/component.go#PushLocal
	DevfileWatchHandler func(common.PushParameters) error
	// This is a channel added to signal readiness of the watch command to the external channel listeners
	StartChan chan bool
	// This is a channel added to terminate the watch command gracefully without passing SIGINT. "Stop" message on this channel terminates WatchAndPush function
	ExtChan chan bool
	// Interval of time before pushing changes to remote(component) pod
	PushDiffDelay int
	// Parameter whether or not to show build logs
	Show bool
}

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
		matched, err := util.IsGlobExpMatch(path, ignores)
		if err != nil {
			return errors.Wrapf(err, "unable to watcher on %s", path)
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
			return err
		}

		if info.IsDir() {
			// If the current directory matches any of the ignore patterns, ignore them so that their contents are also not ignored
			matched, err := util.IsGlobExpMatch(newPath, ignores)
			if err != nil {
				return errors.Wrapf(err, "unable to addRecursiveWatch on %s", newPath)
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

		if matched, _ := util.IsGlobExpMatch(folder, ignores); matched {
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

// ErrUserRequestedWatchExit is returned when the user stops the watch loop
var ErrUserRequestedWatchExit = fmt.Errorf("safely exiting from filesystem watch based on user request")

// WatchAndPush watches path, if something changes in  that path it calls PushLocal
// ignores .git/* by default
// inspired by https://github.com/openshift/origin/blob/e785f76194c57bd0e1674c2f2776333e1e0e4e78/pkg/oc/cli/cmd/rsync/rsync.go#L257
// Parameters:
//	client: occlient instance
//	out: io Writer instance
// 	parameters: WatchParameters
func WatchAndPush(client *occlient.Client, out io.Writer, parameters WatchParameters) error {
	// ToDo reduce number of parameters to this function by extracting them into a struct and passing the struct instance instead of passing each of them separately
	// delayInterval int
	klog.V(4).Infof("starting WatchAndPush, path: %s, component: %s, ignores %s", parameters.Path, parameters.ComponentName, parameters.FileIgnores)

	// these variables must be accessed while holding the changeLock
	// mutex as they are shared between goroutines to communicate
	// sync state/events.
	var (
		changeLock   sync.Mutex
		dirty        bool
		lastChange   time.Time
		watchError   error
		deletedPaths []string
		changedFiles []string
	)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error setting up filesystem watcher: %v", err)
	}
	defer watcher.Close()
	defer close(parameters.ExtChan)

	go func() {
		for {
			select {
			case extMsg := <-parameters.ExtChan:
				if extMsg {
					changeLock.Lock()
					watchError = ErrUserRequestedWatchExit
					changeLock.Unlock()
				}
			case event := <-watcher.Events:
				isIgnoreEvent := false
				changeLock.Lock()
				klog.V(4).Infof("filesystem watch event: %s", event)

				if !(event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename) {
					stat, err := os.Lstat(event.Name)
					if err != nil {
						// Some of the editors like vim and gedit, generate temporary buffer files during update to the file and deletes it soon after exiting from the editor
						// So, its better to log the error rather than feeding it to error handler via `watchError = errors.Wrap(err, "unable to watch changes")`,
						// which will terminate the watch
						klog.V(4).Infof("Failed getting details of the changed file %s. Ignoring the change", event.Name)
					}
					// Some of the editors generate temporary buffer files during update to the file and deletes it soon after exiting from the editor
					// So, its better to log the error rather than feeding it to error handler via `watchError = errors.Wrap(err, "unable to watch changes")`,
					// which will terminate the watch
					if stat == nil {
						klog.V(4).Infof("Ignoring event for file %s as details about the file couldn't be fetched", event.Name)
						isIgnoreEvent = true
					}

					// In windows, every new file created under a sub-directory of the watched directory, raises 2 events:
					// 1. Write event for the directory under which the file was created
					// 2. Create event for the file that was created
					// Ignore 1 to avoid duplicate events.
					if isIgnoreEvent || (stat.IsDir() && event.Op&fsnotify.Write == fsnotify.Write) {
						isIgnoreEvent = true
					}
				}

				// add file name to changedFiles only once
				alreadyInChangedFiles := false
				for _, cfile := range changedFiles {
					if cfile == event.Name {
						alreadyInChangedFiles = true
						break
					}
				}

				// Filter out anything in ignores list from the list of changed files
				// This is important inspite of not watching the
				// ignores paths because, when a directory that is ignored, is deleted,
				// because its parent is watched, the fsnotify automatically raises an event
				// for it.
				matched, err := util.IsGlobExpMatch(event.Name, parameters.FileIgnores)
				klog.V(4).Infof("Matching %s with %s. Matched %v, err: %v", event.Name, parameters.FileIgnores, matched, err)
				if err != nil {
					watchError = errors.Wrap(err, "unable to watch changes")
				}
				if !alreadyInChangedFiles && !matched && !isIgnoreEvent {
					// Append the new file change event to changedFiles if and only if the event is not a file remove event
					if event.Op&fsnotify.Remove != fsnotify.Remove {
						changedFiles = append(changedFiles, event.Name)
					}
				}

				lastChange = time.Now()
				dirty = true
				// Rename operation triggers RENAME event on old path + CREATE event for renamed path so delete old path in case of rename
				// Also weirdly, fsnotify raises a RENAME event for deletion of files/folders with space in their name so even that should be handled here
				if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
					if e := watcher.Remove(event.Name); e != nil {
						klog.V(4).Infof("error removing watch for %s: %v", event.Name, e)
					}
					// append the file to list of deleted files
					// When a file/folder is deleted, it raises 2 events:
					//	a. RENAME with event.Name empty
					//	b. REMOVE with event.Name as file name
					if !alreadyInChangedFiles && !matched && event.Name != "" {
						relPath, err := filepath.Rel(parameters.Path, event.Name)
						if err != nil {
							watchError = errors.Wrapf(err, "failed to propagate delete of file %s as its relative to %s couldn't be found", event.Name, parameters.Path)
						}
						deletedPaths = append(deletedPaths, relPath)
					}
				} else {
					if e := addRecursiveWatch(watcher, event.Name, parameters.FileIgnores); e != nil && watchError == nil {
						watchError = e
					}
				}
				changeLock.Unlock()
			case err := <-watcher.Errors:
				changeLock.Lock()
				watchError = fmt.Errorf("error watching filesystem for changes: %v", err)
				changeLock.Unlock()
			}
		}
	}()
	// adding watch on the root folder and the sub folders recursively
	// so directory and the path in addRecursiveWatch() are the same
	err = addRecursiveWatch(watcher, parameters.Path, parameters.FileIgnores)
	if err != nil {
		return fmt.Errorf("error watching source path %s: %v", parameters.Path, err)
	}

	// Only signal start of watch if invoker is interested
	if parameters.StartChan != nil {
		parameters.StartChan <- true
	}

	delay := time.Duration(parameters.PushDiffDelay) * time.Second
	ticker := time.NewTicker(delay)
	showWaitingMessage := true
	defer ticker.Stop()
	for {
		changeLock.Lock()
		if watchError != nil {
			return watchError
		}
		if showWaitingMessage {
			fmt.Fprintf(out, "Waiting for something to change in %s\n", parameters.Path)
			showWaitingMessage = false
		}
		// if a change happened more than 'delay' seconds ago, sync it now.
		// if a change happened less than 'delay' seconds ago, sleep for 'delay' seconds
		// and see if more changes happen, we don't want to sync when
		// the filesystem is in the middle of changing due to a massive
		// set of changes (such as a local build in progress).
		if dirty && time.Now().After(lastChange.Add(delay)) {
			for _, file := range changedFiles {
				fmt.Fprintf(out, "File %s changed\n", file)
			}
			if len(changedFiles) > 0 || len(deletedPaths) > 0 {
				fmt.Fprintf(out, "Pushing files...\n")
				fileInfo, err := os.Stat(parameters.Path)
				if err != nil {
					return errors.Wrapf(err, "%s: file doesn't exist", parameters.Path)
				}
				if fileInfo.IsDir() {
					klog.V(4).Infof("Copying files %s to pod", changedFiles)
					if parameters.DevfileWatchHandler != nil {
						pushParams := common.PushParameters{
							Path:              parameters.Path,
							WatchFiles:        changedFiles,
							WatchDeletedFiles: deletedPaths,
							IgnoredFiles:      parameters.FileIgnores,
							ForceBuild:        false,
						}

						err = parameters.DevfileWatchHandler(pushParams)
					} else {
						err = parameters.WatchHandler(client, parameters.ComponentName, parameters.ApplicationName, parameters.Path, out, changedFiles, deletedPaths, false, parameters.FileIgnores, parameters.Show)
					}

				} else {
					pathDir := filepath.Dir(parameters.Path)
					klog.V(4).Infof("Copying file %s to pod", parameters.Path)
					if parameters.DevfileWatchHandler != nil {
						pushParams := common.PushParameters{
							Path:              pathDir,
							WatchFiles:        changedFiles,
							WatchDeletedFiles: deletedPaths,
							IgnoredFiles:      parameters.FileIgnores,
							ForceBuild:        false,
						}

						err = parameters.DevfileWatchHandler(pushParams)
					} else {
						err = parameters.WatchHandler(client, parameters.ComponentName, parameters.ApplicationName, pathDir, out, []string{parameters.Path}, deletedPaths, false, parameters.FileIgnores, parameters.Show)
					}

				}
				if err != nil {
					// Intentionally not exiting on error here.
					// We don't want to break watch when push failed, it might be fixed with the next change.
					klog.V(4).Infof("Error from Push: %v", err)
				}
				dirty = false
				showWaitingMessage = true
				// Reset changed files
				changedFiles = []string{}
				// Reset deleted Paths
				deletedPaths = []string{}
			}
		}
		changeLock.Unlock()
		<-ticker.C
	}
}

// DevfileWatchAndPush calls out to the WatchAndPush function.
// As an occlient instance is not needed for devfile components, it sets it to nil
func DevfileWatchAndPush(out io.Writer, parameters WatchParameters) error {
	return WatchAndPush(nil, out, parameters)
}
