// Copyright 2018 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package oauth1

import (
	"context"

	"github.com/jenkins-x/go-scm/scm"
)

// StaticTokenSource returns a TokenSource that always
// returns the same token. Because the provided token t
// is never refreshed, StaticTokenSource is only useful
// for tokens that never expire.
func StaticTokenSource(t *scm.Token) scm.TokenSource {
	return staticTokenSource{t}
}

type staticTokenSource struct {
	token *scm.Token
}

func (s staticTokenSource) Token(context.Context) (*scm.Token, error) {
	return s.token, nil
}

// ContextTokenSource returns a TokenSource that returns
// a token from the http.Request context.
func ContextTokenSource() scm.TokenSource {
	return contextTokenSource{}
}

type contextTokenSource struct {
}

func (s contextTokenSource) Token(ctx context.Context) (*scm.Token, error) {
	token, _ := ctx.Value(scm.TokenKey{}).(*scm.Token)
	return token, nil
}
