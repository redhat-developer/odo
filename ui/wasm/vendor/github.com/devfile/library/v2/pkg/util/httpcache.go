//
// Copyright 2022 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
