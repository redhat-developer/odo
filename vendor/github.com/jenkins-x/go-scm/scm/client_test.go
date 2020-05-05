// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scm

import (
	"net/http"
	"testing"
)

func TestClient(t *testing.T) {
	t.Skip()
}

func TestResponse(t *testing.T) {
	res := newResponse(&http.Response{
		StatusCode: 200,
		Header: http.Header{
			"Link": {`<https://api.github.com/resource?page=4>; rel="next",
				<https://api.github.com/resource?page=2>; rel="prev",
				<https://api.github.com/resource?page=1>; rel="first",
				<https://api.github.com/resource?page=5>; rel="last"`},
		},
	})
	if got, want := res.Status, 200; got != want {
		t.Errorf("Want status code %d, got %d", want, got)
	}
	if got, want := res.Page.First, 1; got != want {
		t.Errorf("Want rel first %d, got %d", want, got)
	}
	if got, want := res.Page.Last, 5; got != want {
		t.Errorf("Want rel last %d, got %d", want, got)
	}
	if got, want := res.Page.Prev, 2; got != want {
		t.Errorf("Want rel prev %d, got %d", want, got)
	}
	if got, want := res.Page.Next, 4; got != want {
		t.Errorf("Want rel next %d, got %d", want, got)
	}
}
