package helper

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	dfutil "github.com/devfile/library/pkg/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// CreateNewContext create new empty temporary directory
func CreateNewContext() string {
	directory, err := ioutil.TempDir("", "")
	Expect(err).NotTo(HaveOccurred())
	fmt.Fprintf(GinkgoWriter, "Created dir: %s\n", directory)
	return directory
}

// DeleteDir deletes the specified path; due to Windows behaviour (for example https://github.com/redhat-developer/odo/issues/3371)
// where Windows temporarily holds a lock on files and folders, we keep trying to delete until the operation passes (or it expires)
func DeleteDir(dir string) {
	attempts := 0

	errorReportedAtLeastOnce := false

	err := RunWithExponentialBackoff(func() error {
		attempts++

		fmt.Fprintf(GinkgoWriter, "Deleting dir: %s\n", dir)
		err := os.RemoveAll(dir)
		if err == nil {
			return nil
		}

		errorReportedAtLeastOnce = true
		fmt.Fprintf(GinkgoWriter, "Unable to delete %s on attempt #%d, trying again...\n", dir, attempts)

		return err
	}, 16, time.Duration(2)*time.Minute)
	Expect(err).NotTo(HaveOccurred())

	if errorReportedAtLeastOnce {
		fmt.Fprintf(GinkgoWriter, "Successfully deleted %s after #%d attempts\n", dir, attempts)
	}
}

// RunWithExponentialBackoff keeps trying to run 'fxn' until it no longer returns an error; if the function never succeeded,
// then the most recent error is returned.
func RunWithExponentialBackoff(fxn func() error, maxDelayInSeconds int, expireDuration time.Duration) error {
	expireTime := time.Now().Add(expireDuration)
	delayInSeconds := 1

	var err error

	for {

		err = fxn()

		if err == nil || time.Now().After(expireTime) {
			break
		}

		delayInSeconds *= 2 // exponential backoff
		if delayInSeconds > maxDelayInSeconds {
			delayInSeconds = maxDelayInSeconds
		}
		time.Sleep(time.Duration(delayInSeconds) * time.Second)

	}
	return err

}

// DeleteFile deletes file
func DeleteFile(filepath string) {
	fmt.Fprintf(GinkgoWriter, "Deleting file: %s\n", filepath)
	err := os.Remove(filepath)
	Expect(err).NotTo(HaveOccurred())
}

// RenameFile renames a file from oldFileName to newFileName
func RenameFile(oldFileName, newFileName string) {
	err := os.Rename(oldFileName, newFileName)
	Expect(err).NotTo(HaveOccurred())
}

// Chdir change current working dir
func Chdir(dir string) {
	fmt.Fprintf(GinkgoWriter, "Setting current dir to: %s\n", dir)
	err := os.Chdir(dir)
	Expect(err).ShouldNot(HaveOccurred())
}

// MakeDir creates a new dir
func MakeDir(dir string) {
	err := os.MkdirAll(dir, 0750)
	Expect(err).ShouldNot(HaveOccurred())
}

// Getwd returns current working dir
func Getwd() string {
	dir, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	fmt.Fprintf(GinkgoWriter, "Current working dir: %s\n", dir)
	return dir
}

// CopyExampleFile copies an example file from tests/examples/<file-path>
// into targetDst
func CopyExampleFile(filePath, targetDst string) {
	// filename of this file
	_, filename, _, _ := runtime.Caller(0)
	// path to the examples directory
	examplesDir := filepath.Join(filepath.Dir(filename), "..", "examples")

	src := filepath.Join(examplesDir, filePath)
	info, err := os.Stat(src)
	Expect(err).NotTo(HaveOccurred())

	err = dfutil.CopyFile(src, targetDst, info)
	Expect(err).NotTo(HaveOccurred())
}

// CopyExample copies an example from tests/examples/<binaryOrSource>/<componentName>/<exampleName> into targetDir
func CopyExample(exampleName string, targetDir string) {
	// filename of this file
	_, filename, _, _ := runtime.Caller(0)
	// path to the examples directory
	examplesDir := filepath.Join(filepath.Dir(filename), "..", "examples")

	src := filepath.Join(examplesDir, exampleName)
	info, err := os.Stat(src)
	Expect(err).NotTo(HaveOccurred())

	err = copyDir(src, targetDir, info)
	Expect(err).NotTo(HaveOccurred())
}

func CopyManifestFile(fileName, targetDst string) {
	// filename of this file
	_, filename, _, _ := runtime.Caller(0)
	// path to the examples directory
	manifestsDir := filepath.Join(filepath.Dir(filename), "..", "examples", "manifests")

	src := filepath.Join(manifestsDir, fileName)
	info, err := os.Stat(src)
	Expect(err).NotTo(HaveOccurred())

	err = dfutil.CopyFile(src, targetDst, info)
	Expect(err).NotTo(HaveOccurred())

}

func GetExamplePath(args ...string) string {
	_, filename, _, _ := runtime.Caller(0)
	path := append([]string{filepath.Dir(filename), "..", "examples"}, args...)
	return filepath.Join(path...)
}

