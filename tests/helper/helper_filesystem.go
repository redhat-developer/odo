package helper

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	dfutil "github.com/devfile/library/v2/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// CreateNewContext create new empty temporary directory
func CreateNewContext() string {
	directory, err := os.MkdirTemp("", "")
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
	}, 16, 2*time.Minute)
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

// CopyExampleDevFile copies an example devfile from tests/examples/source/devfiles/<componentName>/devfile.yaml into targetDst.
// If newName is not empty, it will replace the component name in the target Devfile.
// The Devfile updaters allow to perform operations against the target Devfile, like removing the component name (via DevfileMetadataNameRemover).
func CopyExampleDevFile(devfilePath, targetDst string, newName string, updaters ...DevfileUpdater) {
	// filename of this file
	_, filename, _, _ := runtime.Caller(0)
	// path to the examples directory
	examplesDir := filepath.Join(filepath.Dir(filename), "..", "examples")

	src := filepath.Join(examplesDir, devfilePath)
	info, err := os.Stat(src)
	Expect(err).NotTo(HaveOccurred())

	err = dfutil.CopyFile(src, targetDst, info)
	Expect(err).NotTo(HaveOccurred())

	var devfileUpdaters []DevfileUpdater
	if newName != "" {
		devfileUpdaters = append(devfileUpdaters, DevfileMetadataNameSetter(newName))
	}
	devfileUpdaters = append(devfileUpdaters, updaters...)
	UpdateDevfileContent(targetDst, devfileUpdaters)
}

// FileShouldContainSubstring check if file contains subString
func FileShouldContainSubstring(file string, subString string) {
	data, err := os.ReadFile(file)
	Expect(err).NotTo(HaveOccurred())
	Expect(string(data)).To(ContainSubstring(subString))
}

// FileShouldNotContainSubstring check if file does not contain subString
func FileShouldNotContainSubstring(file string, subString string) {
	data, err := os.ReadFile(file)
	Expect(err).NotTo(HaveOccurred())
	Expect(string(data)).NotTo(ContainSubstring(subString))
}

// ReplaceString replaces oldString with newString in text file
func ReplaceString(filename string, oldString string, newString string) {
	fmt.Fprintf(GinkgoWriter, "Replacing \"%s\" with \"%s\" in %s\n", oldString, newString, filename)

	f, err := os.ReadFile(filename)
	Expect(err).NotTo(HaveOccurred())

	newContent := strings.ReplaceAll(string(f), oldString, newString)

	err = os.WriteFile(filename, []byte(newContent), 0600)
	Expect(err).NotTo(HaveOccurred())
}

// ReplaceStrings replaces oldStrings with newStrings in text file
// two arrays must be of same length, else will fail
func ReplaceStrings(filename string, oldStrings []string, newStrings []string) {
	fmt.Fprintf(GinkgoWriter, "Replacing \"%v\" with \"%v\" in %s\n", oldStrings, newStrings, filename)

	contentByte, err := os.ReadFile(filename)
	Expect(err).NotTo(HaveOccurred())

	newContent := string(contentByte)
	for i := range oldStrings {
		newContent = strings.ReplaceAll(newContent, oldStrings[i], newStrings[i])
	}

	err = os.WriteFile(filename, []byte(newContent), 0600)
	Expect(err).NotTo(HaveOccurred())
}

// copyDir copy one directory to the other
// this function is called recursively info should start as os.Stat(src)
func copyDir(src string, dst string, info os.FileInfo) error {

	if info.IsDir() {
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			file, err := entry.Info()
			if err != nil {
				return err
			}
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
	return CreateFileWithContentAndPerm(path, fileContent, 0600)
}

// CreateFileWithContentAndPerm creates a file at the given path using the given file permissions, and writes the given content.
// path is the path to the required file
// fileContent is the content to be written to the given file
func CreateFileWithContentAndPerm(path string, fileContent string, perm os.FileMode) error {
	// create and open file if not exists
	var file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, perm)
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
	entries, err := os.ReadDir(directoryName)
	Expect(err).ShouldNot(HaveOccurred())
	for _, entry := range entries {
		file, err := entry.Info()
		Expect(err).ShouldNot(HaveOccurred())
		filesInDirectory = append(filesInDirectory, file.Name())
	}
	return filesInDirectory
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

// ReadFile reads the file from the filePath
func ReadFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
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
	err := os.WriteFile(FilePath, content, 0600)
	Expect(err).NotTo(HaveOccurred())

	return FilePath, string(content)
}

func AppendToFile(filepath string, s string) error {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close() // #nosec G307
	if _, err := f.WriteString(s); err != nil {
		return err
	}
	return nil
}
