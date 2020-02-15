package testingutil

import (
	"fmt"
	"github.com/openshift/odo/pkg/testingutil/filesystem"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// TempMkdir creates a temporary directory
func TempMkdir(parentDir string, newDirPrefix string) (string, error) {
	parentDir = filepath.FromSlash(parentDir)
	dir, err := ioutil.TempDir(parentDir, newDirPrefix)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create dir with prefix %s in directory %s", newDirPrefix, parentDir)
	}
	return dir, nil
}

// TempMkFile creates a temporary file.
func TempMkFile(dir string, fileName string) (string, error) {
	dir = filepath.FromSlash(dir)
	f, err := ioutil.TempFile(dir, fileName)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create test file %s in dir %s", fileName, dir)
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return f.Name(), nil
}

// FileType custom type to indicate type of file
type FileType int

const (
	// RegularFile enum to represent regular file
	RegularFile FileType = 0
	// Directory enum to represent directory
	Directory FileType = 1
)

// ModificationType custom type to indicate file modification type
type ModificationType string

const (
	// UPDATE enum representing update operation on a file
	UPDATE ModificationType = "update"
	// CREATE enum representing create operation for a file/folder
	CREATE ModificationType = "create"
	// DELETE enum representing delete operation for a file/folder
	DELETE ModificationType = "delete"
	// APPEND enum representing append operation on a file
	APPEND ModificationType = "append"
)

// FileProperties to contain meta-data of a file like, file/folder name, file/folder parent dir, file type and desired file modification type
type FileProperties struct {
	FilePath         string
	FileParent       string
	FileType         FileType
	ModificationType ModificationType
}

// SimulateFileModifications mock function to simulate requested file/folder operation
// Parameters:
//	basePath: The parent directory for file/folder involved in desired file operation
//	fileModification: Meta-data of file/folder
// Returns:
//	path to file/folder involved in the operation
//	error if any or nil
func SimulateFileModifications(basePath string, fileModification FileProperties) (string, error) {
	// Files/folders intended to be directly under basepath will be indicated by fileModification.FileParent set to empty string
	if fileModification.FileParent != "" {
		// If fileModification.FileParent is not empty, use it to generate file/folder absolute path
		basePath = filepath.Join(basePath, fileModification.FileParent)
	}

	switch fileModification.ModificationType {
	case CREATE:
		if fileModification.FileType == Directory {
			filePath, err := TempMkdir(basePath, fileModification.FilePath)
			// t.Logf("In simulateFileModifications, Attempting to create folder %s in %s. Error : %v", fileModification.filePath, basePath, err)
			return filePath, err
		} else if fileModification.FileType == RegularFile {
			folderPath, err := TempMkFile(basePath, fileModification.FilePath)
			// t.Logf("In simulateFileModifications, Attempting to create file %s in %s", fileModification.filePath, basePath)
			return folderPath, err
		}
	case DELETE:
		if fileModification.FileType == Directory {
			return filepath.Join(basePath, fileModification.FilePath), os.RemoveAll(filepath.Join(basePath, fileModification.FilePath))
		} else if fileModification.FileType == RegularFile {
			return filepath.Join(basePath, fileModification.FilePath), os.Remove(filepath.Join(basePath, fileModification.FilePath))
		}
	case UPDATE:
		if fileModification.FileType == Directory {
			return "", fmt.Errorf("Updating directory %s is not supported", fileModification.FilePath)
		} else if fileModification.FileType == RegularFile {
			f, err := os.Open(filepath.Join(basePath, fileModification.FilePath))
			if err != nil {
				return "", err
			}
			if _, err := f.WriteString("Hello from Odo"); err != nil {
				return "", err
			}
			if err := f.Sync(); err != nil {
				return "", err
			}
			if err := f.Close(); err != nil {
				return "", err
			}
			return filepath.Join(basePath, fileModification.FilePath), nil
		}
	case APPEND:
		if fileModification.FileType == RegularFile {
			err := ioutil.WriteFile(filepath.Join(basePath, fileModification.FilePath), []byte("// Check watch command"), os.ModeAppend)
			if err != nil {
				return "", err
			}
			return filepath.Join(basePath, fileModification.FilePath), nil
		}

		return "", fmt.Errorf("Append not supported for file of type %v", fileModification.FileType)

	default:
		return "", fmt.Errorf("Unsupported file operation %s", fileModification.ModificationType)
	}
	return "", nil
}

func MkFileWithContent(path, content string, fs filesystem.Filesystem) (string, error) {
	path = filepath.FromSlash(path)
	f, err := fs.Create(path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create file in path %s", path)
	}
	_, err = f.WriteString(content)
	if err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return f.Name(), nil
}
