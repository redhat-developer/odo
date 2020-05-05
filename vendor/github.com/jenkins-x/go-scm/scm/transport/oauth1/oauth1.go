// Copyright 2018 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package oauth1

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/transport/internal"
)

// clock provides a interface for current time providers. A Clock can be used
// in place of calling time.Now() directly.
type clock interface {
	Now() time.Time
}

// A noncer provides random nonce strings.
type noncer interface {
	Nonce() string
}

// Transport is an http.RoundTripper that refreshes oauth
// tokens, wrapping a base RoundTripper and refreshing the
// token if expired.
type Transport struct {
	// Consumer Key
	ConsumerKey string

	// Consumer Private Key
	PrivateKey *rsa.PrivateKey

	// Source supplies the Token to add to the request
	// Authorization headers.
	Source scm.TokenSource

	// Base is the base RoundTripper used to make requests.
	// If nil, http.DefaultTransport is used.
	Base http.RoundTripper

	noncer noncer
	clock  clock
}

// RoundTrip authorizes and authenticates the request with
// an access token from the request context.
func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	ctx := r.Context()
	token, err := t.Source.Token(ctx)
	if err != nil {
		return nil, err
	}
	if token == nil {
		return t.base().RoundTrip(r)
	}
	r2 := internal.CloneRequest(r)
	err = t.setRequestAuthHeader(r2, token)
	if err != nil {
		return nil, err
	}
	return t.base().RoundTrip(r2)
}

// base returns the base transport. If no base transport
// is configured, the default transport is returned.
func (t *Transport) base() http.RoundTripper {
	if t.Base != nil {
		return t.Base
	}
	return http.DefaultTransport
}

// setRequestAuthHeader sets the OAuth1 header for making
// authenticated requests with an AccessToken according to
// RFC 5849 3.1.
func (t *Transport) setRequestAuthHeader(r *http.Request, token *scm.Token) error {
	oauthParams := t.commonOAuthParams()
	oauthParams["oauth_token"] = token.Token
	params := collectParameters(r, oauthParams)

	signatureBase := signatureBase(r, params)
	signature, err := sign(t.PrivateKey, signatureBase)
	if err != nil {
		return err
	}
	oauthParams["oauth_signature"] = signature
	r.Header.Set("Authorization", authHeaderValue(oauthParams))
	return nil
}

// commonOAuthParams returns a map of the common OAuth1
// protocol parameters, excluding the oauth_signature.
func (t *Transport) commonOAuthParams() map[string]string {
	return map[string]string{
		"oauth_consumer_key":     t.ConsumerKey,
		"oauth_signature_method": "RSA-SHA1",
		"oauth_timestamp":        strconv.FormatInt(t.epoch(), 10),
		"oauth_nonce":            t.nonce(),
		"oauth_version":          "1.0",
	}
}

// Returns a base64 encoded random 32 byte string.
func (t *Transport) nonce() string {
	if t.noncer != nil {
		return t.noncer.Nonce()
	}
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// Returns the Unix epoch seconds.
func (t *Transport) epoch() int64 {
	if t.clock != nil {
		return t.clock.Now().Unix()
	}
	return time.Now().Unix()
}

// authHeaderValue formats OAuth parameters according to
// RFC 5849 3.5.1.
func authHeaderValue(oauthParams map[string]string) string {
	pairs := sortParameters(encodeParameters(oauthParams), `%s="%s"`)
	return "OAuth " + strings.Join(pairs, ", ")
}

// collectParameters returns a map of request parameter keys
// and values as defined in RFC 5849 3.4.1.3.
func collectParameters(r *http.Request, oauthParams map[string]string) map[string]string {
	params := map[string]string{}
	for key, value := range r.URL.Query() {
		params[key] = value[0]
	}
	for key, value := range oauthParams {
		params[key] = value
	}
	return params
}
