package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"k8s.io/klog"
)

// cleanHttpCache checks cacheDir and deletes all files that were modified more than cacheTime back
func cleanHttpCache(cacheDir string, cacheTime time.Duration) error {
	cacheFiles, err := ioutil.ReadDir(cacheDir)
	if err != nil {
		return err
	}

	for _, f := range cacheFiles {
		if f.ModTime().Add(cacheTime).Before(time.Now()) {
			klog.V(4).Infof("Removing cache file %s, because it is older than %s", f.Name(), cacheTime.String())
			err := os.Remove(filepath.Join(cacheDir, f.Name()))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
