package location

import (
	"os"
	"path/filepath"

	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"
	"github.com/redhat-developer/odo/pkg/util"
)

// possibleDevfileNames contains possivle devfile name that should be checked in the context dir.
var possibleDevfileNames = [...]string{"devfile.yaml", ".devfile.yaml"}

// DevfileFilenamesProvider checks if the context dir contains devfile with possible names from possibleDevfileNames variable,
// else it returns "devfile.yaml" as default value.
func DevfileFilenamesProvider(contexDir string) string {
	for _, devFile := range possibleDevfileNames {
		if util.CheckPathExists(filepath.Join(contexDir, devFile)) {
			return devFile
		}
	}
	return "devfile.yaml"
}

// DevfileLocation returns path to devfile File with approprite devfile File name from possibleDevfileNames variable,
// if contextDir value is provided as argument then DevfileLocation returns devfile path
// else it assumes current dir as contextDir and returns appropriate value.
func DevfileLocation(contexDir string) string {
	if contexDir == "" {
		contexDir = "./"
	}
	devFile := DevfileFilenamesProvider(contexDir)
	return filepath.Join(contexDir, devFile)
}

// IsDevfileName returns true if name is a supported name for a devfile
func IsDevfileName(name string) bool {
	for _, devFile := range possibleDevfileNames {
		if devFile == name {
			return true
		}
	}
	return false
}

// DirectoryContainsDevfile returns true if the given directory contains a devfile with a supported name
func DirectoryContainsDevfile(fsys filesystem.Filesystem, dir string) (bool, error) {
	for _, devFile := range possibleDevfileNames {
		_, err := fsys.Stat(filepath.Join(dir, devFile))
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return false, err
		}
		// path to file does exist
		return true, nil
	}
	return false, nil
}

// DirIsEmpty returns true if the given directory contains no file
func DirIsEmpty(fsys filesystem.Filesystem, path string) (bool, error) {
	files, err := fsys.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(files) == 0, nil
}

// DirContainsOnlyDevfile returns true if the directory contains only one file which is a devfile
// with a supported name
func DirContainsOnlyDevfile(fsys filesystem.Filesystem, path string) (bool, error) {
	files, err := fsys.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(files) == 1 && IsDevfileName(files[0].Name()), nil
}
