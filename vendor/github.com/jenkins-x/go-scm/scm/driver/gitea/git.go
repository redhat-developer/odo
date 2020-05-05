// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gitea

import (
	"context"
	"fmt"
	"time"

	"github.com/jenkins-x/go-scm/scm"
)

type gitService struct {
	client *wrapper
}

func (s *gitService) FindRef(ctx context.Context, repo, ref string) (string, *scm.Response, error) {
	return "", nil, scm.ErrNotSupported
}

func (s *gitService) DeleteRef(ctx context.Context, repo, ref string) (*scm.Response, error) {
	return nil, scm.ErrNotSupported
}

func (s *gitService) FindBranch(ctx context.Context, repo, name string) (*scm.Reference, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/repos/%s/branches/%s", repo, name)
	out := new(branch)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertBranch(out), res, err
}

func (s *gitService) FindCommit(ctx context.Context, repo, ref string) (*scm.Commit, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/repos/%s/git/commits/%s", repo, ref)
	out := new(commitInfo)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertCommitInfo(out), res, err
}

func (s *gitService) FindTag(ctx context.Context, repo, name string) (*scm.Reference, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *gitService) ListBranches(ctx context.Context, repo string, _ scm.ListOptions) ([]*scm.Reference, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/repos/%s/branches", repo)
	out := []*branch{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertBranchList(out), res, err
}

func (s *gitService) ListCommits(ctx context.Context, repo string, _ scm.CommitListOptions) ([]*scm.Commit, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *gitService) ListTags(ctx context.Context, repo string, _ scm.ListOptions) ([]*scm.Reference, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *gitService) ListChanges(ctx context.Context, repo, ref string, _ scm.ListOptions) ([]*scm.Change, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

//
// native data structures
//

type (
	// gitea branch object.
	branch struct {
		Name   string `json:"name"`
		Commit commit `json:"commit"`
	}

	// gitea commit object.
	commit struct {
		ID        string    `json:"id"`
		Sha       string    `json:"sha"`
		Message   string    `json:"message"`
		URL       string    `json:"url"`
		Author    signature `json:"author"`
		Committer signature `json:"committer"`
		Timestamp time.Time `json:"timestamp"`
	}

	// gitea commit info object.
	commitInfo struct {
		Sha       string `json:"sha"`
		Commit    commit `json:"commit"`
		Author    user   `json:"author"`
		Committer user   `json:"committer"`
	}

	// gitea signature object.
	signature struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Username string `json:"username"`
	}
)

//
// native data structure conversion
//

func convertBranchList(src []*branch) []*scm.Reference {
	dst := []*scm.Reference{}
	for _, v := range src {
		dst = append(dst, convertBranch(v))
	}
	return dst
}

func convertBranch(src *branch) *scm.Reference {
	return &scm.Reference{
		Name: scm.TrimRef(src.Name),
		Path: scm.ExpandRef(src.Name, "refs/heads/"),
		Sha:  src.Commit.ID,
	}
}

// func convertCommitList(src []*commit) []*scm.Commit {
// 	dst := []*scm.Commit{}
// 	for _, v := range src {
// 		dst = append(dst, convertCommitInfo(v))
// 	}
// 	return dst
// }

func convertCommitInfo(src *commitInfo) *scm.Commit {
	return &scm.Commit{
		Sha:       src.Sha,
		Link:      src.Commit.URL,
		Message:   src.Commit.Message,
		Author:    convertUserSignature(src.Author),
		Committer: convertUserSignature(src.Committer),
	}
}

func convertSignature(src signature) scm.Signature {
	return scm.Signature{
		Login: src.Username,
		Email: src.Email,
		Name:  src.Name,
	}
}

func convertUserSignature(src user) scm.Signature {
	return scm.Signature{
		Login:  userLogin(&src),
		Email:  src.Email,
		Name:   src.Fullname,
		Avatar: src.Avatar,
	}
}
