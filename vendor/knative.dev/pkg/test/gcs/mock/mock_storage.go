/*
Copyright 2020 The Knative Authors

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

package mock

import (
	"path/filepath"
)

// more friendly type casts for better readability of what some strings are
type project string
type bucket string
type Method string

// mockpath contains the bucket path to an object and the object name
type mockpath struct {
	dir string
	obj string
}

func newMockPath(dir, obj string) mockpath {
	return mockpath{
		dir: dir,
		obj: obj,
	}
}

// toString stringify mockpath
func (m mockpath) toString() string {
	return filepath.Join(m.dir, m.obj)
}

// Fake GCS objects
type object struct {
	name mockpath
	//NOTE: current ObjectAttrs supported:
	//	Size
	//	Bucket
	//	Name
	bkt     string
	content []byte
}

// bucket of objects - structure is flat
type objects struct {
	obj map[mockpath]*object
}

// project with buckets
type buckets struct {
	bkt map[bucket]*objects
}

// Error map to return custom errors for specific methods
type ReturnError struct {
	NumCall uint8
	Err     error
}
