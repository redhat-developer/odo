// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gitlab

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/jenkins-x/go-scm/scm"

	"github.com/google/go-cmp/cmp"
)

func TestWebhooks(t *testing.T) {
	tests := []struct {
		event  string
		before string
		after  string
		obj    interface{}
	}{
		// branch hooks
		{
			event:  "Push Hook",
			before: "testdata/webhooks/branch_create.json",
			after:  "testdata/webhooks/branch_create.json.golden",
			obj:    new(scm.PushHook),
		},
		{
			event:  "Push Hook",
			before: "testdata/webhooks/branch_delete.json",
			after:  "testdata/webhooks/branch_delete.json.golden",
			obj:    new(scm.BranchHook),
		},
		// tag hooks
		{
			event:  "Tag Push Hook",
			before: "testdata/webhooks/tag_create.json",
			after:  "testdata/webhooks/tag_create.json.golden",
			obj:    new(scm.PushHook),
		},
		{
			event:  "Push Hook",
			before: "testdata/webhooks/tag_delete.json",
			after:  "testdata/webhooks/tag_delete.json.golden",
			obj:    new(scm.TagHook),
		},
		// push hooks
		{
			event:  "Push Hook",
			before: "testdata/webhooks/push.json",
			after:  "testdata/webhooks/push.json.golden",
			obj:    new(scm.PushHook),
		},
		// // issue hooks
		// {
		// 	event:  "issues",
		// 	before: "testdata/webhooks/issues_opened.json",
		// 	after:  "testdata/webhooks/issues_opened.json.golden",
		// 	obj:    new(scm.IssueHook),
		// },
		// // issue comment hooks
		// {
		// 	event:  "issue_comment",
		// 	before: "testdata/webhooks/issue_comment_created.json",
		// 	after:  "testdata/webhooks/issue_comment_created.json.golden",
		// 	obj:    new(scm.IssueCommentHook),
		// },
		// pull request hooks
		{
			event:  "Merge Request Hook",
			before: "testdata/webhooks/pull_request_create.json",
			after:  "testdata/webhooks/pull_request_create.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// {
		// 	event:  "Merge Request Hook",
		// 	before: "testdata/webhooks/pull_request_edited.json",
		// 	after:  "testdata/webhooks/pull_request_edited.json.golden",
		// 	obj:    new(scm.PullRequestHook),
		// },
		// {
		// 	event:  "Merge Request Hook",
		// 	before: "testdata/webhooks/pull_request_synchronized.json",
		// 	after:  "testdata/webhooks/pull_request_synchronized.json.golden",
		// 	obj:    new(scm.PullRequestHook),
		// },
		{
			event:  "Merge Request Hook",
			before: "testdata/webhooks/pull_request_close.json",
			after:  "testdata/webhooks/pull_request_close.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		{
			event:  "Merge Request Hook",
			before: "testdata/webhooks/pull_request_reopen.json",
			after:  "testdata/webhooks/pull_request_reopen.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		{
			event:  "Merge Request Hook",
			before: "testdata/webhooks/pull_request_merge.json",
			after:  "testdata/webhooks/pull_request_merge.json.golden",
			obj:    new(scm.PullRequestHook),
		},
		// pull request comment hooks
		{
			event:  "Note Hook",
			before: "testdata/webhooks/pull_request_comment_create.json",
			after:  "testdata/webhooks/pull_request_comment_create.json.golden",
			obj:    new(scm.PullRequestCommentHook),
		},
	}

	for _, test := range tests {
		t.Run(test.before, func(t *testing.T) {
			before, err := ioutil.ReadFile(test.before)
			if err != nil {
				t.Error(err)
				return
			}
			after, err := ioutil.ReadFile(test.after)
			if err != nil {
				t.Error(err)
				return
			}

			buf := bytes.NewBuffer(before)
			r, _ := http.NewRequest("GET", "/", buf)
			r.Header.Set("X-Gitlab-Event", test.event)
			r.Header.Set("X-Gitlab-Token", "9edf3260d727e29d906bdb10c8a099a")
			r.Header.Set("X-Request-Id", "ee8d97b4-1479-43f1-9cac-fbbd1b80da55")

			s := new(webhookService)
			o, err := s.Parse(r, secretFunc)
			if err != nil && err != scm.ErrSignatureInvalid {
				t.Error(err)
				return
			}

			err = json.Unmarshal(after, &test.obj)
			if err != nil {
				t.Error(err)
				return
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
			}
		})
	}
}

func TestWebhook_SignatureValid(t *testing.T) {
	f, _ := ioutil.ReadFile("testdata/webhooks/branch_delete.json")
	r, _ := http.NewRequest("GET", "/", bytes.NewBuffer(f))
	r.Header.Set("X-Gitlab-Event", "Push Hook")
	r.Header.Set("X-Gitlab-Token", "topsecret")
	r.Header.Set("X-Request-Id", "ee8d97b4-1479-43f1-9cac-fbbd1b80da55")

	s := new(webhookService)
	_, err := s.Parse(r, secretFunc)
	if err != nil {
		t.Error(err)
	}
}

func TestWebhook_SignatureInvalid(t *testing.T) {
	f, _ := ioutil.ReadFile("testdata/webhooks/branch_delete.json")
	r, _ := http.NewRequest("GET", "/", bytes.NewBuffer(f))
	r.Header.Set("X-Gitlab-Event", "Push Hook")
	r.Header.Set("X-Gitlab-Token", "void")
	r.Header.Set("X-Request-Id", "ee8d97b4-1479-43f1-9cac-fbbd1b80da55")

	s := new(webhookService)
	_, err := s.Parse(r, secretFunc)
	if err != scm.ErrSignatureInvalid {
		t.Errorf("Expect invalid signature error, got %v", err)
	}
}

func TestWebhook_SignatureMissing(t *testing.T) {
	f, _ := ioutil.ReadFile("testdata/webhooks/branch_delete.json")
	r, _ := http.NewRequest("GET", "/", bytes.NewBuffer(f))
	r.Header.Set("X-Gitlab-Event", "Push Hook")
	r.Header.Set("X-Gitlab-Token", "")
	r.Header.Set("X-Request-Id", "ee8d97b4-1479-43f1-9cac-fbbd1b80da55")

	s := new(webhookService)
	_, err := s.Parse(r, secretFunc)
	if err != scm.ErrSignatureInvalid {
		t.Errorf("Expect invalid signature error, got %v", err)
	}
}

func secretFunc(scm.Webhook) (string, error) {
	return "topsecret", nil
}
