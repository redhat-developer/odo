package helper

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

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

// DeleteDir delete directory
func DeleteDir(dir string) {
	fmt.Fprintf(GinkgoWriter, "Deleating dir: %s\n", dir)
	err := os.RemoveAll(dir)
	Expect(err).NotTo(HaveOccurred())

}

// Chdir change current working dir
func Chdir(dir string) {
	fmt.Fprintf(GinkgoWriter, "Setting current dir to: %s\n", dir)
	err := os.Chdir(dir)
	Expect(err).ShouldNot(HaveOccurred())
}

// Getwd retruns current working dir
func Getwd() string {
	dir, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	fmt.Fprintf(GinkgoWriter, "Current working dir: %s\n", dir)
	return dir
}

// CopyExample copies an example from tests/e2e/examples/<exampleName> into targetDir
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

// FileShouldContainSubstring check if file contains subString
func FileShouldContainSubstring(file string, subString string) {
	data, err := ioutil.ReadFile(file)
	Expect(err).NotTo(HaveOccurred())
	Expect(string(data)).To(ContainSubstring(subString))
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

	dFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dFile.Close()

	sFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sFile.Close()

	if err = os.Chmod(dFile.Name(), info.Mode()); err != nil {
		return err
	}

	_, err = io.Copy(dFile, sFile)
	return err
}

// CreateFileWithContent creates a file at the given path and writes the given content
// path is the path to the required file
// fileContent is the content to be written to the given file
func CreateFileWithContent(path string, fileContent string) error {
	// create and open file if not exists
	var file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	// write to file
	_, err = file.WriteString(fileContent)
	if err != nil {
		return err
	}
	return nil
}