// CopyExampleDevFile copies an example devfile from tests/examples/source/devfiles/<componentName>/devfile.yaml
// into targetDst
func CopyExampleDevFile(devfilePath, targetDst string) {
	// filename of this file
	_, filename, _, _ := runtime.Caller(0)
	// path to the examples directory
	examplesDir := filepath.Join(filepath.Dir(filename), "..", "examples")

	src := filepath.Join(examplesDir, devfilePath)
	info, err := os.Stat(src)
	Expect(err).NotTo(HaveOccurred())

	err = dfutil.CopyFile(src, targetDst, info)
	Expect(err).NotTo(HaveOccurred())
}

// FileShouldContainSubstring check if file contains subString
func FileShouldContainSubstring(file string, subString string) {
	data, err := ioutil.ReadFile(file)
	Expect(err).NotTo(HaveOccurred())
	Expect(string(data)).To(ContainSubstring(subString))
}

// ReplaceString replaces oldString with newString in text file
func ReplaceString(filename string, oldString string, newString string) {
	fmt.Fprintf(GinkgoWriter, "Replacing \"%s\" with \"%s\" in %s\n", oldString, newString, filename)

	f, err := ioutil.ReadFile(filename)
	Expect(err).NotTo(HaveOccurred())

	newContent := strings.ReplaceAll(string(f), oldString, newString)

	err = ioutil.WriteFile(filename, []byte(newContent), 0600)
	Expect(err).NotTo(HaveOccurred())
}

// copyDir copy one directory to the other
// this function is called recursively info should start as os.Stat(src)
func copyDir(src string, dst string, info os.FileInfo) error {

	if info.IsDir() {
		files, err := ioutil.ReadDir(src)
		if err != nil {
			return err
		}

		for _, file := range files {
			dsrt := filepath.Join(src, file.Name())
			ddst := filepath.Join(dst, file.Name())
			if err := copyDir(dsrt, ddst, file); err != nil {
				return err
			}
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return err
	}

	return dfutil.CopyFile(src, dst, info)
}

// CreateFileWithContent creates a file at the given path and writes the given content
// path is the path to the required file
// fileContent is the content to be written to the given file
func CreateFileWithContent(path string, fileContent string) error {
	// create and open file if not exists
	var file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close() // #nosec G307
	// write to file
	_, err = file.WriteString(fileContent)
	if err != nil {
		return err
	}
	return nil
}

// ListFilesInDir lists all the files in the directory
// directoryName is the name of the directory
func ListFilesInDir(directoryName string) []string {
	var filesInDirectory []string
	files, err := ioutil.ReadDir(directoryName)
	Expect(err).ShouldNot(HaveOccurred())

	for _, file := range files {
		filesInDirectory = append(filesInDirectory, file.Name())
	}
	return filesInDirectory
}

// CreateSymLink creates a symlink between the oldFile and the newFile
func CreateSymLink(oldFileName, newFileName string) {
	err := os.Symlink(oldFileName, newFileName)
	Expect(err).NotTo(HaveOccurred())
}

// VerifyFileExists receives a path to a file, and returns whether or not
// it points to an existing file
func VerifyFileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// VerifyFilesExist receives an array of paths to files, and returns whether
// or not they all exist. If any one of the expected files doesn't exist, it
// returns false
func VerifyFilesExist(path string, files []string) bool {
	for _, f := range files {
		if !VerifyFileExists(filepath.Join(path, f)) {
			return false
		}
	}
	return true
}

// ReplaceDevfileField replaces the value of a given field in a specified
// devfile.
// Currently only the first match of the field is replaced. i.e if the
// field is 'type' and there are two types throughout the devfile, only one
// is replaced with the newValue
func ReplaceDevfileField(devfileLocation, field, newValue string) error {
	file, err := ioutil.ReadFile(devfileLocation)
	if err != nil {
		return err
	}
	lines := strings.Split(string(file), "\n")
	for i, line := range lines {
		if strings.Contains(line, field) {
			lineSplit := strings.SplitN(lines[i], ":", 2)
			lineSplit[1] = newValue
			lines[i] = strings.Join(lineSplit, ": ")
			break
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(devfileLocation, []byte(output), 0600)
	if err != nil {
		return err
	}
	return nil
}

// FileIsEmpty checks if the file is empty
func FileIsEmpty(filename string) (bool, error) {
	file, err := os.Stat(filename)
	if err != nil {
		return false, err
	}

	if file.Size() > 0 {
		return false, nil
	}

	return true, nil
}

// ReadFile reads the file from the filePath
func ReadFile(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CreateSimpleFile creates a simple file
// return the file path with random string
func CreateSimpleFile(context, filePrefix, fileExtension string) (string, string) {

	FilePath := filepath.Join(context, filePrefix+RandString(10)+fileExtension)
	content := []byte(RandString(10))
	err := ioutil.WriteFile(FilePath, content, 0600)
	Expect(err).NotTo(HaveOccurred())

	return FilePath, string(content)
}
