//
// Copyright Red Hat, Inc.
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

package parser

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/devfile/library/v2/pkg/testingutil/filesystem"
)

// possibleDevfileNames contains possible filenames for a devfile.
// Those are checked in this priority order from a given context dir.
var possibleDevfileNames = []string{
	"devfile.yaml",
	".devfile.yaml",
	"devfile.yml",
	".devfile.yml",
}

// lookupDevfileFromPath returns the file path to use as devfile filename, by looking at the relative path specified in relPath.
// If relPath is not a directory, it is returned as is.
// For backward compatibility, if relPath is a directory, it will try to detect the first existing devfile filename under relPath,
// based on the list of possible devfile filenames defined in the sorted possibleDevfileNames.
// It returns any error found while interacting with the filesystem, or if no file was found from the list of possible devfile names.
func lookupDevfileFromPath(fsys filesystem.Filesystem, relPath string) (string, error) {
	stat, err := fsys.Stat(relPath)
	if err != nil {
		return "", err
	}

	if !stat.IsDir() {
		return relPath, nil
	}

	for _, possibleDevfileName := range possibleDevfileNames {
		p := filepath.Join(relPath, possibleDevfileName)
		if _, err = fsys.Stat(p); errors.Is(err, fs.ErrNotExist) {
			continue
		}
		return p, nil
	}

	return "", fmt.Errorf(
		"the provided path is not a valid yaml filepath, and no possible devfile could be found in the provided path : %s. Possible filenames for a devfile: %s",
		relPath,
		strings.Join(possibleDevfileNames, ", "))
}
