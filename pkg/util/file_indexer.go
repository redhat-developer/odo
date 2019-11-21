package util

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const fileIndexDirectory = ".odo"
const fileIndexName = "odo-file-index.json"

// FileIndex holds the file index used for storing local file state change
type FileIndex struct {
	metav1.TypeMeta
	Files map[string]FileData
}

// NewFileIndex returns a fileIndex
func NewFileIndex() *FileIndex {

	return &FileIndex{
		TypeMeta: metav1.TypeMeta{
			Kind:       "FileIndex",
			APIVersion: "v1",
		},
		Files: make(map[string]FileData),
	}
}

type FileData struct {
	Size             int64
	LastModifiedDate time.Time
}

// read tries to read the odo index file from the given location and returns the data from the file
// if no such file is present, it means the folder hasn't been walked and thus returns a empty list
func readFileIndex(filePath string) (*FileIndex, error) {
	// Read operation
	var fi FileIndex
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return NewFileIndex(), nil
	}

	byteValue, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	// Unmarshals the byte values and fill up the file read map
	err = json.Unmarshal(byteValue, &fi)
	if err != nil {
		// This is added here for backward compatibility because
		// if the marshalling fails then that means we are trying to
		// read a very old version of the index and hence we can just
		// ignore it and reset index
		// TODO: we need to remove this later
		return NewFileIndex(), nil
	}
	return &fi, nil
}

// resolveIndexFilePath resolves the filepath of the odo index file in the .odo folder
func resolveIndexFilePath(directory string) (string, error) {
	directoryFi, err := os.Stat(filepath.Join(directory))
	if err != nil {
		return "", err
	}

	switch mode := directoryFi.Mode(); {
	case mode.IsDir():
		// do directory stuff
		return filepath.Join(directory, fileIndexDirectory, fileIndexName), nil
	case mode.IsRegular():
		// do file stuff
		// for binary files
		return filepath.Join(filepath.Dir(directory), fileIndexDirectory, fileIndexName), nil
	}

	return directory, nil
}

// gitignoreFilePath gives the filepath of the .gitignore file in the context
func gitIgnoreFilePath(directory string) (string, error) {
	_, err := os.Stat(directory)
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, ".gitignore"), nil
}

// addOdoFileIndex adds odo-file-index.json to .gitignore
func addOdoFileIndex(gitIgnoreFile string) error {
	var data []byte
	file, err := os.OpenFile(gitIgnoreFile, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return errors.Wrap(err, "failed to open .gitignore file")
	}
	defer file.Close()

	if data, err = ioutil.ReadFile(gitIgnoreFile); err != nil {
		return errors.Wrap(err, "failed reading data from .gitignore file")
	}
	// check whether .odo/odo-file-index.json is already in the .gitignore file
	if !strings.Contains(string(data), filepath.Join(fileIndexDirectory, fileIndexName)) {
		if _, err := file.WriteString("\n" + filepath.Join(fileIndexDirectory, fileIndexName)); err != nil {
			return errors.Wrapf(err, "failed to Add %v to .gitignore file", fileIndexName)
		}
	}
	return nil
}

// Check .gitignore file exists or not, if not then create it
func checkGitIgnoreFile(directory string) (string, error) {

	gitIgnoreFile, err := gitIgnoreFilePath(directory)
	if err != nil {
		return "", err
	}
	// err checks the existence of .gitignore and then creates if does not exists
	if _, err := os.Stat(gitIgnoreFile); os.IsNotExist(err) {
		file, err := os.OpenFile(gitIgnoreFile, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return gitIgnoreFile, errors.Wrap(err, "failed to create .gitignore file")
		}
		file.Close()
	}

	return gitIgnoreFile, nil
}

// DeleteIndexFile deletes the index file. It doesn't throw error if it doesn't exist
func DeleteIndexFile(directory string) error {
	indexFile, err := resolveIndexFilePath(directory)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return DeletePath(indexFile)
}

