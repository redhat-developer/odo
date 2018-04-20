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
	log "github.com/sirupsen/logrus"
)

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
		log.Debugf("adding watch on path %s", path)
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
			log.Debugf("ignoring watch for %s", v)
			continue
		}
		log.Debugf("adding watch on path %s", v)
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
func WatchAndPush(client *occlient.Client, componentName string, applicationName, path string, asFile bool, out io.Writer) error {
	log.Debugf("starting WatchAndPush, path: %s, component: %s", path, componentName)

	// it might be better to expose this as argument in the future
	ignores := []string{
		// ignore git as it can change even if no source file changed
		// for example some plugins providing git info in PS1 doing that
		".*\\.git.*",
	}

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
				log.Debugf("filesystem watch event: %s", event)

				// add file name to changedFiles only once
				alreadyInChangedFiles := false
				for _, cfile := range changedFiles {
					if cfile == event.Name {
						alreadyInChangedFiles = true
						break
					}
				}
				if !alreadyInChangedFiles {
					changedFiles = append(changedFiles, event.Name)
				}

				lastChange = time.Now()
				dirty = true
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					if e := watcher.Remove(event.Name); e != nil {
						log.Debugf("error removing watch for %s: %v", event.Name, e)
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

	delay := 1 * time.Second
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
			err := PushLocal(client, componentName, applicationName, path, asFile, out)
			if err != nil {
				// Intentionally not exiting on error here.
				// We don't want to break watch when push failed, it might be fixed with the next change.
				log.Debug("Error from PushLocal: %v", err)
			}
			dirty = false
			showWaitingMessage = true
		}
		changeLock.Unlock()
		<-ticker.C
	}
}
