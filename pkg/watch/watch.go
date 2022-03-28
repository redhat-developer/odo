package watch

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	componentlabels "github.com/redhat-developer/odo/pkg/component/labels"

	"github.com/redhat-developer/odo/pkg/kclient"

	"github.com/devfile/library/pkg/devfile/parser"

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
	kubernetesClient kclient.ClientInterface
}

func NewWatchClient(kubernetesClient kclient.ClientInterface) *WatchClient {
	return &WatchClient{
		kubernetesClient: kubernetesClient,
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
	// This is a channel added to signal readiness of the watch command to the external channel listeners
	StartChan chan bool
	// This is a channel added to terminate the watch command gracefully without passing SIGINT. "Stop" message on this channel terminates WatchAndPush function
	ExtChan chan bool
	// Interval of time before pushing changes to remote(component) pod
	PushDiffDelay int
	// Parameter whether or not to show build logs
	Show bool
	// EnvSpecificInfo contains infomation of env.yaml file
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

// ErrUserRequestedWatchExit is returned when the user stops the watch loop
var ErrUserRequestedWatchExit = fmt.Errorf("safely exiting from filesystem watch based on user request")

// WatchAndPush watches path, if something changes in  that path it calls PushLocal
// ignores .git/* by default
// inspired by https://github.com/openshift/origin/blob/e785f76194c57bd0e1674c2f2776333e1e0e4e78/pkg/oc/cli/cmd/rsync/rsync.go#L257
// Parameters:
//	client: occlient instance
//	out: io Writer instance
// 	parameters: WatchParameters
func (o *WatchClient) WatchAndPush(out io.Writer, parameters WatchParameters, cleanupDone chan bool, ctx context.Context) error {
	// ToDo reduce number of parameters to this function by extracting them into a struct and passing the struct instance instead of passing each of them separately
	// delayInterval int
	klog.V(4).Infof("starting WatchAndPush, path: %s, component: %s, ignores %s", parameters.Path, parameters.ComponentName, parameters.FileIgnores)

	// these variables must be accessed while holding the changeLock
	// mutex as they are shared between goroutines to communicate
	// sync state/events.
	var (
		lastChange                     time.Time
		watchError                     error
		deletedPaths                   []string
		changedFiles                   []string
		hasFirstSuccessfulPushOccurred bool
		wg                             sync.WaitGroup
		showWaitingMessage             bool
		delay                          time.Duration
	)

	delay = 1 * time.Second

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error setting up filesystem watcher: %v", err)
	}
	defer watcher.Close()
	defer close(parameters.ExtChan)

	// adding watch on the root folder and the sub folders recursively
	// so directory and the path in addRecursiveWatch() are the same
	err = addRecursiveWatch(watcher, parameters.Path, parameters.FileIgnores)
	if err != nil {
		return fmt.Errorf("error watching source path %s: %v", parameters.Path, err)
	}

	log.Finfof(out, "\nWatching for changes in the current directory %s\n\nPress Ctrl+c to exit.", parameters.Path)
	// This goroutine listens for either file change events from fsnotify, fs errors, or a terminate signal
	// The results are stored in the variables defined in the var( ... ) block above
	wg.Add(1)
	for {
		select {
		case event := <-watcher.Events:
			// check if the delay between events is more than a second; if not don't push again
			if time.Now().After(lastChange.Add(delay)) {
				lastChange = time.Now()

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
				matched, globErr := dfutil.IsGlobExpMatch(event.Name, parameters.FileIgnores)
				klog.V(4).Infof("Matching %s with %s. Matched %v, err: %v", event.Name, parameters.FileIgnores, matched, globErr)
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
					if e := addRecursiveWatch(watcher, event.Name, parameters.FileIgnores); e != nil && watchError == nil {
						klog.V(4).Infof("Error occurred in addRecursiveWatch, setting watchError to %v", e)
						watchError = e
					}
				}

				deletedPaths = removeDuplicates(deletedPaths)

				for _, file := range removeDuplicates(append(changedFiles, deletedPaths...)) {
					fmt.Fprintf(out, "\n\nFile %s changed\n", file)
				}
				if len(changedFiles) > 0 || len(deletedPaths) > 0 {
					fmt.Fprintf(out, "Pushing files...\n\n")
					fileInfo, err := os.Stat(parameters.Path)
					if err != nil {
						log.Errorf("%s: file doesn't exist: %s", parameters.Path, err.Error())
					}
					showWaitingMessage = true
					if fileInfo.IsDir() {
						klog.V(4).Infof("Copying files %s to pod", changedFiles)

						if parameters.DevfileWatchHandler != nil {
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

							err = parameters.DevfileWatchHandler(pushParams, parameters)

						}
					} else {
						pathDir := filepath.Dir(parameters.Path)
						klog.V(4).Infof("Copying file %s to pod", parameters.Path)

						if parameters.DevfileWatchHandler != nil {
							pushParams := common.PushParameters{
								Path:                     pathDir,
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

							err = parameters.DevfileWatchHandler(pushParams, parameters)
						}
					}
					if err != nil {
						// Log and output, but intentionally not exiting on error here.
						// We don't want to break watch when push failed, it might be fixed with the next change.
						klog.V(4).Infof("Error from Push: %v", err)
						fmt.Fprintf(out, "%s - %s\n\n", PushErrorString, err.Error())
					} else {
						hasFirstSuccessfulPushOccurred = true
					}

					// Reset changed files
					changedFiles = []string{}
					// Reset deleted Paths
					deletedPaths = []string{}
				}
				if showWaitingMessage {
					log.Finfof(out, "\nWatching for changes in the current directory %s\n\nPress Ctrl+c to exit.", parameters.Path)
					showWaitingMessage = false
				}
			}
		case watchErr := <-watcher.Errors:
			log.Errorf("error watching filesystem for changes: %v", watchErr.Error())
			wg.Done()
		case <-ctx.Done():
			devfileObj := parameters.InitialDevfileObj
			err = o.kubernetesClient.DeleteDeployment(componentlabels.GetLabels(devfileObj.GetMetadataName(), "app", false))
			if err != nil {
				fmt.Fprintf(out, "failed to delete inner loop resources: %w", err)
			}
			cleanupDone <- true
			wg.Done()
		}
	}
	wg.Wait()
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
