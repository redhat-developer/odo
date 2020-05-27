// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitbucket

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/jenkins-x/go-scm/scm"
)

type repository struct {
	UUID       string    `json:"uuid"`
	SCM        string    `json:"scm"`
	FullName   string    `json:"full_name"`
	IsPrivate  bool      `json:"is_private"`
	CreatedOn  time.Time `json:"created_on"`
	UpdatedOn  time.Time `json:"updated_on"`
	Mainbranch struct {
		Type string `json:"type"`
		Name string `json:"name"`
	} `json:"mainbranch"`
}

type perms struct {
	Values []*perm `json:"values"`
}

type perm struct {
	Permissions string `json:"permission"`
}

type hooks struct {
	pagination
	Values []*hook `json:"values"`
}

type hook struct {
	Description          string   `json:"description"`
	URL                  string   `json:"url"`
	SkipCertVerification bool     `json:"skip_cert_verification"`
	Active               bool     `json:"active"`
	Events               []string `json:"events"`
	UUID                 string   `json:"uuid"`
}

type hookInput struct {
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Active      bool     `json:"active"`
	Events      []string `json:"events"`
}

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

// Find returns the repository by name.
func (s *repositoryService) Find(ctx context.Context, repo string) (*scm.Repository, *scm.Response, error) {
	path := fmt.Sprintf("2.0/repositories/%s", repo)
	out := new(repository)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertRepository(out), res, err
}

// FindHook returns a repository hook.
func (s *repositoryService) FindHook(ctx context.Context, repo string, id string) (*scm.Hook, *scm.Response, error) {
	path := fmt.Sprintf("2.0/repositories/%s/hooks/%s", repo, id)
	out := new(hook)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertHook(out), res, err
}

// FindPerms returns the repository permissions.
func (s *repositoryService) FindPerms(ctx context.Context, repo string) (*scm.Perm, *scm.Response, error) {
	path := fmt.Sprintf("2.0/user/permissions/repositories?q=repository.full_name=%q", repo)
	out := new(perms)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertPerms(out), res, err
}

// List returns the user repository list.
func (s *repositoryService) List(ctx context.Context, opts scm.ListOptions) ([]*scm.Repository, *scm.Response, error) {
	path := fmt.Sprintf("2.0/repositories?%s", encodeListRoleOptions(opts))
	if opts.URL != "" {
		path = opts.URL
	}
	out := new(repositories)
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	copyPagination(out.pagination, res)
	return convertRepositoryList(out), res, err
}

