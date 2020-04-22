package pipelines

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/openshift/odo/pkg/manifest/yaml"
	"github.com/spf13/afero"
)

func TestWriteResources(t *testing.T) {
	testFs := afero.NewMemMapFs()
	tmpDir := afero.GetTempDir(testFs, "odo")
	resources := map[string]interface{}{
		"01_roles/serviceaccount.yaml": fakeYamlDoc(1),
		"02_tasks/buildah_task.yaml":   fakeYamlDoc(2),
	}

	_, err := yaml.WriteResources(testFs, tmpDir, resources)
	if err != nil {
		t.Fatalf("failed to writeResources: %v", err)
	}
	assertFileContents(t, testFs, filepath.Join(tmpDir, "01_roles/serviceaccount.yaml"), []byte("key1: value1\n"))
	assertFileContents(t, testFs, filepath.Join(tmpDir, "02_tasks/buildah_task.yaml"), []byte("key2: value2\n"))
}

func assertFileContents(t *testing.T, fs afero.Fs, filename string, want []byte) {
	t.Helper()
	body, err := afero.ReadFile(fs, filename)
	if err != nil {
		t.Fatalf("failed to read file: %s", filename)
	}

	if diff := cmp.Diff(body, want); diff != "" {
		t.Fatalf("file %s diff = \n%s\n", filename, diff)
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func fakeYamlDoc(n int) map[string]string {
	return map[string]string{
		fmt.Sprintf("key%d", n): fmt.Sprintf("value%d", n),
	}
}
