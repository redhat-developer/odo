package util

import (
	"encoding/json"
	"fmt"
	"github.com/monochromegane/go-gitignore"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	dfutil "github.com/devfile/library/pkg/util"

	"github.com/redhat-developer/odo/pkg/testingutil/filesystem"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const DotOdoDirectory = ".odo"
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
	RemoteAttribute  string `json:"RemoteAttribute,omitempty"`
}

// ReadFileIndex tries to read the odo index file from the given location and returns the data from the file
// if no such file is present, it means the folder hasn't been walked and thus returns an empty list
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
		return filepath.Join(directory, DotOdoDirectory, fileIndexName), nil
	case mode.IsRegular():
		// do file stuff
		// for binary files
		return filepath.Join(filepath.Dir(directory), DotOdoDirectory, fileIndexName), nil
	}

	return directory, nil
}

// GetIndexFileRelativeToContext returns the index file relative to context i.e.; .odo/odo-file-index.json
func GetIndexFileRelativeToContext() string {
	return filepath.Join(DotOdoDirectory, fileIndexName)
}

// AddOdoFileIndex adds odo-file-index.json to .gitignore
func AddOdoFileIndex(gitIgnoreFile string) error {
	return addOdoFileIndex(gitIgnoreFile, filesystem.DefaultFs{})
}

func addOdoFileIndex(gitIgnoreFile string, fs filesystem.Filesystem) error {
	return addFileToIgnoreFile(gitIgnoreFile, filepath.Join(DotOdoDirectory, fileIndexName), fs)
}

// TouchGitIgnoreFile checks .gitignore file exists or not, if not then create it
func TouchGitIgnoreFile(directory string) (string, error) {
	return touchGitIgnoreFile(directory, filesystem.DefaultFs{})
}

func touchGitIgnoreFile(directory string, fs filesystem.Filesystem) (string, error) {

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
	return dfutil.DeletePath(indexFile)
}

