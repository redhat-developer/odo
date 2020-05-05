// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scm_test

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/driver/github"
)

var ctx context.Context
var db *sql.DB

func ExampleClient() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	// Sets a custom http.Client. This can be used with
	// github.com/golang/oauth2 for authorization.
	client.Client = &http.Client{}
}

func ExampleUser_find() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	user, _, err := client.Users.Find(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(user.Login)
}

func ExampleUser_findLogin() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	user, _, err := client.Users.FindLogin(ctx, "octocat")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(user.Login)
}

func ExampleRepository_find() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	repo, _, err := client.Repositories.Find(ctx, "octocat/Hello-World")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(repo.Namespace, repo.Name)
}

func ExampleRepository_list() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	opts := scm.ListOptions{
		Page: 1,
		Size: 30,
	}

	repos, _, err := client.Repositories.List(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}

	for _, repo := range repos {
		log.Println(repo.Namespace, repo.Name)
	}
}

func ExampleBranch_find() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	branch, _, err := client.Git.FindBranch(ctx, "octocat/Hello-World", "master")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(branch.Name, branch.Sha)
}

func ExampleBranch_list() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	opts := scm.ListOptions{
		Page: 1,
		Size: 30,
	}

	branches, _, err := client.Git.ListBranches(ctx, "octocat/Hello-World", opts)
	if err != nil {
		log.Fatal(err)
	}

	for _, branch := range branches {
		log.Println(branch.Name, branch.Sha)
	}
}

func ExampleTag_find() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	tag, _, err := client.Git.FindTag(ctx, "octocat/Hello-World", "v1.0.0")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(tag.Name, tag.Sha)
}

func ExampleTag_list() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	opts := scm.ListOptions{
		Page: 1,
		Size: 30,
	}

	tags, _, err := client.Git.ListTags(ctx, "octocat/Hello-World", opts)
	if err != nil {
		log.Fatal(err)
	}

	for _, tag := range tags {
		log.Println(tag.Name, tag.Sha)
	}
}

func ExampleCommit_find() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	commit, _, err := client.Git.FindCommit(ctx, "octocat/Hello-World", "6dcb09b5b57875f334f61aebed695e2e4193db5e")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(commit.Sha, commit.Message, commit.Author.Login)
}

func ExampleCommit_list() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	opts := scm.CommitListOptions{
		Ref:  "master",
		Page: 1,
		Size: 30,
	}

	commits, _, err := client.Git.ListCommits(ctx, "octocat/Hello-World", opts)
	if err != nil {
		log.Fatal(err)
	}

	for _, commit := range commits {
		log.Println(commit.Sha, commit.Message, commit.Author.Login)
	}
}

func ExampleCommit_changes() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	opts := scm.ListOptions{
		Page: 1,
		Size: 30,
	}

	changes, _, err := client.Git.ListChanges(ctx, "octocat/Hello-World", "6dcb09b5b57875f334f61aebed695e2e4193db5e", opts)
	if err != nil {
		log.Fatal(err)
	}

	for _, change := range changes {
		log.Println(change.Path, change.Added, change.Deleted, change.Renamed)
	}
}

func ExampleContent_find() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	content, _, err := client.Contents.Find(ctx, "octocat/Hello-World", "README", "6dcb09b5b57875f334f61aebed695e2e4193db5e")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(content.Path, content.Data)
}

func ExampleHook_list() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	opts := scm.ListOptions{
		Page: 1,
		Size: 30,
	}

	hooks, _, err := client.Repositories.ListHooks(ctx, "octocat/Hello-World", opts)
	if err != nil {
		log.Fatal(err)
	}

	for _, hook := range hooks {
		log.Println(hook.ID, hook.Target, hook.Events)
	}
}

func ExampleHook_find() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	hook, _, err := client.Repositories.FindHook(ctx, "octocat/Hello-World", "1")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(hook.ID, hook.Target, hook.Events)
}

func ExampleHook_create() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	input := &scm.HookInput{
		Name:       "CI",
		Target:     "https://ci.example.com",
		Secret:     "topsecret",
		SkipVerify: false,
		Events: scm.HookEvents{
			Branch:             true,
			Issue:              false,
			IssueComment:       false,
			PullRequest:        true,
			PullRequestComment: false,
			Push:               true,
			ReviewComment:      false,
			Tag:                true,
		},
	}

	_, _, err = client.Repositories.CreateHook(ctx, "octocat/Hello-World", input)
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleStatus_list() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	opts := scm.ListOptions{
		Page: 1,
		Size: 30,
	}

	statuses, _, err := client.Repositories.ListStatus(ctx, "octocat/Hello-World", "6dcb09b5b57875f334f61aebed695e2e4193db5e", opts)
	if err != nil {
		log.Fatal(err)
	}

	for _, status := range statuses {
		log.Println(status.State, status.Target)
	}
}

func ExampleStatus_create() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	input := &scm.StatusInput{
		State:  scm.StateSuccess,
		Label:  "continuous-integation",
		Desc:   "Build has completed successfully",
		Target: "https://ci.example.com/octocat/hello-world/1",
	}

	_, _, err = client.Repositories.CreateStatus(ctx, "octocat/Hello-World", "6dcb09b5b57875f334f61aebed695e2e4193db5e", input)
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleIssue_list() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	opts := scm.IssueListOptions{
		Page:   1,
		Size:   30,
		Open:   true,
		Closed: false,
	}

	issues, _, err := client.Issues.List(ctx, "octocat/Hello-World", opts)
	if err != nil {
		log.Fatal(err)
	}

	for _, issue := range issues {
		log.Println(issue.Number, issue.Title)
	}
}

