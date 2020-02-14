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

package boskos

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"knative.dev/pkg/testutils/clustermanager/e2e-tests/common"
)

var (
	fakeHost = "fakehost"
	fakeRes  = "{\"name\": \"res\", \"type\": \"t\", \"state\": \"d\"}"
)

// create a fake server as Boskos server, must close() afterwards
func fakeServer(f func(http.ResponseWriter, *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(f))
}

func TestAcquireGKEProject(t *testing.T) {
	mockJobName := "mockjobname"
	tests := []struct {
		name      string
		serverErr bool
		host      string
		expHost   string
		expErr    bool
	}{
		{"Test boskos server error", true, fakeHost, "fakehost", true},
		{"Test passing host as param", false, fakeHost, "fakehost", false},
		{"Test using default host", false, "", "mockjobname", false},
	}

	oldBoskosURI := boskosURI
	defer func() {
		boskosURI = oldBoskosURI
	}()
	oldGetOSEnv := common.GetOSEnv
	common.GetOSEnv = func(s string) string {
		if s == "JOB_NAME" {
			return mockJobName
		}
		return oldGetOSEnv(s)
	}
	defer func() {
		common.GetOSEnv = oldGetOSEnv
	}()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := fakeServer(func(w http.ResponseWriter, r *http.Request) {
				if tt.serverErr {
					http.Error(w, "Bad Request", http.StatusBadRequest)
				} else {
					// RequestURI for acquire contains a random hash, doing
					// substring matching instead
					for _, s := range []string{"/acquire?", "owner=" + tt.expHost, "state=free", "dest=busy", "type=gke-project"} {
						if !strings.Contains(r.RequestURI, s) {
							t.Fatalf("Request URI = %q, want: %q", r.RequestURI, s)
						}
					}
					// Return a mocked fake http response
					fmt.Fprint(w, fakeRes)
				}
			})
			defer ts.Close()
			boskosURI = ts.URL
			client, err := NewClient(tt.host, /* boskos owner */
				"", /* boskos user */
				"" /* boskos password file */)
			if err != nil {
				t.Fatalf("Failed to create test client %v", err)
			}
			_, err = client.AcquireGKEProject(GKEProjectResource)
			if tt.expErr && (err == nil) {
				t.Fatal("No expected error when acquiring GKE project.")
			}
			if !tt.expErr && (err != nil) {
				t.Fatalf("Unexpected error when acquiring GKE project, '%v'", err)
			}
		})
	}
}

func TestReleaseGKEProject(t *testing.T) {
	mockJobName := "mockjobname"
	tests := []struct {
		name      string
		serverErr bool
		host      string
		resName   string
		expReq    string
		expErr    bool
	}{
		{"Test boskos server error", true, fakeHost, "a", "/release?dest=dirty&name=a&owner=fakehost", true},
		{"Test passing host as param", false, fakeHost, "b", "/release?dest=dirty&name=b&owner=fakehost", false},
		{"Test using default host", false, "", "c", "/release?dest=dirty&name=c&owner=mockjobname", false},
	}
	oldBoskosURI := boskosURI
	defer func() {
		boskosURI = oldBoskosURI
	}()
	oldGetOSEnv := common.GetOSEnv
	common.GetOSEnv = func(s string) string {
		if s == "JOB_NAME" {
			return mockJobName
		}
		return oldGetOSEnv(s)
	}
	defer func() {
		common.GetOSEnv = oldGetOSEnv
	}()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := fakeServer(func(w http.ResponseWriter, r *http.Request) {
				if tt.serverErr {
					http.Error(w, "", http.StatusBadRequest)
				} else if r.RequestURI != tt.expReq {
					t.Fatalf("Request URI doesn't match: want: '%s', got: '%s'", tt.expReq, r.RequestURI)
				} else {
					fmt.Fprint(w, "")
				}
			})
			defer ts.Close()
			boskosURI = ts.URL
			client, err := NewClient(tt.host, /* boskos owner */
				"", /* boskos user */
				"" /* boskos password file */)
			if err != nil {
				t.Fatalf("Failed to create test client %v", err)
			}
			err = client.ReleaseGKEProject(tt.resName)
			if tt.expErr && (err == nil) {
				t.Fatal("No expected error when releasing GKE project.")
			}
			if !tt.expErr && (err != nil) {
				t.Fatalf("Unexpected error when releasing GKE project, '%v'", err)
			}
		})
	}
}
