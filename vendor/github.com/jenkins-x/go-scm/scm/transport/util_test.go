// Copyright 2018 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transport

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCopyRequest(t *testing.T) {
	b := new(bytes.Buffer)
	r1, _ := http.NewRequest("GET", "http://example.com", b)
	r1.Header.Set("Accept", "application/json")
	r1.Header.Set("Etag", "1")
	r2 := cloneRequest(r1)
	if r1 == r2 {
		t.Errorf("Expect http.Request cloned")
	}
	if diff := cmp.Diff(r1.Header, r2.Header); diff != "" {
		t.Errorf("Expect http.Header cloned")
		t.Log(diff)
	}
}