func ExampleIssue_find() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	issue, _, err := client.Issues.Find(ctx, "octocat/Hello-World", 1)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(issue.Number, issue.Title)
}

func ExampleIssue_close() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.Issues.Close(ctx, "octocat/Hello-World", 1)
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleIssue_lock() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.Issues.Lock(ctx, "octocat/Hello-World", 1)
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleIssue_unlock() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.Issues.Lock(ctx, "octocat/Hello-World", 1)
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleOrganization_list() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	opts := scm.ListOptions{
		Page: 1,
		Size: 30,
	}

	orgs, _, err := client.Organizations.List(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}

	for _, org := range orgs {
		log.Println(org.Name)
	}
}

func ExampleOrganization_find() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	org, _, err := client.Organizations.Find(ctx, "github")
	if err != nil {
		log.Fatal(err)
	}

	log.Println(org.Name)
}

func ExampleComment_list() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	opts := scm.ListOptions{
		Page: 1,
		Size: 30,
	}

	comments, _, err := client.Issues.ListComments(ctx, "octocat/Hello-World", 1, opts)
	if err != nil {
		log.Fatal(err)
	}

	for _, comment := range comments {
		log.Println(comment.Body)
	}
}

func ExampleComment_create() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	in := &scm.CommentInput{
		Body: "Found a bug",
	}

	comment, _, err := client.Issues.CreateComment(ctx, "octocat/Hello-World", 1, in)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(comment.ID)
}

func ExampleComment_find() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	comment, _, err := client.Issues.FindComment(ctx, "octocat/Hello-World", 1, 1)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(comment.Body)
}

func ExampleReview_list() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	opts := scm.ListOptions{
		Page: 1,
		Size: 30,
	}

	reviews, _, err := client.Reviews.List(ctx, "octocat/Hello-World", 1, opts)
	if err != nil {
		log.Fatal(err)
	}

	for _, review := range reviews {
		log.Println(
			review.Path,
			review.Line,
			review.Body,
		)
	}
}

func ExampleReview_find() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	review, _, err := client.Reviews.Find(ctx, "octocat/Hello-World", 1, 1)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(
		review.Path,
		review.Line,
		review.Body,
	)
}

func ExampleReview_create() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	in := &scm.ReviewInput{
		Line: 38,
		Path: "main.go",
		Body: "Run gofmt please",
	}

	review, _, err := client.Reviews.Create(ctx, "octocat/Hello-World", 1, in)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(review.ID)
}

func ExamplePullRequest_list() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	opts := scm.PullRequestListOptions{
		Page:   1,
		Size:   30,
		Open:   true,
		Closed: false,
	}

	prs, _, err := client.PullRequests.List(ctx, "octocat/Hello-World", opts)
	if err != nil {
		log.Fatal(err)
	}

	for _, pr := range prs {
		log.Println(pr.Number, pr.Title)
	}
}

func ExamplePullRequest_find() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	pr, _, err := client.PullRequests.Find(ctx, "octocat/Hello-World", 1)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(pr.Number, pr.Title)
}

func ExamplePullRequest_changes() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	opts := scm.ListOptions{
		Page: 1,
		Size: 30,
	}

	changes, _, err := client.PullRequests.ListChanges(ctx, "octocat/Hello-World", 1, opts)
	if err != nil {
		log.Fatal(err)
	}

	for _, change := range changes {
		log.Println(change.Path, change.Added, change.Deleted, change.Renamed)
	}
}

func ExamplePullRequest_close() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.PullRequests.Close(ctx, "octocat/Hello-World", 1)
	if err != nil {
		log.Fatal(err)
	}
}

func ExamplePullRequest_merge() {
	client, err := github.New("https://api.github.com")
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.PullRequests.Merge(ctx, "octocat/Hello-World", 1, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleWebhook() {
	client := github.NewDefault()

	secret := func(webhook scm.Webhook) (string, error) {
		return "topsecret", nil
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		webhook, err := client.Webhooks.Parse(r, secret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		switch event := webhook.(type) {
		case *scm.PushHook:
			log.Println(
				event.Ref,
				event.Commit.Sha,
				event.Commit.Message,
				event.Repo.Namespace,
				event.Repo.Name,
				event.Sender.Login,
			)
		case *scm.BranchHook:
		case *scm.TagHook:
		case *scm.IssueHook:
		case *scm.IssueCommentHook:
		case *scm.PullRequestHook:
		case *scm.PullRequestCommentHook:
		case *scm.ReviewCommentHook:
		}
	}

	http.HandleFunc("/hook", handler)
	http.ListenAndServe(":8000", nil)
}

func ExampleWebhook_lookupSecret() {
	client := github.NewDefault()

	secret := func(webhook scm.Webhook) (secret string, err error) {
		stmt := "SELECT secret FROM repos WHERE id = ?"
		repo := webhook.Repository()
		err = db.QueryRow(stmt, repo.ID).Scan(&secret)
		return
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		webhook, err := client.Webhooks.Parse(r, secret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		switch event := webhook.(type) {
		case *scm.PushHook:
			log.Println(
				event.Ref,
				event.Commit.Sha,
				event.Commit.Message,
				event.Repo.Namespace,
				event.Repo.Name,
				event.Sender.Login,
			)
		}
	}

	http.HandleFunc("/hook", handler)
	http.ListenAndServe(":8000", nil)
}
