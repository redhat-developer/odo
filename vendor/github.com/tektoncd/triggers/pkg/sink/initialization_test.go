/*
Copyright 2019 The Tekton Authors

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

package sink

import (
	"flag"
	"testing"
)

func Test_GetArgs(t *testing.T) {
	if err := flag.Set(name, "elname"); err != nil {
		t.Errorf("Error setting flag el-name: %s", err)
	}
	if err := flag.Set(elNamespace, "elnamespace"); err != nil {
		t.Errorf("Error setting flag el-namespace: %s", err)
	}
	if err := flag.Set(port, "port"); err != nil {
		t.Errorf("Error setting flag port: %s", err)
	}
	sinkArgs, err := GetArgs()
	if err != nil {
		t.Fatalf("GetArgs() returned unexpected error: %s", err)
	}
	if sinkArgs.ElName != "elname" {
		t.Errorf("Error el-name want elname, got %s", sinkArgs.ElName)
	}
	if sinkArgs.ElNamespace != "elnamespace" {
		t.Errorf("Error el-namespace want elnamespace, got %s", sinkArgs.ElNamespace)
	}
	if sinkArgs.Port != "port" {
		t.Errorf("Error port want port, got %s", sinkArgs.Port)
	}
}

func Test_GetArgs_error(t *testing.T) {
	tests := []struct {
		name        string
		elName      string
		elNamespace string
		port        string
	}{{
		name:        "no eventlistener name",
		elName:      "",
		elNamespace: "elnamespace",
		port:        "port",
	}, {
		name:        "no eventlistener namespace",
		elName:      "elname",
		elNamespace: "",
		port:        "port",
	}, {
		name:        "no eventlistener namespace",
		elName:      "elname",
		elNamespace: "elnamespace",
		port:        "",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := flag.Set(name, tt.elName); err != nil {
				t.Errorf("Error setting flag %s: %s", name, err)
			}
			if err := flag.Set("el-namespace", tt.elNamespace); err != nil {
				t.Errorf("Error setting flag %s: %s", namespace, err)
			}
			if err := flag.Set("port", tt.port); err != nil {
				t.Errorf("Error setting flag %s: %s", port, err)
			}
			if sinkArgs, err := GetArgs(); err == nil {
				t.Errorf("GetArgs() did not return error when expected; sinkArgs: %v", sinkArgs)
			}
		})
	}
}
