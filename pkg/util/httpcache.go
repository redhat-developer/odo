package util

import (
	"path/filepath"

	"github.com/devfile/library/v2/pkg/testingutil/filesystem"
	"k8s.io/klog"
)

// CleanDefaultHTTPCacheDir cleans the default directory used for HTTP caching
func CleanDefaultHTTPCacheDir() error {
	return cleanDefaultHTTPCacheDir(filesystem.DefaultFs{})
}

func cleanDefaultHTTPCacheDir(fs filesystem.Filesystem) error {
	cacheFiles, err := fs.ReadDir(httpCacheDir)
	if err != nil {
		return err
	}

	for _, f := range cacheFiles {
		klog.V(4).Infof("Removing cache file %s", f.Name())
		err := fs.Remove(filepath.Join(httpCacheDir, f.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}
