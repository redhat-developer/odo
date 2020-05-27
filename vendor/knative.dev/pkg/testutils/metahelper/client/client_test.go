/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package client supports various needs for running tests
package client

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	fakeArtifactDir = "fakeArtifactDir"
	mockArtifactEnv = "mockArtifactDir"
)

func TestNewClient(t *testing.T) {
	datas := []struct {
		name      string
		customDir string
		expPath   string
		expErr    bool
	}{
		{"default dir", "", "mockArtifactDir/metadata.json", false},
		{"custom dir", "a", "a/metadata.json", false},
	}

	for _, data := range datas {
		t.Run(data.name, func(t *testing.T) {
			dir := data.customDir
			if data.customDir == "" { // use env var
				oriArtifactDir := os.Getenv("ARTIFACTS")
				defer os.Setenv("ARTIFACTS", oriArtifactDir)
				os.Setenv("ARTIFACTS", mockArtifactEnv)
				dir = mockArtifactEnv
			}
			os.RemoveAll(dir)
			defer os.RemoveAll(dir)
			c, err := New(data.customDir)
			if (err == nil && data.expErr) || (err != nil && !data.expErr) {
				t.Errorf("Err = %v; want?: %v", err, data.expErr)
			}
			if c.Path != data.expPath {
				t.Errorf("Path = %q, want: %q", c.Path, data.expPath)
			}
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				t.Errorf("Directory %q wasn't created", dir)
			}
		})
	}
}

func TestSync(t *testing.T) {
	datas := []struct {
		name        string
		fileExist   bool
		content     string
		expMetadata map[string]string
		expErr      bool
	}{
		{"file not exist", false, "", make(map[string]string), false},
		{"file exist but empty", true, "", make(map[string]string), true},
		{"file exist, invalid", true, "{", make(map[string]string), true},
		{"file exist valid", true, "{}", make(map[string]string), false},
	}

	for _, data := range datas {
		t.Run(data.name, func(t *testing.T) {
			c, _ := New(fakeArtifactDir)
			os.Remove(c.Path)
			if data.fileExist {
				defer os.Remove(c.Path)
				ioutil.WriteFile(c.Path, []byte(data.content), 0644)
			}
			err := c.sync()
			if (err == nil && data.expErr) || (err != nil && !data.expErr) {
				t.Errorf("Err = %v, want?: %v", err, data.expErr)
			}
			if got, want := c.metadata, data.expMetadata; !cmp.Equal(got, want) {
				t.Error("Metadata diff(-want,+got):\n", cmp.Diff(want, got))
			}
		})
	}
}

func TestSet(t *testing.T) {
	datas := []struct {
		name        string
		metadata    map[string]string
		content     string
		setKey      string
		setVal      string
		expMetadata map[string]string
		expErr      bool
	}{
		{"sync failed", make(map[string]string), "", "", "", make(map[string]string), true},
		{"set normal key", make(map[string]string), "{}", "a", "b", map[string]string{"a": "b"}, false},
		{"override", make(map[string]string), `{"a":"b"}`, "a", "c", map[string]string{"a": "c"}, false},
		{"ignore old client val", map[string]string{"a": "b"}, "{}", "c", "d", map[string]string{"c": "d"}, false},
	}

	for _, data := range datas {
		t.Run(data.name, func(t *testing.T) {
			c, _ := New(fakeArtifactDir)
			defer os.Remove(c.Path)
			ioutil.WriteFile(c.Path, []byte(data.content), 0644)

			err := c.Set(data.setKey, data.setVal)
			if (err == nil && data.expErr) || (err != nil && !data.expErr) {
				t.Errorf("Error = %v, want?: %v", err, data.expErr)
			}
			if got, want := c.metadata, data.expMetadata; !cmp.Equal(got, want) {
				t.Error("Metadata mismatch(-want,+got):\n", cmp.Diff(want, got))
			}
		})
	}
}

func TestGet(t *testing.T) {
	datas := []struct {
		name        string
		metadata    map[string]string
		content     string
		getKey      string
		expMetadata map[string]string
		expVal      string
		expErr      bool
	}{
		{"sync failed", make(map[string]string), "", "", make(map[string]string), "", true},
		{"get normal key", make(map[string]string), `{"a":"b"}`, "a", map[string]string{"a": "b"}, "b", false},
		{"key not exist", make(map[string]string), `{"a":"b"}`, "c", map[string]string{"a": "b"}, "", true},
		{"ignore old client val", map[string]string{"a": "c"}, `{"a":"b"}`, "a", map[string]string{"a": "b"}, "b", false},
	}

	for _, data := range datas {
		t.Run(data.name, func(t *testing.T) {
			c, _ := New(fakeArtifactDir)
			defer os.Remove(c.Path)
			ioutil.WriteFile(c.Path, []byte(data.content), 0644)

			val, err := c.Get(data.getKey)
			if (err == nil && data.expErr) || (err != nil && !data.expErr) {
				t.Errorf("Error = %v, want?: %v", err, data.expErr)
			}
			if got, want := c.metadata, data.expMetadata; !cmp.Equal(got, want) {
				t.Error("Metadata mismatch (-want,+got):\n", cmp.Diff(want, got))
			}
			if got, want := val, data.expVal; got != want {
				t.Errorf("Value = %q, want: %q", got, want)
			}
		})
	}
}
