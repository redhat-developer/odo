package watch

import (
	"fmt"
	"os"
	"path/filepath"

	dfutil "github.com/devfile/library/pkg/util"
	"github.com/fsnotify/fsnotify"
	"github.com/redhat-developer/odo/pkg/util"
	gitignore "github.com/sabhiram/go-gitignore"
	"k8s.io/klog"
)

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
