// Copyright 2018 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package oauth2

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/h2non/gock"
	"github.com/jenkins-x/go-scm/scm"
)

func TestRefresh(t *testing.T) {
	defer gock.Off()

	gock.New("https://bitbucket.org").
		Post("/site/oauth2/access_token").
		Reply(200).
		BodyString(`
			{
				"access_token": "9698fa6a8113b3",
				"expires_in": 7200,
				"refresh_token":
				"3a2bfce4cb9b0f",
				"token_type": "bearer"
			}
		`)

	before := &scm.Token{
		Refresh: "3a2bfce4cb9b0f",
	}

	r := Refresher{
		ClientID:     "dafe3804960dab",
		ClientSecret: "20e651849b1f12",
		Endpoint:     "https://bitbucket.org/site/oauth2/access_token",
		Source:       StaticTokenSource(before),
	}

	ctx := context.Background()
	after, err := r.Token(ctx)
	if err != nil {
		t.Error(err)
	}

	if after.Token != "9698fa6a8113b3" {
		t.Errorf("Expect access token updated")
	}
	if after.Expires.IsZero() {
		t.Errorf("Expect access token expiry updated")
	}
	if after.Refresh != "3a2bfce4cb9b0f" {
		t.Errorf("Expect refresh token not changed, got %s", after.Refresh)
	}
}

func TestRefresh_Error(t *testing.T) {
	defer gock.Off()

	gock.New("https://bitbucket.org").
		Post("/site/oauth2/access_token").
		Reply(400).
		BodyString(`
			{
				"error_description": "Invalid OAuth client credentials",
				"error": "unauthorized_client"
			}
		`)

	r := Refresher{
		ClientID:     "dafe3804960dab",
		ClientSecret: "20e651849b1f12",
		Endpoint:     "https://bitbucket.org/site/oauth2/access_token",
		Source: StaticTokenSource(&scm.Token{
			Refresh: "3a2bfce4cb9b0f",
		}),
	}

	ctx := context.Background()
	_, got := r.Token(ctx)
	if got == nil {
		t.Errorf("Expect Oauth Error")
		return
	}

	want := &tokenError{
		Code:    "unauthorized_client",
		Message: "Invalid OAuth client credentials",
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Error message")
		t.Log(diff)
	}
	if got, want := got.Error(), want.Message; got != want {
		t.Errorf("Want Error.String() %s, got %s", want, got)
	}
}

func TestRefresh_NotExpired(t *testing.T) {
	before := &scm.Token{
		Token: "6084984dab20e6",
	}
	r := Refresher{
		ClientID:     "dafe3804960dab",
		ClientSecret: "20e651849b1f12",
		Endpoint:     "https://bitbucket.org/site/oauth2/access_token",
		Source:       StaticTokenSource(before),
	}

	ctx := context.Background()
	after, err := r.Token(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	if after == nil {
		t.Errorf("Expected Token, got nil")
		return
	}
	if after.Token != "6084984dab20e6" {
		t.Errorf("Expect Token not refreshed")
	}
}

func TestExpired(t *testing.T) {
	tests := []struct {
		token   *scm.Token
		expired bool
	}{
		{
			expired: false,
			token: &scm.Token{
				Token:   "12345",
				Refresh: "",
			},
		},
		{
			expired: false,
			token: &scm.Token{
				Token:   "12345",
				Refresh: "",
				Expires: time.Now().Add(-time.Hour),
			},
		},
		{
			expired: false,
			token: &scm.Token{
				Token:   "12345",
				Refresh: "54321",
			},
		},
		{
			expired: false,
			token: &scm.Token{
				Token:   "12345",
				Refresh: "54321",
				Expires: time.Now().Add(time.Hour),
			},
		},
		// missing access token
		{
			expired: true,
			token: &scm.Token{
				Token:   "",
				Refresh: "54321",
			},
		},
		// token expired
		{
			expired: true,
			token: &scm.Token{
				Token:   "12345",
				Refresh: "54321",
				Expires: time.Now().Add(-time.Second),
			},
		},
		// this token is not expired, however, it is within
		// the default 1 minute expiry window.
		{
			expired: true,
			token: &scm.Token{
				Token:   "12345",
				Refresh: "54321",
				Expires: time.Now().Add(time.Second * 30),
			},
		},
	}

	for i, test := range tests {
		if got, want := expired(test.token), test.expired; got != want {
			t.Errorf("Want token expired %v, got %v at index %d", want, got, i)
		}
	}
}
