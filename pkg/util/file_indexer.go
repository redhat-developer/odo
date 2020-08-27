package util

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/odo/pkg/testingutil/filesystem"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
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

// ReadFileIndex tries to read the odo index file from the given location and returns the data from the file
// if no such file is present, it means the folder hasn't been walked and thus returns a empty list
func ReadFileIndex(filePath string) (*FileIndex, error) {
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

// ResolveIndexFilePath resolves the filepath of the odo index file in the .odo folder
func ResolveIndexFilePath(directory string) (string, error) {
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

// AddOdoFileIndex adds odo-file-index.json to .gitignore
func AddOdoFileIndex(gitIgnoreFile string) error {
	return addOdoFileIndex(gitIgnoreFile, filesystem.DefaultFs{})
}

func addOdoFileIndex(gitIgnoreFile string, fs filesystem.Filesystem) error {
	return addFileToIgnoreFile(gitIgnoreFile, filepath.Join(fileIndexDirectory, fileIndexName), fs)
}

// CheckGitIgnoreFile checks .gitignore file exists or not, if not then create it
func CheckGitIgnoreFile(directory string) (string, error) {
	return checkGitIgnoreFile(directory, filesystem.DefaultFs{})
}

func checkGitIgnoreFile(directory string, fs filesystem.Filesystem) (string, error) {

	_, err := fs.Stat(directory)
	if err != nil {
		return "", err
	}

	gitIgnoreFile := filepath.Join(directory, ".gitignore")

	// err checks the existence of .gitignore and then creates if does not exists
	if _, err := fs.Stat(gitIgnoreFile); os.IsNotExist(err) {
		file, err := fs.OpenFile(gitIgnoreFile, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			return gitIgnoreFile, errors.Wrap(err, "failed to create .gitignore file")
		}
		file.Close()
	}

	return gitIgnoreFile, nil
}

// DeleteIndexFile deletes the index file. It doesn't throw error if it doesn't exist
func DeleteIndexFile(directory string) error {
	indexFile, err := ResolveIndexFilePath(directory)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return DeletePath(indexFile)
}

// IndexerRet is a struct that represent return value of RunIndexer function
type IndexerRet struct {
	FilesChanged []string
	FilesDeleted []string
	NewFileMap   map[string]FileData
	ResolvedPath string
}

// RunIndexer walks the given directory and finds the files which have changed and which were deleted/renamed
// it reads the odo index file from the .odo folder
// if no such file is present, it means it's the first time the folder is being walked and thus returns a empty list
// after the walk, it stores the list of walked files with some information in a odo index file in the .odo folder
// The filemap stores the values as "relative filepath" => FileData but it the FilesChanged and filesDeleted are absolute paths
// to the files
func RunIndexer(directory string, ignoreRules []string) (ret IndexerRet, err error) {
	directory = filepath.FromSlash(directory)
	ret.ResolvedPath, err = ResolveIndexFilePath(directory)

	if err != nil {
		return ret, err
	}

	// check for .gitignore file and add odo-file-index.json to .gitignore
	gitIgnoreFile, err := CheckGitIgnoreFile(directory)
	if err != nil {
		return ret, err
	}

	// add odo-file-index.json path to .gitignore
	err = AddOdoFileIndex(gitIgnoreFile)
	if err != nil {
		return ret, err
	}

	// read the odo index file
	existingFileIndex, err := ReadFileIndex(ret.ResolvedPath)
	if err != nil {
		return ret, err
	}

	ret.NewFileMap = make(map[string]FileData)
	walk := func(walkFnPath string, fi os.FileInfo, err error) error {

		if err != nil {
			return err
		}
		if fi.IsDir() {

			// if folder is the root folder, don't add it
			if walkFnPath == directory {
				return nil
			}

			match, err := IsGlobExpMatch(walkFnPath, ignoreRules)
			if err != nil {
				return err
			}
			// the folder matches a glob rule and thus should be skipped
			if match {
				return filepath.SkipDir
			}

			if fi.Name() == fileIndexDirectory || fi.Name() == ".git" {
				klog.V(4).Info(".odo or .git directory detected, skipping it")
				return filepath.SkipDir
			}
		}

		relativeFilename, err := CalculateFileDataKeyFromPath(walkFnPath, directory)
		if err != nil {
			return err
		}

		if _, ok := existingFileIndex.Files[relativeFilename]; !ok {
			ret.FilesChanged = append(ret.FilesChanged, walkFnPath)
			klog.V(4).Infof("file added: %s", walkFnPath)
		} else if !fi.ModTime().Equal(existingFileIndex.Files[relativeFilename].LastModifiedDate) {
			ret.FilesChanged = append(ret.FilesChanged, walkFnPath)
			klog.V(4).Infof("last modified date changed: %s", walkFnPath)
		} else if fi.Size() != existingFileIndex.Files[relativeFilename].Size {
			ret.FilesChanged = append(ret.FilesChanged, walkFnPath)
			klog.V(4).Infof("size changed: %s", walkFnPath)
		}

		ret.NewFileMap[relativeFilename] = FileData{
			Size:             fi.Size(),
			LastModifiedDate: fi.ModTime(),
		}
		return nil
	}

	err = filepath.Walk(directory, walk)
	if err != nil {
		return ret, err
	}

	// find files which are deleted/renamed
	for fileName := range existingFileIndex.Files {
		if _, ok := ret.NewFileMap[fileName]; !ok {
			klog.V(4).Infof("Deleting file: %s", fileName)

			// Return the *absolute* path to the file)
			fileAbsolutePath, err := GetAbsPath(filepath.Join(directory, fileName))
			if err != nil {
				return ret, errors.Wrapf(err, "unable to retrieve absolute path of file %s", fileName)
			}
			ret.FilesDeleted = append(ret.FilesDeleted, fileAbsolutePath)
		}
	}

	return ret, nil
}

// DeployRunIndexer walks the given directory and returns all of the files that are found, that don't match the ignore criteria
func DeployRunIndexer(directory string, ignoreRules []string) (files []string, err error) {
	directory = filepath.FromSlash(directory)

	// check for .gitignore file and add odo-file-index.json to .gitignore
	gitIgnoreFile, err := CheckGitIgnoreFile(directory)
	if err != nil {
		return files, err
	}

	// add odo-file-index.json path to .gitignore
	err = AddOdoFileIndex(gitIgnoreFile)
	if err != nil {
		return files, err
	}

	// Create a function to be passed as a parameter to filepath.Walk
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
				klog.V(4).Info(".odo or .git directory detected, skipping it")
				return filepath.SkipDir
			}
		}

		files = append(files, fn)
		return nil
	}

	err = filepath.Walk(directory, walk)
	if err != nil {
		return files, err
	}

	return files, nil
}

