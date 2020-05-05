// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package integration

import (
	"net/http"
	"os"
	"testing"

	"github.com/jenkins-x/go-scm/scm/driver/github"
	"github.com/jenkins-x/go-scm/scm/transport"
)

func TestGitHub(t *testing.T) {
	if os.Getenv("GITHUB_TOKEN") == "" {
		t.Skipf("missing GITHUB_TOKEN environment variable")
		return
	}

	client := github.NewDefault()
	client.Client = &http.Client{
		Transport: &transport.BearerToken{
			Token: os.Getenv("GITHUB_TOKEN"),
		},
	}

	t.Run("Contents", testContents(client))
	t.Run("Git", testGit(client))
	t.Run("Issues", testIssues(client))
	t.Run("Organizations", testOrgs(client))
	t.Run("PullRequests", testPullRequests(client))
	t.Run("Repositories", testRepos(client))
	t.Run("Users", testUsers(client))
}
