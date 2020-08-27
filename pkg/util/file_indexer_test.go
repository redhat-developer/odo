package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/openshift/odo/pkg/testingutil/filesystem"
)

func TestCheckGitIgnoreFile(t *testing.T) {

	// create a fake fs in memory
	fs := filesystem.NewFakeFs()
	// create a context directory on fake fs
	contextDir, err := fs.TempDir(os.TempDir(), "context")
	if err != nil {
		t.Error(err)
	}

	gitignorePath := filepath.Join(contextDir, ".gitignore")

	tests := []struct {
		testName        string
		create          bool
		gitIgnoreCreate func(create bool, contextDir string, fs filesystem.Filesystem) error
		directory       string
		want            string
		wantErr         bool
	}{
		{
			testName:        "Test when .gitignore does not exist",
			create:          true,
			gitIgnoreCreate: mockDirectoryInfo,
			directory:       contextDir,
			want:            gitignorePath,
			wantErr:         false,
		},
		{
			testName:        "Test when .gitignore exists",
			create:          false,
			gitIgnoreCreate: mockDirectoryInfo,
			directory:       contextDir,
			want:            gitignorePath,
			wantErr:         false,
		},
	}

	for _, tt := range tests {

		err := tt.gitIgnoreCreate(tt.create, tt.directory, fs)
		if err != nil {
			t.Error(err)
		}

		t.Run(tt.testName, func(t *testing.T) {

			gitIgnoreFilePath, err := checkGitIgnoreFile(tt.directory, fs)

			if tt.want != gitIgnoreFilePath {
				t.Errorf("checkGitIgnoreFile unexpected error %v, while creating .gitignore file", err)
			}

			if !tt.wantErr == (err != nil) {
				t.Errorf("checkGitIgnoreFile unexpected error %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func TestAddOdoFileIndex(t *testing.T) {

	// create a fake fs in memory
	fs := filesystem.NewFakeFs()
	// create a context directory on fake fs
	contextDir, err := fs.TempDir(os.TempDir(), "context")
	if err != nil {
		t.Error(err)
	}

	gitignorePath := filepath.Join(contextDir, ".gitignore")

	tests := []struct {
		testName        string
		create          bool
		gitIgnoreCreate func(create bool, contextDir string, fs filesystem.Filesystem) error
		directory       string
		wantErr         bool
	}{
		{
			testName:        "Test when odo-file-index.json added to .gitignore",
			create:          false,
			gitIgnoreCreate: mockDirectoryInfo,
			directory:       gitignorePath,
			wantErr:         false,
		},
	}

	for _, tt := range tests {

		err := tt.gitIgnoreCreate(tt.create, tt.directory, fs)
		if err != nil {
			t.Error(err)
		}

		t.Run(tt.testName, func(t *testing.T) {

			err := addOdoFileIndex(tt.directory, fs)

			if !tt.wantErr == (err != nil) {
				t.Errorf("addOdoFileIndex unexpected error %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

func mockDirectoryInfo(create bool, contextDir string, fs filesystem.Filesystem) error {

	if !create {
		err := fs.MkdirAll(filepath.Join(contextDir, ".gitignore"), os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

func TestCalculateFileDataKeyFromPath(t *testing.T) {

	// create a temp dir for the fake component
	directory, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: error creating temporary directory for the indexer: %v", err)
	}

	tests := []struct {
		absolutePath   string
		rootDirectory  string
		expectedResult string
	}{
		{
			absolutePath:   filepath.Join(directory, "/path/file1"),
			rootDirectory:  filepath.Join(directory, "/path"),
			expectedResult: "file1",
		},
		{
			absolutePath:   filepath.Join(directory, "/path/path2/file1"),
			rootDirectory:  filepath.Join(directory, "/path/"),
			expectedResult: "path2/file1",
		},
		{
			absolutePath:   filepath.Join(directory, "/path"),
			rootDirectory:  filepath.Join(directory, "/"),
			expectedResult: "path",
		},
	}

	for _, tt := range tests {

		t.Run("Expect result: "+tt.expectedResult, func(t *testing.T) {

			result, err := CalculateFileDataKeyFromPath(tt.absolutePath, tt.rootDirectory)
			if err != nil {
				t.Fatalf("unexpecter error occurred %v", err)
			}

			if result != filepath.FromSlash(tt.expectedResult) {
				t.Fatalf("unexpected result: %v %v", tt.expectedResult, result)
			}
		})
	}
}

func TestGenerateNewFileDataEntry(t *testing.T) {

	// create a temp dir for the fake component
	directory, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: error creating temporary directory for the indexer: %v", err)
	}

	tests := []struct {
		testName      string
		absolutePath  string
		rootDirectory string
		expectedKey   string
	}{
		{
			absolutePath:  filepath.Join(directory, "/path1/file1"),
			rootDirectory: filepath.Join(directory, "/path1"),
			expectedKey:   "file1",
		},
		{
			absolutePath:  filepath.Join(directory, "/path2/path2/file1"),
			rootDirectory: filepath.Join(directory, "/path2"),
			expectedKey:   "path2/file1",
		},
		{
			absolutePath:  filepath.Join(directory, "/path3"),
			rootDirectory: filepath.Join(directory, "/"),
			expectedKey:   "path3",
		},
	}

	for _, tt := range tests {

		t.Run("Expected key '"+tt.expectedKey+"'", func(t *testing.T) {

			if err := os.MkdirAll(filepath.Dir(tt.absolutePath), 0750); err != nil {
				t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: unable to create directories for %s: %v", tt.absolutePath, err)
			}

			if err := ioutil.WriteFile(tt.absolutePath, []byte("non-empty-string"), 0644); err != nil {
				t.Fatalf("TestUpdateIndexWithWatchChangesLocal error: unable to write to index file path: %v", err)
			}

			key, filedata, err := GenerateNewFileDataEntry(tt.absolutePath, tt.rootDirectory)

			if err != nil {
				t.Fatalf("Unexpected error occurred %v", err)
			}

			// Keys are platform specific, so swap to forward slash for Windows before comparison
			key = strings.ReplaceAll(key, "\\", "/")

			if key != tt.expectedKey {
				t.Fatalf("Key %s did not match expected key %s", key, tt.expectedKey)
			}

			if filedata == nil {
				t.Fatalf("Filedata should not be null")
			}

			if filedata.Size == 0 || filedata.LastModifiedDate.IsZero() {
				t.Fatalf("Invalid filedata values %v %v", filedata.Size, filedata.LastModifiedDate)
			}

		})
	}
}
