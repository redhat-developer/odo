// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gitea

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/jenkins-x/go-scm/scm"
)

type repositoryService struct {
	client *wrapper
}

func (s *repositoryService) Create(context.Context, *scm.RepositoryInput) (*scm.Repository, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *repositoryService) FindCombinedStatus(ctx context.Context, repo, ref string) (*scm.CombinedStatus, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *repositoryService) FindUserPermission(ctx context.Context, repo string, user string) (string, *scm.Response, error) {
	return "", nil, scm.ErrNotSupported
}

func (s *repositoryService) IsCollaborator(ctx context.Context, repo, user string) (bool, *scm.Response, error) {
	return false, nil, scm.ErrNotSupported
}

func (s *repositoryService) ListCollaborators(ctx context.Context, repo string, ops scm.ListOptions) ([]scm.User, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *repositoryService) ListLabels(context.Context, string, scm.ListOptions) ([]*scm.Label, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *repositoryService) Find(ctx context.Context, repo string) (*scm.Repository, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/repos/%s", repo)
	out := new(repository)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertRepository(out), res, err
}

func (s *repositoryService) FindHook(ctx context.Context, repo string, id string) (*scm.Hook, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/repos/%s/hooks/%s", repo, id)
	out := new(hook)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertHook(out), res, err
}

func (s *repositoryService) FindPerms(ctx context.Context, repo string) (*scm.Perm, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/repos/%s", repo)
	out := new(repository)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertRepository(out).Perm, res, err
}

