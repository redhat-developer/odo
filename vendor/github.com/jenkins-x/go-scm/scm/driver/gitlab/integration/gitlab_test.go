// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package integration

import (
	"net/http"
	"os"
	"testing"

	"github.com/jenkins-x/go-scm/scm/driver/gitlab"
	"github.com/jenkins-x/go-scm/scm/transport"
)

func TestGitLab(t *testing.T) {
	if os.Getenv("GITLAB_TOKEN") == "" {
		t.Skipf("missing GITLAB_TOKEN environment variable")
		return
	}

	client, _ := gitlab.New("https://gitlab.com/")
	client.Client = &http.Client{
		Transport: &transport.PrivateToken{
			Token: os.Getenv("GITLAB_TOKEN"),
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
