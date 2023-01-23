package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devfile/library/v2/pkg/testingutil/filesystem"
	dfutil "github.com/devfile/library/v2/pkg/util"
)

func TestCleanDefaultHTTPCacheDir(t *testing.T) {
	fakeFs := filesystem.NewFakeFs()
	filesToGenerate := 10
	for i := 0; i < filesToGenerate; i++ {
		err := fakeFs.WriteFile(filepath.Join(httpCacheDir, dfutil.GenerateRandomString(10)), []byte(dfutil.GenerateRandomString(10)), os.ModePerm)
		if err != nil {
			t.Error(err)
		}
	}
	files, err := fakeFs.ReadDir(httpCacheDir)
	if err != nil {
		t.Error(err)
	}
	// checking the file count before the run
	if len(files) != filesToGenerate {
		t.Error("the file count in the httpCacheDir don't match files generated")
	}
	err = cleanDefaultHTTPCacheDir(fakeFs)
	if err != nil {
		t.Error(err)
	}

	newFiles, err := fakeFs.ReadDir(httpCacheDir)
	if err != nil {
		t.Error(err)
	}

	if len(newFiles) != 0 {
		t.Error("httpCacheDir is not empty after cleanup")
	}

}
