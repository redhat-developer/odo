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

package webhook

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	"github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func TestWebHookInterceptor(t *testing.T) {
	payload := new(bytes.Buffer)
	_ = json.NewEncoder(payload).Encode(map[string]string{
		"eventType": "push",
		"foo":       "bar",
	})
	wantPayload := []byte("fake webhook response")

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantURL := "http://foo.default.svc/"
		if u := r.URL.String(); u != wantURL {
			t.Errorf("URL did not match: got: %s want %s", u, wantURL)
		}
		if r.Header.Get("Param-Header") != "val" {
			http.Error(w, "Expected header does not match", http.StatusBadRequest)
			return
		}
		// Return new values back in the response. It is expected for interceptors
		// to be able to mutate the request.
		w.Header().Set("Foo", "bar")
		_, _ = w.Write(wantPayload)
	}))
	defer ts.Close()
	interceptorURL, _ := url.Parse(ts.URL)
	// Proxy all requests through test server.
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(interceptorURL),
		},
	}
	webhook := &v1alpha1.WebhookInterceptor{
		ObjectRef: &corev1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Service",
			Name:       "foo",
		},
		Header: []pipelinev1.Param{{
			Name: "Param-Header",
			Value: pipelinev1.ArrayOrString{
				Type:      pipelinev1.ParamTypeString,
				StringVal: "val",
			}},
		},
	}
	i := NewInterceptor(webhook, client, "default", nil)

	incoming, _ := http.NewRequest("POST", "http://doesnotmatter.example.com", payload)
	incoming.Header.Add("Content-type", "application/json")
	resp, err := i.ExecuteTrigger(incoming)
	if err != nil {
		t.Fatalf("ExecuteTrigger: %v", err)
	}

	resPayload, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if diff := cmp.Diff(wantPayload, resPayload); diff != "" {
		t.Errorf("response payload: %s", diff)
	}

	// Check headers.
	for k, v := range map[string]string{
		"Param-Header": "",
		"Foo":          "bar",
	} {
		if s := resp.Header.Get(k); s != v {
			t.Errorf("Header[%s] = %s, want %s", k, s, v)
		}
	}
}

func TestWebHookInterceptor_NotOK(t *testing.T) {
	payload := new(bytes.Buffer)
	_ = json.NewEncoder(payload).Encode(map[string]string{
		"eventType": "push",
		"foo":       "bar",
	})

	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()
	interceptorURL, _ := url.Parse(ts.URL)
	// Proxy all requests through test server.
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(interceptorURL),
		},
	}
	webhook := &v1alpha1.WebhookInterceptor{
		ObjectRef: &corev1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Service",
			Name:       "foo",
		},
	}
	i := NewInterceptor(webhook, client, "default", nil)

	incoming, _ := http.NewRequest("POST", "http://doesnotmatter.example.com", payload)
	resp, err := i.ExecuteTrigger(incoming)
	if err == nil || resp.StatusCode != http.StatusAccepted {
		got, _ := httputil.DumpResponse(resp, true)
		t.Fatalf("ExecuteTrigger: expected (Accepted, err), got: %s", string(got))
	}

}

func TestGetURI(t *testing.T) {
	var eventListenerNs = "default"
	tcs := []struct {
		name     string
		ref      corev1.ObjectReference
		expected string
		wantErr  bool
	}{{
		name: "namespace specified",
		ref: corev1.ObjectReference{
			Kind:       "Service",
			Name:       "foo",
			APIVersion: "v1",
			Namespace:  "bar",
		},
		expected: "http://foo.bar.svc/",
		wantErr:  false,
	}, {
		name: "no namespace",
		ref: corev1.ObjectReference{
			Kind:       "Service",
			Name:       "foo",
			APIVersion: "v1",
		},
		expected: "http://foo.default.svc/",
		wantErr:  false,
	}, {
		name: "non services",
		ref: corev1.ObjectReference{
			Kind:       "Blah",
			Name:       "foo",
			APIVersion: "v1",
		},
		expected: "",
		wantErr:  true,
	}}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			url, err := getURI(&tc.ref, eventListenerNs)
			if err != nil {
				if !tc.wantErr {
					t.Errorf("Unexpected error: %v", err)
				}
			} else if diff := cmp.Diff(tc.expected, url.String()); diff != "" {
				t.Errorf("Did not get expected URL: %s", diff)
			}
		})
	}
}

func Test_addInterceptorHeaders(t *testing.T) {
	type args struct {
		header       http.Header
		headerParams []pipelinev1.Param
	}
	tests := []struct {
		name string
		args args
		want http.Header
	}{{
		name: "Empty params",
		args: args{
			header: map[string][]string{
				"Header1": {"val"},
			},
			headerParams: []pipelinev1.Param{},
		},
		want: map[string][]string{
			"Header1": {"val"},
		},
	}, {
		name: "One string param",
		args: args{
			header: map[string][]string{
				"Header1": {"val"},
			},
			headerParams: []pipelinev1.Param{{
				Name: "header2",
				Value: pipelinev1.ArrayOrString{
					Type:      pipelinev1.ParamTypeString,
					StringVal: "val",
				}},
			},
		},
		want: map[string][]string{
			"Header1": {"val"},
			"Header2": {"val"},
		},
	}, {
		name: "One array param",
		args: args{
			header: map[string][]string{
				"Header1": {"val"},
			},
			headerParams: []pipelinev1.Param{{
				Name: "header2",
				Value: pipelinev1.ArrayOrString{
					Type:     pipelinev1.ParamTypeArray,
					ArrayVal: []string{"val1", "val2"},
				}},
			},
		},
		want: map[string][]string{
			"Header1": {"val"},
			"Header2": {"val1", "val2"},
		},
	}, {
		name: "Clobber param",
		args: args{
			header: map[string][]string{
				"Header1": {"val"},
			},
			headerParams: []pipelinev1.Param{{
				Name: "header1",
				Value: pipelinev1.ArrayOrString{
					Type:     pipelinev1.ParamTypeArray,
					ArrayVal: []string{"new_val"},
				}},
			},
		},
		want: map[string][]string{
			"Header1": {"new_val"},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addInterceptorHeaders(tt.args.header, tt.args.headerParams)
			if diff := cmp.Diff(tt.want, tt.args.header); diff != "" {
				t.Errorf("addInterceptorHeaders() Diff: -want +got: %s", diff)
			}
		})
	}
}
