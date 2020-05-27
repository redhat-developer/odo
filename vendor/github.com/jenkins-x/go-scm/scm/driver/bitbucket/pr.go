// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitbucket

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jenkins-x/go-scm/scm"
)

type pullService struct {
	*issueService
}

const debugDump = false

func (s *pullService) Find(ctx context.Context, repo string, number int) (*scm.PullRequest, *scm.Response, error) {
	path := fmt.Sprintf("2.0/repositories/%s/pullrequests/%d", repo, number)
	out := new(pullRequest)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertPullRequest(out), res, err
}

func (s *pullService) List(ctx context.Context, repo string, opts scm.PullRequestListOptions) ([]*scm.PullRequest, *scm.Response, error) {
	path := fmt.Sprintf("2.0/repositories/%s/pullrequests?%s", repo, encodePullRequestListOptions(opts))
	out := new(pullRequests)
	if debugDump {
		var buf bytes.Buffer
		res, err := s.client.do(ctx, "GET", path, nil, &buf)
		fmt.Printf("%s\n", buf.String())
		return nil, res, err
	}
	res, err := s.client.do(ctx, "GET", path, nil, out)
	copyPagination(out.pagination, res)
	return convertPullRequests(out), res, err
}

func (s *pullService) ListChanges(ctx context.Context, repo string, number int, opts scm.ListOptions) ([]*scm.Change, *scm.Response, error) {
	path := fmt.Sprintf("2.0/repositories/%s/pullrequests/%d/diffstat?%s", repo, number, encodeListOptions(opts))
	out := new(diffstats)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	copyPagination(out.pagination, res)
	return convertDiffstats(out), res, err
}

func (s *pullService) ListLabels(context.Context, string, int, scm.ListOptions) ([]*scm.Label, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *pullService) ListEvents(context.Context, string, int, scm.ListOptions) ([]*scm.ListedIssueEvent, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *pullService) DeleteLabel(ctx context.Context, repo string, number int, label string) (*scm.Response, error) {
	return nil, scm.ErrNotSupported
}

func (s *pullService) Merge(ctx context.Context, repo string, number int, options *scm.PullRequestMergeOptions) (*scm.Response, error) {
	path := fmt.Sprintf("2.0/repositories/%s/pullrequests/%d/merge", repo, number)
	res, err := s.client.do(ctx, "POST", path, nil, nil)
	return res, err
}

func (s *pullService) Update(ctx context.Context, repo string, number int, prInput *scm.PullRequestInput) (*scm.PullRequest, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *pullService) Close(ctx context.Context, repo string, number int) (*scm.Response, error) {
	return nil, scm.ErrNotSupported
}

func (s *pullService) Reopen(ctx context.Context, repo string, number int) (*scm.Response, error) {
	return nil, scm.ErrNotSupported
}

func (s *pullService) AssignIssue(ctx context.Context, repo string, number int, logins []string) (*scm.Response, error) {
	return nil, scm.ErrNotSupported
}

func (s *pullService) UnassignIssue(ctx context.Context, repo string, number int, logins []string) (*scm.Response, error) {
	return nil, scm.ErrNotSupported
}

func (s *pullService) Create(ctx context.Context, repo string, input *scm.PullRequestInput) (*scm.PullRequest, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *pullService) RequestReview(ctx context.Context, repo string, number int, logins []string) (*scm.Response, error) {
	return nil, scm.ErrNotSupported
}

func (s *pullService) UnrequestReview(ctx context.Context, repo string, number int, logins []string) (*scm.Response, error) {
	return nil, scm.ErrNotSupported
}

type prCommit struct {
	LatestCommit string `json:"hash"`
}

type prSource struct {
	Commit struct {
		Type   string `json:"type"`
		Ref    string `json:"ref"`
		Commit string `json:"hash"`
	} `json:"commit"`
	Repository repository `json:"repository"`
	Branch     struct {
		Name string `json:"name"`
	} `json:"branch"`
}

type prDestination struct {
	Commit struct {
		Type   string `json:"type"`
		Ref    string `json:"ref"`
		Commit string `json:"Commit"`
	} `json:"commit"`
	Repository repository `json:"repository"`
	Branch     struct {
		Name string `json:"name"`
	} `json:"branch"`
}

type pullRequest struct {
	ID int `json:"id"`
	//Version     int    `json:"version"`
	Title        string        `json:"title"`
	Description  string        `json:"description"`
	State        string        `json:"state"`
	CreatedDate  time.Time     `json:"created_on"`
	UpdatedDate  time.Time     `json:"updated_on"`
	Source       prSource      `json:"source"`
	Destination  prDestination `json:"destination"`
	Locked       bool          `json:"locked"`
	Author       user          `json:"author"`
	Reviewers    []user        `json:"reviewers"`
	Participants []user        `json:"participants"`
	Links        struct {
		Self link `json:"self"`
		HTML link `json:"html"`
	} `json:"links"`
}

type pullRequests struct {
	pagination
	Values []*pullRequest `json:"values"`
}

func convertPullRequest(from *pullRequest) *scm.PullRequest {
	// TODO
	fork := "false"
	closed := strings.ToLower(from.State) != "open"
	return &scm.PullRequest{
		Number:  from.ID,
		Title:   from.Title,
		Body:    from.Description,
		Sha:     from.Source.Commit.Commit,
		Ref:     fmt.Sprintf("refs/pull-requests/%d/from", from.ID),
		Source:  from.Source.Commit.Commit,
		Target:  from.Destination.Commit.Commit,
		Fork:    fork,
		Base:    convertPullRequestBranch(from.Destination.Commit.Ref, from.Destination.Commit.Commit, from.Destination.Repository),
		Head:    convertPullRequestBranch(from.Source.Commit.Ref, from.Source.Commit.Commit, from.Source.Repository),
		Link:    from.Links.HTML.Href,
		State:   strings.ToLower(from.State),
		Closed:  closed,
		Merged:  from.State == "MERGED",
		Created: from.CreatedDate,
		Updated: from.UpdatedDate,
		Author: scm.User{
			Login:  from.Author.GetLogin(),
			Name:   from.Author.DisplayName,
			Email:  from.Author.EmailAddress,
			Link:   from.Author.Links.Self.Href,
			Avatar: from.Author.Links.Avatar.Href,
		},
	}
}

func convertPullRequestBranch(ref string, sha string, repo repository) scm.PullRequestBranch {
	return scm.PullRequestBranch{
		Ref:  ref,
		Sha:  sha,
		Repo: *convertRepository(&repo),
	}
}

func convertPullRequests(from *pullRequests) []*scm.PullRequest {
	answer := []*scm.PullRequest{}
	for _, pr := range from.Values {
		answer = append(answer, convertPullRequest(pr))
	}
	return answer
}
