// Copyright 2018 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transport

import (
	"net/http"
	"testing"

	"github.com/h2non/gock"
)

func TestPrivateToken(t *testing.T) {
	defer gock.Off()

	gock.New("https://gitlab.com").
		Get("/api/v4/user").
		MatchHeader("Private-Token", "5d41402abc4b").
		Reply(200)

	client := &http.Client{
		Transport: &PrivateToken{
			Token: "5d41402abc4b",
		},
	}

	res, err := client.Get("https://gitlab.com/api/v4/user")
	if err != nil {
		t.Error(err)
		return
	}
	defer res.Body.Close()
}

func TestPrivateToken_DontOverwriteHeader(t *testing.T) {
	defer gock.Off()

	gock.New("https://gitlab.com").
		Get("/api/v4/user").
		MatchHeader("Private-Token", "5d41402abc4b").
		Reply(200)

	client := &http.Client{
		Transport: &PrivateToken{
			Token: "9d911017c592",
		},
	}

	req, err := http.NewRequest("GET", "https://gitlab.com/api/v4/user", nil)
	if err != nil {
		t.Error(err)
		return
	}
	req.Header.Set("Private-Token", "5d41402abc4b")
	res, err := client.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	defer res.Body.Close()
}