// CalculateFileDataKeyFromPath converts an absolute path to relative (and converts to OS-specific paths) for use
// as a map key in IndexerRet and FileIndex
func CalculateFileDataKeyFromPath(absolutePath string, rootDirectory string) (string, error) {

	rootDirectory = filepath.FromSlash(rootDirectory)

	relativeFilename, err := filepath.Rel(rootDirectory, absolutePath)
	if err != nil {
		return "", err
	}

	return relativeFilename, nil
}

// GenerateNewFileDataEntry creates a new FileData entry for use by IndexerRet and/or FileIndex
func GenerateNewFileDataEntry(absolutePath string, rootDirectory string) (string, *FileData, error) {

	relativeFilename, err := CalculateFileDataKeyFromPath(absolutePath, rootDirectory)
	if err != nil {
		return "", nil, err
	}

	fi, err := os.Stat(absolutePath)

	if err != nil {
		return "", nil, err
	}
	return relativeFilename, &FileData{
		Size:             fi.Size(),
		LastModifiedDate: fi.ModTime(),
	}, nil
}

// write writes the map of walked files and info about them, in a file
// filePath is the location of the file to which it is supposed to be written
func write(filePath string, fi *FileIndex) error {
	jsonData, err := json.Marshal(fi)
	if err != nil {
		return err
	}
	// 0600 is the mask used when a file is created using os.Create hence defaulting
	return ioutil.WriteFile(filePath, jsonData, 0600)
}

// WriteFile writes a file map to a file, the file map is given by
// newFileMap param and the file location is resolvedPath param
func WriteFile(newFileMap map[string]FileData, resolvedPath string) error {
	newfi := NewFileIndex()
	newfi.Files = newFileMap
	err := write(resolvedPath, newfi)

	return err
}