// RunIndexer walks the given directory and finds the files which have changed and which were deleted/renamed
// it reads the odo index file from the .odo folder
// if no such file is present, it means it's the first time the folder is being walked and thus returns a empty list
// after the walk, it stores the list of walked files with some information in a odo index file in the .odo folder
// The filemap stores the values as "relative filepath" => FileData but it the filesChanged and filesDeleted are absolute paths
// to the files
func RunIndexer(directory string, ignoreRules []string) (filesChanged []string, filesDeleted []string, err error) {
	resolvedPath, err := resolveIndexFilePath(directory)
	if err != nil {
		return filesChanged, filesDeleted, err
	}

	// check for .gitignore file and add odo-file-index.json to .gitignore
	gitIgnoreFile, err := checkGitIgnoreFile(directory)
	if err != nil {
		return filesChanged, filesDeleted, err
	}

	// add odo-file-index.json path to .gitignore
	err = addOdoFileIndex(gitIgnoreFile)
	if err != nil {
		return filesChanged, filesDeleted, err
	}

	// read the odo index file
	existingFileIndex, err := readFileIndex(resolvedPath)
	if err != nil {
		return filesChanged, filesDeleted, err
	}

	newFileMap := make(map[string]FileData)
	walk := func(fn string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {

			// if folder is the root folder, don't add it
			if fn == directory {
				return nil
			}

			match, err := IsGlobExpMatch(fn, ignoreRules)
			if err != nil {
				return err
			}
			// the folder matches a glob rule and thus should be skipped
			if match {
				return filepath.SkipDir
			}

			if fi.Name() == fileIndexDirectory || fi.Name() == ".git" {
				glog.V(4).Info(".odo or .git directory detected, skipping it")
				return filepath.SkipDir
			}
		}

		relFn, err := filepath.Rel(directory, fn)
		if err != nil {
			return err
		}

		if _, ok := existingFileIndex.Files[relFn]; !ok {
			filesChanged = append(filesChanged, fn)
			glog.V(4).Infof("file added: %s", fn)
		} else if !fi.ModTime().Equal(existingFileIndex.Files[relFn].LastModifiedDate) {
			filesChanged = append(filesChanged, fn)
			glog.V(4).Infof("last modified date changed: %s", fn)
		} else if fi.Size() != existingFileIndex.Files[relFn].Size {
			filesChanged = append(filesChanged, fn)
			glog.V(4).Infof("size changed: %s", fn)
		}

		newFileMap[relFn] = FileData{
			Size:             fi.Size(),
			LastModifiedDate: fi.ModTime(),
		}
		return nil
	}

	err = filepath.Walk(directory, walk)
	if err != nil {
		return filesChanged, filesDeleted, err
	}

	// find files which are deleted/renamed
	for fileName := range existingFileIndex.Files {
		if _, ok := newFileMap[fileName]; !ok {
			glog.V(4).Infof("file deleted: %s", fileName)
			// we return the absolute path of the files eventhough we store relative
			filesDeleted = append(filesDeleted, filepath.Join(directory, fileName))
		}

	}

	// if there are added/deleted/modified/renamed files or folders, write it to the odo index file
	if len(filesChanged) > 0 || len(filesDeleted) > 0 {
		newfi := NewFileIndex()
		newfi.Files = newFileMap
		err = write(resolvedPath, newfi)
		if err != nil {
			return filesChanged, filesDeleted, err
		}
	}

	return filesChanged, filesDeleted, nil
}

// writes the map of walked files and info about them, in a file
// filepath is the location of the file to which it is supposed to be written
func write(filePath string, fi *FileIndex) error {
	jsonData, err := json.Marshal(fi)
	if err != nil {
		return err
	}
	// 0666 is the mask used when a file is created using os.Create hence defaulting
	return ioutil.WriteFile(filePath, jsonData, 0666)
}
