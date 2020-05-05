// Copyright 2018 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package oauth1

import (
	"net/http"
	"net/url"
	"testing"
)

func TestBaseURL(t *testing.T) {
	tests := []struct {
		before string
		after  string
	}{
		{
			before: "HTTP://EXAMPLE.COM:80/r%20v/X?id=123",
			after:  "http://example.com/r%20v/X",
		},
		{
			before: "http://example.com:80",
			after:  "http://example.com",
		},
		{
			before: "https://example.com:443",
			after:  "https://example.com",
		},
		{
			before: "http://www.example.com:8080/?q=1",
			after:  "http://www.example.com:8080/",
		},
	}
	for _, test := range tests {
		r := new(http.Request)
		r.URL, _ = url.Parse(test.before)
		if got, want := baseURI(r), test.after; got != want {
			t.Errorf("Want url %s, got %s", want, got)
		}
	}
}