// IndexerRet is a struct that represent return value of RunIndexer function
type IndexerRet struct {
	FilesChanged  []string
	FilesDeleted  []string
	RemoteDeleted []string
	NewFileMap    map[string]FileData
	ResolvedPath  string
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
	// While 0666 is the mask used when a file is created using os.Create,
	// gosec objects, so use 0600 instead
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

// RunIndexerWithRemote reads the existing index from the given directory and runs the indexer on it
// with the given ignore rules
// it also adds the file index to the .gitignore file and resolves the path
func RunIndexerWithRemote(directory string, ignoreRules []string, remoteDirectories map[string]string) (ret IndexerRet, err error) {
	directory = filepath.FromSlash(directory)
	ret.ResolvedPath, err = ResolveIndexFilePath(directory)
	if err != nil {
		return ret, err
	}

	// check for .gitignore file and add odo-file-index.json to .gitignore
	gitIgnoreFile, err := TouchGitIgnoreFile(directory)
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

	returnedIndex, err := runIndexerWithExistingFileIndex(directory, ignoreRules, remoteDirectories, existingFileIndex)
	if err != nil {
		return IndexerRet{}, err
	}
	returnedIndex.ResolvedPath = ret.ResolvedPath
	return returnedIndex, nil
}

// runIndexerWithExistingFileIndex visits the given directory and creates the new index data
// it ignores the files and folders satisfying the ignoreRules
func runIndexerWithExistingFileIndex(directory string, ignoreRules []string, remoteDirectories map[string]string, existingFileIndex *FileIndex) (ret IndexerRet, err error) {
	destPath := ""
	srcPath := directory

	ret.NewFileMap = make(map[string]FileData)

	fileChanged := make(map[string]bool)
	filesDeleted := make(map[string]bool)
	fileRemoteChanged := make(map[string]bool)

	if len(remoteDirectories) == 0 {
		// The file could be a regular file or even a folder, so use recursiveTar which handles symlinks, regular files and folders
		pathOptions := recursiveCheckerPathOptions{directory, filepath.Dir(srcPath), filepath.Base(srcPath), filepath.Dir(destPath), filepath.Base(destPath)}
		innerRet, err := recursiveChecker(pathOptions, ignoreRules, remoteDirectories, *existingFileIndex)

		if err != nil {
			return IndexerRet{}, err
		}

		for k, v := range innerRet.NewFileMap {
			ret.NewFileMap[k] = v
		}

		for _, remote := range innerRet.FilesChanged {
			fileChanged[remote] = true
		}

		for _, remote := range innerRet.RemoteDeleted {
			fileRemoteChanged[remote] = true
		}

		for _, remote := range innerRet.FilesDeleted {
			filesDeleted[remote] = true
		}
	}

	for remoteAttribute := range remoteDirectories {
		matches, err := filepath.Glob(filepath.Join(directory, remoteAttribute))
		if err != nil {
			return IndexerRet{}, err
		}
		if len(matches) == 0 {
			return IndexerRet{}, fmt.Errorf("path %q doens't exist", remoteAttribute)
		}
		for _, fileName := range matches {
			if checkFileExist(fileName) {
				// Fetch path of source file relative to that of source base path so that it can be passed to recursiveTar
				// which uses path relative to base path for taro header to correctly identify file location when untarred

				// Yes, now that the file exists, now we need to get the absolute path.. if we don't, then when we pass in:
				// 'odo push --context foobar' instead of 'odo push --context ~/foobar' it will NOT work..
				fileAbsolutePath, err := dfutil.GetAbsPath(fileName)
				if err != nil {
					return IndexerRet{}, err
				}
				klog.V(4).Infof("Got abs path: %s", fileAbsolutePath)
				klog.V(4).Infof("Making %s relative to %s", srcPath, fileAbsolutePath)

				// We use "FromSlash" to make this OS-based (Windows uses \, Linux & macOS use /)
				// we get the relative path by joining the two
				destFile, err := filepath.Rel(filepath.FromSlash(srcPath), filepath.FromSlash(fileAbsolutePath))
				if err != nil {
					return IndexerRet{}, err
				}

				// Now we get the source file and join it to the base directory.
				srcFile := filepath.Join(filepath.Base(srcPath), destFile)

				if value, ok := remoteDirectories[filepath.ToSlash(destFile)]; ok {
					destFile = value
				}

				klog.V(4).Infof("makeTar srcFile: %s", srcFile)
				klog.V(4).Infof("makeTar destFile: %s", destFile)

				// The file could be a regular file or even a folder, so use recursiveTar which handles symlinks, regular files and folders
				pathOptions := recursiveCheckerPathOptions{directory, filepath.Dir(srcPath), srcFile, filepath.Dir(destPath), destFile}
				innerRet, err := recursiveChecker(pathOptions, ignoreRules, remoteDirectories, *existingFileIndex)
				if err != nil {
					return IndexerRet{}, err
				}

				for k, v := range innerRet.NewFileMap {
					ret.NewFileMap[k] = v
				}

				for _, remote := range innerRet.FilesChanged {
					fileChanged[remote] = true
				}

				for _, remote := range innerRet.RemoteDeleted {
					fileRemoteChanged[remote] = true
				}

				for _, remote := range innerRet.FilesDeleted {
					filesDeleted[remote] = true
				}
			} else {
				return IndexerRet{}, fmt.Errorf("path %q doens't exist", fileName)
			}
		}
	}

	// find files which are deleted/renamed
	for fileName, value := range existingFileIndex.Files {
		if _, ok := ret.NewFileMap[fileName]; !ok {
			klog.V(4).Infof("Deleting file: %s", fileName)

			if value.RemoteAttribute != "" {
				currentRemote := value.RemoteAttribute
				for _, remote := range findRemoteFolderForDeletion(currentRemote, remoteDirectories) {
					fileRemoteChanged[remote] = true
				}
			} else {
				// check the *absolute* path to the file for glob rules
				fileAbsolutePath, err := dfutil.GetAbsPath(filepath.Join(directory, fileName))
				if err != nil {
					return ret, errors.Wrapf(err, "unable to retrieve absolute path of file %s", fileName)
				}

				matched, err := dfutil.IsGlobExpMatch(fileAbsolutePath, ignoreRules)
				if err != nil {
					return IndexerRet{}, err
				}
				if matched {
					continue
				}
				filesDeleted[fileName] = true
			}
		}
	}

	if len(fileRemoteChanged) > 0 {
		ret.RemoteDeleted = []string{}
	}
	if len(fileChanged) > 0 {
		ret.FilesChanged = []string{}
	}
	if len(filesDeleted) > 0 {
		ret.FilesDeleted = []string{}
	}
	for remote := range fileRemoteChanged {
		ret.RemoteDeleted = append(ret.RemoteDeleted, remote)
	}
	for remote := range fileChanged {
		ret.FilesChanged = append(ret.FilesChanged, remote)
	}
	for remote := range filesDeleted {
		ret.FilesDeleted = append(ret.FilesDeleted, remote)
	}

	return ret, nil
}

// recursiveCheckerPathOptions are the path options for the recursiveChecker function
type recursiveCheckerPathOptions struct {
	// directory of the component
	// srcBase is the base of the file/folder
	// srcFile is the file name
	// destBase is the base of the file's/folder's destination
	// destFile is the base of the file's destination
	directory, srcBase, srcFile, destBase, destFile string
}

// recursiveChecker visits the current source and it's inner files and folders, if any
// the destination values are used to record the appropriate remote location for file or folder
// ignoreRules are used to ignore file and folders
// remoteDirectories are used to find the remote destination of the file/folder and to delete files/folders left behind after the attributes are changed
// existingFileIndex is used to check for file/folder changes
func recursiveChecker(pathOptions recursiveCheckerPathOptions, ignoreRules []string, remoteDirectories map[string]string, existingFileIndex FileIndex) (IndexerRet, error) {
	klog.V(4).Infof("recursiveTar arguments: srcBase: %s, srcFile: %s, destBase: %s, destFile: %s", pathOptions.srcBase, pathOptions.srcFile, pathOptions.destBase, pathOptions.destFile)

	// The destination is a LINUX container and thus we *must* use ToSlash in order
	// to get the copying over done correctly..
	pathOptions.destBase = filepath.ToSlash(pathOptions.destBase)
	pathOptions.destFile = filepath.ToSlash(pathOptions.destFile)
	klog.V(4).Infof("Corrected destinations: base: %s file: %s", pathOptions.destBase, pathOptions.destFile)

	joinedPath := filepath.Join(pathOptions.srcBase, pathOptions.srcFile)
	matchedPathsDir, err := filepath.Glob(joinedPath)
	if err != nil {
		return IndexerRet{}, err
	}

	if len(matchedPathsDir) == 0 {
		return IndexerRet{}, fmt.Errorf("path %q doens't exist", joinedPath)
	}

	joinedRelPath, err := filepath.Rel(pathOptions.directory, joinedPath)
	if err != nil {
		return IndexerRet{}, err
	}

	var ret IndexerRet
	ret.NewFileMap = make(map[string]FileData)

	fileChanged := make(map[string]bool)
	fileRemoteChanged := make(map[string]bool)

	var ignoreMatcher gitignore.IgnoreMatcher
	ignoreMatcher, err = GetIgnoreMatcherFromRules(pathOptions.directory, ignoreRules)
	if err != nil {
		return IndexerRet{}, fmt.Errorf("could not create ignore matcher: %w", err)
	}

	for _, matchedPath := range matchedPathsDir {

		// check if it matches a ignore rule
		match := ignoreMatcher.Match(matchedPath, false)
		//match, err := dfutil.IsGlobExpMatch(matchedPath, ignoreRules)
		//if err != nil {
		//	return IndexerRet{}, err
		//}
		// the folder matches a glob rule and thus should be skipped
		if match {
			return IndexerRet{}, nil
		}

		stat, err := os.Stat(matchedPath)
		if err != nil {
			return IndexerRet{}, err
		}

		if joinedRelPath != "." {
			// check for changes in the size and the modified date of the file or folder
			// and if the file is newly added
			if _, ok := existingFileIndex.Files[joinedRelPath]; !ok {
				fileChanged[matchedPath] = true
				klog.V(4).Infof("file added: %s", matchedPath)
			} else if !stat.ModTime().Equal(existingFileIndex.Files[joinedRelPath].LastModifiedDate) {
				fileChanged[matchedPath] = true
				klog.V(4).Infof("last modified date changed: %s", matchedPath)
			} else if stat.Size() != existingFileIndex.Files[joinedRelPath].Size {
				fileChanged[matchedPath] = true
				klog.V(4).Infof("size changed: %s", matchedPath)
			}
		}

		if stat.IsDir() {

			if stat.Name() == DotOdoDirectory || stat.Name() == ".git" {
				return IndexerRet{}, nil
			}

			if joinedRelPath != "." {
				folderData, folderChangedData, folderRemoteChangedData := handleRemoteDataFolder(pathOptions.destFile, matchedPath, joinedRelPath, remoteDirectories, existingFileIndex)
				folderData.Size = stat.Size()
				folderData.LastModifiedDate = stat.ModTime()
				ret.NewFileMap[joinedRelPath] = folderData

				for data, value := range folderChangedData {
					fileChanged[data] = value
				}

				for data, value := range folderRemoteChangedData {
					fileRemoteChanged[data] = value
				}
			}

			// read the current folder and read inner files and folders
			files, err := ioutil.ReadDir(matchedPath)
			if err != nil {
				return IndexerRet{}, err
			}
			if len(files) == 0 {
				continue
			}
			for _, f := range files {
				if _, ok := remoteDirectories[filepath.Join(joinedRelPath, f.Name())]; ok {
					continue
				}

				opts := recursiveCheckerPathOptions{pathOptions.directory, pathOptions.srcBase, filepath.Join(pathOptions.srcFile, f.Name()), pathOptions.destBase, filepath.Join(pathOptions.destFile, f.Name())}
				innerRet, err := recursiveChecker(opts, ignoreRules, remoteDirectories, existingFileIndex)
				if err != nil {
					return IndexerRet{}, err
				}

				for k, v := range innerRet.NewFileMap {
					ret.NewFileMap[k] = v
				}

				for _, remote := range innerRet.FilesChanged {
					fileChanged[remote] = true
				}
				for _, remote := range innerRet.RemoteDeleted {
					fileRemoteChanged[remote] = true
				}
			}
		} else {
			fileData, fileChangedData, fileRemoteChangedData := handleRemoteDataFile(pathOptions.destFile, matchedPath, joinedRelPath, remoteDirectories, existingFileIndex)
			fileData.Size = stat.Size()
			fileData.LastModifiedDate = stat.ModTime()
			ret.NewFileMap[joinedRelPath] = fileData

			for data, value := range fileChangedData {
				fileChanged[data] = value
			}

			for data, value := range fileRemoteChangedData {
				fileRemoteChanged[data] = value
			}
		}
	}

	// remove duplicates in the records
	if len(fileRemoteChanged) > 0 {
		ret.RemoteDeleted = []string{}
	}
	if len(fileChanged) > 0 {
		ret.FilesChanged = []string{}
	}
	for remote := range fileRemoteChanged {
		ret.RemoteDeleted = append(ret.RemoteDeleted, remote)
	}
	for file := range fileChanged {
		ret.FilesChanged = append(ret.FilesChanged, file)
	}

	return ret, nil
}

// handleRemoteDataFile handles remote addition, deletion etc for the given file
func handleRemoteDataFile(destFile, path, relPath string, remoteDirectories map[string]string, existingFileIndex FileIndex) (FileData, map[string]bool, map[string]bool) {
	destFile = filepath.ToSlash(destFile)
	fileChanged := make(map[string]bool)
	fileRemoteChanged := make(map[string]bool)

	remoteDeletionRequired := false

	remoteAttribute := destFile
	if len(remoteDirectories) == 0 {
		// if no remote attributes specified currently
		remoteAttribute = ""
		if existingFileIndex.Files[relPath].RemoteAttribute != "" && existingFileIndex.Files[relPath].RemoteAttribute != destFile {
			// remote attribute for the file exists in the index
			// but the value doesn't match the current relative path
			// we need to push the current file again and delete the previous location from the container

			fileChanged[path] = true
			if existingFileIndex.Files[relPath].RemoteAttribute != "" {
				remoteDeletionRequired = true
			}
		}
	} else {
		if value, ok := remoteDirectories[relPath]; ok {
			remoteAttribute = value
		}

		if existingFileData, ok := existingFileIndex.Files[relPath]; !ok {
			// if the file data doesn't exist in the existing index, we mark the file for pushing
			fileChanged[path] = true
		} else {
			// if the remote attribute is different in the file data from the existing index
			// and the remote attribute is not same as the current relative path
			// we mark the file for pushing and delete the remote paths
			if existingFileData.RemoteAttribute != remoteAttribute && (remoteAttribute != relPath || existingFileData.RemoteAttribute != "") {
				fileChanged[path] = true
				remoteDeletionRequired = true
			}
		}
	}

	if remoteDeletionRequired {
		// if remote deletion is required but the remote attribute is empty
		// we use the relative path for deletion
		currentRemote := existingFileIndex.Files[relPath].RemoteAttribute
		if currentRemote == "" {
			currentRemote = relPath
		}
		fileRemoteChanged[currentRemote] = true
		for _, remote := range findRemoteFolderForDeletion(currentRemote, remoteDirectories) {
			fileRemoteChanged[remote] = true
		}
	}

	return FileData{
		RemoteAttribute: filepath.ToSlash(remoteAttribute),
	}, fileChanged, fileRemoteChanged
}

// handleRemoteDataFolder handles remote addition, deletion etc for the given folder
func handleRemoteDataFolder(destFile, path, relPath string, remoteDirectories map[string]string, existingFileIndex FileIndex) (FileData, map[string]bool, map[string]bool) {
	destFile = filepath.ToSlash(destFile)
	remoteAttribute := destFile

	fileChanged := make(map[string]bool)
	fileRemoteChanged := make(map[string]bool)

	remoteChanged := false

	if len(remoteDirectories) == 0 {
		remoteAttribute = ""
		// remote attribute for the folder exists in the index
		// but the value doesn't match the current relative path
		// we need to push the current folder again and delete the previous location from the container

		if existingFileIndex.Files[relPath].RemoteAttribute != "" && existingFileIndex.Files[relPath].RemoteAttribute != destFile {
			fileChanged[path] = true
			if existingFileIndex.Files[relPath].RemoteAttribute != "" {
				remoteChanged = true
			}
		}
	} else {
		if value, ok := remoteDirectories[relPath]; ok {
			remoteAttribute = value
		}

		if existingFileData, ok := existingFileIndex.Files[relPath]; !ok {
			fileChanged[path] = true
		} else {
			// if the remote attribute is different in the file data from the existing index
			// and the remote attribute is not same as the current relative path
			// we mark the file for pushing and delete the remote paths
			if existingFileData.RemoteAttribute != remoteAttribute && (remoteAttribute != relPath || existingFileData.RemoteAttribute != "") {
				fileChanged[path] = true
				remoteChanged = true
			}
		}
	}

	if remoteChanged {
		// if remote deletion is required but the remote attribute is empty
		// we use the relative path for deletion
		currentRemote := existingFileIndex.Files[relPath].RemoteAttribute
		if currentRemote == "" {
			currentRemote = relPath
		}
		fileRemoteChanged[currentRemote] = true
		for _, remote := range findRemoteFolderForDeletion(currentRemote, remoteDirectories) {
			fileRemoteChanged[remote] = true
		}
	}

	return FileData{
		RemoteAttribute: filepath.ToSlash(remoteAttribute),
	}, fileChanged, fileRemoteChanged
}

// checkFileExist check if given file exists or not
func checkFileExist(fileName string) bool {
	_, err := os.Stat(fileName)
	return !os.IsNotExist(err)
}

// findRemoteFolderForDeletion finds the remote directories which can be deleted by checking the remoteDirectories map
func findRemoteFolderForDeletion(currentRemote string, remoteDirectories map[string]string) []string {
	var remoteDelete []string
	currentRemote = filepath.ToSlash(currentRemote)
	for currentRemote != "" && currentRemote != "." && currentRemote != "/" {

		found := false
		for _, remote := range remoteDirectories {
			if strings.HasPrefix(remote, currentRemote+"/") || remote == currentRemote {
				found = true
				break
			}
		}
		if !found {
			remoteDelete = append(remoteDelete, currentRemote)
		}
		currentRemote = filepath.ToSlash(filepath.Clean(filepath.Dir(currentRemote)))
	}
	return remoteDelete
}
