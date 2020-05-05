// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/stretchr/testify/assert"

	"github.com/google/go-cmp/cmp"
)

func TestWebhooks(t *testing.T) {
	tests := []struct {
		name   string
		event  string
		before string
		after  string
		obj    interface{}
	}{
		//
		// push events
		//

		// fork
		{
			name:   "fork",
			event:  "fork",
			before: "testdata/webhooks/fork.json",
			after:  "testdata/webhooks/fork.json.golden",
			obj:    new(scm.ForkHook),
		},

		// repository
		{
			name:   "repository",
			event:  "repository",
			before: "testdata/webhooks/repository.json",
			after:  "testdata/webhooks/repository.json.golden",
			obj:    new(scm.RepositoryHook),
		},

		// installation_repositories
		{
			name:   "installation_repositories",
			event:  "installation_repositories",
			before: "testdata/webhooks/installation_repository.json",
			after:  "testdata/webhooks/installation_repository.json.golden",
			obj:    new(scm.InstallationRepositoryHook),
		},

		// check_suite
		{
			name:   "check_suite",
			event:  "check_suite",
			before: "testdata/webhooks/check_suite_created.json",
			after:  "testdata/webhooks/check_suite_created.json.golden",
			obj:    new(scm.CheckSuiteHook),
		},

		// deployment_status
		{
			name:   "deployment_status",
			event:  "deployment_status",
			before: "testdata/webhooks/deployment_status.json",
			after:  "testdata/webhooks/deployment_status.json.golden",
			obj:    new(scm.DeploymentStatusHook),
		},

		// release
		{
			name:   "release",
			event:  "release",
			before: "testdata/webhooks/release.json",
			after:  "testdata/webhooks/release.json.golden",
			obj:    new(scm.ReleaseHook),
		},

		// status
		{
			name:   "status",
			event:  "status",
			before: "testdata/webhooks/status.json",
			after:  "testdata/webhooks/status.json.golden",
			obj:    new(scm.StatusHook),
		},

		// label
		{
			name:   "label",
			event:  "label",
			before: "testdata/webhooks/label_deleted.json",
			after:  "testdata/webhooks/label_deleted.json.golden",
			obj:    new(scm.LabelHook),
		},

		// ping
		{
			name:   "ping",
			event:  "ping",
			before: "testdata/webhooks/ping.json",
			after:  "testdata/webhooks/ping.json.golden",
			obj:    new(scm.PingHook),
		},

		// push hooks
		{
			name:   "push",
			event:  "push",
			before: "testdata/webhooks/push.json",
			after:  "testdata/webhooks/push.json.golden",
			obj:    new(scm.PushHook),
		},
		// push tag create hooks
		{
			name:   "push_tag_create",
			event:  "push",
			before: "testdata/webhooks/push_tag.json",
			after:  "testdata/webhooks/push_tag.json.golden",
			obj:    new(scm.PushHook),
		},
		// push tag delete hooks
		{
			name:   "push_tag_delete",
			event:  "push",
			before: "testdata/webhooks/push_tag_delete.json",
			after:  "testdata/webhooks/push_tag_delete.json.golden",
			obj:    new(scm.PushHook),
		},
		// push branch create
		{
			name:   "push_branch_create",
			event:  "push",
			before: "testdata/webhooks/push_branch_create.json",
			after:  "testdata/webhooks/push_branch_create.json.golden",
			obj:    new(scm.PushHook),
		},
		// push branch delete
		{
			name:   "push_branch_delete",
			event:  "push",
			before: "testdata/webhooks/push_branch_delete.json",
			after:  "testdata/webhooks/push_branch_delete.json.golden",
			obj:    new(scm.PushHook),
		},

		//
		// branch events
		//

		// branch create
		{
			name:   "branch_create",
			event:  "create",
			before: "testdata/webhooks/branch_create.json",
			after:  "testdata/webhooks/branch_create.json.golden",
			obj:    new(scm.BranchHook),
		},
		// branch delete
		{
			name:   "branch_delete",
			event:  "delete",
			before: "testdata/webhooks/branch_delete.json",
			after:  "testdata/webhooks/branch_delete.json.golden",
			obj:    new(scm.BranchHook),
		},

		//
		// tag events
		//

		// tag create
		{
			name:   "tag_create",
			event:  "create",
			before: "testdata/webhooks/tag_create.json",
			after:  "testdata/webhooks/tag_create.json.golden",
			obj:    new(scm.TagHook),
		},
		// tag delete
		{
			name:   "tag_delete",
			event:  "delete",
			before: "testdata/webhooks/tag_delete.json",
			after:  "testdata/webhooks/tag_delete.json.golden",
			obj:    new(scm.TagHook),
		},

		//
		// pull request events
		//

		// pull request synced
		{
			name:   "pr_sync",
			event:  "pull_request",
			before: "testdata/webhooks/pr_sync.json",
			after:  "testdata/webhooks/pr_sync.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request opened
		{
			name:   "pr_opened",
			event:  "pull_request",
			before: "testdata/webhooks/pr_opened.json",
			after:  "testdata/webhooks/pr_opened.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request closed
		{
			name:   "pr_closed",
			event:  "pull_request",
			before: "testdata/webhooks/pr_closed.json",
			after:  "testdata/webhooks/pr_closed.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request reopened
		{
			name:   "pr_reopened",
			event:  "pull_request",
			before: "testdata/webhooks/pr_reopened.json",
			after:  "testdata/webhooks/pr_reopened.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request ready for d
		{
			name:   "pr_ready_for_review",
			event:  "pull_request",
			before: "testdata/webhooks/pr_ready_for_review.json",
			after:  "testdata/webhooks/pr_ready_for_review.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request edited
		{
			name:   "pr_edited",
			event:  "pull_request",
			before: "testdata/webhooks/pr_edited.json",
			after:  "testdata/webhooks/pr_edited.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request labeled
		{
			name:   "pr_labeled",
			event:  "pull_request",
			before: "testdata/webhooks/pr_labeled.json",
			after:  "testdata/webhooks/pr_labeled.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request unlabeled
		{
			name:   "pr_unlabeled",
			event:  "pull_request",
			before: "testdata/webhooks/pr_unlabeled.json",
			after:  "testdata/webhooks/pr_unlabeled.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request comment
		{
			name:   "pr_review_comment",
			event:  "pull_request_review_comment",
			before: "testdata/webhooks/pr_comment.json",
			after:  "testdata/webhooks/pr_comment.json.golden",
			obj:    new(scm.PullRequestCommentHook),
		},
		// issue comment
		{
			name:   "issue_comment",
			event:  "issue_comment",
			before: "testdata/webhooks/issue_comment.json",
			after:  "testdata/webhooks/issue_comment.json.golden",
			obj:    new(scm.IssueCommentHook),
		},
		// deployment
		{
			name:   "deployemnt",
			event:  "deployment",
			before: "testdata/webhooks/deployment.json",
			after:  "testdata/webhooks/deployment.json.golden",
			obj:    new(scm.DeployHook),
		},

		// installation of GitHub App
		{
			name:   "installation",
			event:  "installation",
			before: "testdata/webhooks/installation.json",
			after:  "testdata/webhooks/installation.json.golden",
			obj:    new(scm.InstallationHook),
		},
		// delete installation of GitHub App
		{
			name:   "installation_delete",
			event:  "installation",
			before: "testdata/webhooks/installation_delete.json",
			after:  "testdata/webhooks/installation_delete.json.golden",
			obj:    new(scm.InstallationHook),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			before, err := ioutil.ReadFile(test.before)
			if err != nil {
				t.Fatal(err)
			}
			after, err := ioutil.ReadFile(test.after)
			if err != nil {
				t.Fatal(err)
			}

			buf := bytes.NewBuffer(before)
			r, _ := http.NewRequest("GET", "/", buf)
			r.Header.Set("X-GitHub-Event", test.event)
			r.Header.Set("X-Hub-Signature", "sha1=380f462cd2e160b84765144beabdad2e930a7ec5")
			r.Header.Set("X-GitHub-Delivery", "f2467dea-70d6-11e8-8955-3c83993e0aef")

			s := new(webhookService)
			o, err := s.Parse(r, secretFunc)
			if err != nil && err != scm.ErrSignatureInvalid {
				t.Logf("failed to parse webhook for test %s", test.event)
				t.Fatal(err)
			}

			err = json.Unmarshal(after, test.obj)
			if err != nil {
				t.Fatal(err, "failed to unmarshal", test.after, "for test", test.event)
			}

			if diff := cmp.Diff(test.obj, o); diff != "" {
				t.Errorf("Error unmarshaling %s", test.before)
				t.Log(diff)

				// debug only. remove once implemented
				json.NewEncoder(os.Stdout).Encode(o)
			}

			switch event := o.(type) {
			case *scm.PushHook:
				if !strings.HasPrefix(event.Ref, "refs/") {
					t.Errorf("Push hook reference must start with refs/")
				}
			case *scm.BranchHook:
				if strings.HasPrefix(event.Ref.Name, "refs/") {
					t.Errorf("Branch hook reference must not start with refs/")
				}
			case *scm.TagHook:
				if strings.HasPrefix(event.Ref.Name, "refs/") {
					t.Errorf("Branch hook reference must not start with refs/")
				}
			case *scm.InstallationHook:
				assert.NotNil(t, event.Installation, "InstallationHook.Installation")
				assert.NotNil(t, event.GetInstallationRef(), "InstallationHook.GetInstallationRef()")
				assert.NotEmpty(t, event.GetInstallationRef().ID, "InstallationHook.GetInstallationRef().ID")
			}
		})
	}
}

func TestWebhook_ErrUnknownEvent(t *testing.T) {
	f, _ := ioutil.ReadFile("testdata/webhooks/push.json")
	r, _ := http.NewRequest("GET", "/", bytes.NewBuffer(f))
	r.Header.Set("X-GitHub-Delivery", "ee8d97b4-1479-43f1-9cac-fbbd1b80da55")
	r.Header.Set("X-Hub-Signature", "sha1=380f462cd2e160b84765144beabdad2e930a7ec5")

	s := new(webhookService)
	_, err := s.Parse(r, secretFunc)
	if !scm.IsUnknownWebhook(err) {
		t.Errorf("Expect unknown event error, got %v", err)
	}
}

func TestWebhookInvalid(t *testing.T) {
	f, _ := ioutil.ReadFile("testdata/webhooks/push.json")
	r, _ := http.NewRequest("GET", "/", bytes.NewBuffer(f))
	r.Header.Set("X-GitHub-Event", "push")
	r.Header.Set("X-GitHub-Delivery", "ee8d97b4-1479-43f1-9cac-fbbd1b80da55")
	r.Header.Set("X-Hub-Signature", "sha1=380f462cd2e160b84765144beabdad2e930a7ec5")

	s := new(webhookService)
	_, err := s.Parse(r, secretFunc)
	if err != scm.ErrSignatureInvalid {
		t.Errorf("Expect invalid signature error, got %v", err)
	}
}

func TestWebhookValid(t *testing.T) {
	// the sha can be recalculated with the below command
	// openssl dgst -sha1 -hmac <secret> <file>

	f, _ := ioutil.ReadFile("testdata/webhooks/push.json")
	r, _ := http.NewRequest("GET", "/", bytes.NewBuffer(f))
	r.Header.Set("X-GitHub-Event", "push")
	r.Header.Set("X-GitHub-Delivery", "ee8d97b4-1479-43f1-9cac-fbbd1b80da55")
	r.Header.Set("X-Hub-Signature", "sha1=e9c4409d39729236fda483f22e7fb7513e5cd273")

	s := new(webhookService)
	_, err := s.Parse(r, secretFunc)
	if err != nil {
		t.Errorf("Expect valid signature, got %v", err)
	}
}

func secretFunc(scm.Webhook) (string, error) {
	return "topsecret", nil
}