func (s *repositoryService) List(ctx context.Context, _ scm.ListOptions) ([]*scm.Repository, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/user/repos")
	out := []*repository{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertRepositoryList(out), res, err
}

func (s *repositoryService) ListOrganisation(ctx context.Context, org string, opts scm.ListOptions) ([]*scm.Repository, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/orgs/%s/repos", org)
	out := []*repository{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertRepositoryList(out), res, err
}

func (s *repositoryService) ListUser(ctx context.Context, username string, opts scm.ListOptions) ([]*scm.Repository, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/users/%s/repos", username)
	out := []*repository{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertRepositoryList(out), res, err
}

func (s *repositoryService) ListHooks(ctx context.Context, repo string, _ scm.ListOptions) ([]*scm.Hook, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/repos/%s/hooks", repo)
	out := []*hook{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertHookList(out), res, err
}

func (s *repositoryService) ListStatus(ctx context.Context, repo string, ref string, _ scm.ListOptions) ([]*scm.Status, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/repos/%s/statuses/%s", repo, ref)
	out := []*status{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertStatusList(out), res, err
}

func (s *repositoryService) CreateHook(ctx context.Context, repo string, input *scm.HookInput) (*scm.Hook, *scm.Response, error) {
	target, err := url.Parse(input.Target)
	if err != nil {
		return nil, nil, err
	}
	params := target.Query()
	params.Set("secret", input.Secret)
	target.RawQuery = params.Encode()

	path := fmt.Sprintf("api/v1/repos/%s/hooks", repo)
	in := new(hook)
	in.Type = "gitea"
	in.Active = true
	in.Config.Secret = input.Secret
	in.Config.ContentType = "json"
	in.Config.URL = target.String()
	in.Events = append(
		input.NativeEvents,
		convertHookEvent(input.Events)...,
	)
	out := new(hook)
	res, err := s.client.do(ctx, "POST", path, in, out)
	return convertHook(out), res, err
}

func (s *repositoryService) CreateStatus(ctx context.Context, repo string, ref string, input *scm.StatusInput) (*scm.Status, *scm.Response, error) {
	path := fmt.Sprintf("api/v1/repos/%s/statuses/%s", repo, ref)
	in := &statusInput{
		State:       convertFromState(input.State),
		Context:     input.Label,
		Description: input.Desc,
		TargetURL:   input.Target,
	}
	out := new(status)
	res, err := s.client.do(ctx, "POST", path, in, out)
	return convertStatus(out), res, err
}

func (s *repositoryService) DeleteHook(ctx context.Context, repo string, id string) (*scm.Response, error) {
	path := fmt.Sprintf("api/v1/repos/%s/hooks/%s", repo, id)
	return s.client.do(ctx, "DELETE", path, nil, nil)
}

//
// native data structures
//

type (
	// gitea repository resource.
	repository struct {
		ID            int       `json:"id"`
		Owner         user      `json:"owner"`
		Name          string    `json:"name"`
		FullName      string    `json:"full_name"`
		Private       bool      `json:"private"`
		Fork          bool      `json:"fork"`
		HTMLURL       string    `json:"html_url"`
		SSHURL        string    `json:"ssh_url"`
		CloneURL      string    `json:"clone_url"`
		DefaultBranch string    `json:"default_branch"`
		CreatedAt     time.Time `json:"created_at"`
		UpdatedAt     time.Time `json:"updated_at"`
		Permissions   perm      `json:"permissions"`
	}

	// gitea permissions details.
	perm struct {
		Admin bool `json:"admin"`
		Push  bool `json:"push"`
		Pull  bool `json:"pull"`
	}

	// gitea hook resource.
	hook struct {
		ID     int        `json:"id"`
		Type   string     `json:"type"`
		Events []string   `json:"events"`
		Active bool       `json:"active"`
		Config hookConfig `json:"config"`
	}

	// gitea hook configuration details.
	hookConfig struct {
		URL         string `json:"url"`
		ContentType string `json:"content_type"`
		Secret      string `json:"secret"`
	}

	// gitea status resource.
	status struct {
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		State       string    `json:"status"`
		TargetURL   string    `json:"target_url"`
		Description string    `json:"description"`
		Context     string    `json:"context"`
	}

	// gitea status creation request.
	statusInput struct {
		State       string `json:"state"`
		TargetURL   string `json:"target_url"`
		Description string `json:"description"`
		Context     string `json:"context"`
	}
)

//
// native data structure conversion
//

func convertRepositoryList(src []*repository) []*scm.Repository {
	var dst []*scm.Repository
	for _, v := range src {
		dst = append(dst, convertRepository(v))
	}
	return dst
}

func convertRepository(src *repository) *scm.Repository {
	return &scm.Repository{
		ID:        strconv.Itoa(src.ID),
		Namespace: userLogin(&src.Owner),
		Name:      src.Name,
		Perm:      convertPerm(src.Permissions),
		Branch:    src.DefaultBranch,
		Private:   src.Private,
		Clone:     src.CloneURL,
		CloneSSH:  src.SSHURL,
	}
}

func convertPerm(src perm) *scm.Perm {
	return &scm.Perm{
		Push:  src.Push,
		Pull:  src.Pull,
		Admin: src.Admin,
	}
}

func convertHookList(src []*hook) []*scm.Hook {
	var dst []*scm.Hook
	for _, v := range src {
		dst = append(dst, convertHook(v))
	}
	return dst
}

func convertHook(from *hook) *scm.Hook {
	return &scm.Hook{
		ID:     strconv.Itoa(from.ID),
		Active: from.Active,
		Target: from.Config.URL,
		Events: from.Events,
	}
}

func convertHookEvent(from scm.HookEvents) []string {
	var events []string
	if from.PullRequest {
		events = append(events, "pull_request")
	}
	if from.Issue {
		events = append(events, "issues")
	}
	if from.IssueComment || from.PullRequestComment {
		events = append(events, "issue_comment")
	}
	if from.Branch || from.Tag {
		events = append(events, "create")
		events = append(events, "delete")
	}
	if from.Push {
		events = append(events, "push")
	}
	return events
}

func convertStatusList(src []*status) []*scm.Status {
	var dst []*scm.Status
	for _, v := range src {
		dst = append(dst, convertStatus(v))
	}
	return dst
}

func convertStatus(from *status) *scm.Status {
	return &scm.Status{
		State:  convertState(from.State),
		Label:  from.Context,
		Desc:   from.Description,
		Target: from.TargetURL,
	}
}

func convertState(from string) scm.State {
	switch from {
	case "error":
		return scm.StateError
	case "failure":
		return scm.StateFailure
	case "pending":
		return scm.StatePending
	case "success":
		return scm.StateSuccess
	default:
		return scm.StateUnknown
	}
}

func convertFromState(from scm.State) string {
	switch from {
	case scm.StatePending, scm.StateRunning:
		return "pending"
	case scm.StateSuccess:
		return "success"
	case scm.StateFailure:
		return "failure"
	default:
		return "error"
	}
}