func (s *repositoryService) ListOrganisation(context.Context, string, scm.ListOptions) ([]*scm.Repository, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *repositoryService) ListUser(context.Context, string, scm.ListOptions) ([]*scm.Repository, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

// ListHooks returns a list or repository hooks.
func (s *repositoryService) ListHooks(ctx context.Context, repo string, opts scm.ListOptions) ([]*scm.Hook, *scm.Response, error) {
	path := fmt.Sprintf("2.0/repositories/%s/hooks?%s", repo, encodeListOptions(opts))
	out := new(hooks)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	copyPagination(out.pagination, res)
	return convertHookList(out), res, err
}

// ListStatus returns a list of commit statuses.
func (s *repositoryService) ListStatus(ctx context.Context, repo, ref string, opts scm.ListOptions) ([]*scm.Status, *scm.Response, error) {
	path := fmt.Sprintf("2.0/repositories/%s/commit/%s/statuses?%s", repo, ref, encodeListOptions(opts))
	out := new(statuses)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	copyPagination(out.pagination, res)
	return convertStatusList(out), res, err
}

// CreateHook creates a new repository webhook.
func (s *repositoryService) CreateHook(ctx context.Context, repo string, input *scm.HookInput) (*scm.Hook, *scm.Response, error) {
	target, err := url.Parse(input.Target)
	if err != nil {
		return nil, nil, err
	}
	params := target.Query()
	params.Set("secret", input.Secret)
	target.RawQuery = params.Encode()

	path := fmt.Sprintf("2.0/repositories/%s/hooks", repo)
	in := new(hookInput)
	in.URL = target.String()
	in.Active = true
	in.Description = input.Name
	in.Events = append(
		input.NativeEvents,
		convertHookEvents(input.Events)...,
	)
	out := new(hook)
	res, err := s.client.do(ctx, "POST", path, in, out)
	return convertHook(out), res, err
}

// CreateStatus creates a new commit status.
func (s *repositoryService) CreateStatus(ctx context.Context, repo, ref string, input *scm.StatusInput) (*scm.Status, *scm.Response, error) {
	path := fmt.Sprintf("2.0/repositories/%s/commit/%s/statuses/build", repo, ref)
	in := &status{
		State: convertFromState(input.State),
		Desc:  input.Desc,
		Key:   input.Label,
		Name:  input.Label,
		URL:   input.Target,
	}
	out := new(status)
	res, err := s.client.do(ctx, "POST", path, in, out)
	return convertStatus(out), res, err
}

// DeleteHook deletes a repository webhook.
func (s *repositoryService) DeleteHook(ctx context.Context, repo string, id string) (*scm.Response, error) {
	path := fmt.Sprintf("2.0/repositories/%s/hooks/%s", repo, id)
	return s.client.do(ctx, "DELETE", path, nil, nil)
}

// helper function to convert from the gogs repository list to
// the common repository structure.
func convertRepositoryList(from *repositories) []*scm.Repository {
	to := []*scm.Repository{}
	for _, v := range from.Values {
		to = append(to, convertRepository(v))
	}
	return to
}

// helper function to convert from the gogs repository structure
// to the common repository structure.
func convertRepository(from *repository) *scm.Repository {
	namespace, name := scm.Split(from.FullName)
	return &scm.Repository{
		ID:        from.UUID,
		Name:      name,
		Namespace: namespace,
		FullName:  from.FullName,
		Link:      fmt.Sprintf("https://bitbucket.org/%s", from.FullName),
		Branch:    from.Mainbranch.Name,
		Private:   from.IsPrivate,
		Clone:     fmt.Sprintf("https://bitbucket.org/%s.git", from.FullName),
		CloneSSH:  fmt.Sprintf("git@bitbucket.org:%s.git", from.FullName),
		Created:   from.CreatedOn,
		Updated:   from.UpdatedOn,
	}
}

func convertPerms(from *perms) *scm.Perm {
	to := new(scm.Perm)
	if len(from.Values) != 1 {
		return to
	}
	switch from.Values[0].Permissions {
	case "admin":
		to.Pull = true
		to.Push = true
		to.Admin = true
	case "write":
		to.Pull = true
		to.Push = true
	default:
		to.Pull = true
	}
	return to
}

func convertHookList(from *hooks) []*scm.Hook {
	to := []*scm.Hook{}
	for _, v := range from.Values {
		to = append(to, convertHook(v))
	}
	return to
}

func convertHook(from *hook) *scm.Hook {
	return &scm.Hook{
		ID:         from.UUID,
		Name:       from.Description,
		Active:     from.Active,
		Target:     from.URL,
		Events:     from.Events,
		SkipVerify: from.SkipCertVerification,
	}
}

func convertHookEvents(from scm.HookEvents) []string {
	var events []string
	if from.Push {
		events = append(events, "repo:push")
	}
	if from.PullRequest {
		events = append(events, "pullrequest:updated")
		events = append(events, "pullrequest:unapproved")
		events = append(events, "pullrequest:approved")
		events = append(events, "pullrequest:rejected")
		events = append(events, "pullrequest:fulfilled")
		events = append(events, "pullrequest:created")
	}
	if from.PullRequestComment {
		events = append(events, "pullrequest:comment_created")
		events = append(events, "pullrequest:comment_updated")
		events = append(events, "pullrequest:comment_deleted")
	}
	if from.Issue {
		events = append(events, "issues")
		events = append(events, "issue:created")
		events = append(events, "issue:updated")
	}
	if from.IssueComment {
		events = append(events, "issue:comment_created")
	}
	return events
}

type repositories struct {
	pagination
	Values []*repository `json:"values"`
}

type statuses struct {
	pagination
	Values []*status `json:"values"`
}

type status struct {
	State string `json:"state"`
	Key   string `json:"key"`
	Name  string `json:"name,omitempty"`
	URL   string `json:"url"`
	Desc  string `json:"description,omitempty"`
}

func convertStatusList(from *statuses) []*scm.Status {
	to := []*scm.Status{}
	for _, v := range from.Values {
		to = append(to, convertStatus(v))
	}
	return to
}

func convertStatus(from *status) *scm.Status {
	return &scm.Status{
		State:  convertState(from.State),
		Label:  from.Key,
		Desc:   from.Desc,
		Target: from.URL,
	}
}

func convertState(from string) scm.State {
	switch from {
	case "FAILED":
		return scm.StateFailure
	case "INPROGRESS":
		return scm.StatePending
	case "SUCCESSFUL":
		return scm.StateSuccess
	default:
		return scm.StateUnknown
	}
}

func convertFromState(from scm.State) string {
	switch from {
	case scm.StatePending, scm.StateRunning:
		return "INPROGRESS"
	case scm.StateSuccess:
		return "SUCCESSFUL"
	default:
		return "FAILED"
	}
}
