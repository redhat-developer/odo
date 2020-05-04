package pipelines

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/openshift/odo/pkg/pipelines/ioutils"
	"github.com/spf13/afero"
)

func TestAddEnv(t *testing.T) {
	fakeFs := ioutils.NewMapFilesystem()
	gitopsPath := afero.GetTempDir(fakeFs, "test")

	manifestFile := filepath.Join(gitopsPath, pipelinesFile)
	envParameters := EnvParameters{
		ManifestFilename: manifestFile,
		EnvName:          "dev",
	}
	afero.WriteFile(fakeFs, manifestFile, []byte("environments:"), 0644)

	if err := AddEnv(&envParameters, fakeFs); err != nil {
		t.Fatalf("AddEnv() failed :%s", err)
	}

	wantedPaths := []string{
		"environments/dev/env/base/kustomization.yaml",
		"environments/dev/env/base/dev-environment.yaml",
		"environments/dev/env/overlays/kustomization.yaml",
	}
	for _, path := range wantedPaths {
		t.Run(fmt.Sprintf("checking path %s already exists", path), func(rt *testing.T) {
			assertFileExists(rt, fakeFs, filepath.Join(gitopsPath, path))
		})
	}
}

func TestAddEnvWithExistingName(t *testing.T) {
	fakeFs := ioutils.NewMapFilesystem()
	gitopsPath := afero.GetTempDir(fakeFs, "test")

	manifestFile := filepath.Join(gitopsPath, pipelinesFile)
	envParameters := EnvParameters{
		ManifestFilename: manifestFile,
		EnvName:          "dev",
	}
	afero.WriteFile(fakeFs, manifestFile, []byte("environments:\n - name: dev\n"), 0644)

	if err := AddEnv(&envParameters, fakeFs); err == nil {
		t.Fatal("AddEnv() did not fail with duplicate environment")
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
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
