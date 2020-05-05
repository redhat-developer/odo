// Copyright 2018 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package oauth2

import (
	"context"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
)

func TestContextTokenSource(t *testing.T) {
	source := ContextTokenSource()
	want := new(scm.Token)

	ctx := context.Background()
	ctx = context.WithValue(ctx, scm.TokenKey{}, want)
	got, err := source.Token(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	if got != want {
		t.Errorf("Expect token retrieved from Context")
	}
}

func TestContextTokenSource_Nil(t *testing.T) {
	source := ContextTokenSource()

	ctx := context.Background()
	token, err := source.Token(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	if token != nil {
		t.Errorf("Expect nil token from Context")
	}
}
