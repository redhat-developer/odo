package component

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/redhat-developer/odo/pkg/occlient"

	"github.com/fsnotify/fsnotify"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

// isRegExpMatch compiles strToMatch against each of the passed regExps
// Parameters: a string strToMatch and a list of regexp patterns to match strToMatch with
// Returns: true if there is any match else false
func isRegExpMatch(strToMatch string, regExps []string) (bool, error) {
	for _, regExp := range regExps {
		matched, err := regexp.MatchString(regExp, strToMatch)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}
	return false, nil
}

// addRecursiveWatch handles adding watches recursively for the path provided
// and its subdirectories.  If a non-directory is specified, this call is a no-op.
// Files matching glob pattern defined in ignores will be ignored.
// Taken from https://github.com/openshift/origin/blob/85eb37b34f0657631592356d020cef5a58470f8e/pkg/util/fsnotification/fsnotification.go
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
		glog.V(4).Infof("adding watch on path %s", path)
		err = watcher.Add(path)
		if err != nil {
			return fmt.Errorf("error adding watcher for path %s: %v", path, err)
		}
		return nil
	}

	folders := []string{}
	err = filepath.Walk(path, func(newPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// If the current directory matches any of the ignore patterns, ignore them so that their contents are also not ignored
			matched, err := isRegExpMatch(newPath, ignores)
			if err != nil {
				return errors.Wrapf(err, "unable to addRecursiveWatch on %s", newPath)
			}
			if matched {
				glog.V(4).Infof("ignoring watch on path %s", newPath)
				return filepath.SkipDir
			}
			// Append the folder we just walked on
			folders = append(folders, newPath)
		}
		return nil
	})
	for _, v := range folders {
		ignore := false
		for _, pattern := range ignores {
			if matched, _ := regexp.MatchString(pattern, v); matched {
				ignore = true
				break
			}
		}
		if ignore {
			glog.V(4).Infof("ignoring watch for %s", v)
			continue
		}
		glog.V(4).Infof("adding watch on path %s", v)
		err = watcher.Add(v)
		if err != nil {
			// Linux "no space left on device" issues are usually resolved via
			// $ sudo sysctl fs.inotify.max_user_watches=65536
			// BSD / OSX: "too many open files" issues are ussualy resolved via
			// $ sysctl variables "kern.maxfiles" and "kern.maxfilesperproc",
			return fmt.Errorf("error adding watcher for path %s: %v", v, err)
		}
	}
	return nil
}

// WatchAndPush watches path, if something changes in  that path it calls PushLocal
// ignores .git/* by default
// inspired by https://github.com/openshift/origin/blob/e785f76194c57bd0e1674c2f2776333e1e0e4e78/pkg/oc/cli/cmd/rsync/rsync.go#L257
func WatchAndPush(client *occlient.Client, componentName string, applicationName, path string, out io.Writer, ignores []string, delayInterval int) error {
	glog.V(4).Infof("starting WatchAndPush, path: %s, component: %s, ignores %s", path, componentName, ignores)

	// these variables must be accessed while holding the changeLock
	// mutex as they are shared between goroutines to communicate
	// sync state/events.
	var (
		changeLock   sync.Mutex
		dirty        bool
		lastChange   time.Time
		watchError   error
		changedFiles []string
	)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error setting up filesystem watcher: %v", err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				changeLock.Lock()
				glog.V(4).Infof("filesystem watch event: %s", event)

				stat, err := os.Lstat(event.Name)
				if err != nil {
					glog.Errorf("Failed getting details of the changed file %s", event.Name)
					watchError = errors.Wrap(err, "unable to watch changes")
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
				matched, err := isRegExpMatch(event.Name, ignores)
				if err != nil {
					watchError = errors.Wrap(err, "unable to watch changes")
				}
				if !alreadyInChangedFiles && !matched {
					// In windows, every new file created under a sub-directory of the watched directory, raises 2 events:
					// 1. Write event for the directory under which the file was created
					// 2. Create event for the file that was created
					// Ignore 1 to avoid duplicate events.
					if !(stat.IsDir() && event.Op&fsnotify.Write == fsnotify.Write) {
						changedFiles = append(changedFiles, event.Name)
					}
				}

				lastChange = time.Now()
				dirty = true
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					if e := watcher.Remove(event.Name); e != nil {
						glog.V(4).Infof("error removing watch for %s: %v", event.Name, e)
					}
				} else {
					if e := addRecursiveWatch(watcher, event.Name, ignores); e != nil && watchError == nil {
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
	err = addRecursiveWatch(watcher, path, ignores)
	if err != nil {
		return fmt.Errorf("error watching source path %s: %v", path, err)
	}

	delay := time.Duration(delayInterval) * time.Second
	ticker := time.NewTicker(delay)
	showWaitingMessage := true
	defer ticker.Stop()
	for {
		changeLock.Lock()
		if watchError != nil {
			return watchError
		}
		if showWaitingMessage {
			fmt.Fprintf(out, "Waiting for something to change in %s\n", path)
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
			fmt.Fprintf(out, "Pushing files...\n")
			fileInfo, err := os.Stat(path)
			if err != nil {
				return errors.Wrapf(err, "%s: file doesn't exist", path)
			}
			if fileInfo.IsDir() {
				glog.V(4).Infof("Copying files %s to pod", changedFiles)
				err = PushLocal(client, componentName, applicationName, path, out, changedFiles)
			} else {
				pathDir := filepath.Dir(path)
				glog.V(4).Infof("Copying file %s to pod", path)
				err = PushLocal(client, componentName, applicationName, pathDir, out, []string{path})
			}
			if err != nil {
				// Intentionally not exiting on error here.
				// We don't want to break watch when push failed, it might be fixed with the next change.
				glog.V(4).Infof("Error from PushLocal: %v", err)
			}
			dirty = false
			showWaitingMessage = true
		}
		changeLock.Unlock()
		<-ticker.C
	}
}
