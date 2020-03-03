package parser

import (
	"os"
	"testing"

	"github.com/openshift/odo/pkg/testingutil/filesystem"
)

func TestSetDevfileContent(t *testing.T) {

	const (
		TempJsonDevfilePrefix = "odo-devfile.*.json"
		InvalidDevfilePath    = "/invalid/path"
		InvalidDevfileContent = ":: invalid :: content"
	)

	// createTempDevfile helper creates temp devfile
	createTempDevfile := func(t *testing.T, content []byte, fakeFs filesystem.Filesystem) (f filesystem.File) {

		t.Helper()

		// Create tempfile
		f, err := fakeFs.TempFile(os.TempDir(), TempJsonDevfilePrefix)
		if err != nil {
			t.Errorf("failed to create temp devfile, %v", err)
			return f
		}

		// Write content to devfile
		if _, err := f.Write(content); err != nil {
			t.Errorf("failed to write to temp devfile")
			return f
		}

		// Successful
		return f
	}

	t.Run("valid file", func(t *testing.T) {

		var (
			fakeFs      = filesystem.NewFakeFs()
			tempDevfile = createTempDevfile(t, validJsonRawContent100(), fakeFs)
			d           = DevfileCtx{
				absPath: tempDevfile.Name(),
				Fs:      fakeFs,
			}
		)
		defer os.Remove(tempDevfile.Name())

		err := d.SetDevfileContent()

		if err != nil {
			t.Errorf("unexpected error '%v'", err)
		}

		if err := tempDevfile.Close(); err != nil {
			t.Errorf("failed to close temp devfile")
		}
	})

	t.Run("invalid content", func(t *testing.T) {

		var (
			fakeFs      = filesystem.NewFakeFs()
			tempDevfile = createTempDevfile(t, []byte(InvalidDevfileContent), fakeFs)
			d           = DevfileCtx{
				absPath: tempDevfile.Name(),
				Fs:      fakeFs,
			}
		)
		defer os.Remove(tempDevfile.Name())

		err := d.SetDevfileContent()

		if err == nil {
			t.Errorf("expected error, didn't get one ")
		}

		if err := tempDevfile.Close(); err != nil {
			t.Errorf("failed to close temp devfile")
		}
	})

	t.Run("invalid filepath", func(t *testing.T) {

		var (
			fakeFs = filesystem.NewFakeFs()
			d      = DevfileCtx{
				absPath: InvalidDevfilePath,
				Fs:      fakeFs,
			}
		)

		err := d.SetDevfileContent()

		if err == nil {
			t.Errorf("expected an error, didn't get one")
		}
	})
}
