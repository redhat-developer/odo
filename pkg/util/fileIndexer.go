package util

import (
	"encoding/json"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const fileIndexDirectory = ".odo"
const fileIndexName = "odo-file-index.json"

type FileData struct {
	Size             int64
	LastModifiedDate time.Time
}

// read tries to read the odo index file from the given location and returns the data from the file
// if no such file is present, it means the folder hasn't been walked and thus returns a empty list
func read(filePath string) (map[string]FileData, error) {
	// read operation
	fileReadMap := make(map[string]FileData)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return map[string]FileData{}, nil
	}

	jsonFileRead, err := os.Open(filePath)
	// if we os.Open returns an error then handle it
	if err != nil {
		return nil, err
	}
	defer jsonFileRead.Close()

	byteValue, _ := ioutil.ReadAll(jsonFileRead)
	// unmarshals the byte values and fill up the file read map
	err = json.Unmarshal(byteValue, &fileReadMap)
	if err != nil {
		return nil, err
	}
	return fileReadMap, nil
}

// resolveFilePath resolves the filepath of the odo index file in the .odo folder
func resolveFilePath(directory string) (string, error) {
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

// Run walks the given directory and finds the files which have changed and which were deleted/renamed
// it reads the odo index file from the .odo folder
// if no such file is present, it means it's the first time the folder is being walked and thus returns a empty list
// after the walk, it stores the list of walked files with some information in a odo index file in the .odo folder
func Run(directory string, ignoreRules []string) (filesChanged []string, filesDeleted []string, err error) {
	resolvedPath, err := resolveFilePath(directory)
	if err != nil {
		return filesChanged, filesDeleted, err
	}

	// read the odo index file
	readData, err := read(resolvedPath)
	if err != nil {
		return filesChanged, filesDeleted, err
	}

	filesMap := make(map[string]FileData)
	walk := func(fn string, fi os.FileInfo, err error) error {
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

		if _, ok := readData[fn]; !ok {
			filesChanged = append(filesChanged, fn)
			glog.V(4).Infof("file added: %s", fn)
		} else if !fi.ModTime().Equal(readData[fn].LastModifiedDate) {
			filesChanged = append(filesChanged, fn)
			glog.V(4).Infof("last modified date changed: %s", fn)
		} else if fi.Size() != readData[fn].Size {
			filesChanged = append(filesChanged, fn)
			glog.V(4).Infof("size changed: %s", fn)
		}

		filesMap[fn] = FileData{
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
	for fileName, _ := range readData {
		if _, ok := filesMap[fileName]; !ok {
			glog.V(4).Infof("file deleted: %s", fileName)
			filesDeleted = append(filesDeleted, fileName)
		}

	}

	// if there are added/deleted/modified/renamed files or folders, write it to the odo index file
	if len(filesChanged) > 0 || len(filesDeleted) > 0 {
		err = write(resolvedPath, filesMap)
		if err != nil {
			return filesChanged, filesDeleted, err
		}
	}

	return filesChanged, filesDeleted, nil
}

// writes the map of walked files and info about them, in a file
// filepath is the location of the file to which it is supposed to be written
func write(filePath string, writeMap map[string]FileData) error {
	jsonData, err := json.Marshal(writeMap)

	jsonFile, err := os.Create(filePath)

	if err != nil {
		return err
	}
	defer jsonFile.Close()

	_, err = jsonFile.Write(jsonData)
	if err != nil {
		return err
	}
	return nil
}
