package manifest

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/openshift/odo/pkg/manifest/ioutils"
	"github.com/spf13/afero"
)

func TestEnv(t *testing.T) {

	fakeFs := ioutils.NewMapFilesystem()
	gitopsPath := afero.GetTempDir(fakeFs, "test")
	envParameters := EnvParameters{
		EnvName: "dev",
		Output:  gitopsPath,
	}
	if err := Env(&envParameters, fakeFs); err != nil {
		t.Fatalf("Env() failed :%s", err)
	}

	wantedPaths := []string{
		"environments/dev/base/kustomization.yaml",
		"environments/dev/base/namespace.yaml",
		"environments/dev/base/rolebinding.yaml",
		"environments/dev/overlays/kustomization.yaml",
	}

	for _, path := range wantedPaths {
		t.Run(fmt.Sprintf("checking path %s already exists", path), func(rt *testing.T) {
			assertFileExists(rt, fakeFs, filepath.Join(gitopsPath, path))
		})
	}
}

func assertFileExists(t *testing.T, testFs afero.Fs, path string) {
	t.Helper()
	exists, err := afero.Exists(testFs, path)
	assertNoError(t, err)
	if !exists {
		t.Fatalf("unable to find file %q", path)
	}
	isDir, err := afero.DirExists(testFs, path)
	assertNoError(t, err)
	if isDir {
		t.Fatalf("%q is a directory", path)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
