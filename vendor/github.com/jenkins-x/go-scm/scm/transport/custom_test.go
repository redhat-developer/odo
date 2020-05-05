// Copyright 2018 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transport

import (
	"net/http"
	"testing"

	"github.com/h2non/gock"
)

func TestCustomTransport(t *testing.T) {
	defer gock.Off()

	gock.New("https://try.gogs.io").
		Get("/api/user").
		MatchHeader("Authorization", "token mF_9.B5f-4.1JqM").
		Reply(200)

	client := &http.Client{
		Transport: &Custom{
			Before: func(r *http.Request) {
				r.Header.Set("Authorization", "token mF_9.B5f-4.1JqM")
			},
		},
	}

	res, err := client.Get("https://try.gogs.io/api/user")
	if err != nil {
		t.Error(err)
		return
	}
	defer res.Body.Close()
}
