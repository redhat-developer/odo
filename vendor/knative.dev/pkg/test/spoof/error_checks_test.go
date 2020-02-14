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

// spoof contains logic to make polling HTTP requests against an endpoint with optional host spoofing.

package spoof

import (
	"errors"
	"net/http"
	"testing"
)

func TestDNSError(t *testing.T) {
	client := &http.Client{}

	for _, tt := range []struct {
		name     string
		url      string
		dnsError bool
	}{{
		name:     "url does not exist",
		url:      "http://this.url.does.not.exist",
		dnsError: true,
	}, {
		name:     "ip address",
		url:      "http://127.0.0.1",
		dnsError: false,
	}, {
		name:     "localhost",
		url:      "http://localhost:8080",
		dnsError: false,
	}, {
		name:     "no error",
		url:      "http://google.com",
		dnsError: false,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			_, err := client.Do(req)
			if dnsError := isDNSError(err); tt.dnsError != dnsError {
				t.Errorf("Expected dnsError=%v, got %v", tt.dnsError, dnsError)
			}
		})
	}
}

func TestConnectionRefused(t *testing.T) {
	client := &http.Client{}

	for _, tt := range []struct {
		name        string
		url         string
		connRefused bool
	}{{
		name:        "nothing listening",
		url:         "http://localhost:60001",
		connRefused: true,
	}, {
		name:        "dns error",
		url:         "http://this.url.does.not.exist",
		connRefused: false,
	}, {
		name:        "google.com",
		url:         "https://google.com",
		connRefused: false,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			_, err := client.Do(req)
			if connRefused := isConnectionRefused(err); tt.connRefused != connRefused {
				t.Errorf("Expected connRefused=%v, got %v", tt.connRefused, connRefused)
			}
		})
	}
}

func TestConnectionReset(t *testing.T) {
	for _, tt := range []struct {
		name      string
		err       error
		connReset bool
	}{{
		name:      "error matching",
		err:       errors.New("read tcp 10.60.2.57:47882->104.154.144.94:80: read: connection reset by peer"),
		connReset: true,
	}, {
		name:      "error not matching",
		err:       errors.New("dial tcp: lookup this.url.does.not.exist on 127.0.0.1:53: no such host"),
		connReset: false,
	}, {
		name:      "nil error",
		err:       nil,
		connReset: false,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			if connReset := isConnectionReset(tt.err); tt.connReset != connReset {
				t.Errorf("Expected connReset=%v, got %v", tt.connReset, connReset)
			}
		})
	}
}

func TestTCPTimeout(t *testing.T) {
	client := &http.Client{}

	// We have no positive test for TCP timeout, but we do have a few negative tests.
	for _, tt := range []struct {
		name       string
		url        string
		tcpTimeout bool
	}{{
		name:       "nothing listening",
		url:        "http://localhost:60001",
		tcpTimeout: false,
	}, {
		name:       "dns error",
		url:        "http://this.url.does.not.exist",
		tcpTimeout: false,
	}, {
		name:       "google.com",
		url:        "https://google.com",
		tcpTimeout: false,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			_, err := client.Do(req)
			if tcpTimeout := isTCPTimeout(err); tt.tcpTimeout != tcpTimeout {
				t.Errorf("Expected tcpTimeout=%v, got %v", tt.tcpTimeout, tcpTimeout)
			}
		})
	}
}
