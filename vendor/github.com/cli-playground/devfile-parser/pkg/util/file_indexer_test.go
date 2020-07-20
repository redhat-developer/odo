package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cli-playground/devfile-parser/pkg/testingutil/filesystem"
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
