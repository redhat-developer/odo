package parser

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestSetDevfileContent(t *testing.T) {

	const (
		TempJsonDevfilePrefix = "odo-devfile.*.json"
		InvalidDevfilePath    = "/invalid/path"
		InvalidDevfileContent = ":: invalid :: content"
	)

	// createTempDevfile helper creates temp devfile
	createTempDevfile := func(t *testing.T, content []byte) (f *os.File) {

		t.Helper()

		// Create tempfile
		f, err := ioutil.TempFile(os.TempDir(), TempJsonDevfilePrefix)
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
			tempDevfile = createTempDevfile(t, validJsonRawContent100())
			d           = DevfileCtx{absPath: tempDevfile.Name()}
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
			tempDevfile = createTempDevfile(t, []byte(InvalidDevfileContent))
			d           = DevfileCtx{absPath: tempDevfile.Name()}
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
			d = DevfileCtx{absPath: InvalidDevfilePath}
		)

		err := d.SetDevfileContent()

		if err == nil {
			t.Errorf("expected an error, didn't get one")
		}
	})
}
