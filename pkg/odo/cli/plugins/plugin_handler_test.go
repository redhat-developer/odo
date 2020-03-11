package plugins

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

var sampleScript = []byte(`
#!/bin/sh
echo 'hello'
`)

func TestDefaultHandlerLookup(t *testing.T) {
	tempDir, cleanup := makeTempDir(t)
	defer cleanup()

	origPath := os.Getenv("PATH")
	defer func() {
		os.Setenv("PATH", origPath)
	}()
	os.Setenv("PATH", fmt.Sprintf("%s:%s", origPath, tempDir))
	scriptName := path.Join(tempDir, "tst-script")
	err := ioutil.WriteFile(scriptName, sampleScript, 0755)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		prefix string
		want   string
	}{
		{"tst", scriptName},
		{"unk", ""},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("prefix:%s", tt.prefix), func(rt *testing.T) {
			h := NewDefaultHandler(tt.prefix)
			found := h.Lookup("script")
			if tt.want != found {
				rt.Errorf("found script name got %s, want %s", tt.want, found)
			}
		})
	}
}

func makeTempDir(t *testing.T) (string, func()) {
	dir, err := ioutil.TempDir(os.TempDir(), "odo")
	if err != nil {
		t.Fatal(err)
	}
	return dir, func() {
		err := os.RemoveAll(dir)
		if err != nil {
			t.Fatal(err)
		}
	}
}
