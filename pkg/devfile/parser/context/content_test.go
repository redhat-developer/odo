package parser

import (
	"os"
	"testing"

	"github.com/openshift/odo/pkg/testingutil/filesystem"
)

const (
	TempJSONDevfilePrefix = "odo-devfile.*.json"
	InvalidDevfileContent = ":: invalid :: content"
	validJson200          = `{ "schemaVersion": "2.0.0", "metadata": { "name": "java-maven", "version": "1.0.0" }, "components": [ { "container": { "name": "tools", "image": "quay.io/eclipse/che-java11-maven:nightly", "memoryLimit": "512Mi", "mountSources": true, "endpoints": [ { "name": "http-8080", "targetPort": 8080 } ], "volumeMounts": [ { "name": "m2", "path": "/home/user/.m2" } ] } }, { "volume": { "name": "m2" } } ], "commands": [ { "exec": { "id": "mvn-package", "component": "tools", "commandLine": "mvn package", "group": { "kind": "build", "isDefault": true } } }, { "exec": { "id": "run", "component": "tools", "commandLine": "java -jar target/*.jar", "group": { "kind": "run", "isDefault": true } } }, { "exec": { "id": "debug", "component": "tools", "commandLine": "java -Xdebug -Xrunjdwp:server=y,transport=dt_socket,address=${DEBUG_PORT},suspend=n -jar target/*.jar", "group": { "kind": "debug", "isDefault": true } } } ] }`
)

func validJsonRawContent200() []byte {
	return []byte(validJson200)
}
func TestSetDevfileContent(t *testing.T) {

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

	t.Run("valid file", func(t *testing.T) {

		var (
			fakeFs      = filesystem.NewFakeFs()
			tempDevfile = createTempDevfile(t, validJsonRawContent200(), fakeFs)
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

func TestSetDevfileContentFromBytes(t *testing.T) {

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
			tempDevfile = createTempDevfile(t, validJsonRawContent200(), fakeFs)
			d           = DevfileCtx{
				absPath: tempDevfile.Name(),
				Fs:      fakeFs,
			}
		)

		defer os.Remove(tempDevfile.Name())

		err := d.SetDevfileContentFromBytes(validJsonRawContent200())

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
			tempDevfile = createTempDevfile(t, []byte(validJsonRawContent200()), fakeFs)
			d           = DevfileCtx{
				absPath: tempDevfile.Name(),
				Fs:      fakeFs,
			}
		)
		defer os.Remove(tempDevfile.Name())

		err := d.SetDevfileContentFromBytes([]byte(InvalidDevfileContent))

		if err == nil {
			t.Errorf("expected error, didn't get one ")
		}

		if err := tempDevfile.Close(); err != nil {
			t.Errorf("failed to close temp devfile")
		}
	})
}
