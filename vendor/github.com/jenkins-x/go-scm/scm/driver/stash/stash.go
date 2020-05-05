// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package stash implements a Bitbucket Server client.
package stash

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/url"
	"strings"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/driver/internal/null"
)

// Reference API Documentation:
//   https://docs.atlassian.com/bitbucket-server/rest/5.11.1/bitbucket-rest.html

// New returns a new Stash API client.
func New(uri string) (*scm.Client, error) {
	base, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(base.Path, "/") {
		base.Path = base.Path + "/"
	}
	client := &wrapper{new(scm.Client)}
	client.BaseURL = base
	// initialize services
	client.Driver = scm.DriverStash
	client.Contents = &contentService{client}
	client.Git = &gitService{client}
	client.Issues = &issueService{client}
	client.Organizations = &organizationService{client}
	client.PullRequests = &pullService{client}
	client.Repositories = &repositoryService{client}
	client.Reviews = &reviewService{client}
	client.Users = &userService{client}
	client.Webhooks = &webhookService{client}
	return client.Client, nil
}

// NewDefault returns a new Stash API client.
func NewDefault() *scm.Client {
	client, _ := New("http://localhost:7990")
	return client
}

// wraper wraps the Client to provide high level helper functions
// for making http requests and unmarshaling the response.
type wrapper struct {
	*scm.Client
}

// do wraps the Client.Do function by creating the Request and
// unmarshalling the response.
func (c *wrapper) do(ctx context.Context, method, path string, in, out interface{}) (*scm.Response, error) {
	req := &scm.Request{
		Method: method,
		Path:   path,
	}
	// if we are posting or putting data, we need to
	// write it to the body of the request.
	if in != nil {
		buf := new(bytes.Buffer)
		json.NewEncoder(buf).Encode(in)
		req.Header = map[string][]string{
			"Content-Type": {"application/json"},
		}
		req.Body = buf
	}

	// execute the http request
	res, err := c.Client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// if an error is encountered, unmarshal and return the
	// error response.
	if res.Status == 401 {
		return res, scm.ErrNotAuthorized
	} else if res.Status > 300 {
		err := new(Error)
		json.NewDecoder(res.Body).Decode(err)
		return res, err
	}

	if out == nil {
		return res, nil
	}

	// if raw output is expected, copy to the provided
	// buffer and exit.
	if w, ok := out.(io.Writer); ok {
		io.Copy(w, res.Body)
		return res, nil
	}

	// if a json response is expected, parse and return
	// the json response.
	return res, json.NewDecoder(res.Body).Decode(out)
}

// pagination represents Bitbucket pagination properties
// embedded in list responses.
type pagination struct {
	Start    null.Int  `json:"start"`
	Size     null.Int  `json:"size"`
	Limit    null.Int  `json:"limit"`
	LastPage null.Bool `json:"isLastPage"`
	NextPage null.Int  `json:"nextPageStart"`
}

// Error represents a Stash error.
type Error struct {
	Errors []struct {
		Message         string `json:"message"`
		ExceptionName   string `json:"exceptionName"`
		CurrentVersion  int    `json:"currentVersion"`
		ExpectedVersion int    `json:"expectedVersion"`
	} `json:"errors"`
}

func (e *Error) Error() string {
	if len(e.Errors) == 0 {
		return "No message available"
	}
	return e.Errors[0].Message
}
