package pipelines

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestWriteResources(t *testing.T) {
	tmpDir, cleanUp := makeTempDir(t)
	defer cleanUp()
	resources := map[string]interface{}{
		"01_roles/serviceaccount.yaml": fakeYamlDoc(1),
		"02_tasks/buildah_task.yaml":   fakeYamlDoc(2),
	}

	_, err := writeResources(tmpDir, resources)
	if err != nil {
		t.Fatalf("failed to writeResources: %v", err)
	}
	assertFileContents(t, filepath.Join(tmpDir, "01_roles/serviceaccount.yaml"), []byte("key1: value1\n---\n"))
	assertFileContents(t, filepath.Join(tmpDir, "02_tasks/buildah_task.yaml"), []byte("key2: value2\n---\n"))
}

func assertFileContents(t *testing.T, filename string, want []byte) {
	t.Helper()
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file: %s", filename)
	}

	if diff := cmp.Diff(body, want); diff != "" {
		t.Fatalf("file %s diff = \n%s\n", filename, diff)
	}
}

func makeTempDir(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := ioutil.TempDir(os.TempDir(), "test")
	assertNoError(t, err)
	return dir, func() {
		err := os.RemoveAll(dir)
		assertNoError(t, err)
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
