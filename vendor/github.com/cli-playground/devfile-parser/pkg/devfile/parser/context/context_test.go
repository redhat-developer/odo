package parser

import (
	"os"
	"testing"

	"github.com/cli-playground/devfile-parser/pkg/testingutil/filesystem"
)

func TestPopulateFromBytes(t *testing.T) {

	const (
		InvalidDevfilePath = "/invalid/path"
	)

	// createTempDevfile helper creates temp devfile
	createTempDevfile := func(t *testing.T, content []byte, fakeFs filesystem.Filesystem) (f filesystem.File) {

		t.Helper()

		// Create tempfile
		f, err := fakeFs.TempFile(os.TempDir(), TempJSONDevfilePrefix)
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

	t.Run("valid data passed", func(t *testing.T) {

		var (
			fakeFs      = filesystem.NewFakeFs()
			tempDevfile = createTempDevfile(t, validJsonRawContent100(), fakeFs)
			d           = DevfileCtx{
				absPath: tempDevfile.Name(),
				Fs:      fakeFs,
			}
		)
		defer os.Remove(tempDevfile.Name())

		err := d.PopulateFromBytes(validJsonRawContent100())

		if err != nil {
			t.Errorf("unexpected error '%v'", err)
		}

		if err := tempDevfile.Close(); err != nil {
			t.Errorf("failed to close temp devfile")
		}
	})

	t.Run("invalid data passed", func(t *testing.T) {

		var (
			fakeFs      = filesystem.NewFakeFs()
			tempDevfile = createTempDevfile(t, []byte(validJsonRawContent100()), fakeFs)
			d           = DevfileCtx{
				absPath: tempDevfile.Name(),
				Fs:      fakeFs,
			}
		)
		defer os.Remove(tempDevfile.Name())

		err := d.PopulateFromBytes([]byte(InvalidDevfileContent))

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
