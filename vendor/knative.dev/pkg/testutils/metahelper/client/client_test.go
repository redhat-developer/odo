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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
)

const (
	fakeArtifactDir = "fakeArtifactDir"
	mockArtifactEnv = "mockArtifactDir"
)

func TestNewClient(t *testing.T) {
	datas := []struct {
		customDir string
		expPath   string
		expErr    bool
	}{
		{ // default dir
			"", "mockArtifactDir/metadata.json", false,
		}, { // custom dir
			"a", "a/metadata.json", false,
		},
	}

	for _, data := range datas {
		dir := data.customDir
		if data.customDir == "" { // use env var
			oriArtifactDir := os.Getenv("ARTIFACTS")
			defer os.Setenv("ARTIFACTS", oriArtifactDir)
			os.Setenv("ARTIFACTS", mockArtifactEnv)
			dir = mockArtifactEnv
		}
		os.RemoveAll(dir)
		defer os.RemoveAll(dir)
		c, err := NewClient(data.customDir)
		errMsg := fmt.Sprintf("Testing new client with dir: %q", data.customDir)
		if (err == nil && data.expErr) || (err != nil && !data.expErr) {
			log.Fatalf("%s\ngot: '%v', want: '%v'", errMsg, err, data.expErr)
		}
		if c.Path != data.expPath {
			log.Fatalf("%s\ngot: %q, want: %q", errMsg, c.Path, data.expPath)
		}
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			log.Fatalf("%s\nDirectory %q wasn't created", errMsg, dir)
		}

	}
}

func TestSync(t *testing.T) {
	datas := []struct {
		fileExist   bool
		content     string
		expMetadata map[string]string
		expErr      bool
	}{
		{ // file not exist
			false, "", make(map[string]string), false,
		}, { // file exist but empty
			true, "", make(map[string]string), true,
		}, { // file exist, invalid
			true, "{", make(map[string]string), true,
		}, { // file exist valid
			true, "{}", make(map[string]string), false,
		},
	}

	for _, data := range datas {
		c, _ := NewClient(fakeArtifactDir)
		os.Remove(c.Path)
		if data.fileExist {
			defer os.Remove(c.Path)
			ioutil.WriteFile(c.Path, []byte(data.content), 0644)
		}
		err := c.sync()
		errMsg := fmt.Sprintf("Testing syncing with file exist: %v, content: %q", data.fileExist, data.content)
		if (err == nil && data.expErr) || (err != nil && !data.expErr) {
			log.Fatalf("%s\ngot: '%v', want: '%v'", errMsg, err, data.expErr)
		}
		if !reflect.DeepEqual(c.MetaData, data.expMetadata) {
			log.Fatalf("%s\ngot: '%v', want: '%v'", errMsg, c.MetaData, data.expMetadata)
		}
	}
}

func TestSet(t *testing.T) {
	datas := []struct {
		metadata    map[string]string
		content     string
		setKey      string
		setVal      string
		expMetadata map[string]string
		expErr      bool
	}{
		{ // sync failed
			make(map[string]string), "", "", "", make(map[string]string), true,
		}, { // set normal key
			make(map[string]string), "{}", "a", "b", map[string]string{"a": "b"}, false,
		}, { // override
			make(map[string]string), "{\"a\":\"b\"}", "a", "c", map[string]string{"a": "c"}, false,
		}, { // ignore old client val
			map[string]string{"a": "b"}, "{}", "c", "d", map[string]string{"c": "d"}, false,
		},
	}

	for _, data := range datas {
		c, _ := NewClient(fakeArtifactDir)
		defer os.Remove(c.Path)
		ioutil.WriteFile(c.Path, []byte(data.content), 0644)

		err := c.Set(data.setKey, data.setVal)
		errMsg := fmt.Sprintf("Testing set %q:%q, with metadata: %v, content: %q",
			data.setKey, data.setVal, data.metadata, data.content)
		if (err == nil && data.expErr) || (err != nil && !data.expErr) {
			log.Fatalf("%s\ngot: '%v', want: '%v'", errMsg, err, data.expErr)
		}
		if !reflect.DeepEqual(c.MetaData, data.expMetadata) {
			log.Fatalf("%s\ngot: '%v', want: '%v'", errMsg, c.MetaData, data.expMetadata)
		}
	}
}

func TestGet(t *testing.T) {
	datas := []struct {
		metadata    map[string]string
		content     string
		getKey      string
		expMetadata map[string]string
		expVal      string
		expErr      bool
	}{
		{ // sync failed
			make(map[string]string), "", "", make(map[string]string), "", true,
		}, { // get normal key
			make(map[string]string), "{\"a\":\"b\"}", "a", map[string]string{"a": "b"}, "b", false,
		}, { // key not exist
			make(map[string]string), "{\"a\":\"b\"}", "c", map[string]string{"a": "b"}, "", true,
		}, { // ignore old client val
			map[string]string{"a": "c"}, "{\"a\":\"b\"}", "a", map[string]string{"a": "b"}, "b", false,
		},
	}

	for _, data := range datas {
		c, _ := NewClient(fakeArtifactDir)
		defer os.Remove(c.Path)
		ioutil.WriteFile(c.Path, []byte(data.content), 0644)

		val, err := c.Get(data.getKey)
		errMsg := fmt.Sprintf("Testing get %q, with metadata: %v, content: %q",
			data.getKey, data.metadata, data.content)
		if (err == nil && data.expErr) || (err != nil && !data.expErr) {
			log.Fatalf("%s\ngot: '%v', want: '%v'", errMsg, err, data.expErr)
		}
		if !reflect.DeepEqual(c.MetaData, data.expMetadata) {
			log.Fatalf("%s\ngot: '%v', want: '%v'", errMsg, c.MetaData, data.expMetadata)
		}
		if val != data.expVal {
			log.Fatalf("%s\ngot: %q, want: %q", errMsg, val, data.expVal)
		}
	}
}
